package editor

import (
	"fmt"
	"path/filepath"
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

// Segment represents a colored text chunk in the status bar.
type Segment struct {
	Text string
	Fg   termbox.Attribute
	Bg   termbox.Attribute
}

// Editor manages the viewport, input dispatching, and screen rendering.
type Editor struct {
	buf           *buffer.Buffer
	mode          Mode
	rows          int
	cols          int
	curRow        int
	curCol        int
	offsetRow     int
	offsetCol     int
	statusMessage string
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

// DisplayStatusBar renders a sleek, minimalist status bar with unified base styling.
func (e *Editor) DisplayStatusBar() {
	var modeText string
	var modeFg termbox.Attribute

	if e.mode == ModeEdit {
		modeText = " EDIT "
		modeFg = termbox.ColorGreen | termbox.AttrReverse | termbox.AttrBold
	} else {
		modeText = " VIEW "
		modeFg = termbox.ColorCyan | termbox.AttrReverse | termbox.AttrBold
	}

	barAttr := termbox.ColorDefault | termbox.AttrReverse
	barBold := termbox.ColorDefault | termbox.AttrReverse | termbox.AttrBold

	leftSegments := []Segment{
		{Text: modeText, Fg: modeFg, Bg: termbox.ColorDefault},
		{Text: " " + e.buf.SourceFile, Fg: barBold, Bg: termbox.ColorDefault},
	}

	if e.buf.Modified {
		leftSegments = append(leftSegments, Segment{
			Text: " [+] ",
			Fg:   termbox.ColorYellow | termbox.AttrReverse | termbox.AttrBold,
			Bg:   termbox.ColorDefault,
		})
	} else {
		leftSegments = append(leftSegments, Segment{
			Text: " ",
			Fg:   barAttr,
			Bg:   termbox.ColorDefault,
		})
	}

	rightText := fmt.Sprintf(" %s │ Ln %d, Col %d │ %s [%dL] ",
		e.detectFileType(),
		e.curRow+1,
		e.curCol+1,
		e.calculatePercentage(),
		e.buf.LineCount(),
	)

	rightSegments := []Segment{
		{Text: rightText, Fg: barAttr, Bg: termbox.ColorDefault},
	}

	leftWidth := 0
	for _, seg := range leftSegments {
		leftWidth += runewidth.StringWidth(seg.Text)
	}

	rightWidth := 0
	for _, seg := range rightSegments {
		rightWidth += runewidth.StringWidth(seg.Text)
	}

	// 1. Draw Left Segments
	curCol := 0
	for _, seg := range leftSegments {
		if curCol < e.cols {
			curCol = e.drawSegment(curCol, e.rows, seg)
		}
	}

	// 2. Draw Center / Message area
	rightStartCol := e.cols - rightWidth
	if rightStartCol < curCol {
		rightStartCol = curCol
	}

	msg := ""
	if e.statusMessage != "" {
		msg = " " + e.statusMessage + " "
	}
	msgWidth := runewidth.StringWidth(msg)

	if curCol+msgWidth <= rightStartCol {
		curCol = e.drawSegment(curCol, e.rows, Segment{Text: msg, Fg: barAttr, Bg: termbox.ColorDefault})
	}

	// Fill remaining gap with bar background
	for col := curCol; col < rightStartCol; col++ {
		termbox.SetCell(col, e.rows, ' ', barAttr, termbox.ColorDefault)
	}

	// 3. Draw Right Segments
	if rightStartCol >= 0 {
		rCol := rightStartCol
		for _, seg := range rightSegments {
			if rCol < e.cols {
				rCol = e.drawSegment(rCol, e.rows, seg)
			}
		}
	}
}

// drawSegment writes a Segment at a given row and column, returning the next column index.
func (e *Editor) drawSegment(col, row int, seg Segment) int {
	for _, ch := range seg.Text {
		termbox.SetCell(col, row, ch, seg.Fg, seg.Bg)
		col += runewidth.RuneWidth(ch)
	}
	return col
}

// calculatePercentage returns TOP, BOT, or an approximate percentage into the buffer.
func (e *Editor) calculatePercentage() string {
	total := e.buf.LineCount()
	if total <= 1 || e.curRow == 0 {
		return "TOP"
	}
	if e.curRow >= total-1 {
		return "BOT"
	}
	pct := int(float64(e.curRow) / float64(total-1) * 100)
	return fmt.Sprintf("%2d%%", pct)
}

// detectFileType returns the file extension or basename without dot.
func (e *Editor) detectFileType() string {
	ext := filepath.Ext(e.buf.SourceFile)
	if ext != "" {
		return strings.ToLower(strings.TrimPrefix(ext, "."))
	}
	base := filepath.Base(e.buf.SourceFile)
	if base != "" && base != "." {
		return strings.ToLower(base)
	}
	return "txt"
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
		e.statusMessage = ""
		return true
	}

	if event.Ch != 0 {
		if e.mode == ModeEdit {
			e.curCol = e.buf.InsertRune(e.curRow, e.curCol, event.Ch)
			e.statusMessage = ""
		} else {
			switch event.Ch {
			case 'q':
				return false
			case 'e':
				e.buf.PushSnapshot(e.curRow, e.curCol)
				e.mode = ModeEdit
				e.statusMessage = ""
			case 'w':
				if err := e.buf.WriteFile(e.buf.SourceFile); err != nil {
					e.statusMessage = "Error: " + err.Error()
				} else {
					e.statusMessage = fmt.Sprintf("\"%s\" %dL written", e.buf.SourceFile, e.buf.LineCount())
				}
			case 'c':
				e.copyLine()
			case 'p':
				e.pasteLine()
			case 'd':
				e.cutLine()
			case 'u':
				newRow, newCol, ok := e.buf.Undo(e.curRow, e.curCol)
				if ok {
					e.curRow, e.curCol = newRow, newCol
					e.statusMessage = "1 change undone"
				} else {
					e.statusMessage = "Already at oldest change"
				}
			case 'U':
				newRow, newCol, ok := e.buf.Redo(e.curRow, e.curCol)
				if ok {
					e.curRow, e.curCol = newRow, newCol
					e.statusMessage = "1 change redone"
				} else {
					e.statusMessage = "Already at newest change"
				}
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
			newRow, newCol, ok := e.buf.Redo(e.curRow, e.curCol)
			if ok {
				e.curRow, e.curCol = newRow, newCol
				e.statusMessage = "1 change redone"
			} else {
				e.statusMessage = "Already at newest change"
			}
		}
	case termbox.KeyEnter:
		if e.mode == ModeEdit {
			e.curRow, e.curCol = e.buf.InsertLine(e.curRow, e.curCol)
			e.statusMessage = ""
		}
	case termbox.KeyBackspace, termbox.KeyBackspace2:
		if e.mode == ModeEdit {
			e.curRow, e.curCol = e.buf.DeleteRune(e.curRow, e.curCol)
			e.statusMessage = ""
		}
	case termbox.KeyTab:
		if e.mode == ModeEdit {
			for range 4 {
				e.curCol = e.buf.InsertRune(e.curRow, e.curCol, ' ')
			}
			e.statusMessage = ""
		}
	case termbox.KeySpace:
		if e.mode == ModeEdit {
			e.curCol = e.buf.InsertRune(e.curRow, e.curCol, ' ')
			e.statusMessage = ""
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
		e.statusMessage = "1 line yanked"
	}
}

func (e *Editor) cutLine() {
	e.buf.PushSnapshot(e.curRow, e.curCol)
	newRow, newCol, cutLine := e.buf.CutLine(e.curRow)
	e.curRow = newRow
	e.curCol = newCol
	if cutLine != nil {
		_ = clipboard.WriteAll(string(cutLine))
		e.statusMessage = "1 line cut"
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
	if len(toPaste) == 1 {
		e.statusMessage = "1 line pasted"
	} else {
		e.statusMessage = fmt.Sprintf("%d lines pasted", len(toPaste))
	}
}
