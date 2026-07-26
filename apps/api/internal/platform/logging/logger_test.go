package logging

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestLevelFilterWriter_DropsBelowMin(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	w := levelFilterWriter{Writer: &buf, MinLevel: zerolog.ErrorLevel}

	if _, err := w.WriteLevel(zerolog.InfoLevel, []byte(`{"level":"info"}`+"\n")); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected info to be dropped, got %q", buf.String())
	}

	if _, err := w.WriteLevel(zerolog.ErrorLevel, []byte(`{"level":"error"}`+"\n")); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "error") {
		t.Fatalf("expected error to be written, got %q", buf.String())
	}
}
