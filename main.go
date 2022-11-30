package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

var errInvalidNumberRange = errors.New("invalid number range")

func main() {
	var nrange string

	flag.StringVar(&nrange, "num-range", "", "Numeric range to iterator over in format n..m")

	flag.Parse()
	if err := generate(nrange, flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "code-template: %s\n", err.Error())
		os.Exit(1)
	}

}

type Values struct {
	Num int
}

type numRange struct {
	from, to int
	step     int
}

func (nr *numRange) undefined() bool {
	return nr.from == 0 && nr.to == 0 && nr.step == 0
}

func must[T any](t T, e error) T {
	if e != nil {
		panic(e)
	}
	return t
}

var numRangeRegexp = regexp.MustCompile(`^([0-9]+)\.\.([0-9]+)$`)

func parseNumRange(nrange string) (result numRange, err error) {
	if nrange == "" {
		return
	}
	var matches []string
	if matches := numRangeRegexp.FindStringSubmatch(nrange); matches == nil {
		err = fmt.Errorf("%w: %s", errInvalidNumberRange, nrange)
		return
	}
	result.from = must(strconv.Atoi(matches[1]))
	result.to = must(strconv.Atoi(matches[2]))
	if result.to < result.from {
		result.step = -1
	} else {
		result.step = 1
	}
	return
}

func generate(nrange string, templateFiles []string) (err error) {
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
		textContent = string(content)
		if tpl, err = template.New("base").Funcs(sprig.FuncMap()).Parse(textContent); err != nil {
			err = fmt.Errorf("%s: %w", file, err)
			return
		}
		if numr.undefined() {
			values := Values{}
			if err = writeTemplate(tpl, &values, fmt.Sprintf("%s_%s.go", noext, ext)); err != nil {
				err = fmt.Errorf("%s: %w", file, err)
			}
		} else {
			for n := numr.from; n != numr.from; n += numr.step {
				values := Values{Num: n}
				if err = writeTemplate(tpl, &values, fmt.Sprintf("%s_%s_%d.go", ext, noext, n)); err != nil {
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
