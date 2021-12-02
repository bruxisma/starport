package protocode

import (
	"bytes"
	"errors"
	"io"

	"github.com/emicklei/proto"
	"github.com/emicklei/proto-contrib/pkg/protofmt"
)

type File = proto.Proto

var ErrMissingPackage = errors.New("'package' statement missing")

func Parse(reader io.Reader, filename string) (*File, error) {
	parser := proto.NewParser(reader)
	parser.Filename(filename)
	return parser.Parse()
}

func Write(tree *File) (io.Reader, error) {
	buffer := &bytes.Buffer{}
	if err := Fprint(buffer, tree); err != nil {
		return nil, err
	}
	return buffer, nil
}

func Fprint(w io.Writer, root *File) error {
	formatter := protofmt.NewFormatter(w, "\t")
	formatter.Format(root)
	return nil
}

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

func (of *OrganizedFile) AsFile() *File {
	length := len(of.Imports) +
		len(of.Options) +
		len(of.Enums) +
		len(of.Messages) +
		len(of.Procedures) +
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
