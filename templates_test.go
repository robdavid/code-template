package main

import (
	"strings"
	"testing"

	"github.com/robdavid/genutil-go/errors/result"
	"github.com/stretchr/testify/assert"
)

func must[T any](t *testing.T, r result.Result[T]) T {
	if !assert.NoError(t, r.GetErr()) {
		t.FailNow()
	}
	return r.Get()
}

func check(t *testing.T, err error) {
	if !assert.NoError(t, err) {
		t.FailNow()
	}
}

func TestGenerate(t *testing.T) {
	template := `{{ mapTpl "{{.}}" (seq 1 .max.value) | join .delim}}`
	expected := "1-2-3-4-5"
	var output strings.Builder
	valuesStr := map[string]string{"max.value": "5", "delim": "-"}
	var cache outputCache
	cache.preOpen("output", &output)
	exe := must(t, result.From(parseFromText("test", template)))
	exe.values = must(t, result.From(mapValues(valuesStr)))
	exe.output = "output"
	check(t, exe.execute(&cache))
	assert.Equal(t, expected, output.String())
}
