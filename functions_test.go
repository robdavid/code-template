package main

import (
	"strings"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func runTemplate(tmpl string, data any) string {
	ptmpl, err := template.New("test").Funcs(tmplFuncs).Parse(tmpl)
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
	actual := runTemplate(`{{seq 3 6 | enumerate}}`, nil)
	assert.Equal(t, "[(0,3) (1,4) (2,5) (3,6)]", actual)
}

func TestAbs(t *testing.T) {
	i := runTemplate(`{{abs -6}}`, nil)
	assert.Equal(t, "6", i)
	f := runTemplate(`{{abs -6.6}}`, nil)
	assert.Equal(t, "6.6", f)
}
