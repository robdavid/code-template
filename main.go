package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	flag "github.com/spf13/pflag"
)

var errInvalidNumberRange = errors.New("invalid number range")
var errKeyConflict = errors.New("conflicting property key")
var errBadNrange = errors.New("expected number range in the format var.name=<range>")

type options struct {
	optValues
	files []string
}

func main() {
	var opts options
	var help bool

	flag.BoolVarP(&help, "help", "h", false, "Display help")
	flag.StringVar(&opts.nrangeSpec, "num-range", "", "Numeric range to iterator over in format var.name=n..m")
	flag.StringToStringVar(&opts.values, "set", nil, "Set a value to place within template values")
	flag.Parse()

	if help {
		fmt.Fprintln(os.Stderr, "Usage: code-template [options] files...")
		flag.PrintDefaults()
		os.Exit(0)
	}
	opts.files = flag.Args()

	var err error
	if err = opts.ParseNRange(); err == nil {
		err = generateForOpts(&opts)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "code-template: %s\n", err.Error())
		os.Exit(1)
	}
}

func generateForOpts(opts *options) (err error) {
	for _, s := range opts.files {
		var errFile string
		if glob, gerr := filepath.Glob(s); gerr != nil && len(glob) > 0 {
			for _, gs := range glob {
				if err = generateFiles(&opts.optValues, gs); err != nil {
					errFile = gs
					break
				}
			}
		} else {
			errFile = s
			err = generateFiles(&opts.optValues, s)
		}
		if err != nil {
			return fmt.Errorf("%s: %w", errFile, err)
		}
	}
	return nil
}
