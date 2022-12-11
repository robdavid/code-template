package main

import (
	"errors"
	"fmt"
	"os"

	flag "github.com/spf13/pflag"
)

var errInvalidNumberRange = errors.New("invalid number range")
var errKeyConflict = errors.New("conflicting property key")
var errBadNrange = errors.New("expected number range in the format var.name=<range>")

type options struct {
	optValues
	files    []string
	includes []string
}

func main() {
	var opts options
	var help bool

	flag.BoolVarP(&help, "help", "h", false, "Display help")
	flag.StringVar(&opts.nrangeSpec, "repeat", "", "Numeric range to iterator over in format var.name=n..m")
	flag.StringToStringVar(&opts.values, "set", nil, "Set a value to place within template values")
	flag.StringVarP(&opts.output, "output", "o", "", "Send output to specified file, - for standard out")
	flag.StringArrayVarP(&opts.includes, "include", "i", nil, "Include specified files in each template execution")
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

func generateForOpts(opts *options) error {
	var cache outputCache
	defer cache.close()
	for _, s := range expandGlob(opts.files) {
		if err := runTemplate(&opts.optValues, &cache, s, opts.includes); err != nil {
			return fmt.Errorf("%s: %w", s, err)
		}
	}
	return nil
}
