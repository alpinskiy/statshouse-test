package main

import (
	_ "embed"
)

type java struct{ client }

//go:embed template_java.txt
var javaTemplate string

func (*java) sourceFileTemplate() string {
	return javaTemplate
}

func (*java) sourceFileName() string {
	return "src/main/java/test.java"
}

func (l *java) configure(d any) error {
	if err := l.exec("git", "clone", "-b", "malpinskiy/dev", "--depth=1", "--no-tags", "git@github.com:VKCOM/statshouse-java.git", "."); err != nil {
		return err
	}
	if err := l.client.configure(d); err != nil {
		return err
	}
	l.cd("src/main/java")
	if err := l.exec("javac", "test.java"); err != nil {
		return err
	}
	return nil
}

func (l *java) run() error {
	return l.exec("java", "test")
}
