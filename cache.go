package main

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	"github.com/apex/log"
)

var (
	cacheArchives map[string][]Game = make(map[string][]Game)
	userArchives  map[string][]string
	eTags         map[string]string

	_cacheDir     string
	cacheDisabled bool
)

const (
	dirName          = "sh.echo.chess"
	userArchivesFile = "userarchives.json"
	eTagsFile        = "etags.json"
)

func cacheDir() string {
	if cacheDisabled {
		return ""
	}

	if _cacheDir != "" {
		return _cacheDir
	}

	base, err := os.UserCacheDir()
	if err != nil {
		log.WithError(err).
			Warn("OS did not provide a cache directory, using /tmp")
		base = os.TempDir()
	}

	_cacheDir = filepath.Join(base, dirName)
	if err = os.MkdirAll(_cacheDir, 0755); err != nil {
		log.WithError(err).WithField("dir", _cacheDir).
			Error("Could not create cache directory, caching disabled")
		cacheDisabled = true
		return ""
	}

	log.WithField("dir", _cacheDir).Info("Cache enabled")
	return _cacheDir
}

func LoadETag(archiveID string) string {
	if eTags == nil {
		loadETags()
	}

	return eTags[archiveID]
}

func SaveETag(archiveID, eTag string) {
	if eTags == nil {
		loadETags()
	}

	eTags[archiveID] = eTag
	saveETags()
}

func loadETags() {
	baseDir := cacheDir()
	if cacheDisabled || baseDir == "" {
		eTags = make(map[string]string)
		return
	}

	path := filepath.Join(baseDir, eTagsFile)
	f, err := os.Open(path)
	if err != nil {
		log.WithError(err).
			Warn("Could not open cached ETags")
		eTags = make(map[string]string)
		return
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&eTags); err != nil {
		log.WithError(err).WithField("path", path).
			Warn("Could not read cached ETags")
		eTags = make(map[string]string)
		return
	}

	if len(eTags) == 0 {
		log.WithField("path", path).Warn("Loaded ETags file but it was empty")
		eTags = make(map[string]string)
	}

	log.WithFields(log.Fields{
		"path":  path,
		"count": len(eTags),
	}).Info("Loaded cached ETags")
}

func saveETags() {
	baseDir := cacheDir()
	if cacheDisabled || baseDir == "" {
		return
	}

	data, err := json.MarshalIndent(eTags, "", "  ")
	if err != nil {
		log.WithError(err).Warn("Could not marshal ETags")
		return
	}

	path := filepath.Join(baseDir, eTagsFile)
	if err = ioutil.WriteFile(path, data, 0644); err != nil {
		log.WithError(err).WithField("path", path).
			Warn("Could not write ETags to file")
		return
	}

	log.WithFields(log.Fields{
		"path":  path,
		"count": len(eTags),
	}).Info("Saved ETags to file")
}

func LoadUserArchives(user string) []string {
	if userArchives == nil {
		loadUserArchives()
	}

	return userArchives[user]
}

func SaveUserArchives(user string, archives []string) {
	if userArchives == nil {
		loadUserArchives()
	}

	userArchives[user] = archives
	saveUserArchives()
}

func loadUserArchives() {
	baseDir := cacheDir()
	if cacheDisabled || baseDir == "" {
		userArchives = make(map[string][]string)
		return
	}

	path := filepath.Join(baseDir, userArchivesFile)
	f, err := os.Open(path)
	if err != nil {
		log.WithError(err).
			Warn("Could not open cached user archives")
		userArchives = make(map[string][]string)
		return
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&userArchives); err != nil {
		log.WithError(err).WithField("path", path).
			Warn("Could not read cached user archives")
		userArchives = make(map[string][]string)
		return
	}

	if len(userArchives) == 0 {
		log.WithField("path", path).
			Warn("Loaded user archives file but it was empty")
		userArchives = make(map[string][]string)
	}

	log.WithFields(log.Fields{
		"path":  path,
		"count": len(userArchives),
	}).Info("Loaded cached user archives")
}

func saveUserArchives() {
	baseDir := cacheDir()
	if cacheDisabled || baseDir == "" {
		return
	}

	data, err := json.MarshalIndent(userArchives, "", "  ")
	if err != nil {
		log.WithError(err).Warn("Could not marshal user archives")
		return
	}

	path := filepath.Join(baseDir, userArchivesFile)
	if err = ioutil.WriteFile(path, data, 0644); err != nil {
		log.WithError(err).WithField("path", path).
			Warn("Could not write user archives to file")
		return
	}

	log.WithFields(log.Fields{
		"path":  path,
		"count": len(userArchives),
	}).Info("Saved user archives to file")
}

func createArchiveFilename(archiveID string) string {
	//return strings.ReplaceAll(archiveID, "/", "_")
	return url.QueryEscape(archiveID) + ".json"
}

func LoadArchive(archiveID string) ([]Game, bool) {
	archive, cached := cacheArchives[archiveID]
	if cached {
		return archive, true
	}

	// try load from file
	baseDir := cacheDir()
	if cacheDisabled || baseDir == "" {
		return nil, false
	}

	path := filepath.Join(baseDir, createArchiveFilename(archiveID))
	f, err := os.Open(path)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"archive": archiveID,
			"path":    path,
		}).Warn("Could not open archive file")
		return nil, false
	}
	defer f.Close()

	var games []Game
	if err := json.NewDecoder(f).Decode(&games); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"archive": archiveID,
			"path":    path,
		}).Warn("Could not read archive file")
		return nil, false
	}

	if len(games) == 0 {
		log.WithError(err).WithFields(log.Fields{
			"archive": archiveID,
			"path":    path,
		}).Warn("Loaded archive file but it was empty")
	}

	cacheArchives[archiveID] = games
	log.WithError(err).WithFields(log.Fields{
		"archive": archiveID,
		"path":    path,
		"count":   len(games),
	}).Info("Loaded cached archive")
	return games, true
}

func SaveArchive(archiveID string, games []Game) {
	baseDir := cacheDir()
	if cacheDisabled || baseDir == "" {
		return
	}

	path := filepath.Join(baseDir, createArchiveFilename(archiveID))
	tmpPath := path + ".tmp"

	f, err := os.Create(tmpPath)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"archive": archiveID,
			"path":    tmpPath,
		}).Warn("Could not create temporary archive file")
		return
	}
	defer f.Close()

	e := json.NewEncoder(f)
	e.SetIndent("", "  ")
	if err = e.Encode(games); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"archive": archiveID,
			"path":    tmpPath,
		}).Warn("Could not write archive to temporary file")
		return
	}

	f.Close()
	if err = os.Rename(tmpPath, path); err != nil {
		log.WithError(err).WithFields(log.Fields{
			"archive":  archiveID,
			"tmp_path": tmpPath,
			"path":     path,
		}).Warn("Could not rename temporary file to archive")
		return
	}

	log.WithFields(log.Fields{
		"archive": archiveID,
		"path":    path,
		"count":   len(games),
	}).Info("Saved archive to file")
}
