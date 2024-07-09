package main

import (
	_ "embed"
	"strings"
)

type rust struct{ client }

//go:embed template_rust.txt
var rustTemplate string

func (*rust) sourceFileTemplate() string {
	return rustTemplate
}

func (*rust) sourceFileName() string {
	return "main.rs"
}

func (l *rust) make() error {
	l.binFile = strings.TrimSuffix(l.srcFile, ".rs")
	return l.exec("rustc", "-o", l.binFile, l.srcFile)
}
