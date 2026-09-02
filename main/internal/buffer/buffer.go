package buffer

import (
	"bufio"
	"os"
)

// Snapshot records a point-in-time state of the buffer and cursor position.
type Snapshot struct {
	Lines [][]rune
	Row   int
	Col   int
}

// Buffer manages text lines, modifications, and multi-level undo/redo history.
type Buffer struct {
	Lines      [][]rune
	SourceFile string
	Modified   bool
	UndoStack  []Snapshot
	RedoStack  []Snapshot
	CopyBuffer []rune
}

// New creates a Buffer from a file, or creates an empty one if filename is empty or doesn't exist.
func New(fileName string) *Buffer {
	b := &Buffer{
		Lines:      [][]rune{},
		UndoStack:  []Snapshot{},
		RedoStack:  []Snapshot{},
		CopyBuffer: []rune{},
	}

	if fileName == "" {
		b.SourceFile = "out.txt"
		b.Lines = append(b.Lines, []rune{})
		return b
	}

	b.SourceFile = fileName
	b.ReadFile(fileName)
	return b
}

// ReadFile loads text content from the specified file into the buffer.
func (b *Buffer) ReadFile(fileName string) {
	b.Lines = [][]rune{}
	file, err := os.Open(fileName)
	if err != nil {
		b.SourceFile = fileName
		b.Lines = append(b.Lines, []rune{})
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		line := scanner.Text()
		b.Lines = append(b.Lines, []rune{})
		for _, ch := range line {
			b.Lines[lineNumber] = append(b.Lines[lineNumber], ch)
		}
		lineNumber++
	}

	if lineNumber == 0 {
		b.Lines = append(b.Lines, []rune{})
	}
}

// WriteFile saves the buffer contents to the specified file.
func (b *Buffer) WriteFile(fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for _, line := range b.Lines {
		if _, err := writer.WriteString(string(line) + "\n"); err != nil {
			return err
		}
	}

	if err := writer.Flush(); err != nil {
		return err
	}

	b.Modified = false
	return nil
}

// LineCount returns the number of lines in the buffer.
func (b *Buffer) LineCount() int {
	return len(b.Lines)
}

// LineLen returns the number of runes on the given row.
func (b *Buffer) LineLen(row int) int {
	if row < 0 || row >= len(b.Lines) {
		return 0
	}
	return len(b.Lines[row])
}

// GetRune returns the rune at row, col.
func (b *Buffer) GetRune(row, col int) (rune, bool) {
	if row < 0 || row >= len(b.Lines) || col < 0 || col >= len(b.Lines[row]) {
		return 0, false
	}
	return b.Lines[row][col], true
}

// InsertRune inserts a rune at the specified cursor position and returns the new column.
func (b *Buffer) InsertRune(row, col int, ch rune) int {
	if row < 0 || row >= len(b.Lines) {
		return col
	}

	if col > len(b.Lines[row]) {
		col = len(b.Lines[row])
	}

	newLine := make([]rune, len(b.Lines[row])+1)
	copy(newLine[:col], b.Lines[row][:col])
	newLine[col] = ch
	copy(newLine[col+1:], b.Lines[row][col:])
	b.Lines[row] = newLine
	b.Modified = true
	return col + 1
}

// DeleteRune deletes the rune before (row, col) or merges with the previous line.
// Returns the new row and col.
func (b *Buffer) DeleteRune(row, col int) (int, int) {
	if row < 0 || row >= len(b.Lines) {
		return row, col
	}

	if col > 0 {
		col--
		delLine := make([]rune, len(b.Lines[row])-1)
		copy(delLine[:col], b.Lines[row][:col])
		copy(delLine[col:], b.Lines[row][col+1:])
		b.Lines[row] = delLine
		b.Modified = true
		return row, col
	}

	if row > 0 {
		prevRow := row - 1
		newCol := len(b.Lines[prevRow])
		appendLine := make([]rune, len(b.Lines[row]))
		copy(appendLine, b.Lines[row])

		mergedLine := make([]rune, len(b.Lines[prevRow])+len(appendLine))
		copy(mergedLine[:len(b.Lines[prevRow])], b.Lines[prevRow])
		copy(mergedLine[len(b.Lines[prevRow]):], appendLine)

		newBuffer := make([][]rune, len(b.Lines)-1)
		copy(newBuffer[:prevRow], b.Lines[:prevRow])
		newBuffer[prevRow] = mergedLine
		copy(newBuffer[prevRow+1:], b.Lines[row+1:])
		b.Lines = newBuffer
		b.Modified = true
		return prevRow, newCol
	}

	return row, col
}

// InsertLine splits the current line at (row, col) and inserts the remainder on the next line.
// Returns the new row and col.
func (b *Buffer) InsertLine(row, col int) (int, int) {
	if row < 0 || row >= len(b.Lines) {
		return row, col
	}

	if col > len(b.Lines[row]) {
		col = len(b.Lines[row])
	}

	leftLine := make([]rune, len(b.Lines[row][:col]))
	copy(leftLine, b.Lines[row][:col])

	rightLine := make([]rune, len(b.Lines[row][col:]))
	copy(rightLine, b.Lines[row][col:])

	b.Lines[row] = leftLine
	newBuffer := make([][]rune, len(b.Lines)+1)
	copy(newBuffer[:row+1], b.Lines[:row+1])
	newBuffer[row+1] = rightLine
	copy(newBuffer[row+2:], b.Lines[row+1:])
	b.Lines = newBuffer
	b.Modified = true

	return row + 1, 0
}

