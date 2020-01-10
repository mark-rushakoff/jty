package jty_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/influxdata/jty/pkg/jty"
	"github.com/spf13/afero"
)

// JY is a Jsonnet->YAML pairing,
// where J is a standalone Jsonnet script
// and Y is the expected resulting YAML doc.
type JY struct {
	J string
	Y string
}

// WriteJ writes jy.J to the given path in the given fs,
// calling t.Fatal if there is an error.
func (jy JY) WriteJ(t *testing.T, fs afero.Fs, path string) {
	t.Helper()
	if err := afero.WriteFile(fs, path, []byte(jy.J), 0600); err != nil {
		t.Fatal(err)
	}
}

// ExpectY reads path from fs and compares its content to jy.Y.
// If there is an error reading the file, or if the content does not match,
// it calls t.Fatal.
func (jy JY) ExpectY(t *testing.T, fs afero.Fs, path string) {
	t.Helper()

	got, err := afero.ReadFile(fs, path)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != jy.Y {
		t.Fatalf("expected file content of %s to be %q; got %q", path, jy.Y, got)
	}
}

// Some simple Jsonnet->YAML pairings so tests don't have to repeat these values.
var (
	JYOneTwo = JY{
		J: `local one = {one: 1};
[
  one,
  one + {two: 2},
]
`,
		Y: `---
one: 1
---
one: 1
two: 2
...
`,
	}

	JYSeq = JY{
		J: `[ [n for n in std.range(1, 5)] ]`,
		Y: `---
- 1
- 2
- 3
- 4
- 5
...
`,
	}
)

// TestCommand wraps a jty.Command with stdout and stderr exposed as *bytes.Buffer.
type TestCommand struct {
	Cmd *jty.Command

	Stdout *bytes.Buffer
	Stderr *bytes.Buffer

	// The same FS as in Cmd, exposed here for shorthand.
	FS afero.Fs
}

// NewTestCommand returns a TestCommand that wraps a jty.Command
// with an in-memory filesystem and a new empty bytes.Buffer for stdout and stderr.
// The stdin string is provided as the standard input to the command.
func NewTestCommand(stdin string) *TestCommand {
	fs := afero.NewMemMapFs()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	return &TestCommand{
		Cmd: &jty.Command{
			Stdin:  strings.NewReader(stdin),
			Stdout: stdout,
			Stderr: stderr,

			FS: fs,
		},

		Stdout: stdout,
		Stderr: stderr,

		FS: fs,
	}
}
