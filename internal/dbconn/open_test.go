package dbconn

import (
	"strings"
	"testing"
)

func TestWithSimpleProtocol(t *testing.T) {
	t.Parallel()
	const want = "default_query_exec_mode=simple_protocol"

	t.Run("appends to postgres url", func(t *testing.T) {
		t.Parallel()
		in := "postgres://user:pass@host:5432/db?sslmode=require"
		got := WithSimpleProtocol(in)
		if got == in || !strings.Contains(got, want) {
			t.Fatalf("got %q, want %q appended", got, want)
		}
	})

	t.Run("unchanged when already set", func(t *testing.T) {
		t.Parallel()
		in := "postgres://host/db?" + want
		if got := WithSimpleProtocol(in); got != in {
			t.Fatalf("got %q, want unchanged %q", got, in)
		}
	})
}
