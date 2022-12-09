package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
)

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

func newNumRange(from, to, step int) numRange {
	return numRange{from, to, step}
}

func mapValues(strValues map[string]string) (map[string]any, error) {
	output := make(map[string]any)
	for name, value := range strValues {
		var parsed any
		if i, err := strconv.ParseInt(value, 10, 32); err == nil {
			parsed = int32(i)
		} else if i, err := strconv.ParseInt(value, 10, 64); err == nil {
			parsed = int64(i)
		} else if f, err := strconv.ParseFloat(value, 64); err == nil {
			parsed = f
		} else if b, err := strconv.ParseBool(value); err == nil {
			parsed = b
		} else {
			parsed = value
		}
		if err := insertPath(name, parsed, output); err != nil {
			return nil, err
		}
	}
	return output, nil
}

func insertPath(path string, value any, top map[string]any) error {
	pathList := strings.Split(path, ".")
	m := top
	for i, s := range pathList {
		if i == len(pathList)-1 {
			if n, ok := m[s]; ok {
				if _, ok := n.(map[string]any); ok {
					return fmt.Errorf("%w at %s", errKeyConflict, path)
				}
			}
			m[s] = value
		} else {
			if n, ok := m[s]; ok {
				if nm, okm := n.(map[string]any); okm {
					m = nm
				} else {
					return fmt.Errorf("%w at %s", errKeyConflict, strings.Join(pathList[:i+1], "."))
				}
			} else {
				n := make(map[string]any)
				m[s] = n
				m = n
			}
		}
	}
	return nil
}

func generate(opts *optValues, templateName string, templateContent string, output io.Writer) (err error) {
	var numr numRange
	var values map[string]any
	if numr, err = parseNumRange(opts.nrange); err != nil {
		return
	}
	if values, err = mapValues(opts.values); err != nil {
		return err
	}

	var tpl *template.Template
	include := func(templateName string, values any) (string, error) {
		var result strings.Builder
		if err := tpl.ExecuteTemplate(&result, templateName, values); err != nil {
			return "", nil
		} else {
			return result.String(), nil
		}
	}

	includeFuncMap := map[string]any{"include": include}
	if tpl, err = template.New(filepath.Base(templateName)).
		Funcs(sprig.FuncMap()).
		Funcs(tmplFuncs).
		Funcs(includeFuncMap).
		Parse(templateContent); err != nil {
		return
	}
	if numr.undefined() {
		err = tpl.Execute(output, values)
	} else {
		for n := numr.from; numr.inRange(n); n += numr.step {
			insertPath(opts.nrangeVar, n, values)
			if err = tpl.Execute(output, values); err != nil {
				return
			}
		}
	}
	return
}

func generateFiles(fname string, opts *optValues) error {
	var ext string
	var source []byte
	var err error
	noext := fname
	if ext = filepath.Ext(noext); ext != "" {
		noext = noext[:len(noext)-len(ext)]
	}
	if source, err = os.ReadFile(fname); err != nil {
		return err
	}
	target := fmt.Sprintf("%s_%s.go", noext, ext[1:])
	if output, err := os.Create(target); err != nil {
		return err
	} else {
		defer output.Close()
		return generate(opts, fname, string(source), output)
	}
}
