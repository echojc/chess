package main

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
)

func init() {
	log.SetHandler(text.Default)
}

func main() {
	//fmt.Println(FetchArchives("echojc"))
	//fmt.Println(FetchArchive("/pub/player/echojc/games/2020/12", `W/"4b90087c0b31967a38b259a2a2d735fd"`))
	//SetETag("/pub/player/echojc/games/2020/12", `W/"4b90087c0b31967a38b259a2a2d735fd"`)
	OpenArchive("/pub/player/echojc/games/2020/12", false, false)
	return
}
