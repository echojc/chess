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

var (
	user           = flag.String("u", "", "User whose games to load. (required)")
	isRefresh      = flag.Bool("r", false, "Check server for new data for user.")
	isForce        = flag.Bool("f", false, "Force refresh all data for user.")
	length         = flag.Int("n", 20, "Number of games to display.")
	searchMoves    = flag.String("q", "", "Only display games with these initial moves (space-separated algebraic notation).")
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

	games, err := ListGames(*user, cacheOnly, forceFetch)
	if err != nil {
		log.WithError(err).WithField("user", user).Fatal("Could not get games")
	}

	//if *searchMoves != "" {
	//	moves := strings.Split(*searchMoves, " ")
	//}

	for i := 0; i < *length && i < len(games); i++ {
		fmt.Println(formatGame(games[i], *user))
	}
	return
}

func formatGame(g Game, user string) string {
	var url string
	if g.URL != nil {
		url = g.URL.String()
	}

	rating := g.White.Rating
	if g.Black.Username == user {
		rating = g.Black.Rating
	}

	t := chess.NewGame()
	parsedGame, err := g.Game()
	if err == nil {
		moves := parsedGame.Moves()
		for i := 0; i < 6*2 && i < len(moves); i++ {
			t.Move(moves[i])
		}
	} else {
		log.WithError(err).WithField("url", url).Warn("Could not parse game")
	}

	return fmt.Sprintf("%s [%s] (%d) %s",
		g.EndTime.Format("2006/01/02"),
		url,
		rating,
		strings.TrimSpace(t.String()),
	)
}
