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

func TestCommand_DryRun(t *testing.T) {
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
