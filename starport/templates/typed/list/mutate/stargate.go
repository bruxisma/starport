package mutate

import (
	"fmt"
	"strings"

	"github.com/emicklei/proto"
	"github.com/tendermint/starport/starport/pkg/protocode"
	"github.com/tendermint/starport/starport/templates/typed"
)

func StargateProtoTx(tree *protocode.File, opts *typed.Options) (*protocode.File, error) {
	/* Mutate All Imports */
	if idx := tree.IndexOfImportf("%s/%s.proto", opts.ModuleName, opts.TypeName.Snake); idx > 0 {
		tree.RemoveImportAt(idx)
	}
	tree.AppendImportf("%s/%s.proto", opts.ModuleName, opts.TypeName.Snake)

	for _, path := range opts.Fields.ProtoImports() {
		tree.AppendImport(path)
	}

	for _, field := range opts.Fields.Custom() {
		tree.AppendImportf("%s/%s.proto", opts.ModuleName, field)
	}

	service, err := tree.FindService("Msg")
	if err != nil {
		return nil, err
	}
	for _, action := range []string{"Create", "Update", "Delete"} {
		rpc := &proto.RPC{
			Name:        fmt.Sprintf("%[1]s%[2]s", action, opts.TypeName.UpperCamel),
			RequestType: fmt.Sprintf("Msg%[1]s%[2]s", action, opts.TypeName.UpperCamel),
			ReturnsType: fmt.Sprintf("Msg%[1]s%[2]sResponse", action, opts.TypeName.UpperCamel),
		}
		service.AppendRPC(rpc)
	}

	msgCreateResponse := protocode.CreateMessagef("MsgCreate%sResponse", opts.TypeName.UpperCamel)
	msgUpdateResponse := protocode.CreateMessagef("MsgUpdate%sResponse", opts.TypeName.UpperCamel)
	msgDeleteResponse := protocode.CreateMessagef("MsgDelete%sResponse", opts.TypeName.UpperCamel)
	msgCreate := protocode.CreateMessagef("MsgCreate%s", opts.TypeName.UpperCamel)
	msgUpdate := protocode.CreateMessagef("MsgUpdate%s", opts.TypeName.UpperCamel)
	msgDelete := protocode.CreateMessagef("MsgDelete%s", opts.TypeName.UpperCamel)

	signerField := &proto.Field{Name: opts.MsgSigner.LowerCamel, Type: "string"}
	idField := &proto.Field{Name: "id", Type: "uint64"}

	msgCreateResponse.AppendField(idField)
	msgDelete.AppendFields(signerField, idField)
	msgUpdate.AppendFields(signerField, idField)
	msgCreate.AppendFields(signerField)

	for _, msg := range []*protocode.Message{msgCreate, msgUpdate} {
		for _, field := range opts.Fields {
			msg.AppendField(&proto.Field{
				Name: field.Name.LowerCamel,
				Type: string(field.DatatypeName),
			})
		}
	}

	tree.Messages = append(
		tree.Messages,
		msgCreate,
		msgCreateResponse,
		msgUpdate,
		msgUpdateResponse,
		msgDelete,
		msgDeleteResponse)

	return tree, nil
}

func StargateProtoQuery(tree *protocode.File, opts *typed.Options) (*protocode.File, error) {
	if idx := tree.IndexOfImport("gogoproto/gogo.proto"); idx > 0 {
		tree.RemoveImportAt(idx)
	}
	if idx := tree.IndexOfImportf("%s/%s.proto", opts.ModuleName, opts.TypeName.Snake); idx > 0 {
		tree.RemoveImportAt(idx)
	}
	tree.PrependImport("gogoproto/gogo.proto")
	tree.AppendImportf("%s/%s.proto", opts.ModuleName, opts.TypeName.Snake)

	service, err := tree.FindService("Query")
	if err != nil {
		return nil, err
	}

	stargateProtoQueryService(service, opts)
	stargateProtoQueryMessages(tree, opts)

	return tree, nil
}

func stargateProtoQueryService(service *protocode.Service, opts *typed.Options) {
	queryPathComponents := []string{
		opts.OwnerName,
		opts.AppName,
		opts.ModuleName,
		opts.TypeName.LowerCamel,
	}
	queryPathPrefix := fmt.Sprintf("%s", strings.Join(queryPathComponents, "/"))
	queryProcedures := []*proto.RPC{
		&proto.RPC{
			Name:        opts.TypeName.UpperCamel,
			Comment:     protocode.Commentf("Queries a %s by id", opts.TypeName.LowerCamel),
			RequestType: fmt.Sprintf("QueryGet%sRequest", opts.TypeName.UpperCamel),
			ReturnsType: fmt.Sprintf("QueryGet%sResponse", opts.TypeName.UpperCamel),
			Elements: []proto.Visitee{
				&proto.Option{
					Name:     "(google.api.http).get",
					Constant: protocode.String("/%v/{id}", queryPathPrefix),
				},
			},
		},
		&proto.RPC{
			Name:        fmt.Sprintf("%vAll", opts.TypeName.UpperCamel),
			Comment:     protocode.Commentf("Queries a list of %v items", opts.TypeName.LowerCamel),
			RequestType: fmt.Sprintf("QueryAll%vRequest", opts.TypeName.UpperCamel),
			ReturnsType: fmt.Sprintf("QueryAll%vResponse", opts.TypeName.UpperCamel),
			Elements: []proto.Visitee{
				&proto.Option{
					Name:     "(google.api.http).get",
					Constant: protocode.String("/%v", queryPathPrefix),
				},
			},
		},
	}

	service.AppendRPCs(queryProcedures...)
}

func stargateProtoQueryMessages(tree *protocode.File, opts *typed.Options) {
	msgGetResponse := protocode.CreateMessagef("QueryGet%vResponse", opts.TypeName.UpperCamel)
	msgAllResponse := protocode.CreateMessagef("QueryAll%vResponse", opts.TypeName.UpperCamel)
	msgGetRequest := protocode.CreateMessagef("QueryGet%vRequest", opts.TypeName.UpperCamel)
	msgAllRequest := protocode.CreateMessagef("QueryAll%vRequest", opts.TypeName.UpperCamel)

	paginationField := &proto.Field{
		Name: "pagination",
		Type: "cosmos.base.query.v1beta1.PageResponse",
	}
	namedField := &proto.Field{
		Name: opts.TypeName.UpperCamel,
		Type: opts.TypeName.UpperCamel,
		Options: []*proto.Option{
			{Name: "(gogoproto.nullable)", Constant: protocode.False()},
		},
	}

	msgGetRequest.AppendField(&proto.Field{Name: "id", Type: "uint64"})

	msgGetResponse.AppendField(namedField)

	msgAllRequest.AppendField(paginationField)

	msgAllResponse.AppendRepeatedField(namedField)
	msgAllResponse.AppendField(paginationField)

	tree.Messages = append(
		tree.Messages,
		msgGetRequest,
		msgGetResponse,
		msgAllRequest,
		msgAllResponse,
	)
}
