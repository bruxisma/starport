package protocode

import (
	"fmt"

	"github.com/emicklei/proto"
)

// TODO: Complete wrapping this so it can just *be* the type that is actually
// used, instead of proto.Proto
type OrganizedFile struct {
	Filename string
	Syntax   *proto.Syntax
	Package  *proto.Package
	Imports  []*proto.Import
	Options  []*proto.Option
	Enums    []*proto.Enum
	Messages []*proto.Message
	Services []*proto.Service
}

func NewOrganizedFile(file *File) *OrganizedFile {
	of := &OrganizedFile{Filename: file.Filename}
	proto.Walk(file, WithFile(of))
	return of
}

func WithFile(file *OrganizedFile) proto.Handler {
	return func(target proto.Visitee) {
		switch value := target.(type) {
		case *proto.Syntax:
			file.setSyntax(value)
		case *proto.Package:
			file.setPackage(value)
		case *proto.Import:
			file.appendImport(value)
		case *proto.Option:
			file.appendOption(value)
		case *proto.Enum:
			file.appendEnum(value)
		case *proto.Message:
			file.appendMessage(value)
		case *proto.Service:
			file.appendService(value)
		}
	}
}

func (of *OrganizedFile) IndexOfImport(filename string) int {
	for idx, item := range of.Imports {
		if item.Filename == filename {
			return idx
		}
	}
	return -1
}

func (of *OrganizedFile) IndexOfMessage(name string) int {
	for idx, item := range of.Messages {
		if item.Name == name {
			return idx
		}
	}
	return -1
}

func (of *OrganizedFile) RemoveImportAt(idx int) {
	of.Imports = append(of.Imports[:idx], of.Imports[idx+1:]...)
}

func (of *OrganizedFile) FindImportf(format string, args ...interface{}) (*proto.Import, error) {
	return of.FindImport(fmt.Sprintf(format, args...))
}

func (of *OrganizedFile) FindMessagef(format string, args ...interface{}) (*proto.Message, error) {
	return of.FindMessage(fmt.Sprintf(format, args...))
}

func (of *OrganizedFile) FindImport(filename string) (*proto.Import, error) {
	for _, node := range of.Imports {
		if node.Filename == filename {
			return node, nil
		}
	}
	return nil, fmt.Errorf("%w %q", ErrImportNotFound, filename)
}

func (of *OrganizedFile) FindMessage(name string) (*proto.Message, error) {
	for _, message := range of.Messages {
		if message.Name == name {
			return message, nil
		}
	}
	return nil, fmt.Errorf("%w %q", ErrMessageNotFound, name)
}

func (of *OrganizedFile) AppendImportf(format string, args ...interface{}) {
	of.AppendImport(fmt.Sprintf(format, args...))
}

func (of *OrganizedFile) AppendImport(filename string) {
	of.Imports = append(of.Imports, &proto.Import{Filename: filename})
}

func (of *OrganizedFile) PrependImportf(format string, args ...interface{}) {
	of.PrependImport(fmt.Sprintf(format, args...))
}

func (of *OrganizedFile) PrependImport(filename string) {
	nodes := []*proto.Import{
		&proto.Import{Filename: filename},
	}
	of.Imports = append(nodes, of.Imports...)
}

func (of *OrganizedFile) AsFile() *File {
	length := len(of.Imports) +
		len(of.Options) +
		len(of.Enums) +
		len(of.Messages) +
		len(of.Services) +
		2
	elements := make([]proto.Visitee, 0, length)
	elements = append(elements, of.Syntax, of.Package)

	for _, item := range of.Imports {
		elements = append(elements, item)
	}

	for _, item := range of.Options {
		elements = append(elements, item)
	}

	for _, item := range of.Enums {
		elements = append(elements, item)
	}

	for _, item := range of.Messages {
		elements = append(elements, item)
	}

	for _, item := range of.Services {
		elements = append(elements, item)
	}

	return &File{
		Filename: of.Filename,
		Elements: elements,
	}
}

func (file *OrganizedFile) setSyntax(syntax *proto.Syntax) {
	file.Syntax = syntax
}

func (file *OrganizedFile) setPackage(pkg *proto.Package) {
	file.Package = pkg
}

func (file *OrganizedFile) appendImport(path *proto.Import) {
	file.Imports = append(file.Imports, path)
}

func (file *OrganizedFile) appendOption(option *proto.Option) {
	file.Options = append(file.Options, option)
}

func (file *OrganizedFile) appendEnum(enum *proto.Enum) {
	file.Enums = append(file.Enums, enum)
}

func (file *OrganizedFile) appendMessage(message *proto.Message) {
	file.Messages = append(file.Messages, message)
}

func (file *OrganizedFile) appendService(service *proto.Service) {
	file.Services = append(file.Services, service)
}
