package utilz

import (
	htmltemplate "html/template"
	"path/filepath"
	texttemplate "text/template"
)

func NewHTMLTemplateFromFile(path string) (*htmltemplate.Template, error) {
	return htmltemplate.ParseFiles(path)
}
func NewHTMLTemplateFromFileWithFuncs(path string, funcs htmltemplate.FuncMap) (*htmltemplate.Template, error) {
	return htmltemplate.
		New(filepath.Base(path)).
		Funcs(funcs).
		ParseFiles(path)
}

func NewTextTemplateFromString(name string, tpl string) (*texttemplate.Template, error) {
	return texttemplate.
		New(name).
		Parse(tpl)
}
func NewTextTemplateFromFile(path string) (*texttemplate.Template, error) {
	return texttemplate.ParseFiles(path)
}
func NewTextTemplateFromFileWithFuncs(path string, funcs texttemplate.FuncMap) (*texttemplate.Template, error) {
	return texttemplate.
		New(filepath.Base(path)).
		Funcs(funcs).
		ParseFiles(path)
}
