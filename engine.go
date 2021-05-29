package main

import (
	"bufio"
	"fmt"
	"io"
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
	//e.send("go movetime 2000\n")

	data := e.readUntil("bestmove")

	bestmove := data[len(data)-1]
	info := data[len(data)-2]

	idx1 := strings.Index(info, " cp ") + 4
	idx2 := strings.Index(info[idx1:], " ")

	var f float64
	if idx1 >= 0 || idx2 >= 0 {
		f, e.err = strconv.ParseFloat(info[idx1:idx1+idx2], 64)
	}

	return Result{
		Score:    f / 100,
		BestMove: strings.Split(bestmove, " ")[1],
		Time:     time.Since(start),
	}
}

type Engine struct {
	searchCmd string

	cmd     *exec.Cmd
	stdin   io.WriteCloser
	stdout  io.ReadCloser
	scanner *bufio.Scanner

	err error
}

func NewEngine(depth int) (*Engine, error) {
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
