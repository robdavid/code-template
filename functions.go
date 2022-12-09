package main

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"text/template"
)

var errBadType = errors.New("bad type")

type ordered interface {
	int8 | int16 | int32 | int64 | int |
		uint8 | uint16 | uint32 | uint64 | uint |
		float32 | float64
}

func abs[T ordered](t T) T {
	if t >= 0 {
		return t
	} else {
		return -t
	}
}

// Like the sprig seq function, except returns a slice
// of ints rather than a string.
func seq(params ...int) []int {
	var nr numRange
	nr.step = 1
	np := len(params)
	switch {
	case np == 1:
		nr.to = params[0]
	case np == 3:
		nr.step = params[2]
		fallthrough
	case np == 2:
		nr.from = params[0]
		nr.to = params[1]
	}
	result := make([]int, 0, abs(nr.to-nr.from)/nr.step)
	for i := nr.from; nr.inRange(i); i += nr.step {
		result = append(result, i)
	}
	return result
}

type Enumerated struct {
	Value any
	Index int
}

func (e Enumerated) String() string {
	return fmt.Sprintf("(%d,%#v)", e.Index, e.Value)
}

func enumerate(in any) (out []Enumerated) {
	value := reflect.ValueOf(in)
	if value.Kind() == reflect.Slice {
		len := value.Len()
		out = make([]Enumerated, len)
		for i := 0; i < len; i++ {
			out[i].Index = i
			out[i].Value = value.Index(i).Interface()
		}
	}
	return
}

func absTmpl(u any) (v any, err error) {
	switch o := u.(type) {
	case int:
		v = abs(o)
	case int8:
		v = abs(o)
	case int16:
		v = abs(o)
	case int32:
		v = abs(o)
	case int64:
		v = abs(o)
	case uint8:
		v = abs(o)
	case uint16:
		v = abs(o)
	case uint32:
		v = abs(o)
	case uint64:
		v = abs(o)
	case float32:
		v = abs(o)
	case float64:
		v = abs(o)
	default:
		err = fmt.Errorf("%w: %T", errBadType, u)
	}
	return
}

func tplMap(tmpl string, items any) (result []string, err error) {
	itemVal := reflect.ValueOf(items)
	if itemVal.Kind() == reflect.Slice || itemVal.Kind() == reflect.Array {
		len := itemVal.Len()
		result = make([]string, len)
		for i := 0; i < len; i++ {
			if result[i], err = tplFunc(tmpl, itemVal.Index(i).Interface()); err != nil {
				return
			}
		}
	} else {
		err = fmt.Errorf("%w: got %T, wanted slice or array", errBadType, items)
	}
	return
}

func tplFunc(templateText string, data any) (string, error) {
	t, err := template.New("tplFunc").Parse(templateText)
	if err != nil {
		return "", err
	}
	var result strings.Builder
	if err := t.ExecuteTemplate(&result, "tplFunc", data); err != nil {
		return "", err
	} else {
		return result.String(), nil
	}
}

var tmplFuncs = map[string]any{
	"seq":       seq,
	"enumerate": enumerate,
	"abs":       absTmpl,
	"tplMap":    tplMap,
	"tpl":       tplFunc,
}
