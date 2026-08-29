# monkey

Monkey is a lightweight terminal-based text editor written in Go, built completely from scratch.

The project is focused on understanding how text editors work internally while learning Go, terminal programming, input handling, text rendering, and file manipulation.

## Keybinds
---
```txt
   ESC: enter the 'VIEW' mode
     e: enter the 'EDIT' mode
     q: quit from the text editor
     w: write file to disk
     d: cut current line
     c: copy current line to copy buffer
     p: paste line from copy buffer
     s: push text buffer to undo buffer
     l: pull text buffer from undo buffer
Arrows: move cursor
PgDown: scroll 1/4 of the screen downwards
  PgUp: scroll 1/4 of the screen upwards
  HOME: move cursor to the beginning of the current line
   END: move cursor to the end of the current line
```
---
## Tech Stack
| Component       | Technology          |
| --------------- | ------------------- |
| Language        | Go                  |
| Terminal UI     | Termbox             |
| Unicode Support | go-runewidth        |
| Build           | Bash + Go toolchain |
## Build from source
```
git clone git@github.com:akshatgitgud/monkey.git
cd monkey
go mod tidy
go build -o monkey monkey.go
```

Run the editor:
```
./monkey
``` 
Or open a file directly:
```
./monkey filename.txt
```
---
## Requirements
- Go 1.18 or later
- A Unix-like terminal environment
- Git
## License
MIT License. see `LICENSE` for more information.

---
Built from scratch with Go.:)

