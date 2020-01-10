package jty_test

import (
	"strings"
	"testing"

	"github.com/influxdata/jty/pkg/jty"
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
			}); err != jty.ErrOddInputFiles {
				t.Fatalf("expected ErrOddInputFiles, got %v", err)
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
			}); err != jty.ErrOddInputFiles {
				t.Fatalf("expected ErrOddInputFiles, got %v", err)
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
