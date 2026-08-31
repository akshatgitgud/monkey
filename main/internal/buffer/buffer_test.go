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

func TestBufferUndoSnapshot(t *testing.T) {
	b := New("")
	b.InsertRune(0, 0, 'A')
	b.PushSnapshot()

	b.InsertRune(0, 1, 'B')
	if string(b.Lines[0]) != "AB" {
		t.Fatalf("expected 'AB', got '%s'", string(b.Lines[0]))
	}

	row, col := b.PullSnapshot(0, 2)
	if row != 0 || col != 1 {
		t.Fatalf("expected cursor clamped to (0, 1), got (%d, %d)", row, col)
	}
	if string(b.Lines[0]) != "A" {
		t.Fatalf("expected 'A' after undo, got '%s'", string(b.Lines[0]))
	}
}
