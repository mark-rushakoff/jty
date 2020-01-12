package jty

import (
	"path/filepath"

	"github.com/spf13/pflag"
)

// Flags are the command-line flags jty supports.
type Flags struct {
	Args []string // The positional arguments.

	DryRun    bool
	FromStdin bool
	Zero      bool

	HelpRequested bool

	// Parsed --jpath values (in given order)
	// prefixed with JSONNET_PATH environment variable values in reverse order.
	// Same behavior as official jsonnet tool.
	JPaths []string
}

// AddToFlagSet associates f with the given FlagSet.
func (f *Flags) AddToFlagSet(s *pflag.FlagSet) {
	s.BoolVarP(&f.DryRun, "dry-run", "n", false, "Print to stdout what processing would be done, without touching any files on disk.")
	s.BoolVarP(&f.FromStdin, "stdin", "i", false, "Read the input-output pairs of files from stdin.")
	s.BoolVarP(&f.Zero, "zero", "z", false, "Expect NUL-separated input-output pairs from stdin. Implies -i.")
	s.BoolVarP(&f.HelpRequested, "help", "h", false, "Show help.")

	s.StringArrayVarP(&f.JPaths, "jpath", "J", nil, "Additional library search paths (rightmost wins).")
}

// FinishParse sets any default values that are implied by another option,
// and parses any supplied environment values.
//
// jsonnetPathEnv is the value of environment variable JSONNET_PATH.
func (f *Flags) FinishParse(jsonnetPathEnv string) {
	if f.Zero {
		f.FromStdin = true
	}

	e := filepath.SplitList(jsonnetPathEnv)

	// Reverse the list. https://github.com/golang/go/wiki/SliceTricks#reversing
	for i := len(e)/2 - 1; i >= 0; i-- {
		opp := len(e) - 1 - i
		e[i], e[opp] = e[opp], e[i]
	}

	f.JPaths = append(e, f.JPaths...)
}
