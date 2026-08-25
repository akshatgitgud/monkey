package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

var ROWS, COLS int
var offsetX, offsetY int
var source_file string
var text_buffer = [][]rune{}

func read_file(fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		source_file = fileName
		text_buffer = append(text_buffer, []rune{})
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		line := scanner.Text()
		text_buffer = append(text_buffer, []rune{})

		for i := 0; i < len(line); i++ {
			text_buffer[lineNumber] = append(text_buffer[lineNumber], rune(line[i]))
		}
		lineNumber++
	}
	if lineNumber == 0 {
		text_buffer = append(text_buffer, []rune{})
	}
}
func display_text_buffer() {
	var row, col int
	for row = 0; row < ROWS; row++ {
		text_bufferRow := row + offsetY
		for col = 0; col < COLS; col++ {
			text_bufferCol := col + offsetX
			if text_bufferRow >= 0 && text_bufferRow < len(text_buffer) && text_bufferCol < len(text_buffer[text_bufferRow]) {
				if text_buffer[text_bufferRow][text_bufferCol] != '\t' {
					termbox.SetChar(col, row, text_buffer[text_bufferRow][text_bufferCol])
				} else {
					termbox.SetCell(col, row, rune(' '), termbox.ColorDefault, termbox.ColorDefault)
				}
			} else if row+offsetY > len(text_buffer)-1 {
				termbox.SetCell(0, row, rune('*'), termbox.ColorBlue, termbox.ColorDefault)
			}
		}
		termbox.SetChar(col, row, '\n')
	}
}

func print_message(col, row int, fg, bg termbox.Attribute, message string) {
	for _, ch := range message {
		termbox.SetCell(col, row, ch, fg, bg)
		col += runewidth.RuneWidth(ch)
	}
}
func run_editor() {
	err := termbox.Init()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if len(os.Args) > 1 {
		source_file := os.Args[1]
		read_file(source_file)
	} else {
		source_file = "out.txt"
		text_buffer = append(text_buffer, []rune{})
	}

	for {
		COLS, ROWS = termbox.Size()
		ROWS--
		if COLS < 78 {
			COLS = 78
		}
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		display_text_buffer()
		// print_message(25, 11, termbox.ColorDefault, termbox.ColorDefault, "Monkey - The text editor")
		termbox.Flush()
		event := termbox.PollEvent()
		// This is to make sure that just the Escape key exits the editor
		if event.Type == termbox.EventKey && event.Key == termbox.KeyEsc {
			termbox.Close()
			break
		}
	}
}

func main() {
	run_editor()
}
