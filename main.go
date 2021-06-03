package main

import (
	"flag"
	"fmt"
	"math"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/notnil/chess"
)

func init() {
	log.SetHandler(text.Default)
}

type config struct {
	user   string
	output string

	// data consistency
	cacheOnly  bool
	forceFetch bool

	// search
	limit int
	query string

	// analyse
	analyze   string
	depth     int
	timeout   time.Duration
	threshold float64
}

func main() {
	var (
		logLevelString = flag.String("l", "info", "Log level.")
		user           = flag.String("u", "", "User whose games to load. (required)")
		output         = flag.String("o", "", "Output format: url (default), pgn")

		isRefresh = flag.Bool("r", false, "Check server for new data for user.")
		isForce   = flag.Bool("f", false, "Force refresh all data for user.")

		limit = flag.Int("n", 20, "Number of games to display.")
		query = flag.String("q", "", "Only display games with these initial moves (space-separated algebraic notation).")

		analyze   = flag.String("a", "", "ID of game to analyse.")
		depth     = flag.Int("d", 20, "Depth to analyse each position.")
		timeout   = flag.Duration("t", time.Second, "Timeout when analysing each position.")
		threshold = flag.Float64("th", 1.8, "Threshold for annotating inaccurate moves (delta in position score).")
	)
	flag.Parse()

	if *user == "" {
		flag.PrintDefaults()
		os.Exit(2)
	}

	logLevel, err := log.ParseLevel(*logLevelString)
	if err != nil {
		log.WithField("level", *logLevelString).
			Warn("Invalid log level, defaulting to WARN")
		logLevel = log.WarnLevel
	}
	log.SetLevel(logLevel)

	// read arguments into config
	cfg := config{
		user:       *user,
		output:     *output,
		cacheOnly:  !*isRefresh,
		forceFetch: *isForce,
		limit:      *limit,
		query:      *query,
		analyze:    *analyze,
		depth:      *depth,
		timeout:    *timeout,
		threshold:  *threshold,
	}
	log.WithField("cfg", cfg).Debug("Loaded arguments")

	// check with server if either refresh or force refresh are set
	if !cfg.cacheOnly || cfg.forceFetch {
		_, err := RefreshCache(cfg.user, cfg.forceFetch)
		if err != nil {
			log.WithError(err).WithField("user", cfg.user).
				Warn("Could not refresh cache")
		}
	}

	// main function
	if cfg.analyze != "" {
		Analyze(cfg)
	} else {
		Search(cfg)
	}
}

