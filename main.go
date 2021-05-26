package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/notnil/chess"
)

func init() {
	log.SetHandler(text.Default)
}

const (
	CharWhiteKing = '\u2654'
	CharBlackKing = '\u265A'
)

var (
	user           = flag.String("u", "", "User whose games to load. (required)")
	isRefresh      = flag.Bool("r", false, "Check server for new data for user.")
	isForce        = flag.Bool("f", false, "Force refresh all data for user.")
	length         = flag.Int("n", 20, "Number of games to display.")
	query          = flag.String("q", "", "Only display games with these initial moves (space-separated algebraic notation).")
	logLevelString = flag.String("l", "warn", "Log level.")
)

func main() {
	flag.Parse()

	if *user == "" {
		flag.PrintDefaults()
		os.Exit(2)
	}
	cacheOnly := !*isRefresh
	forceFetch := *isForce

	logLevel, err := log.ParseLevel(*logLevelString)
	if err != nil {
		log.WithField("level", *logLevelString).
			Warn("Invalid log level, defaulting to WARN")
		logLevel = log.WarnLevel
	}
	log.SetLevel(logLevel)

	// validate moves in query string
	var searchMoves []*chess.Move
	if *query != "" {
		algebraicMoves := strings.Split(*query, " ")
		searchBoard := chess.NewGame()
		for _, m := range algebraicMoves {
			err := searchBoard.MoveStr(m)
			if err != nil {
				log.WithError(err).WithFields(log.Fields{
					"q":    *query,
					"move": m,
				}).Fatal("Invalid move in query string")
			}
		}
		searchMoves = searchBoard.Moves()
	}

	games, err := ListGames(*user, cacheOnly, forceFetch)
	if err != nil {
		log.WithError(err).WithField("user", user).Fatal("Could not get games")
	}

	for i, n := 0, 0; i < len(games) && n < *length; i++ {
		if searchMoves != nil && !movesMatch(games[i], searchMoves) {
			continue
		}

		fmt.Println(formatGame(games[i], *user))
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
	rating := g.White.Rating
	king := CharWhiteKing
	if g.Black.Username == user {
		rating = g.Black.Rating
		king = CharBlackKing
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

	return fmt.Sprintf("%s [%s] (%c%d) %s",
		g.EndTime.Format("2006/01/02"),
		g.URL,
		king,
		rating,
		strings.TrimSpace(t.String()),
	)
}
