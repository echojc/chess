package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"math"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/apex/log"
)

const (
	processName = "stockfish"
)

type Result struct {
	Score    float64
	BestMove string
	Time     time.Duration
	Err      error
}

func (e *Engine) Analyze(fen string) Result {
	if e.err != nil {
		return Result{}
	}

	start := time.Now()

	e.send("ucinewgame\n")
	e.send("position fen ")
	e.send(fen)
	e.send("\n")
	e.send(e.searchCmd)

	data := e.readUntilWithTimeout("bestmove")

	// parse cp out of info string
	var f float64
	var ferr error
	curKey := ""
	infoStr := data[len(data)-2]
	for _, v := range strings.Split(infoStr, " ") {
		switch v {
		case "cp", "mate":
			curKey = v
			continue
		}

		switch curKey {
		case "cp":
			f, ferr = strconv.ParseFloat(v, 64)
		case "mate":
			f, ferr = strconv.ParseFloat(v, 64)
			f = (f / math.Abs(f)) * 10000
		}

		curKey = ""
	}

	// parse best move
	bestMoveStr := data[len(data)-1]
	bestMoveArr := strings.Split(bestMoveStr, " ")
	bestMove := "(none)"
	if len(bestMoveArr) >= 2 {
		bestMove = bestMoveArr[1]
	}

	return Result{
		Score:    f / 100,
		BestMove: bestMove,
		Time:     time.Since(start),
		Err:      ferr,
	}
}

type Engine struct {
	searchCmd string
	timeout   time.Duration

	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	scanner *bufio.Scanner

	err error
}

func NewEngine(depth int, timeout time.Duration) (*Engine, error) {
	cmd := exec.Command(processName)

	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	e := &Engine{
		cmd:       cmd,
		stdin:     in,
		stdout:    out,
		scanner:   bufio.NewScanner(out),
		searchCmd: fmt.Sprintf("go depth %d\n", depth),
		timeout:   timeout,
	}

	e.send("uci\n")
	e.readUntil("uciok")

	e.send("setoption name Threads value 8\n")
	e.send("setoption name UCI_AnalyseMode value true\n")
	e.send("setoption name Use NNUE value false\n")
	e.send("isready\n")
	e.readUntil("readyok")

	return e, e.err
}

func (e *Engine) Err() error {
	return e.err
}

func (e *Engine) send(data string) {
	if e.err != nil {
		return
	}
	_, e.err = io.WriteString(e.stdin, data)
	log.WithField("engine", "tx").Debug(data)
}

func (e *Engine) readUntil(prefix string) []string {
	if e.err != nil {
		return nil
	}

	var out []string

	var line string
	for !strings.HasPrefix(line, prefix) {
		e.scanner.Scan()
		line = e.scanner.Text()
		out = append(out, line)
		log.WithField("engine", "rx").Debug(line)
	}

	e.err = e.scanner.Err()
	return out
}

func (e *Engine) readUntilWithTimeout(prefix string) []string {
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	c := make(chan []string)
	go func() {
		c <- e.readUntil(prefix)
		close(c)
	}()

	select {
	case <-ctx.Done():
		e.send("stop\n")
		return <-c
	case data := <-c:
		return data
	}
}
