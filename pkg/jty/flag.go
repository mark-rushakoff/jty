package jty

import "github.com/spf13/pflag"

// Flags are the command-line flags jty supports.
type Flags struct {
	Args []string // The positional arguments.

	DryRun bool
	// TODO:
	// FromStdin bool
	// Zero      bool
}

// AddToFlagSet associates f with the given FlagSet.
func (f *Flags) AddToFlagSet(s *pflag.FlagSet) {
	s.BoolVarP(&f.DryRun, "dry-run", "n", false, "Print to stdout what processing would be done, without touching any files on disk.")
	// TODO:
	// s.BoolVarP(&f.FromStdin, "stdin", "i", false, "Read the input-output pairs of files from stdin.")
	// s.BoolVarP(&f.Zero, "", "z", false, "Expect NUL-separated input-output pairs from stdin. Implies -i.")
}
