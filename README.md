# monkey

Monkey is a lightweight terminal-based text editor written in Go, built completely from scratch.

The project is focused on understanding how text editors work internally while learning Go, terminal programming, input handling, text rendering, and file manipulation.

## Keybinds

### View Mode
```txt
       ESC: Enter VIEW mode
         e: Enter EDIT mode
         q: Quit editor
         w: Save file to disk
         d: Cut current line (copied to clipboard)
         c: Copy current line to clipboard
         p: Paste line from clipboard / copy buffer
         u: Undo last change
    U / ^r: Redo last undone change
   h,j,k,l: Move cursor (Left, Down, Up, Right)
    Arrows: Move cursor
    PgDown: Scroll 1/4 screen downwards
      PgUp: Scroll 1/4 screen upwards
      HOME: Move cursor to beginning of current line
       END: Move cursor to end of current line
```

### Edit Mode
```txt
       ESC: Return to VIEW mode
 Backspace: Delete character / merge lines
     Enter: Insert newline
       Tab: Insert 4 spaces
```
---
## Tech Stack
| Component       | Technology          |
| --------------- | ------------------- |
| Language        | Go                  |
| Terminal UI     | Termbox             |
| Unicode Support | go-runewidth        |
| Build           | Bash + Go toolchain |
## Install from source
```
git clone git@github.com:akshatgitgud/monkey.git
cd monkey && make install
```

Run the editor:
```
monkey filename.txt
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

