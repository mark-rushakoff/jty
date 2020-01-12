package jty_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/influxdata/jty/pkg/jty"
	"github.com/spf13/afero"
)

func TestCommand_PositionalArgs(t *testing.T) {
	tc := NewTestCommand("")
	JYOneTwo.WriteJ(t, tc.FS, "in1.jsonnet")

	if err := tc.Cmd.Run(&jty.Flags{
		Args: []string{"in1.jsonnet", "out1.yml"},
	}); err != nil {
		t.Fatal(err)
	}

	JYOneTwo.ExpectY(t, tc.FS, "out1.yml")

	if tc.Stdout.String() != "" {
		t.Fatalf("expected no standard output, got %q", tc.Stdout.String())
	}
	if tc.Stderr.String() != "" {
		t.Fatalf("expected no standard error, got %q", tc.Stderr.String())
	}
}

func TestCommand_PositionalArgs_Odd(t *testing.T) {
	tc := NewTestCommand("")

	err := tc.Cmd.Run(&jty.Flags{
		Args: []string{"in1.jsonnet"},
	})
	if err != jty.ErrOddInputFiles {
		t.Fatalf("expected ErrOddInputFiles, got %v", err)
	}
}

func TestCommand_PositionalArgs_Missing(t *testing.T) {
	tc := NewTestCommand("")

	err := tc.Cmd.Run(new(jty.Flags))
	if err != jty.ErrNoInputFiles {
		t.Fatalf("expected ErrNoInputFiles, got %v", err)
	}
}

func TestCommand_PositionalArgs_DryRun(t *testing.T) {
	tc := NewTestCommand("")
	JYOneTwo.WriteJ(t, tc.FS, "in1.jsonnet")

	if err := tc.Cmd.Run(&jty.Flags{
		Args:   []string{"in1.jsonnet", "out1.yml"},
		DryRun: true,
	}); err != nil {
		t.Fatal(err)
	}

	if tc.Stderr.String() != "" {
		t.Fatalf("expected no standard error, got %q", tc.Stderr.String())
	}

	want := "would process in1.jsonnet and save YAML output to out1.yml\n"
	if !strings.Contains(tc.Stdout.String(), want) {
		t.Fatalf("expected stdout to contain %q but it didn't", want)
	}
}

func TestCommand_StdinArgs(t *testing.T) {
	stdinArgs := []string{
		"in1.jsonnet", "out1.yml",
		"in2.jsonnet", "out2.yml",
		"", // So that a separatator lands after the last output parameter.
	}

	for name, delim := range map[string]string{
		"newline": "\n",
		"zero":    "\x00",
	} {
		t.Run(name, func(t *testing.T) {
			stdin := strings.Join(stdinArgs, delim)

			tc := NewTestCommand(stdin)
			JYOneTwo.WriteJ(t, tc.FS, "in1.jsonnet")
			JYSeq.WriteJ(t, tc.FS, "in2.jsonnet")

			if err := tc.Cmd.Run(&jty.Flags{
				FromStdin: true,
				Zero:      delim == "\x00",
			}); err != nil {
				t.Fatal(err)
			}

			JYOneTwo.ExpectY(t, tc.FS, "out1.yml")
			JYSeq.ExpectY(t, tc.FS, "out2.yml")

			if tc.Stdout.String() != "" {
				t.Fatalf("expected no standard output, got %q", tc.Stdout.String())
			}
			if tc.Stderr.String() != "" {
				t.Fatalf("expected no standard error, got %q", tc.Stderr.String())
			}
		})
	}
}

// Be strict and require terminator following last output file.
func TestCommand_StdinArgs_MissingFinalTerminator(t *testing.T) {
	stdinArgs := []string{
		"in1.jsonnet", "out1.yml",
		"in2.jsonnet", "out2.yml",
	}

	for name, delim := range map[string]string{
		"newline": "\n",
		"zero":    "\x00",
	} {
		t.Run(name, func(t *testing.T) {
			stdin := strings.Join(stdinArgs, delim)

			tc := NewTestCommand(stdin)
			JYOneTwo.WriteJ(t, tc.FS, "in1.jsonnet")
			JYSeq.WriteJ(t, tc.FS, "in2.jsonnet")

			if err := tc.Cmd.Run(&jty.Flags{
				FromStdin: true,
				Zero:      delim == "\x00",
			}); err != jty.ErrOddInputFilesStdin {
				t.Fatalf("expected ErrOddInputFilesStdin, got %v", err)
			}
		})
	}
}

func TestCommand_StdinArgs_Odd(t *testing.T) {
	stdinArgs := []string{
		"in1.jsonnet",
		"", // So that a separatator lands after the last output parameter.
	}

	for name, delim := range map[string]string{
		"newline": "\n",
		"zero":    "\x00",
	} {
		t.Run(name, func(t *testing.T) {
			stdin := strings.Join(stdinArgs, delim)

			tc := NewTestCommand(stdin)

			if err := tc.Cmd.Run(&jty.Flags{
				FromStdin: true,
				Zero:      delim == "\x00",
			}); err != jty.ErrOddInputFilesStdin {
				t.Fatalf("expected ErrOddInputFilesStdin, got %v", err)
			}
		})
	}
}

