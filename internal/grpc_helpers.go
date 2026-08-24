package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

// GRPCResponse represents the result of a gRPC call
type GRPCResponse struct {
	Message  proto.Message
	Metadata metadata.MD
	Code     codes.Code
	Error    error
}

// ExecuteGRPCRequest performs a gRPC unary call
func ExecuteGRPCRequest(
	t *testing.T,
	step Step,
	serverAddr string,
	ctxMap map[string]any,
) *GRPCResponse {
	t.Helper()
	const op = "ExecuteGRPCRequest"

	if step.GRPCRequest == nil {
		grpcErr := NewError(ErrGRPC, op, "grpcRequest is nil").
			WithContext("step", step.Name)
		t.Fatalf("%+v", grpcErr)
	}

	// Render request with context
	grpcReq := renderGRPCRequest(*step.GRPCRequest, ctxMap)

	// Create connection
	conn, err := grpc.NewClient(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		grpcErr := NewError(ErrGRPC, op, "failed to create gRPC client").
			WithContext("step", step.Name).
			WithContext("error", err.Error())
		t.Fatalf("%+v", grpcErr)
	}
	defer conn.Close()

	// Create context with metadata
	ctx := context.Background()
	if len(grpcReq.Metadata) > 0 {
		md := metadata.New(grpcReq.Metadata)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	// Find the method descriptor
	fullMethodName := fmt.Sprintf("/%s/%s", grpcReq.Service, grpcReq.Method)
	methodDesc, err := findMethodDescriptor(grpcReq.Service, grpcReq.Method)
	if err != nil {
		grpcErr := NewError(ErrGRPC, op, "failed to find method descriptor").
			WithContext("step", step.Name).
			WithContext("service", grpcReq.Service).
			WithContext("method", grpcReq.Method).
			WithContext("error", err.Error())
		t.Fatalf("%+v", grpcErr)
	}

	// Create request message from map
	reqMsg, err := mapToProtoMessage(grpcReq.Message, methodDesc.Input())
	if err != nil {
		grpcErr := NewError(ErrGRPC, op, "failed to create request message").
			WithContext("step", step.Name).
			WithContext("error", err.Error())
		t.Fatalf("%+v", grpcErr)
	}

	// Create response message
	respMsg := dynamicpb.NewMessage(methodDesc.Output())

	// Invoke the method
	var header, trailer metadata.MD
	err = conn.Invoke(
		ctx,
		fullMethodName,
		reqMsg,
		respMsg,
		grpc.Header(&header),
		grpc.Trailer(&trailer),
	)

	// Extract status code
	grpcStatus := status.Code(err)

	return &GRPCResponse{
		Message:  respMsg,
		Metadata: header,
		Code:     grpcStatus,
		Error:    err,
	}
}

// AssertGRPCResponse validates gRPC response
func AssertGRPCResponse(
	t *testing.T,
	response *GRPCResponse,
	expected GRPCResponseSpec,
	ctxMap map[string]any,
) {
	t.Helper()
	AssertStepGRPCResponse(t, Step{}, response, expected, ctxMap)
}

// AssertStepGRPCResponse validates gRPC response and includes step context in failure diagnostics.
func AssertStepGRPCResponse(
	t *testing.T,
	step Step,
	response *GRPCResponse,
	expected GRPCResponseSpec,
	ctxMap map[string]any,
) {
	t.Helper()
	const op = "AssertGRPCResponse"

	// Check status code
	expectedCode := parseGRPCCode(expected.Code)
	if response.Code != expectedCode {
		grpcErr := NewError(ErrGRPC, op, "unexpected gRPC status code").
			WithContext("step", step.Name).
			WithContext("expected", expected.Code).
			WithContext("actual", response.Code.String()).
			WithContext("!diagnostic", formatGRPCDiagnostic(step, response))
		t.Fatalf("%+v", grpcErr)
	}

	// If error expected, we're done
	if expectedCode != codes.OK {
		return
	}

	// Check metadata
	if len(expected.Metadata) > 0 {
		for k, v := range expected.Metadata {
			values := response.Metadata.Get(k)
			found := false
			for _, val := range values {
				if val == v {
					found = true
					break
				}
			}
			if !found {
				grpcErr := NewError(ErrGRPC, op, "unexpected metadata value").
					WithContext("step", step.Name).
					WithContext("key", k).
					WithContext("expected", v).
					WithContext("actual", values).
					WithContext("!diagnostic", formatGRPCDiagnostic(step, response))
				t.Fatalf("%+v", grpcErr)
			}
		}
	}

	// Check message body
	if expected.Message != "" {
		// Convert proto message to JSON
		jsonBytes, err := protojson.Marshal(response.Message)
		if err != nil {
			grpcErr := NewError(ErrGRPC, op, "failed to marshal response to JSON").
				WithContext("step", step.Name).
				WithContext("error", err.Error())
			t.Fatalf("%+v", grpcErr)
		}

		// Render variables in expected message using context from previous steps
		renderedExpected := RenderTemplate(expected.Message, ctxMap)

		// Assert JSON
		assertJSONWithDiagnostic(
			t,
			string(jsonBytes),
			renderedExpected,
			"gRPC response assertion context",
			formatGRPCDiagnostic(step, response),
		)
	}
}

// Helper functions

func renderGRPCRequest(req GRPCRequestSpec, ctx map[string]any) GRPCRequestSpec {
	req.Service = RenderTemplate(req.Service, ctx)
	req.Method = RenderTemplate(req.Method, ctx)
	req.Message = RenderAny(req.Message, ctx).(map[string]any)

	for k, v := range req.Metadata {
		req.Metadata[k] = RenderTemplate(v, ctx)
	}

	return req
}

func parseGRPCCode(codeStr string) codes.Code {
	switch codeStr {
	case "OK":
		return codes.OK
	case "CANCELLED", "Cancelled":
		return codes.Canceled
	case "UNKNOWN", "Unknown":
		return codes.Unknown
	case "INVALID_ARGUMENT", "InvalidArgument":
		return codes.InvalidArgument
	case "DEADLINE_EXCEEDED", "DeadlineExceeded":
		return codes.DeadlineExceeded
	case "NOT_FOUND", "NotFound":
		return codes.NotFound
	case "ALREADY_EXISTS", "AlreadyExists":
		return codes.AlreadyExists
	case "PERMISSION_DENIED", "PermissionDenied":
		return codes.PermissionDenied
	case "RESOURCE_EXHAUSTED", "ResourceExhausted":
		return codes.ResourceExhausted
	case "FAILED_PRECONDITION", "FailedPrecondition":
		return codes.FailedPrecondition
	case "ABORTED", "Aborted":
		return codes.Aborted
	case "OUT_OF_RANGE", "OutOfRange":
		return codes.OutOfRange
	case "UNIMPLEMENTED", "Unimplemented":
		return codes.Unimplemented
	case "INTERNAL", "Internal":
		return codes.Internal
	case "UNAVAILABLE", "Unavailable":
		return codes.Unavailable
	case "DATA_LOSS", "DataLoss":
		return codes.DataLoss
	case "UNAUTHENTICATED", "Unauthenticated":
		return codes.Unauthenticated
	default:
		return codes.Unknown
	}
}

func findMethodDescriptor(serviceName, methodName string) (protoreflect.MethodDescriptor, error) {
	// Find service descriptor
	fullServiceName := protoreflect.FullName(serviceName)
	desc, err := protoregistry.GlobalFiles.FindDescriptorByName(fullServiceName)
	if err != nil {
		return nil, fmt.Errorf("service not found: %s: %w", serviceName, err)
	}

	serviceDesc, ok := desc.(protoreflect.ServiceDescriptor)
	if !ok {
		return nil, fmt.Errorf("%s is not a service", serviceName)
	}

	// Find method
	methods := serviceDesc.Methods()
	for i := 0; i < methods.Len(); i++ {
		method := methods.Get(i)
		if string(method.Name()) == methodName {
			return method, nil
		}
	}

	return nil, fmt.Errorf("method not found: %s in service %s", methodName, serviceName)
}

func mapToProtoMessage(data map[string]any, msgDesc protoreflect.MessageDescriptor) (proto.Message, error) {
	msg := dynamicpb.NewMessage(msgDesc)

	// Convert map to JSON first
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal map to JSON: %w", err)
	}

	// Unmarshal JSON to proto message
	err = protojson.Unmarshal(jsonBytes, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to proto: %w", err)
	}

	return msg, nil
}
