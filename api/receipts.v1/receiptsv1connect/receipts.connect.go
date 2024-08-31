// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: receipts.v1/receipts.proto

package receiptsv1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	receipts_v1 "github.com/manzanit0/mcduck/api/receipts.v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_13_0

const (
	// ReceiptsServiceName is the fully-qualified name of the ReceiptsService service.
	ReceiptsServiceName = "receipts.v1.ReceiptsService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ReceiptsServiceCreateReceiptProcedure is the fully-qualified name of the ReceiptsService's
	// CreateReceipt RPC.
	ReceiptsServiceCreateReceiptProcedure = "/receipts.v1.ReceiptsService/CreateReceipt"
	// ReceiptsServiceUpdateReceiptProcedure is the fully-qualified name of the ReceiptsService's
	// UpdateReceipt RPC.
	ReceiptsServiceUpdateReceiptProcedure = "/receipts.v1.ReceiptsService/UpdateReceipt"
	// ReceiptsServiceDeleteReceiptProcedure is the fully-qualified name of the ReceiptsService's
	// DeleteReceipt RPC.
	ReceiptsServiceDeleteReceiptProcedure = "/receipts.v1.ReceiptsService/DeleteReceipt"
)

// These variables are the protoreflect.Descriptor objects for the RPCs defined in this package.
var (
	receiptsServiceServiceDescriptor             = receipts_v1.File_receipts_v1_receipts_proto.Services().ByName("ReceiptsService")
	receiptsServiceCreateReceiptMethodDescriptor = receiptsServiceServiceDescriptor.Methods().ByName("CreateReceipt")
	receiptsServiceUpdateReceiptMethodDescriptor = receiptsServiceServiceDescriptor.Methods().ByName("UpdateReceipt")
	receiptsServiceDeleteReceiptMethodDescriptor = receiptsServiceServiceDescriptor.Methods().ByName("DeleteReceipt")
)

// ReceiptsServiceClient is a client for the receipts.v1.ReceiptsService service.
type ReceiptsServiceClient interface {
	CreateReceipt(context.Context, *connect.Request[receipts_v1.CreateReceiptRequest]) (*connect.Response[receipts_v1.CreateReceiptResponse], error)
	UpdateReceipt(context.Context, *connect.Request[receipts_v1.UpdateReceiptRequest]) (*connect.Response[receipts_v1.UpdateReceiptResponse], error)
	DeleteReceipt(context.Context, *connect.Request[receipts_v1.DeleteReceiptRequest]) (*connect.Response[receipts_v1.DeleteReceiptResponse], error)
}

