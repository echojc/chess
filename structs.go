package main

import (
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/notnil/chess"
)

type Player struct {
	Username string
	Rating   int
	Result   string
	URL      *url.URL
}

type PlayerT struct {
	Username string `json:"username"`
	Rating   int    `json:"rating"`
	Result   string `json:"result"`
	URL      string `json:"@id"`
}

func (p *Player) UnmarshalJSON(data []byte) error {
	var t PlayerT
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	url, err := url.Parse(t.URL)
	if err != nil {
		log.WithError(err).WithField("url", t.URL).
			Warn("Player has invalid URL")
	}

	p.Username = t.Username
	p.Rating = t.Rating
	p.Result = t.Result
	p.URL = url
	return nil
}

func (p Player) MarshalJSON() ([]byte, error) {
	var url string
	if p.URL != nil {
		url = p.URL.String()
	}

	t := PlayerT{
		Username: p.Username,
		Rating:   p.Rating,
		Result:   p.Result,
		URL:      url,
	}
	return json.Marshal(t)
}

func (p Player) NormalizedResult() string {
	switch p.Result {
	case "win":
		return "win"
	case "abandoned":
		return "abandoned"
	case "agreed", "repetition", "stalemate", "insufficient", "50move",
		"timevsinsufficient":
		return "draw"
	default:
		return "lose"
	}
}

type Game struct {
	URL       *url.URL
	EndTime   time.Time
	Rated     bool
	TimeClass string
	Rules     string
	White     Player
	Black     Player

	pgn  string
	game *chess.Game
}

type GameT struct {
	URL       string `json:"url"`
	PGN       string `json:"pgn"`
	EndTime   int64  `json:"end_time"`
	Rated     bool   `json:"rated"`
	TimeClass string `json:"time_class"`
	Rules     string `json:"rules"`
	White     Player `json:"white"`
	Black     Player `json:"black"`
}

func (g *Game) Game() (*chess.Game, error) {
	if g.game != nil {
		return g.game, nil
	}

	pgn, err := chess.PGN(strings.NewReader(g.pgn))
	if err != nil {
		return nil, err
	}

	g.game = chess.NewGame(pgn)
	return g.game, nil
}

func (g *Game) UnmarshalJSON(data []byte) error {
	var t GameT
	if err := json.Unmarshal(data, &t); err != nil {
		return err
	}

	url, err := url.Parse(t.URL)
	if err != nil {
		log.WithError(err).WithField("url", t.URL).
			Warn("Game has invalid URL")
	}

	g.URL = url
	g.EndTime = time.Unix(t.EndTime, 0)
	g.Rated = t.Rated
	g.TimeClass = t.TimeClass
	g.Rules = t.Rules
	g.White = t.White
	g.Black = t.Black
	g.pgn = t.PGN
	return nil
}

func (g Game) MarshalJSON() ([]byte, error) {
	var url string
	if g.URL != nil {
		url = g.URL.String()
	}

	t := GameT{
		URL:       url,
		PGN:       g.pgn,
		EndTime:   g.EndTime.Unix(),
		Rated:     g.Rated,
		TimeClass: g.TimeClass,
		Rules:     g.Rules,
		White:     g.White,
		Black:     g.Black,
	}
	return json.Marshal(t)
}
