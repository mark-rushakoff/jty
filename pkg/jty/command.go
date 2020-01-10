package jty

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"runtime"

	"github.com/spf13/afero"
)

var (
	ErrOddInputFiles = errors.New("odd number of file arguments received; must be given in pairs")
	ErrNoInputFiles  = errors.New("at least one input-output pair must be given")
)

// Command represents a running CLI environment.
type Command struct {
	Stdin          io.Reader
	Stdout, Stderr io.Writer

	FS afero.Fs
}

// Run inputs all the CLI-specified files to a new Processor.
func (c *Command) Run(f *Flags) error {
	if f.FromStdin {
		if len(f.Args) > 0 {
			panic("error here")
		}
	} else {
		if len(f.Args) == 0 {
			return ErrNoInputFiles
		}
		if len(f.Args)%2 != 0 {
			return ErrOddInputFiles
		}
	}

	p := NewProcessor(runtime.GOMAXPROCS(-1), c.FS, c.Stderr)
	if f.DryRun {
		p.DryRunDest = c.Stdout
	}

	if f.FromStdin {
		if err := c.processFromStdin(f, p); err != nil {
			return err
		}
	} else {
		// Iterate through command line arguments.
		for i := 0; i < len(f.Args); i += 2 {
			p.Process(f.Args[i], f.Args[i+1])
		}
	}

	p.Close()

	// Don't need to take lock, as we have finished all goroutines which may access the field.
	if p.didLogError {
		return fmt.Errorf("encountered errors during processing; failing")
	}

	return nil
}

func (c *Command) processFromStdin(f *Flags, p *Processor) error {
	s := bufio.NewScanner(c.Stdin)

	if f.Zero {
		s.Split(splitNul)
	} else {
		s.Split(splitLF)
	}

	processed := false
	for s.Scan() {
		inPath := s.Text()

		if !s.Scan() {
			return ErrOddInputFiles
		}
		outPath := s.Text()

		p.Process(inPath, outPath)
		processed = true
	}

	if !processed {
		return ErrNoInputFiles
	}

	return nil
}

func splitNul(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\x00'); i >= 0 {
		// Full nul-terminated entry.
		return i + 1, data[0:i], nil
	}

	if atEOF {
		// Entry wasn't terminated. Don't return it.
		return len(data), nil, nil
	}
	// Request more data.
	return 0, nil, nil
}

func splitLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// Full LF-terminated entry.
		if i > 0 && data[i] == '\r' {
			return i + 1, data[0 : i-1], nil
		}
		return i + 1, data[0:i], nil
	}

	if atEOF {
		// Entry wasn't terminated. Don't return it.
		return len(data), nil, nil
	}
	// Request more data.
	return 0, nil, nil
}
