package main

import (
	"flag"
	"fmt"
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
	user string

	// search
	cacheOnly  bool
	forceFetch bool
	limit      int
	query      string

	// analyse
	analyze string
	depth   int
	timeout time.Duration
}

func main() {
	var (
		logLevelString = flag.String("l", "warn", "Log level.")
		user           = flag.String("u", "", "User whose games to load. (required)")

		isRefresh = flag.Bool("r", false, "Check server for new data for user.")
		isForce   = flag.Bool("f", false, "Force refresh all data for user.")
		limit     = flag.Int("n", 20, "Number of games to display.")
		query     = flag.String("q", "", "Only display games with these initial moves (space-separated algebraic notation).")

		analyze = flag.String("a", "", "ID of game to analyse.")
		depth   = flag.Int("d", 20, "Depth to analyse each position.")
		timeout = flag.Duration("t", time.Second, "Timeout when analysing each position.")
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
		cacheOnly:  !*isRefresh,
		forceFetch: *isForce,
		limit:      *limit,
		query:      *query,
		analyze:    *analyze,
		depth:      *depth,
		timeout:    *timeout,
	}
	log.WithField("cfg", cfg).Debug("Loaded arguments")

	Search(cfg)
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

	games, err := ListGames(cfg.user, cfg.cacheOnly, cfg.forceFetch)
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
		break
	case g.Black.Username:
		rating = g.Black.Rating
		icon = '♚'
		result = unicode.ToUpper([]rune(g.Black.NormalizedResult())[0])
		break
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
