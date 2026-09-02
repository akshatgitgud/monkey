package buffer

import (
	"testing"
)

func TestBufferInsertAndGetRune(t *testing.T) {
	b := New("")
	if b.LineCount() != 1 {
		t.Fatalf("expected 1 line, got %d", b.LineCount())
	}

	col := b.InsertRune(0, 0, 'H')
	col = b.InsertRune(0, col, 'i')

	if b.LineLen(0) != 2 {
		t.Fatalf("expected line len 2, got %d", b.LineLen(0))
	}

	if string(b.Lines[0]) != "Hi" {
		t.Fatalf("expected 'Hi', got '%s'", string(b.Lines[0]))
	}
}

func TestBufferDeleteRune(t *testing.T) {
	b := New("")
	b.InsertRune(0, 0, 'A')
	b.InsertRune(0, 1, 'B')

	row, col := b.DeleteRune(0, 2)
	if row != 0 || col != 1 {
		t.Fatalf("expected (0, 1), got (%d, %d)", row, col)
	}

	if string(b.Lines[0]) != "A" {
		t.Fatalf("expected 'A', got '%s'", string(b.Lines[0]))
	}
}

func TestBufferInsertLineAndMerge(t *testing.T) {
	b := New("")
	for _, ch := range "Hello World" {
		b.InsertRune(0, len(b.Lines[0]), ch)
	}

	// Split at 'Hello ' | 'World' (col 6)
	row, col := b.InsertLine(0, 6)
	if row != 1 || col != 0 {
		t.Fatalf("expected (1, 0), got (%d, %d)", row, col)
	}
	if b.LineCount() != 2 {
		t.Fatalf("expected 2 lines, got %d", b.LineCount())
	}
	if string(b.Lines[0]) != "Hello " || string(b.Lines[1]) != "World" {
		t.Fatalf("unexpected lines: '%s', '%s'", string(b.Lines[0]), string(b.Lines[1]))
	}

	// Merge back by deleting at row 1, col 0
	row, col = b.DeleteRune(1, 0)
	if row != 0 || col != 6 {
		t.Fatalf("expected (0, 6), got (%d, %d)", row, col)
	}
	if b.LineCount() != 1 || string(b.Lines[0]) != "Hello World" {
		t.Fatalf("merge failed: '%s'", string(b.Lines[0]))
	}
}

func TestBufferMultiLevelUndoRedo(t *testing.T) {
	b := New("")
	b.InsertRune(0, 0, 'A')

	// Checkpoint 1
	b.PushSnapshot(0, 1)
	b.InsertRune(0, 1, 'B')

	// Checkpoint 2
	b.PushSnapshot(0, 2)
	b.InsertRune(0, 2, 'C')

	if string(b.Lines[0]) != "ABC" {
		t.Fatalf("expected 'ABC', got '%s'", string(b.Lines[0]))
	}

	// Undo to Checkpoint 2 ('AB')
	row, col, ok := b.Undo(0, 3)
	if !ok || string(b.Lines[0]) != "AB" || row != 0 || col != 2 {
		t.Fatalf("expected undo to 'AB' at (0, 2), got '%s' at (%d, %d)", string(b.Lines[0]), row, col)
	}

	// Undo to Checkpoint 1 ('A')
	row, col, ok = b.Undo(row, col)
	if !ok || string(b.Lines[0]) != "A" || row != 0 || col != 1 {
		t.Fatalf("expected undo to 'A' at (0, 1), got '%s' at (%d, %d)", string(b.Lines[0]), row, col)
	}

	// No more undo
	_, _, ok = b.Undo(row, col)
	if ok {
		t.Fatalf("expected undo to fail when stack is empty")
	}

	// Redo to Checkpoint 2 ('AB')
	row, col, ok = b.Redo(row, col)
	if !ok || string(b.Lines[0]) != "AB" || row != 0 || col != 2 {
		t.Fatalf("expected redo to 'AB' at (0, 2), got '%s' at (%d, %d)", string(b.Lines[0]), row, col)
	}

	// Redo to 'ABC'
	row, col, ok = b.Redo(row, col)
	if !ok || string(b.Lines[0]) != "ABC" || row != 0 || col != 3 {
		t.Fatalf("expected redo to 'ABC' at (0, 3), got '%s' at (%d, %d)", string(b.Lines[0]), row, col)
	}
}
