package term

import "testing"

func TestPaint(t *testing.T) {
	got := Paint("ok", ColorGreen, []Style{StyleBold})
	want := "\x1b[1;32mok\x1b[0m"
	if got != want {
		t.Errorf("Paint() = %q, want %q", got, want)
	}
}

func TestTerminalPaintWithoutColor(t *testing.T) {
	term := &Terminal{ColorSupported: false}
	if got := term.Paint("ok", ColorGreen, []Style{StyleBold}); got != "ok" {
		t.Errorf("Paint() = %q, want %q", got, "ok")
	}
}

func TestNewTerminal(t *testing.T) {
	// Assert only that construction never panics; the detected value depends
	// on the environment.
	_ = NewTerminal()
}
