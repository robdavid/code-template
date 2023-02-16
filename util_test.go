package main

import (
	"testing"

	"github.com/robdavid/genutil-go/errors/test"
	"github.com/stretchr/testify/assert"
)

func TestInsertPathOne(t *testing.T) {
	m := make(map[string]any)
	insertPath("a.b", 123, m)
	assert.Equal(t, map[string]any{
		"a": map[string]any{"b": 123},
	}, m)
}

func TestInsertPathTwo(t *testing.T) {
	m := make(map[string]any)
	assert.NoError(t, insertPath("a.b", 123, m))
	assert.NoError(t, insertPath("a.c", 456, m))
	assert.Equal(t, map[string]any{
		"a": map[string]any{"b": 123, "c": 456},
	}, m)
}

func TestInsertPathFourDeep(t *testing.T) {
	m := make(map[string]any)
	assert.NoError(t, insertPath("a.b.c.d", 123, m))
	assert.Equal(t, map[string]any{
		"a": map[string]any{
			"b": map[string]any{
				"c": map[string]any{
					"d": 123,
				},
			},
		},
	}, m)
}

func TestInsertPathConflictLeaf(t *testing.T) {
	m := make(map[string]any)
	assert.NoError(t, insertPath("a.b.c.d", 123, m))
	err := insertPath("a.b.c", 456, m)
	assert.EqualError(t, err, "conflicting property key at a.b.c")
}

func TestInsertPathConflictInterior(t *testing.T) {
	m := make(map[string]any)
	assert.NoError(t, insertPath("a.b.c", 123, m))
	err := insertPath("a.b.c.d", 456, m)
	assert.EqualError(t, err, "conflicting property key at a.b.c")
}

func TestMapValues(t *testing.T) {
	input := map[string]string{
		"one.two.three": "123",
		"one.two.four":  "12.4",
		"one.two.lots":  "12345678900987",
		"val.bool":      "true",
		"val.str":       "hello",
	}
	expected := map[string]any{
		"one": map[string]any{
			"two": map[string]any{
				"three": int32(123),
				"four":  12.4,
				"lots":  int64(12345678900987),
			},
		},
		"val": map[string]any{
			"bool": true,
			"str":  "hello",
		},
	}
	actual, err := mapValues(input)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestParseNRangeEmpty(t *testing.T) {
	var opts optValues
	opts.nrangeSpec = ""
	test.Check(t, opts.ParseNRange())
	assert.Equal(t, "", opts.nrange)
	assert.Equal(t, "", opts.nrangeVar)
}

func TestParseNRangeErr(t *testing.T) {
	var opts optValues
	opts.nrangeSpec = "var"
	assert.ErrorIs(t, opts.ParseNRange(), errBadNrange)
}

func TestParseNRange(t *testing.T) {
	var opts optValues
	opts.nrangeSpec = "var=1..10"
	test.Check(t, opts.ParseNRange())
	assert.Equal(t, "var", opts.nrangeVar)
	assert.Equal(t, "1..10", opts.nrange)
}

func TestParseNumRangeEmpty(t *testing.T) {
	numRange := test.Result(parseNumRange("")).Must(t)
	assert.Equal(t, 0, numRange.from)
	assert.Equal(t, 0, numRange.to)
	assert.Equal(t, 0, numRange.step)
	assert.True(t, numRange.undefined())
}

func TestParseNumRange(t *testing.T) {
	numRange := test.Result(parseNumRange("1..10")).Must(t)
	assert.Equal(t, 1, numRange.from)
	assert.Equal(t, 10, numRange.to)
	assert.Equal(t, 1, numRange.step)
	assert.False(t, numRange.undefined())
}

func TestParseNumRangeReverse(t *testing.T) {
	numRange := test.Result(parseNumRange("10..1")).Must(t)
	assert.Equal(t, 10, numRange.from)
	assert.Equal(t, 1, numRange.to)
	assert.Equal(t, -1, numRange.step)
	assert.False(t, numRange.undefined())
}

func TestParseNumRangeNegative(t *testing.T) {
	numRange := test.Result(parseNumRange("-10..-1")).Must(t)
	assert.Equal(t, -10, numRange.from)
	assert.Equal(t, -1, numRange.to)
	assert.Equal(t, 1, numRange.step)
	assert.False(t, numRange.undefined())
}

func TestParseNumRangeInvalid(t *testing.T) {
	_, err := parseNumRange("-10.0..-1.0")
	assert.ErrorContains(t, err, "invalid number range: -10.0..-1.0")
}