func Analyze(cfg config) {
	var data Game
	var err error

	if cfg.analyze == "latest" {
		games, err := ListCachedGames(cfg.user)
		if err != nil {
			log.WithError(err).WithField("user", cfg.user).
				Fatal("Could not get games")
		}
		data = games[0]
	} else {
		data, err = OpenGame(cfg.user, cfg.analyze)
		if err != nil {
			log.WithError(err).WithFields(log.Fields{
				"user": cfg.user,
				"id":   cfg.analyze,
			}).Fatal("Could not find game to analyse")
		}
	}

	g, err := data.Game()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"user": cfg.user,
			"id":   cfg.analyze,
		}).Fatal("Could not parse game to analyse")
	}

	e, err := NewEngine(cfg.depth, cfg.timeout)
	if err != nil {
		log.WithError(err).Fatal("Could not initialise analysis engine")
	}

	// evaluate all board positions
	positions := g.Positions()
	log.WithFields(log.Fields{
		"url":       data.URL.String(),
		"count":     len(positions),
		"depth":     cfg.depth,
		"timeout":   cfg.timeout,
		"threshold": cfg.threshold,
	}).Info("Starting analysis")
	var results = make([]Result, len(positions))
	for i, p := range positions {
		fen := p.String()

		r := e.Analyze(fen)
		if err = e.Err(); err != nil {
			log.WithError(err).WithFields(log.Fields{"i": i, "fen": fen}).
				Warn("Could not analyse board state")
			continue
		}

		// engine returns score from current player's perspective
		if i%2 == 1 {
			r.Score *= -1
		}
		if r.Err != nil {
			log.WithError(r.Err).WithFields(log.Fields{
				"i":        i,
				"fen":      fen,
				"score":    r.Score,
				"bestmove": r.BestMove,
			}).Warn("Could not parse engine result")
		}

		results[i] = r
		log.WithFields(log.Fields{"i": i, "t": r.Time, "d": r.Depth}).
			Info("Analysed position")
	}

	nalg := chess.AlgebraicNotation{}
	nuci := chess.UCINotation{}

	buf := &strings.Builder{}
	for i, gameMove := range g.Moves() {
		var turn string
		if i%2 == 0 {
			turn = fmt.Sprintf("%d.", i/2+1)
		} else {
			turn = fmt.Sprintf("%d...", i/2+1)
		}

		gameMoveStr := nalg.Encode(positions[i], gameMove)
		fmt.Fprintf(buf, "%s %s ", turn, gameMoveStr)

		var bestMoveStr = results[i].BestMove
		bestMove, err := nuci.Decode(positions[i], bestMoveStr)
		if err != nil {
			log.WithError(err).WithField("move", bestMove).
				Warn("Could not decode best move")
		} else {
			bestMoveStr = nalg.Encode(positions[i], bestMove)
		}

		if gameMoveStr == bestMoveStr {
			fmt.Fprint(buf, "{★} ")
		}

		delta := results[i+1].Score - results[i].Score
		if math.Abs(delta) > cfg.threshold {
			fmt.Fprintf(buf, "{ %+.2f } (%s %s) ", delta, turn, bestMoveStr)
		}
	}

	switch cfg.output {
	case "url":
		fmt.Printf("https://chess.com/analysis?pgn=%s\n", url.QueryEscape(buf.String()))
	default:
		fmt.Println(buf.String())
	}
}

func Search(cfg config) {
	// validate moves in query string
	var searchMoves []*chess.Move
	if cfg.query != "" {
		algebraicMoves := strings.Split(cfg.query, " ")
		searchBoard := chess.NewGame()
		for _, m := range algebraicMoves {
			err := searchBoard.MoveStr(m)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"q":    cfg.query,
					"move": m,
				}).Fatal("Invalid move in query string")
			}
		}
		searchMoves = searchBoard.Moves()
	}

	games, err := ListCachedGames(cfg.user)
	if err != nil {
		log.WithError(err).WithField("user", cfg.user).Fatal("Could not get games")
	}

	for i, n := 0, 0; i < len(games) && n < cfg.limit; i++ {
		if searchMoves != nil && !movesMatch(games[i], searchMoves) {
			continue
		}

		fmt.Println(formatGame(games[i], cfg.user))
		n++
	}
}

func movesMatch(g Game, searchMoves []*chess.Move) bool {
	game, err := g.Game()
	if err != nil {
		log.WithError(err).WithField("url", g.URL).Warn("Could not parse game")
		return false
	}
	gameMoves := game.Moves()

	if len(gameMoves) < len(searchMoves) {
		return false
	}

	for i := range searchMoves {
		if gameMoves[i].String() != searchMoves[i].String() {
			return false
		}
	}

	return true
}

func formatGame(g Game, user string) string {
	var rating int
	var icon rune
	var result rune

	switch user {
	case g.White.Username:
		rating = g.White.Rating
		icon = '♔'
		result = unicode.ToUpper([]rune(g.White.NormalizedResult())[0])
	case g.Black.Username:
		rating = g.Black.Rating
		icon = '♚'
		result = unicode.ToUpper([]rune(g.Black.NormalizedResult())[0])
	}

	t := chess.NewGame()
	parsedGame, err := g.Game()
	if err == nil {
		moves := parsedGame.Moves()
		for i := 0; i < 6*2 && i < len(moves); i++ {
			t.Move(moves[i])
		}
	} else {
		log.WithError(err).WithField("url", g.URL).Warn("Could not parse game")
	}

	return fmt.Sprintf("%s [%s] %c%4d%c %s",
		g.EndTime.Format("02/01"),
		g.URL,
		icon,
		rating,
		result,
		strings.TrimSpace(t.String()),
	)
}
