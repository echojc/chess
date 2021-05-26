package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/apex/log"
)

const (
	APIHost = "https://api.chess.com"
)

func FetchArchives(user string) ([]string, error) {
	log.WithField("user", user).Info("Fetching available archives")

	s, err := http.Get(fmt.Sprintf(
		"%s/pub/player/%s/games/archives", APIHost, url.QueryEscape(user)))
	if err != nil {
		return nil, err
	}
	defer s.Body.Close()

	if s.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected response %d %s", s.StatusCode, s.Status)
	}

	var t struct {
		Archives []string
	}
	if err = json.NewDecoder(s.Body).Decode(&t); err != nil {
		return nil, err
	}

	var archives []string
	for _, s := range t.Archives {
		// validate archives are valid urls
		u, err := url.Parse(s)
		if err != nil {
			log.WithField("url", s).Warn("Skipping archive with invalid URL")
			continue
		}
		archives = append(archives, u.Path)
	}

	return archives, nil
}

type Archive struct {
	ETag  string
	Games []Game
}

func FetchArchive(archiveID, oldETag string) (Archive, error) {
	log.WithField("archive", archiveID).Info("Fetching archive")
	var a Archive

	r, err := http.NewRequest("GET", APIHost+archiveID, http.NoBody)
	if err != nil {
		return a, err
	}

	if oldETag != "" {
		log.WithFields(log.Fields{
			"archive": archiveID,
			"etag":    oldETag,
		}).Info("Using ETag with request")
		r.Header.Add("If-None-Match", oldETag)
	}

	s, err := http.DefaultClient.Do(r)
	if err != nil {
		return a, err
	}
	defer s.Body.Close()

	// existing data is latest
	if s.StatusCode == http.StatusNotModified {
		log.WithFields(log.Fields{
			"archive": archiveID,
			"etag":    oldETag,
		}).Info("Cached data is up to date")
		a.ETag = oldETag
		return a, nil
	}

	if s.StatusCode != http.StatusOK {
		return a, fmt.Errorf("Unexpected response %d %s", s.StatusCode, s.Status)
	}

	// got new data
	a.ETag = s.Header.Get("Etag")
	log.WithFields(log.Fields{
		"archive": archiveID,
		"etag":    a.ETag,
	}).Info("Got new data")

	var data struct {
		Games []Game
	}
	if err = json.NewDecoder(s.Body).Decode(&data); err != nil {
		return a, err
	}

	// filter out non-standard games
	for _, g := range data.Games {
		if g.Rules != "chess" {
			log.WithFields(log.Fields{
				"archive": archiveID,
				"game":    g.URL,
				"rules":   g.Rules,
			}).Debug("Skipped non-standard game")
			continue
		}
		a.Games = append(a.Games, g)
	}

	log.WithFields(log.Fields{
		"archive":     archiveID,
		"etag":        a.ETag,
		"total_count": len(data.Games),
		"kept_count":  len(a.Games),
	}).Info("Fetched archive")
	return a, nil
}
