package main

import (
	"strings"
	"testing"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/stretchr/testify/assert"
)

func testRunTemplate(tmpl string, data any) string {
	ptmpl, err := template.New("test").Funcs(sprig.FuncMap()).Funcs(tmplFuncs).Parse(tmpl)
	if err != nil {
		panic(err)
	}
	var result strings.Builder
	if err := ptmpl.Execute(&result, data); err != nil {
		panic(err)
	}
	return result.String()
}

func TestEnumerate(t *testing.T) {
	actual := testRunTemplate(`{{seq 3 6 | enumerate}}`, nil)
	assert.Equal(t, "[(0,3) (1,4) (2,5) (3,6)]", actual)
}

func TestAbs(t *testing.T) {
	i := testRunTemplate(`{{abs -6}}`, nil)
	assert.Equal(t, "6", i)
	f := testRunTemplate(`{{abs -6.6}}`, nil)
	assert.Equal(t, "6.6", f)
}

func TestTplFunc(t *testing.T) {
	s := testRunTemplate(`{{ tpl "{{.hello}} {{.world}}" .}}`,
		map[string]string{"hello": "Hello", "world": "World"})
	assert.Equal(t, "Hello World", s)
}

func TestTplMap(t *testing.T) {
	s := testRunTemplate(`{{ mapTpl "{{.}}" (seq 1 5) | join "," }}`, nil)
	assert.Equal(t, "1,2,3,4,5", s)
}
