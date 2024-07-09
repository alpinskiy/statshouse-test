package main

import (
	_ "embed"
)

type python struct{ client }

//go:embed template_python.txt
var pythonTemplate string

func (*python) sourceFileTemplate() string {
	return pythonTemplate
}

func (*python) sourceFileName() string {
	return "main.py"
}

func (l *python) configure(d any) error {
	if err := l.exec("git", "clone", "--depth=1", "--no-tags", "git@github.com:VKCOM/statshouse-py.git", "."); err != nil {
		return err
	}
	if err := l.client.configure(d); err != nil {
		return err
	}
	return nil
}

func (l *python) run() error {
	return l.exec("python3", "main.py")
}
