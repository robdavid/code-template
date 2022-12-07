package main

import (
	"strings"
	"testing"

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

func TestGenerateRange(t *testing.T) {
	template := `{{ .my.range.value }}{{ .my.range.delim }}`
	expected := "1-2-3-4-5-"
	var output strings.Builder
	opts := optValues{nrangeVar: "my.range.value", nrange: "1..5", values: map[string]string{"my.range.delim": "-"}}
	err := generate(&opts, "test", template, &output)
	assert.NoError(t, err)
	assert.Equal(t, expected, output.String())
}
