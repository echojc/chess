# chess

CLI tool to interact with Chess.com API. Retrieved data gets cached locally.

## retrieve

Fetches all games played on the given account.

Uses ETags with requests so only new games are downloaded.

```
$ ./chess -u echojc -r
```

## analyse

Analyse and annotate important moves in a game. Outputs in PGN format by default.

```
$ ./chess -u echojc -a 20686778771
1. e4 1... d5 2. exd5 {★} 2... Qxd5 {★} 3. Nc3 {★} 3... Qe6+ 4. Be2 {★} 4... Qd7 5. Nf3 {★} 5... Qd8 6. d4 {★} 6... g6 7. O-O 7... Nf6 {★} 8. Be3 8... Bg7 {★} 9. Qd2 9... O-O {★} 10. Bh6 10... Bxh6 11. Qxh6 {★} 11... Qd6 { +3.16 } (11... Bg4) 12. Ng5 {★} 12... Qd5 { -5.19 } (12... Nbd7) 13. f3 { +5.69 } (13. Nxd5) 13... Qf5 { +2.29 } (13... Qa5) 14. Nce4 { -1.82 } (14. Bc4) 14... Bd7 { -6.16 } (14... Nbd7) 15. Nxf6+ 15... Qxf6 16. Qxh7#
```

Or, use the keyword `latest` as the game-id to analyse the last game on the account. I typically run it like this:

```
$ ./chess -u echojc -r -a latest
```

## search

Lists games played by on the given account, optionally filtered by opening moves.

```
$ ./chess -u echojc -q 'd4 d5 Bf4'
2021/05/24 [https://www.chess.com/game/live/15571917027] (♔1192) 1.d4 d5 2.Bf4 Bf5 3.c4 e6 4.Nc3 Bb4 5.Nf3 Bxc3+ 6.bxc3 dxc4  *
2021/05/24 [https://www.chess.com/game/live/15570200907] (♔1209) 1.d4 d5 2.Bf4 Nc6 3.e3 a6 4.Nf3 Nf6 5.Bd3 e6 6.c3 Bd6  *
2021/05/24 [https://www.chess.com/game/live/15567164085] (♔1223) 1.d4 d5 2.Bf4 a6 3.Nc3 Nc6 4.e3 Nf6 5.Bd3 Bf5 6.Bxf5 e5  *
2021/05/24 [https://www.chess.com/game/live/15565980183] (♔1215) 1.d4 d5 2.Bf4 Nc6 3.e3 Bf5 4.Qd2 e6 5.c3 Nf6 6.Be2 a5  *
2021/05/24 [https://www.chess.com/game/live/15565320497] (♔1222) 1.d4 d5 2.Bf4 Nc6 3.e3 Bf5 4.Nf3 e6 5.Nc3 Nb4 6.e4 dxe4  *
2021/05/16 [https://www.chess.com/game/live/14882841747] (♚1116) 1.d4 d5 2.Bf4 Nf6 3.e3 Nbd7 4.c4 e6 5.Nc3 Bb4 6.Qb3 Bxc3+  *
2021/05/15 [https://www.chess.com/game/live/14784997913] (♚1093) 1.d4 d5 2.Bf4 Nc6 3.Nf3 f6 4.e3 Bg4 5.Be2 Bxf3 6.Bxf3 e5  *
```

## usage

```
$ ./chess
  -a string
        ID of game to analyse.
  -d int
        Depth to analyse each position. (default 20)
  -f    Force refresh all data for user.
  -l string
        Log level. (default "info")
  -n int
        Number of games to display. (default 20)
  -o string
        Output format: pgn (default), url
  -q string
        Only display games with these initial moves (space-separated algebraic notation).
  -r    Check server for new data for user.
  -t duration
        Timeout when analysing each position. (default 3s)
  -th float
        Threshold for annotating inaccurate moves (delta in position score). (default 1.8)
  -u string
        User whose games to load. (required)
```
