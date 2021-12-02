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
	Syntax   *proto.Syntax
	Imports  []*proto.Import
	Packages []*proto.Package
	Options  []*proto.Option
	Messages []*proto.Message
	Enums    []*proto.Enum
	Services []*proto.Service
}

func WithFile(file *OrganizedFile) proto.Handler {
	return func(target proto.Visitee) {
		switch value := target.(type) {
		case *proto.Syntax:
			file.setSyntax(value)
		case *proto.Import:
			file.appendImport(value)
		case *proto.Package:
			file.appendPackage(value)
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

func (file *OrganizedFile) setSyntax(syntax *proto.Syntax) {
	file.Syntax = syntax
}

func (file *OrganizedFile) appendImport(path *proto.Import) {
	file.Imports = append(file.Imports, path)
}

func (file *OrganizedFile) appendPackage(pkg *proto.Package) {
	file.Packages = append(file.Packages, pkg)
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
