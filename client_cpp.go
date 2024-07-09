package main

import (
	_ "embed"
	"strings"
)

type cpp struct{ client }
type cppTransport struct{ cpp }
type cppRegistry struct{ cpp }

//go:embed template_cpp_transport.txt
var cppTransportTemplate string

func (*cppTransport) sourceFileTemplate() string {
	return cppTransportTemplate
}

//go:embed template_cpp_registry.txt
var cppRegistryTemplate string

func (*cppRegistry) sourceFileTemplate() string {
	return cppRegistryTemplate
}

func (*cpp) sourceFileName() string {
	return "main.cpp"
}

func (l *cpp) make() error {
	l.binFile = strings.TrimSuffix(l.srcFile, ".cpp")
	return l.exec("g++", "-o", l.binFile, l.srcFile)
}