func TestCommand_StdinArgs_Empty(t *testing.T) {
	for name, delim := range map[string]string{
		"newline": "\n",
		"zero":    "\x00",
	} {
		t.Run(name, func(t *testing.T) {
			stdin := ""

			tc := NewTestCommand(stdin)

			if err := tc.Cmd.Run(&jty.Flags{
				FromStdin: true,
				Zero:      delim == "\x00",
			}); err != jty.ErrNoInputFiles {
				t.Fatalf("expected ErrNoInputFiles, got %v", err)
			}
		})
	}
}

func TestCommand_StdinArgs_DryRun(t *testing.T) {
	stdinArgs := []string{
		"in1.jsonnet", "out1.yml",
		"in2.jsonnet", "out2.yml",
		"", // So that a separatator lands after the last output parameter.
	}

	for name, delim := range map[string]string{
		"newline": "\n",
		"zero":    "\x00",
	} {
		t.Run(name, func(t *testing.T) {
			stdin := strings.Join(stdinArgs, delim)

			tc := NewTestCommand(stdin)
			JYOneTwo.WriteJ(t, tc.FS, "in1.jsonnet")
			JYSeq.WriteJ(t, tc.FS, "in2.jsonnet")

			if err := tc.Cmd.Run(&jty.Flags{
				FromStdin: true,
				Zero:      delim == "\x00",
				DryRun:    true,
			}); err != nil {
				t.Fatal(err)
			}

			want := "would process in1.jsonnet and save YAML output to out1.yml\n"
			if !strings.Contains(tc.Stdout.String(), want) {
				t.Fatalf("expected stdout to contain %q but it didn't", want)
			}
			want = "would process in2.jsonnet and save YAML output to out2.yml\n"
			if !strings.Contains(tc.Stdout.String(), want) {
				t.Fatalf("expected stdout to contain %q but it didn't", want)
			}

			if tc.Stderr.String() != "" {
				t.Fatalf("expected no standard error, got %q", tc.Stderr.String())
			}
		})
	}
}

// Regression test for a data race.
func TestCommand_MultipleLogs(t *testing.T) {
	tc := NewTestCommand("")

	if err := tc.Cmd.Run(&jty.Flags{
		// Multiple input files that don't exist.
		Args: []string{"1.jsonnet", "1.yml", "2.jsonnet", "2.yml", "3.jsonnet", "3.yml"},
	}); err != jty.ErrEncounteredErrors {
		t.Fatalf("expected ErrEncounteredErrors, got %v", err)
	}
}

func TestCommand_Imports(t *testing.T) {
	tc := NewTestCommand("")

	// The Command is hardcoded to instantiate a Jsonnet VM with a FileImporter,
	// so the imports must reside on disk even though the source files are on an in-mem afero fs.
	libdir, err := ioutil.TempDir("", "jty-imports-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(libdir)

	highDir := filepath.Join(libdir, "high-priority")
	lowDir := filepath.Join(libdir, "low-priority")
	if err := os.Mkdir(highDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(lowDir, 0700); err != nil {
		t.Fatal(err)
	}

	// both.libsonnet is present but different in both directories.
	// Expect X=1.
	if err := ioutil.WriteFile(
		filepath.Join(highDir, "both.libsonnet"),
		[]byte("{X: 1}"),
		0600,
	); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(
		filepath.Join(lowDir, "both.libsonnet"),
		[]byte("{X: 2}"),
		0600,
	); err != nil {
		t.Fatal(err)
	}

	if err := ioutil.WriteFile(
		filepath.Join(highDir, "up.libsonnet"),
		[]byte("{up: true}"),
		0600,
	); err != nil {
		t.Fatal(err)
	}
	if err := ioutil.WriteFile(
		filepath.Join(lowDir, "down.libsonnet"),
		[]byte("{down: true}"),
		0600,
	); err != nil {
		t.Fatal(err)
	}

	// Now write the in-mem fs source.
	if err := afero.WriteFile(tc.FS, "in.jsonnet", []byte(`
local both = import 'both.libsonnet';
local down = import 'down.libsonnet';
local up = import 'up.libsonnet';

[
both + down + up,
]
`), 0600); err != nil {
		t.Fatal(err)
	}

	// Run the command.
	if err := tc.Cmd.Run(&jty.Flags{
		Args: []string{"in.jsonnet", "out.yml"},

		// Rightmost wins.
		JPaths: []string{lowDir, highDir},
	}); err != nil {
		t.Fatal(err)
	}

	// Even though the imports were from disk, the output is in the in-mem fs.
	got, err := afero.ReadFile(tc.FS, "out.yml")
	if err != nil {
		t.Fatal(err)
	}

	expYAML := `---
X: 1
down: true
up: true
...
`
	if string(got) != expYAML {
		t.Fatalf("expected file content of out.yml to be %q; got %q", expYAML, got)
	}
}
