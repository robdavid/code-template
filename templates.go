package main

import (
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	. "github.com/robdavid/genutil-go/errors/handler"
)

type exeTemplate struct {
	tpl    *template.Template
	values map[string]any
	output string
}

func newExeTemplate(tpl *template.Template, values map[string]any, output string) exeTemplate {
	return exeTemplate{tpl, values, output}
}

func parseFromFiles(files []string) (result exeTemplate, err error) {
	defer Catch(&err)
	for i, file := range files {
		content := string(Try(os.ReadFile(file)))
		if i == 0 {
			result.tpl = Try(result.newTpl(file).Parse(content))
		} else {
			Try(result.tpl.New(file).Parse(content))
		}
	}
	return
}

func parseFromText(name string, content string) (result exeTemplate, err error) {
	result.tpl, err = result.newTpl(name).Parse(content)
	return
}

func (et *exeTemplate) newTpl(name string) (tpl *template.Template) {
	includeFuncMap := map[string]any{"include": et.includeFunc}
	return template.New(name).
		Funcs(sprig.FuncMap()).
		Funcs(tmplFuncs).
		Funcs(includeFuncMap)
}

func (et *exeTemplate) includeFunc(templateName string, values any) (result string, err error) {
	var resultBuilder strings.Builder
	if err = et.tpl.ExecuteTemplate(&resultBuilder, templateName, values); err == nil {
		result = resultBuilder.String()
	}
	return
}

func parseFromFile(file string) (result exeTemplate, err error) {
	return parseFromFiles([]string{file})
}

func (et exeTemplate) outputIO(cache *outputCache) (output io.Writer, err error) {
	var outputBuf strings.Builder
	defer Catch(&err)
	Check(Try(template.New("outputName").Parse(et.output)).Execute(&outputBuf, et.values))
	return cache.get(outputBuf.String())
}

func (et exeTemplate) execute(cache *outputCache) (err error) {
	defer Catch(&err)
	return et.tpl.Execute(Try(et.outputIO(cache)), et.values)
}
