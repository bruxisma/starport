package protocode

import (
	"fmt"

	"github.com/emicklei/proto"
)

// Service is a wrapper around the proto.Service type to support easier editing
// and construction of protobuf AST elements.
type Service struct {
	*proto.Service
	procedures []*proto.RPC
	options    []*proto.Option
}

// NewService constructs a Service from a proto.Service
func NewService(input *proto.Service) *Service {
	procedures := []*proto.RPC{}
	options := []*proto.Option{}
	for _, item := range input.Elements {
		if procedure, ok := item.(*proto.RPC); ok {
			procedures = append(procedures, procedure)
		} else if option, ok := item.(*proto.Option); ok {
			options = append(options, option)
		}
	}
	service := &Service{input, procedures, options}
	return service
}

// FindRPC returns an RPC with the provided name inside the received service
func (service *Service) FindRPC(name string) (*proto.RPC, error) {
	idx := service.IndexOfRPC(name)
	if idx == -1 {
		return nil, fmt.Errorf("%w %q", ErrRemoteProcedureNotFound, name)
	}
	return service.procedures[idx], nil
}

// IndexOfRPC returns the index of an RPC with the provided name inside of the
// received service
func (service *Service) IndexOfRPC(name string) int {
	for idx, item := range service.procedures {
		if item.Name == name {
			return idx
		}
	}
	return -1
}

// AppendRPCs appends the provided RPCs to the service
func (service *Service) AppendRPCs(rpcs ...*proto.RPC) {
	for _, rpc := range rpcs {
		service.AppendRPC(rpc)
	}
}

// AppendRPC appends a single RPC to the service
func (service *Service) AppendRPC(rpc *proto.RPC) {
	service.procedures = append(service.procedures, rpc)
}

// RemoveRPCAt removes an RPC at the provided index. If the index is -1, this
// function will panic.
func (service *Service) RemoveRPCAt(idx int) {
	service.procedures = append(service.procedures[:idx], service.procedures[idx+1:]...)
}

// Proto returns a proto.Service constructed from the received Service
func (service *Service) Proto() *proto.Service {
	length := len(service.procedures) + len(service.options)
	elements := make([]proto.Visitee, 0, length)
	for _, item := range service.options {
		elements = append(elements, item)
	}
	for _, item := range service.procedures {
		elements = append(elements, item)
	}
	service.Service.Elements = elements
	return service.Service
}
