package position

import (
	"fmt"
	"strings"
)

// MARK: Namespace

type Namespace struct {
	namespace []string
}

func NewNamespace(namespace []string) *Namespace { return &Namespace{namespace: namespace} }

func (namespace *Namespace) ToString() string {
	var result strings.Builder
	if len(namespace.namespace) > 0 {
		for _, name := range namespace.namespace {
			result.WriteString(name)
			result.WriteString(`\`)
		}
	}
	return result.String()
}

// MARK: File

type File struct {
	Namespace    *Namespace
	Filename     string
	IsStrictType bool
}

func NewFile(filename string) *File {
	// TODO position - Set IsStrictType default value to true/false depending on future ini setting for a strict mode
	return &File{Filename: filename, Namespace: nil, IsStrictType: false}
}

func (file *File) GetNamespaceStr() string {
	if file.Namespace == nil {
		return ""
	}
	return file.Namespace.ToString()
}

// MARK: Position

type Position struct {
	File   *File
	Line   int
	Column int
}

func NewPosition(file *File, line int, column int) *Position {
	return &Position{File: file, Line: line, Column: column}
}

func (pos *Position) ToPosString() string {
	if pos.File == nil {
		return ""
	}
	return fmt.Sprintf("%s:%d:%d", pos.File.Filename, pos.Line, pos.Column)
}

func (pos *Position) String() string {
	return fmt.Sprintf(`{Position - file: "%s", ln: %d, col: %d}`, pos.File.Filename, pos.Line, pos.Column)
}
