package protocode

import (
	"fmt"

	"github.com/emicklei/proto"
)

type Service struct {
	*proto.Service
	procedures []*proto.RPC
	options    []*proto.Option
}

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

func (service *Service) FindRPC(name string) (*proto.RPC, error) {
	idx := service.IndexOfRPC(name)
	if idx == -1 {
		return nil, fmt.Errorf("%w %q", ErrRemoteProcedureNotFound, name)
	}
	return service.procedures[idx], nil
}

func (service *Service) IndexOfRPC(name string) int {
	for idx, item := range service.procedures {
		if item.Name == name {
			return idx
		}
	}
	return -1
}

func (service *Service) AppendRPCs(rpcs ...*proto.RPC) {
	for _, rpc := range rpcs {
		service.AppendRPC(rpc)
	}
}

func (service *Service) AppendRPC(rpc *proto.RPC) {
	service.procedures = append(service.procedures, rpc)
}

func (service *Service) RemoveRPCAt(idx int) {
	service.procedures = append(service.procedures[:idx], service.procedures[idx+1:]...)
}

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
