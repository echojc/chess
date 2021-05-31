package main

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/apex/log"
)

func OpenGame(user string, id string) (Game, error) {
	games, err := listGames(user, true, false)
	if err != nil {
		return Game{}, err
	}

	for _, g := range games {
		if strings.HasSuffix(g.URL.String(), id) {
			return g, nil
		}
	}

	return Game{}, fmt.Errorf("Game not found (%s - %s)", user, id)
}

func ListCachedGames(user string) ([]Game, error) {
	return listGames(user, true, false)
}

func RefreshCache(user string, forceFetch bool) ([]Game, error) {
	return listGames(user, false, forceFetch)
}

func listGames(user string, cacheOnly bool, forceFetch bool) ([]Game, error) {
	archives, err := ListArchives(user, cacheOnly && !forceFetch)
	if err != nil {
		return nil, err
	}

	var games []Game
	for _, a := range archives {
		archiveGames, err := OpenArchive(a, cacheOnly, forceFetch)
		if err != nil {
			log.WithError(err).WithField("archive", a).
				Warn("Could not open archive")
		}
		games = append(games, archiveGames...)
	}

	// newest games first
	sort.Slice(games, func(i, j int) bool {
		return games[i].EndTime.After(games[j].EndTime)
	})

	return games, nil
}

func ListArchives(user string, cacheOnly bool) ([]string, error) {
	if cacheOnly {
		return LoadUserArchives(user), nil
	}

	archives, err := FetchArchives(user)
	if err != nil {
		return nil, err
	}

	SaveUserArchives(user, archives)
	return archives, nil
}

func OpenArchive(archiveID string, cacheOnly bool, forceFetch bool) ([]Game, error) {
	if cacheOnly {
		games, ok := LoadArchive(archiveID)
		if !ok {
			return nil, errors.New("Archive is not cached")
		}
		return games, nil
	}

	var cachedETag string
	if !forceFetch {
		cachedETag = LoadETag(archiveID)
	}

	a, err := FetchArchive(archiveID, cachedETag)
	if err != nil {
		return nil, err
	}

	// cached copy is latest
	if !forceFetch && a.ETag == cachedETag {
		games, ok := LoadArchive(archiveID)
		// if failed to load, force fetch
		if !ok {
			log.WithField("archive", archiveID).
				Error("Could not open cached archive, will force fetch")
			return OpenArchive(archiveID, false, true)
		}
		return games, nil
	}

	SaveArchive(archiveID, a.Games)
	SaveETag(archiveID, a.ETag)
	return a.Games, nil
}
