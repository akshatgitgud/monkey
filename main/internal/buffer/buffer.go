package buffer

import (
	"bufio"
	"os"
)

// Buffer manages text lines, modifications, and snapshots.
type Buffer struct {
	Lines      [][]rune
	SourceFile string
	Modified   bool
	UndoBuffer [][]rune
	CopyBuffer []rune
}

// New creates a Buffer from a file, or creates an empty one if filename is empty or doesn't exist.
func New(fileName string) *Buffer {
	b := &Buffer{
		Lines:      [][]rune{},
		UndoBuffer: [][]rune{},
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

	for row, line := range b.Lines {
		newLine := "\n"
		if row == len(b.Lines)-1 {
			newLine = ""
		}

		writeLine := string(line) + newLine
		if _, err := writer.WriteString(writeLine); err != nil {
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

// PushSnapshot creates a deep copy of Lines into UndoBuffer.
func (b *Buffer) PushSnapshot() {
	b.UndoBuffer = make([][]rune, len(b.Lines))
	for i, line := range b.Lines {
		b.UndoBuffer[i] = make([]rune, len(line))
		copy(b.UndoBuffer[i], line)
	}
}

// PullSnapshot restores Lines from UndoBuffer and clamps cursor. Returns clamped (row, col).
func (b *Buffer) PullSnapshot(curRow, curCol int) (int, int) {
	if len(b.UndoBuffer) == 0 {
		return curRow, curCol
	}

	b.Lines = make([][]rune, len(b.UndoBuffer))
	for i, line := range b.UndoBuffer {
		b.Lines[i] = make([]rune, len(line))
		copy(b.Lines[i], line)
	}

	if curRow >= len(b.Lines) {
		curRow = len(b.Lines) - 1
	}
	if curRow < 0 {
		curRow = 0
	}
	if curCol > len(b.Lines[curRow]) {
		curCol = len(b.Lines[curRow])
	}

	b.Modified = true
	return curRow, curCol
}
