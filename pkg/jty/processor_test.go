package jty_test

import (
	"bytes"
	"runtime"
	"strings"
	"testing"

	"github.com/influxdata/jty/pkg/jty"
	"github.com/spf13/afero"
)

func TestProcessor_DryRun(t *testing.T) {
	fs := afero.NewMemMapFs()
	log := new(bytes.Buffer)
	p := jty.NewProcessor(runtime.GOMAXPROCS(-1), fs, log)

	dryRunOut := new(bytes.Buffer)
	p.DryRunDest = dryRunOut

	p.Process("in1.jsonnet", "out1.yml")
	p.Process("in2.jsonnet", "out2.yml")
	p.Close()

	if got := log.String(); got != "" {
		t.Errorf("expected empty log, got %q", got)
	}

	out := dryRunOut.String()
	want1 := "would process in1.jsonnet and save YAML output to out1.yml\n"
	if !strings.Contains(out, want1) {
		t.Errorf("expected output %q to contain %q but it didn't", out, want1)
	}
	want2 := "would process in2.jsonnet and save YAML output to out2.yml\n"
	if !strings.Contains(out, want2) {
		t.Errorf("expected output %q to contain %q but it didn't", out, want2)
	}
}

func TestProcessor_Process(t *testing.T) {
	fs := afero.NewMemMapFs()
	log := new(bytes.Buffer)
	p := jty.NewProcessor(runtime.GOMAXPROCS(-1), fs, log)

	JYOneTwo.WriteJ(t, fs, "in1.jsonnet")

	p.Process("in1.jsonnet", "out1.yml")
	p.Close()

	JYOneTwo.ExpectY(t, fs, "out1.yml")

	if got := log.String(); got != "" {
		t.Errorf("expected empty log, got %q", got)
	}
}
