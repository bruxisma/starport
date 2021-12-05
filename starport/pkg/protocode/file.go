package protocode

import (
	"fmt"

	"github.com/emicklei/proto"
)

// File represents an "unwrapped" protobuf file that can be reodered safely as
// top-level declarations don't matter once users are past declaring imports
// and options
type File struct {
	Filename string
	Syntax   *proto.Syntax
	Package  *proto.Package
	Imports  []*proto.Import
	Options  []*proto.Option
	Enums    []*proto.Enum
	Messages []*Message
	Services []*Service
}

// NewFile returns a File from a proto.Proto
func NewFile(p *proto.Proto) *File {
	file := &File{Filename: p.Filename}
	proto.Walk(p, withFile(file, p))
	return file
}

func withFile(file *File, root *proto.Proto) proto.Handler {
	return func(target proto.Visitee) {
		switch value := target.(type) {
		case *proto.Syntax:
			file.setSyntax(value)
		case *proto.Package:
			file.setPackage(value)
		case *proto.Import:
			file.appendImport(value)
		case *proto.Service:
			file.appendService(value)
		case *proto.Option:
			if value.Parent == root {
				file.appendOption(value)
			}
		case *proto.Enum:
			if value.Parent == root {
				file.appendEnum(value)
			}
		case *proto.Message:
			if value.Parent == root {
				file.AppendMessage(value)
			}
		}
	}
}

func (file *File) IndexOfImportf(format string, args ...interface{}) int {
	return file.IndexOfImport(fmt.Sprintf(format, args...))
}

func (file *File) IndexOfImport(filename string) int {
	for idx, item := range file.Imports {
		if item.Filename == filename {
			return idx
		}
	}
	return -1
}

func (file *File) IndexOfMessage(name string) int {
	for idx, item := range file.Messages {
		if item.Name == name {
			return idx
		}
	}
	return -1
}

// RemoveServiceAt removes the service located at the index provided.
//
// This function will panic if the provided index is invalid
func (file *File) RemoveServiceAt(idx int) {
	file.Services = append(file.Services[:idx], file.Services[idx+1:]...)
}

// RemoveImportAt removes the import located at the index provided.
//
// This function will panic if the provided index is invalid
func (file *File) RemoveImportAt(idx int) {
	file.Imports = append(file.Imports[:idx], file.Imports[idx+1:]...)
}

// RemoveMessageAt removes the message located at the index provided.
//
// This function will panic if the provided index is invalid
func (file *File) RemoveMessageAt(idx int) {
	file.Messages = append(file.Messages[:idx], file.Messages[idx+1:]...)
}

// FindServicef takes the provided format specifier and arguments and calls
// FindService with the built string
func (file *File) FindServicef(format string, args ...interface{}) (*Service, error) {
	return file.FindService(fmt.Sprintf(format, args...))
}

// FindImportf takes the provided format specifier and arguments and calls
// FindImport with the built string
func (file *File) FindImportf(format string, args ...interface{}) (*proto.Import, error) {
	return file.FindImport(fmt.Sprintf(format, args...))
}

// FindMessagef takes the provided format specifier and arguments and calls
// FindMessage with the built string
func (file *File) FindMessagef(format string, args ...interface{}) (*Message, error) {
	return file.FindMessage(fmt.Sprintf(format, args...))
}

// FindService attempts to return the service with the provided name
func (file *File) FindService(name string) (*Service, error) {
	for _, service := range file.Services {
		if service.Name == name {
			return service, nil
		}
	}
	return nil, fmt.Errorf("%w %q", ErrServiceNotFound, name)
}

// FindImport attempts to return the import with the provided name
func (file *File) FindImport(filename string) (*proto.Import, error) {
	for _, node := range file.Imports {
		if node.Filename == filename {
			return node, nil
		}
	}
	return nil, fmt.Errorf("%w %q", ErrImportNotFound, filename)
}

// FindMessage attempts to return the message with the provided name
func (file *File) FindMessage(name string) (*Message, error) {
	for _, message := range file.Messages {
		if message.Name == name {
			return message, nil
		}
	}
	return nil, fmt.Errorf("%w %q", ErrMessageNotFound, name)
}

// FindRemoteProcedure attempts to return the RPC with the name provided,
// located inside the name of the service that is also provided.
func (file *File) FindRemoteProcedure(serviceName, name string) (*proto.RPC, error) {
	service, err := file.FindService(serviceName)
	if err != nil {
		return nil, err
	}
	rpc, err := service.FindRPC(name)
	if err != nil {
		return nil, fmt.Errorf("%w in service %q", err, serviceName)
	}
	return rpc, nil
}

// AppendImportf calls AppendImport with the formatted string
func (file *File) AppendImportf(format string, args ...interface{}) {
	file.AppendImport(fmt.Sprintf(format, args...))
}

// AppendImport appends an import with the provided filename
func (file *File) AppendImport(filename string) {
	file.Imports = append(file.Imports, &proto.Import{Filename: filename})
}

// PrependImportf calls PrependImport with the formatted string
func (file *File) PrependImportf(format string, args ...interface{}) {
	file.PrependImport(fmt.Sprintf(format, args...))
}

// PrependImport prepends an import with the provided filename
func (file *File) PrependImport(filename string) {
	nodes := []*proto.Import{
		{Filename: filename},
	}
	file.Imports = append(nodes, file.Imports...)
}

// Proto returns a proto.Proto type for interaction with the API File currently
// wraps
func (file *File) Proto() *proto.Proto {
	length := len(file.Imports) +
		len(file.Options) +
		len(file.Enums) +
		len(file.Messages) +
		len(file.Services) +
		2
	elements := make([]proto.Visitee, 0, length)
	elements = append(elements, file.Syntax, file.Package)

	for _, item := range file.Imports {
		elements = append(elements, item)
	}

	for _, item := range file.Options {
		elements = append(elements, item)
	}

	for _, item := range file.Enums {
		elements = append(elements, item)
	}

	for _, item := range file.Messages {
		elements = append(elements, item.Proto())
	}

	for _, item := range file.Services {
		elements = append(elements, item.Proto())
	}

	return &proto.Proto{
		Filename: file.Filename,
		Elements: elements,
	}
}

func (file *File) setSyntax(syntax *proto.Syntax) {
	file.Syntax = syntax
}

func (file *File) setPackage(pkg *proto.Package) {
	file.Package = pkg
}

func (file *File) appendImport(path *proto.Import) {
	file.Imports = append(file.Imports, path)
}

func (file *File) appendOption(option *proto.Option) {
	file.Options = append(file.Options, option)
}

func (file *File) appendEnum(enum *proto.Enum) {
	file.Enums = append(file.Enums, enum)
}

// AppendMessage appends the provided message to the file
func (file *File) AppendMessage(message *proto.Message) {
	file.Messages = append(file.Messages, NewMessage(message))
}

func (file *File) appendService(service *proto.Service) {
	file.Services = append(file.Services, NewService(service))
}
