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

func toInt(n any) (int, error) {
	switch v := n.(type) {
	case int:
		return v, nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case byte:
		return int(v), nil
	}
	return 0, fmt.Errorf("%w: cannot convert %T to int", errBadType, n)
}

// Like the sprig seq function, except returns a slice
// of ints rather than a string.
func seq(params ...any) (result []int, err error) {
	var nr numRange
	nr.step = 1
	np := len(params)
	switch {
	case np == 1:
		if nr.to, err = toInt(params[0]); err != nil {
			return
		}
	case np == 3:
		if nr.step, err = toInt(params[2]); err != nil {
			return
		}
		fallthrough
	case np == 2:
		if nr.from, err = toInt(params[0]); err != nil {
			return
		}
		if nr.to, err = toInt(params[1]); err != nil {
			return
		}
	}
	result = make([]int, 0, abs(nr.to-nr.from)/nr.step)
	for i := nr.from; nr.inRange(i); i += nr.step {
		result = append(result, i)
	}
	return result, err
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

func mapTpl(tmpl string, items any) (result []string, err error) {
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
	"mapTpl":    mapTpl,
	"tpl":       tplFunc,
}
