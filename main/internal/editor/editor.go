package editor

import (
	"strconv"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"

	"monkey/main/internal/buffer"
)

// Mode represents the current editor operational mode.
type Mode int

const (
	ModeView Mode = 0
	ModeEdit Mode = 1
)

// Editor manages the viewport, input dispatching, and screen rendering.
type Editor struct {
	buf       *buffer.Buffer
	mode      Mode
	rows      int
	cols      int
	curRow    int
	curCol    int
	offsetRow int
	offsetCol int
}

// New creates and initializes an Editor instance with the given file.
func New(filename string) *Editor {
	return &Editor{
		buf:  buffer.New(filename),
		mode: ModeView,
	}
}

// Run starts the termbox event loop.
func (e *Editor) Run() error {
	err := termbox.Init()
	if err != nil {
		return err
	}
	defer termbox.Close()

	for {
		cols, rows := termbox.Size()
		rows--
		if rows < 1 {
			rows = 1
		}
		e.cols = cols
		e.rows = rows

		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		e.Scroll()
		e.DisplayTextBuffer()
		e.DisplayStatusBar()
		termbox.SetCursor(e.curCol-e.offsetCol, e.curRow-e.offsetRow)

		if err := termbox.Flush(); err != nil {
			return err
		}

		if !e.ProcessKeypress() {
			break
		}
	}

	return nil
}

// Scroll adjusts offsetRow and offsetCol so the cursor is always visible in the viewport.
func (e *Editor) Scroll() {
	if e.curRow < e.offsetRow {
		e.offsetRow = e.curRow
	}
	if e.curCol < e.offsetCol {
		e.offsetCol = e.curCol
	}
	if e.curRow >= e.offsetRow+e.rows {
		e.offsetRow = e.curRow - e.rows + 1
	}
	if e.curCol >= e.offsetCol+e.cols {
		e.offsetCol = e.curCol - e.cols + 1
	}
}

// DisplayTextBuffer renders the visible slice of lines onto the termbox screen.
func (e *Editor) DisplayTextBuffer() {
	for row := 0; row < e.rows; row++ {
		textBufferRow := row + e.offsetRow
		if textBufferRow >= e.buf.LineCount() {
			termbox.SetCell(0, row, rune('*'), termbox.ColorBlue, termbox.ColorDefault)
			continue
		}
		for col := 0; col < e.cols; col++ {
			textBufferCol := col + e.offsetCol
			if ch, ok := e.buf.GetRune(textBufferRow, textBufferCol); ok {
				if ch != '\t' {
					termbox.SetChar(col, row, ch)
				} else {
					termbox.SetCell(col, row, rune(' '), termbox.ColorDefault, termbox.ColorDefault)
				}
			}
		}
	}
}

// DisplayStatusBar renders the mode, filename, line count, copy/undo state, and cursor coordinate.
func (e *Editor) DisplayStatusBar() {
	var modeStatus string
	if e.mode == ModeEdit {
		modeStatus = "EDIT: "
	} else {
		modeStatus = "VIEW: "
	}

	fileStatus := e.buf.SourceFile + " - " + strconv.Itoa(e.buf.LineCount()) + " lines"
	if e.buf.Modified {
		fileStatus += " modified"
	} else {
		fileStatus += " saved"
	}

	cursorStatus := " Row " + strconv.Itoa(e.curRow) + ", Col " + strconv.Itoa(e.curCol) + " "

	var copyStatus string
	if len(e.buf.CopyBuffer) > 0 {
		copyStatus = " [Copy]"
	}

	var undoStatus string
	if e.buf.CanUndo() {
		undoStatus = " [Undo]"
	}

	var redoStatus string
	if e.buf.CanRedo() {
		redoStatus = " [Redo]"
	}

	usedSpace := runewidth.StringWidth(modeStatus + fileStatus + copyStatus + undoStatus + redoStatus + cursorStatus)
	spaces := ""
	if e.cols > usedSpace {
		spaces = strings.Repeat(" ", e.cols-usedSpace)
	}

	message := modeStatus + fileStatus + copyStatus + undoStatus + redoStatus + spaces + cursorStatus
	if totalWidth := runewidth.StringWidth(message); totalWidth < e.cols {
		message += strings.Repeat(" ", e.cols-totalWidth)
	}

	e.PrintMessage(0, e.rows, termbox.ColorDefault|termbox.AttrReverse, termbox.ColorDefault, message)
}

// PrintMessage writes a string with specified attributes at a given row and column.
func (e *Editor) PrintMessage(col, row int, fg, bg termbox.Attribute, message string) {
	for _, ch := range message {
		termbox.SetCell(col, row, ch, fg, bg)
		col += runewidth.RuneWidth(ch)
	}
}