// CopyLine saves a copy of the line at row to CopyBuffer and returns it.
func (b *Buffer) CopyLine(row int) []rune {
	if row < 0 || row >= len(b.Lines) {
		return nil
	}
	copied := make([]rune, len(b.Lines[row]))
	copy(copied, b.Lines[row])
	b.CopyBuffer = copied
	return copied
}

// CutLine copies and deletes the line at row. Returns new (row, col) and copied line.
func (b *Buffer) CutLine(row int) (int, int, []rune) {
	copied := b.CopyLine(row)
	if row < 0 || row >= len(b.Lines) || len(b.Lines) < 2 {
		return row, 0, copied
	}

	newBuffer := make([][]rune, len(b.Lines)-1)
	copy(newBuffer[:row], b.Lines[:row])
	copy(newBuffer[row:], b.Lines[row+1:])
	b.Lines = newBuffer
	b.Modified = true

	newRow := row
	if newRow > 0 {
		newRow--
	}
	return newRow, 0, copied
}

// PasteLines inserts one or more lines after row. Returns new row and col.
func (b *Buffer) PasteLines(row int, lines [][]rune) (int, int) {
	if len(lines) == 0 {
		return row, 0
	}

	insertAt := row + 1
	if insertAt > len(b.Lines) {
		insertAt = len(b.Lines)
	}

	toInsert := make([][]rune, len(lines))
	for i, l := range lines {
		toInsert[i] = make([]rune, len(l))
		copy(toInsert[i], l)
	}

	newBuffer := make([][]rune, 0, len(b.Lines)+len(toInsert))
	newBuffer = append(newBuffer, b.Lines[:insertAt]...)
	newBuffer = append(newBuffer, toInsert...)
	newBuffer = append(newBuffer, b.Lines[insertAt:]...)
	b.Lines = newBuffer
	b.Modified = true

	return insertAt + len(toInsert) - 1, 0
}

func copyLines(lines [][]rune) [][]rune {
	copied := make([][]rune, len(lines))
	for i, l := range lines {
		copied[i] = make([]rune, len(l))
		copy(copied[i], l)
	}
	return copied
}

func linesEqual(a, b [][]rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}

// PushSnapshot creates a point-in-time snapshot with cursor position and pushes it to UndoStack.
func (b *Buffer) PushSnapshot(curRow, curCol int) {
	if len(b.UndoStack) > 0 && linesEqual(b.UndoStack[len(b.UndoStack)-1].Lines, b.Lines) {
		return
	}
	if len(b.UndoStack) >= 200 {
		b.UndoStack = b.UndoStack[1:]
	}
	b.UndoStack = append(b.UndoStack, Snapshot{
		Lines: copyLines(b.Lines),
		Row:   curRow,
		Col:   curCol,
	})
	b.RedoStack = nil
}

// Undo restores the previous state from UndoStack, pushing current state to RedoStack.
// Returns the restored (row, col) and true if an undo occurred.
func (b *Buffer) Undo(curRow, curCol int) (int, int, bool) {
	if len(b.UndoStack) == 0 {
		return curRow, curCol, false
	}

	lastIdx := len(b.UndoStack) - 1
	snapshot := b.UndoStack[lastIdx]
	b.UndoStack = b.UndoStack[:lastIdx]

	b.RedoStack = append(b.RedoStack, Snapshot{
		Lines: copyLines(b.Lines),
		Row:   curRow,
		Col:   curCol,
	})

	b.Lines = copyLines(snapshot.Lines)
	b.Modified = true

	row := snapshot.Row
	col := snapshot.Col
	if row >= len(b.Lines) {
		row = len(b.Lines) - 1
	}
	if row < 0 {
		row = 0
	}
	if col > len(b.Lines[row]) {
		col = len(b.Lines[row])
	}

	return row, col, true
}

// Redo restores the state from RedoStack, pushing current state to UndoStack.
// Returns the restored (row, col) and true if a redo occurred.
func (b *Buffer) Redo(curRow, curCol int) (int, int, bool) {
	if len(b.RedoStack) == 0 {
		return curRow, curCol, false
	}

	lastIdx := len(b.RedoStack) - 1
	snapshot := b.RedoStack[lastIdx]
	b.RedoStack = b.RedoStack[:lastIdx]

	b.UndoStack = append(b.UndoStack, Snapshot{
		Lines: copyLines(b.Lines),
		Row:   curRow,
		Col:   curCol,
	})

	b.Lines = copyLines(snapshot.Lines)
	b.Modified = true

	row := snapshot.Row
	col := snapshot.Col
	if row >= len(b.Lines) {
		row = len(b.Lines) - 1
	}
	if row < 0 {
		row = 0
	}
	if col > len(b.Lines[row]) {
		col = len(b.Lines[row])
	}

	return row, col, true
}

// CanUndo returns true if an undo step is available.
func (b *Buffer) CanUndo() bool {
	return len(b.UndoStack) > 0
}

// CanRedo returns true if a redo step is available.
func (b *Buffer) CanRedo() bool {
	return len(b.RedoStack) > 0
}
