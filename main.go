package main

import (
	"fmt"

	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
)

func init() {
	log.SetHandler(text.Default)
}

func main() {
	user := "echojc"
	useCache := true

	//fmt.Println(FetchArchives("echojc"))
	//fmt.Println(FetchArchive("/pub/player/echojc/games/2020/12", `W/"4b90087c0b31967a38b259a2a2d735fd"`))
	//SetETag("/pub/player/echojc/games/2020/12", `W/"4b90087c0b31967a38b259a2a2d735fd"`)
	//OpenArchive("/pub/player/echojc/games/2020/12", false, false)

	archives, err := ListArchives(user, useCache)
	if err != nil {
		log.WithError(err).WithField("user", user).
			Fatal("Could not list user archives")
	}

	var games []Game
	for _, a := range archives {
		archiveGames, err := OpenArchive(a, useCache, false)
		if err != nil {
			log.WithError(err).WithField("archive", a).
				Warn("Could not open archive")
		}
		games = append(games, archiveGames...)
	}

	fmt.Println(len(games))

	return
}
