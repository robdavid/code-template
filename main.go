package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	flag "github.com/spf13/pflag"
)

var errInvalidNumberRange = errors.New("invalid number range")
var errKeyConflict = errors.New("conflicting property key")

type options struct {
	nrange string
	values map[string]string
	files  []string
}

func main() {
	var opts options

	flag.StringVar(&opts.nrange, "num-range", "", "Numeric range to iterator over in format n..m")
	flag.StringToStringVar(&opts.values, "set", nil, "Set a value to place within template values")

	flag.Parse()
	opts.files = flag.Args()

	if err := run(&opts); err != nil {
		fmt.Fprintf(os.Stderr, "code-template: %s\n", err.Error())
		os.Exit(1)
	}
}

func run(opts *options) (err error) {
	var values map[string]any
	if values, err = mapValues(opts.values); err != nil {
		return err
	}
	err = generate(opts.nrange, values, opts.files)
	return
}

func generate(nrange string, values map[string]any, templateFiles []string) (err error) {
	var numr numRange
	if numr, err = parseNumRange(nrange); err != nil {
		return
	}
	for _, file := range templateFiles {
		var tpl *template.Template
		var ext string
		noext := file
		if ext = filepath.Ext(noext); ext != "" {
			noext = noext[:len(noext)-len(ext)]
		}
		var content []byte
		var textContent string
		if content, err = os.ReadFile(file); err != nil {
			err = fmt.Errorf("%s: %w", file, err)
			return
		}
		include := func(templateName string, values any) (string, error) {
			var result strings.Builder
			if err := tpl.ExecuteTemplate(&result, templateName, values); err != nil {
				return "", nil
			} else {
				return result.String(), nil
			}
		}
		includeFuncMap := map[string]any{"include": include}
		textContent = string(content)
		if tpl, err = template.New(filepath.Base(file)).
			Funcs(sprig.FuncMap()).
			Funcs(tmplFuncs).
			Funcs(includeFuncMap).
			Parse(textContent); err != nil {
			err = fmt.Errorf("%s: %w", file, err)
			return
		}
		if numr.undefined() {
			if err = writeTemplate(tpl, &values, fmt.Sprintf("%s_%s.go", noext, ext[1:])); err != nil {
				err = fmt.Errorf("%s: %w", file, err)
			}
		} else {
			for n := numr.from; numr.inRange(n); n += numr.step {
				values["Num"] = n
				if err = writeTemplate(tpl, &values, fmt.Sprintf("%s_%s_%d.go", noext, ext[1:], n)); err != nil {
					err = fmt.Errorf("%s: %w", file, err)
				}
			}
		}
	}
	return
}

func writeTemplate(tpl *template.Template, values any, outfile string) error {
	if out, err := os.Create(outfile); err != nil {
		return err
	} else {
		defer out.Close()
		return tpl.Execute(out, values)
	}
}
