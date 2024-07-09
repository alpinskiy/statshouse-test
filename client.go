package main

import (
	_ "embed"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
)

type lib interface {
	sourceFileTemplate() string
	sourceFileName() string
	init(lib, options) error
	configure(any) error
	make() error
	run() error
	cleanup()
}

type client struct {
	lib
	options
	rootDir string
	execDir string
	srcFile string
	binFile string
}

func (l *client) init(e lib, o options) (err error) {
	l.lib = e
	l.options = o
	l.rootDir, err = os.MkdirTemp("", typeName(e))
	l.execDir = l.rootDir
	return err
}

func (l *client) configure(d any) error {
	t, err := template.New(typeName(l.lib)).Parse(l.sourceFileTemplate())
	if err != nil {
		return err
	}
	l.srcFile = filepath.Join(l.rootDir, l.sourceFileName())
	if err = os.MkdirAll(filepath.Dir(l.srcFile), os.ModePerm); err != nil {
		return err
	}
	srcFile, err := os.Create(l.srcFile)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	if err := t.Execute(srcFile, d); err != nil {
		return err
	}
	if l.viewCode {
		l.exec("code", l.srcFile)
	}
	return nil
}

func (l *client) make() error {
	return nil // nop, works for interpreted languages
}

func (l *client) run() error {
	return l.exec(l.binFile) // works for compiled languages
}

func (l *client) cleanup() {
	if l.keepTemp || l.rootDir == "" {
		return
	}
	os.RemoveAll(l.rootDir)
}

func (l *client) cd(path string) {
	l.execDir = filepath.Join(l.rootDir, path)
	log.Printf("$ cd %s", l.execDir)
}

func (l *client) exec(args ...string) error {
	log.Printf("$ %s\n", strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = l.execDir
	return cmd.Run()
}

func runClient(l lib, d any, o options) error {
	if err := l.init(l, o); err != nil {
		return err
	}
	defer l.cleanup()
	if err := l.configure(d); err != nil {
		return err
	}
	if err := l.make(); err != nil {
		return err
	}
	return l.run()
}

func typeName(l lib) string {
	return reflect.ValueOf(l).Type().Elem().Name()
}
