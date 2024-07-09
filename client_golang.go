package main

import (
	_ "embed"
)

type golang struct{ client }

//go:embed template_golang.txt
var golangTemplate string

func (*golang) sourceFileTemplate() string {
	return golangTemplate
}

func (*golang) sourceFileName() string {
	return "main.go"
}

func (l *golang) configure(d any) error {
	if err := l.client.configure(d); err != nil {
		return err
	}
	if err := l.exec("go", "mod", "init", "main"); err != nil {
		return err
	}
	if err := l.exec("go", "get"); err != nil {
		return err
	}
	return nil
}

func (l *golang) run() error {
	return l.exec("go", "run", "main.go")
}
