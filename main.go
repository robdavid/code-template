package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"gopkg.in/yaml.v2"

	"github.com/Masterminds/sprig/v3"
	flag "github.com/spf13/pflag"
)

var errInvalidNumberRange = errors.New("invalid number range")

func main() {
	var nrange string
	var values map[string]string

	flag.StringVar(&nrange, "num-range", "", "Numeric range to iterator over in format n..m")
	flag.StringToStringVar(&values, "set", nil, "set a value to place define for template .")

	flag.Parse()
	if err := generate(nrange, mapValues(values), flag.Args()); err != nil {
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

func (nr *numRange) inRange(n int) bool {
	if nr.step < 0 {
		return n <= nr.from && n >= nr.to
	} else {
		return n >= nr.from && n <= nr.to
	}
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
	if matches = numRangeRegexp.FindStringSubmatch(nrange); matches == nil {
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

func mapValues(strValues map[string]string) (output map[string]any) {
	output = make(map[string]any)
	for name, value := range strValues {
		var v any
		if json.Unmarshal([]byte(value), &v) != nil {
			v = value
		}
		if yaml.Unmarshal([]byte(value), &v) != nil {
			v = value
		}
		output[name] = v
	}
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
