package mutate

import (
	"fmt"
	"strings"

	"github.com/dave/dst"
	"github.com/emicklei/proto"
	"github.com/tendermint/starport/starport/pkg/gocode"
	"github.com/tendermint/starport/starport/pkg/protocode"
	"github.com/tendermint/starport/starport/templates/typed"
)

//StargateHandlerInsertMsgServer inserts a NewMsgServerImpl if (and ONLY IF) it
//is not defined in the function NewHandler.  It does this by first finding the
//assignment statement. If it is *not* found, it will insert it *before* the
//switch statement in the function.
func StargateHandlerInsertMsgServer(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := gocode.FindFunction(tree, "NewHandler")
	if err != nil {
		return nil, err
	}
	for _, stmt := range fn.Body.List {
		if assignment, ok := stmt.(*dst.AssignStmt); !ok {
			continue
		} else if identifier, ok := assignment.Lhs[0].(*dst.Ident); !ok || identifier.Name != "msgServer" {
			continue
		} else {
			gocode.Apply(fn, func(cursor *gocode.Cursor) bool {
				if _, ok := cursor.Node().(*dst.SwitchStmt); !ok {
					return true
				}
				call := gocode.Call("keeper", "NewMsgServerImpl").
					WithArgument("k").
					Node()
				assignment := gocode.DefineVariable("msgServer").To(call)
				cursor.InsertBefore(assignment)
				return false
			})
			break
		}
	}
	return tree, nil
}

// StargateHandlerInsertCases inserts several case clauses into the NewHandler function
func StargateHandlerInsertCases(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := gocode.FindFunction(tree, "NewHandler")
	if err != nil {
		return nil, err
	}

	gocode.Apply(fn, func(cursor *gocode.Cursor) bool {
		if clause, ok := cursor.Node().(*dst.CaseClause); !ok || len(clause.List) != 0 {
			return true
		}
		for _, action := range []string{"Create", "Update", "Delete"} {
			cursor.InsertBefore(stargateCreateCaseClause(action, opts))
		}
		return false
	})
	return tree, nil
}

func StargateClientCliTxInsertCommands(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := gocode.FindFunction(tree, "GetTxCmd")
	if err != nil {
		return nil, err
	}

	gocode.Apply(fn, func(cursor *gocode.Cursor) bool {
		if _, ok := cursor.Node().(*dst.ReturnStmt); !ok {
			return true
		}
		for _, action := range []string{"Create", "Update", "Delete"} {
			stmt := gocode.Call("cmd", "AddCommand").
				WithParameters(gocode.Callf("Cmd%v%v", action, opts.TypeName.UpperCamel).Node()).
				AsStatement()
			cursor.InsertBefore(stmt)
		}
		return false
	})
	return tree, nil
}

func StargateClientCliQueryInserCommands(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := gocode.FindFunction(tree, "GetQueryCmd")
	if err != nil {
		return nil, err
	}

	gocode.Apply(fn, func(cursor *gocode.Cursor) bool {
		if _, ok := cursor.Node().(*dst.ReturnStmt); !ok {
			return true
		}
		for _, action := range []string{"List", "Show"} {
			stmt := gocode.Call("cmd", "AddCommand").
				WithParameters(gocode.Callf("Cmd%v%v", action, opts.TypeName.UpperCamel).Node()).
				AsStatement()
			cursor.InsertBefore(stmt)
		}
		return false
	})

	return tree, nil
}

func StargateInsertGRPCGateway(tree *dst.File, opts *typed.Options) (*dst.File, error) {
	fn, err := gocode.FindFunction(tree, "RegisterGRPCGatewayRoutes")
	if err != nil {
		return nil, err
	}
	if len(fn.Body.List) != 0 {
		return tree, nil
	}

	background := gocode.Call("context", "Background").Node()
	client := gocode.Call("types", "NewQueryClient").WithArgument("clientCtx").Node()

	register := gocode.Call("types", "RegisterQueryHandlerClient").
		WithParameters(background).
		WithArgument("mux").
		WithParameters(client).
		AsStatement()

	fn.Body.List = append(fn.Body.List, register)
	return tree, nil
}

// StargateProtoTx modifies the tx.proto file
func StargateProtoTx(tree *protocode.File, opts *typed.Options) (*protocode.File, error) {
	/* Mutate All Imports */
	if idx := tree.IndexOfImportf("%s/%s.proto", opts.ModuleName, opts.TypeName.Snake); idx >= 0 {
		tree.RemoveImportAt(idx)
	}
	tree.AppendImportf("%s/%s.proto", opts.ModuleName, opts.TypeName.Snake)

	for _, path := range opts.Fields.ProtoImports() {
		if idx := tree.IndexOfImport(path); idx >= 0 {
			tree.RemoveImportAt(idx)
		}
		tree.AppendImport(path)
	}

	for _, field := range opts.Fields.Custom() {
		if idx := tree.IndexOfImportf("%s/%s.proto", opts.ModuleName, field); idx >= 0 {
			tree.RemoveImportAt(idx)
		}
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

// StargateProtoQuery modifies the query.go file
func StargateProtoQuery(tree *protocode.File, opts *typed.Options) (*protocode.File, error) {
	if idx := tree.IndexOfImport("gogoproto/gogo.proto"); idx >= 0 {
		tree.RemoveImportAt(idx)
	}
	if idx := tree.IndexOfImportf("%s/%s.proto", opts.ModuleName, opts.TypeName.Snake); idx >= 0 {
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
		{
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
		{
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

// case *types.MsgCreate%[2]v:
//	res, err := msgServer.Create%[2]v(sdk.WrapSDKContext(ctx), msg)
//	return sdk.WrapServiceResult(ctx, res, err)

func stargateCreateCaseClause(action string, opts *typed.Options) *dst.CaseClause {
	name := fmt.Sprintf("%v%v", action, opts.TypeName.UpperCamel)
	typename := gocode.Identifier("types", fmt.Sprintf("Msg%v", name))

	wrapSDKContext := gocode.Call("msgServer", name).
		WithParameters(gocode.Call("sdk", "WrapSDKContext").WithArgument("ctx").Node()).
		WithArgument("msg").
		Node()
	wrapServiceResult := gocode.Call("sdk", "WrapServiceResult").
		WithArgument("ctx").
		WithArgument("res").
		WithArgument("err").
		Node()

	return &dst.CaseClause{
		Decs: dst.CaseClauseDecorations{NodeDecs: dst.NodeDecs{After: dst.EmptyLine}},
		List: []dst.Expr{&dst.StarExpr{X: typename}},
		Body: []dst.Stmt{
			gocode.DefineCheck("res").To(wrapSDKContext),
			&dst.ReturnStmt{Results: []dst.Expr{wrapServiceResult}},
		},
	}
}
