# chess

CLI tool to interact with Chess.com API. Retrieved data gets cached locally.

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
  -u string
        User whose games to load. (required)
  -r    Check server for new data for user.
  -f    Force refresh all data for user.
  -q string
        Only display games with these initial moves (space-separated algebraic notation).
  -n int
        Number of games to display. (default 20)
  -l string
        Log level. (default "warn")
```
