package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	. "github.com/robdavid/genutil-go/errors/handler"
)

type optValues struct {
	nrangeSpec string
	nrange     string
	nrangeVar  string
	values     map[string]string
	output     string
}

func (opts *optValues) ParseNRange() error {
	if opts.nrangeSpec == "" {
		opts.nrange = ""
		opts.nrangeVar = ""
		return nil
	} else if pos := strings.Index(opts.nrangeSpec, "="); pos < 0 {
		return errBadNrange
	} else {
		opts.nrange = strings.TrimSpace(opts.nrangeSpec[pos+1:])
		opts.nrangeVar = strings.TrimSpace(opts.nrangeSpec[:pos])
		return nil
	}
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
	result.from = Must(strconv.Atoi(matches[1]))
	result.to = Must(strconv.Atoi(matches[2]))
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

func defaultOutput(input string, suffix string) string {
	var ext string
	noext := input
	if ext = filepath.Ext(noext); ext != "" {
		noext = noext[:len(noext)-len(ext)]
	}
	if suffix == "" {
		return fmt.Sprintf("%s_%s.go", noext, ext[1:])
	} else {
		return fmt.Sprintf("%s_%s_%s.go", noext, ext[1:], suffix)
	}
}

func expandGlob(files []string) (result []string) {
	result = make([]string, 0, len(files))
	for _, f := range files {
		if glob, err := filepath.Glob(f); err == nil && len(glob) > 0 {
			result = append(result, glob...)
		} else {
			result = append(result, f)
		}
	}
	return
}

func runTemplate(opts *optValues, cache *outputCache, file string, includes []string) (err error) {
	defer Catch(&err)
	output := opts.output
	if output == "" {
		output = defaultOutput(file, "")
	}
	templateFiles := []string{file}
	templateFiles = append(templateFiles, expandGlob(includes)...)
	numr := Try(parseNumRange(opts.nrange))
	te := Try(parseFromFiles(templateFiles))
	te.output = output
	te.values = Try(mapValues(opts.values))
	if numr.undefined() {
		Check(te.execute(cache))
	} else {
		for n := numr.from; numr.inRange(n); n += numr.step {
			insertPath(opts.nrangeVar, n, te.values)
			Check(te.execute(cache))
		}
	}
	return
}
