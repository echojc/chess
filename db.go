package main

import (
	"errors"

	"github.com/apex/log"
)

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