// NewReceiptsServiceClient constructs a client for the receipts.v1.ReceiptsService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewReceiptsServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) ReceiptsServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	return &receiptsServiceClient{
		createReceipt: connect.NewClient[receipts_v1.CreateReceiptRequest, receipts_v1.CreateReceiptResponse](
			httpClient,
			baseURL+ReceiptsServiceCreateReceiptProcedure,
			connect.WithSchema(receiptsServiceCreateReceiptMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		updateReceipt: connect.NewClient[receipts_v1.UpdateReceiptRequest, receipts_v1.UpdateReceiptResponse](
			httpClient,
			baseURL+ReceiptsServiceUpdateReceiptProcedure,
			connect.WithSchema(receiptsServiceUpdateReceiptMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
		deleteReceipt: connect.NewClient[receipts_v1.DeleteReceiptRequest, receipts_v1.DeleteReceiptResponse](
			httpClient,
			baseURL+ReceiptsServiceDeleteReceiptProcedure,
			connect.WithSchema(receiptsServiceDeleteReceiptMethodDescriptor),
			connect.WithClientOptions(opts...),
		),
	}
}

// receiptsServiceClient implements ReceiptsServiceClient.
type receiptsServiceClient struct {
	createReceipt *connect.Client[receipts_v1.CreateReceiptRequest, receipts_v1.CreateReceiptResponse]
	updateReceipt *connect.Client[receipts_v1.UpdateReceiptRequest, receipts_v1.UpdateReceiptResponse]
	deleteReceipt *connect.Client[receipts_v1.DeleteReceiptRequest, receipts_v1.DeleteReceiptResponse]
}

// CreateReceipt calls receipts.v1.ReceiptsService.CreateReceipt.
func (c *receiptsServiceClient) CreateReceipt(ctx context.Context, req *connect.Request[receipts_v1.CreateReceiptRequest]) (*connect.Response[receipts_v1.CreateReceiptResponse], error) {
	return c.createReceipt.CallUnary(ctx, req)
}

// UpdateReceipt calls receipts.v1.ReceiptsService.UpdateReceipt.
func (c *receiptsServiceClient) UpdateReceipt(ctx context.Context, req *connect.Request[receipts_v1.UpdateReceiptRequest]) (*connect.Response[receipts_v1.UpdateReceiptResponse], error) {
	return c.updateReceipt.CallUnary(ctx, req)
}

// DeleteReceipt calls receipts.v1.ReceiptsService.DeleteReceipt.
func (c *receiptsServiceClient) DeleteReceipt(ctx context.Context, req *connect.Request[receipts_v1.DeleteReceiptRequest]) (*connect.Response[receipts_v1.DeleteReceiptResponse], error) {
	return c.deleteReceipt.CallUnary(ctx, req)
}

// ReceiptsServiceHandler is an implementation of the receipts.v1.ReceiptsService service.
type ReceiptsServiceHandler interface {
	CreateReceipt(context.Context, *connect.Request[receipts_v1.CreateReceiptRequest]) (*connect.Response[receipts_v1.CreateReceiptResponse], error)
	UpdateReceipt(context.Context, *connect.Request[receipts_v1.UpdateReceiptRequest]) (*connect.Response[receipts_v1.UpdateReceiptResponse], error)
	DeleteReceipt(context.Context, *connect.Request[receipts_v1.DeleteReceiptRequest]) (*connect.Response[receipts_v1.DeleteReceiptResponse], error)
}

// NewReceiptsServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewReceiptsServiceHandler(svc ReceiptsServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	receiptsServiceCreateReceiptHandler := connect.NewUnaryHandler(
		ReceiptsServiceCreateReceiptProcedure,
		svc.CreateReceipt,
		connect.WithSchema(receiptsServiceCreateReceiptMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	receiptsServiceUpdateReceiptHandler := connect.NewUnaryHandler(
		ReceiptsServiceUpdateReceiptProcedure,
		svc.UpdateReceipt,
		connect.WithSchema(receiptsServiceUpdateReceiptMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	receiptsServiceDeleteReceiptHandler := connect.NewUnaryHandler(
		ReceiptsServiceDeleteReceiptProcedure,
		svc.DeleteReceipt,
		connect.WithSchema(receiptsServiceDeleteReceiptMethodDescriptor),
		connect.WithHandlerOptions(opts...),
	)
	return "/receipts.v1.ReceiptsService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ReceiptsServiceCreateReceiptProcedure:
			receiptsServiceCreateReceiptHandler.ServeHTTP(w, r)
		case ReceiptsServiceUpdateReceiptProcedure:
			receiptsServiceUpdateReceiptHandler.ServeHTTP(w, r)
		case ReceiptsServiceDeleteReceiptProcedure:
			receiptsServiceDeleteReceiptHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedReceiptsServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedReceiptsServiceHandler struct{}

func (UnimplementedReceiptsServiceHandler) CreateReceipt(context.Context, *connect.Request[receipts_v1.CreateReceiptRequest]) (*connect.Response[receipts_v1.CreateReceiptResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("receipts.v1.ReceiptsService.CreateReceipt is not implemented"))
}

func (UnimplementedReceiptsServiceHandler) UpdateReceipt(context.Context, *connect.Request[receipts_v1.UpdateReceiptRequest]) (*connect.Response[receipts_v1.UpdateReceiptResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("receipts.v1.ReceiptsService.UpdateReceipt is not implemented"))
}

func (UnimplementedReceiptsServiceHandler) DeleteReceipt(context.Context, *connect.Request[receipts_v1.DeleteReceiptRequest]) (*connect.Response[receipts_v1.DeleteReceiptResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("receipts.v1.ReceiptsService.DeleteReceipt is not implemented"))
}