// ProcessKeypress polls a key event and executes the appropriate command. Returns false on exit.
func (e *Editor) ProcessKeypress() bool {
	event := termbox.PollEvent()
	if event.Type == termbox.EventError {
		panic(event.Err)
	}
	if event.Type != termbox.EventKey {
		return true
	}

	if event.Key == termbox.KeyEsc {
		e.mode = ModeView
		return true
	}

	if event.Ch != 0 {
		if e.mode == ModeEdit {
			e.curCol = e.buf.InsertRune(e.curRow, e.curCol, event.Ch)
		} else {
			switch event.Ch {
			case 'q':
				return false
			case 'e':
				e.buf.PushSnapshot(e.curRow, e.curCol)
				e.mode = ModeEdit
			case 'w':
				_ = e.buf.WriteFile(e.buf.SourceFile)
			case 'c':
				e.copyLine()
			case 'p':
				e.pasteLine()
			case 'd':
				e.cutLine()
			case 'u':
				e.curRow, e.curCol, _ = e.buf.Undo(e.curRow, e.curCol)
			case 'U':
				e.curRow, e.curCol, _ = e.buf.Redo(e.curRow, e.curCol)
			case 'j':
				if e.curRow < e.buf.LineCount()-1 {
					e.curRow++
				}
			case 'k':
				if e.curRow > 0 {
					e.curRow--
				}
			case 'h':
				if e.curCol > 0 {
					e.curCol--
				} else if e.curRow > 0 {
					e.curRow--
					e.curCol = e.buf.LineLen(e.curRow)
				}
			case 'l':
				if e.curCol < e.buf.LineLen(e.curRow) {
					e.curCol++
				} else if e.curRow < e.buf.LineCount()-1 {
					e.curRow++
					e.curCol = 0
				}
			}
		}
		return true
	}

	switch event.Key {
	case termbox.KeyCtrlR:
		if e.mode == ModeView {
			e.curRow, e.curCol, _ = e.buf.Redo(e.curRow, e.curCol)
		}
	case termbox.KeyEnter:
		if e.mode == ModeEdit {
			e.curRow, e.curCol = e.buf.InsertLine(e.curRow, e.curCol)
		}
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		if e.mode == ModeEdit {
			e.curRow, e.curCol = e.buf.DeleteRune(e.curRow, e.curCol)
		}
	case termbox.KeyTab:
		if e.mode == ModeEdit {
			for range 4 {
				e.curCol = e.buf.InsertRune(e.curRow, e.curCol, ' ')
			}
		}
	case termbox.KeySpace:
		if e.mode == ModeEdit {
			e.curCol = e.buf.InsertRune(e.curRow, e.curCol, ' ')
		}
	case termbox.KeyHome:
		e.curCol = 0
	case termbox.KeyEnd:
		e.curCol = e.buf.LineLen(e.curRow)
	case termbox.KeyPgup:
		step := e.rows / 4
		if e.curRow-step >= 0 {
			e.curRow -= step
		} else {
			e.curRow = 0
		}
	case termbox.KeyPgdn:
		step := e.rows / 4
		if e.curRow+step < e.buf.LineCount() {
			e.curRow += step
		} else if e.buf.LineCount() > 0 {
			e.curRow = e.buf.LineCount() - 1
		}
	case termbox.KeyArrowUp:
		if e.curRow > 0 {
			e.curRow--
		}
	case termbox.KeyArrowDown:
		if e.curRow < e.buf.LineCount()-1 {
			e.curRow++
		}
	case termbox.KeyArrowLeft:
		if e.curCol > 0 {
			e.curCol--
		} else if e.curRow > 0 {
			e.curRow--
			e.curCol = e.buf.LineLen(e.curRow)
		}
	case termbox.KeyArrowRight:
		if e.curCol < e.buf.LineLen(e.curRow) {
			e.curCol++
		} else if e.curRow < e.buf.LineCount()-1 {
			e.curRow++
			e.curCol = 0
		}
	}

	if e.curCol > e.buf.LineLen(e.curRow) {
		e.curCol = e.buf.LineLen(e.curRow)
	}

	return true
}

func (e *Editor) copyLine() {
	copied := e.buf.CopyLine(e.curRow)
	if copied != nil {
		_ = clipboard.WriteAll(string(copied))
	}
}

func (e *Editor) cutLine() {
	e.buf.PushSnapshot(e.curRow, e.curCol)
	newRow, newCol, cutLine := e.buf.CutLine(e.curRow)
	e.curRow = newRow
	e.curCol = newCol
	if cutLine != nil {
		_ = clipboard.WriteAll(string(cutLine))
	}
}

func (e *Editor) pasteLine() {
	clipText, err := clipboard.ReadAll()
	var toPaste [][]rune

	if err == nil && len(clipText) > 0 {
		clipText = strings.ReplaceAll(clipText, "\r\n", "\n")
		lines := strings.Split(clipText, "\n")
		for _, l := range lines {
			toPaste = append(toPaste, []rune(l))
		}
	} else if len(e.buf.CopyBuffer) > 0 {
		toPaste = [][]rune{e.buf.CopyBuffer}
	} else {
		return
	}

	e.buf.PushSnapshot(e.curRow, e.curCol)
	e.curRow, e.curCol = e.buf.PasteLines(e.curRow, toPaste)
}
