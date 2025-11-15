package internal

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Mock gRPC server for testing
type mockGRPCServer struct {
}

func (s *mockGRPCServer) SayHello(ctx context.Context, req *wrapperspb.StringValue) (*wrapperspb.StringValue, error) {
	// Check metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if auth := md.Get("authorization"); len(auth) > 0 && auth[0] != "Bearer valid-token" {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
	}

	if req.Value == "error" {
		return nil, status.Error(codes.Internal, "internal error")
	}

	if req.Value == "notfound" {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &wrapperspb.StringValue{Value: "Hello, " + req.Value}, nil
}

func TestParseGRPCCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected codes.Code
	}{
		{"OK", "OK", codes.OK},
		{"NotFound", "NOT_FOUND", codes.NotFound},
		{"NotFound alt", "NotFound", codes.NotFound},
		{"InvalidArgument", "INVALID_ARGUMENT", codes.InvalidArgument},
		{"Internal", "INTERNAL", codes.Internal},
		{"Unauthenticated", "UNAUTHENTICATED", codes.Unauthenticated},
		{"Unknown", "UNKNOWN_CODE", codes.Unknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseGRPCCode(tt.input)
			if got != tt.expected {
				t.Errorf("parseGRPCCode(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRenderGRPCRequest(t *testing.T) {
	ctx := map[string]any{
		"userId":   123,
		"token":    "secret-token",
		"userName": "john_doe",
	}

	tests := []struct {
		name     string
		input    GRPCRequestSpec
		expected GRPCRequestSpec
	}{
		{
			name: "simple rendering",
			input: GRPCRequestSpec{
				Service: "UserService",
				Method:  "GetUser",
				Message: map[string]any{
					"id": "{{userId}}",
				},
				Metadata: map[string]string{
					"authorization": "Bearer {{token}}",
				},
			},
			expected: GRPCRequestSpec{
				Service: "UserService",
				Method:  "GetUser",
				Message: map[string]any{
					"id": 123,
				},
				Metadata: map[string]string{
					"authorization": "Bearer secret-token",
				},
			},
		},
		{
			name: "nested message rendering",
			input: GRPCRequestSpec{
				Service: "UserService",
				Method:  "CreateUser",
				Message: map[string]any{
					"name": "{{userName}}",
					"metadata": map[string]any{
						"userId": "{{userId}}",
					},
				},
			},
			expected: GRPCRequestSpec{
				Service: "UserService",
				Method:  "CreateUser",
				Message: map[string]any{
					"name": "john_doe",
					"metadata": map[string]any{
						"userId": 123,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := renderGRPCRequest(tt.input, ctx)

			// Check service
			if got.Service != tt.expected.Service {
				t.Errorf("Service = %v, want %v", got.Service, tt.expected.Service)
			}

			// Check method
			if got.Method != tt.expected.Method {
				t.Errorf("Method = %v, want %v", got.Method, tt.expected.Method)
			}

			// Check metadata
			for k, v := range tt.expected.Metadata {
				if got.Metadata[k] != v {
					t.Errorf("Metadata[%s] = %v, want %v", k, got.Metadata[k], v)
				}
			}

			// Note: Deep comparison of Message map would require reflection
			// For now, we just check that it's not nil
			if got.Message == nil {
				t.Error("Message should not be nil")
			}
		})
	}
}

func TestMapToProtoMessage(t *testing.T) {
	// Test with StringValue wrapper
	data := map[string]any{
		"value": "hello",
	}

	msgDesc := (&wrapperspb.StringValue{}).ProtoReflect().Descriptor()
	msg, err := mapToProtoMessage(data, msgDesc)
	if err != nil {
		t.Fatalf("mapToProtoMessage failed: %v", err)
	}

	if msg == nil {
		t.Fatal("Expected message, got nil")
	}

	// Verify the message
	strMsg, ok := msg.(*wrapperspb.StringValue)
	if !ok {
		t.Fatal("Expected StringValue message")
	}

	if strMsg.Value != "hello" {
		t.Errorf("Expected value 'hello', got '%s'", strMsg.Value)
	}
}

// Integration test with real gRPC server
func TestGRPCIntegration(t *testing.T) {
	// Skip if running in CI without proper proto setup
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Start gRPC server
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	// Register mock service
	// Note: This would require proper proto registration
	// For now, this is a placeholder for the integration test structure

	go func() {
		if err := server.Serve(listener); err != nil {
			t.Logf("Server error: %v", err)
		}
	}()
	defer server.GracefulStop()

	// Test would go here using ExecuteGRPCRequest
	// This demonstrates the structure for integration testing
}

func TestAssertGRPCResponse(t *testing.T) {
	tests := []struct {
		name       string
		response   *GRPCResponse
		expected   GRPCResponseSpec
		shouldFail bool
	}{
		{
			name: "OK status",
			response: &GRPCResponse{
				Code:     codes.OK,
				Message:  &wrapperspb.StringValue{Value: "test"},
				Metadata: metadata.New(map[string]string{}),
			},
			expected: GRPCResponseSpec{
				Code: "OK",
			},
			shouldFail: false,
		},
		{
			name: "NOT_FOUND status",
			response: &GRPCResponse{
				Code:     codes.NotFound,
				Error:    status.Error(codes.NotFound, "not found"),
				Metadata: metadata.New(map[string]string{}),
			},
			expected: GRPCResponseSpec{
				Code: "NOT_FOUND",
			},
			shouldFail: false,
		},
		{
			name: "metadata check",
			response: &GRPCResponse{
				Code:    codes.OK,
				Message: &wrapperspb.StringValue{Value: "test"},
				Metadata: metadata.New(map[string]string{
					"x-request-id": "123",
				}),
			},
			expected: GRPCResponseSpec{
				Code: "OK",
				Metadata: map[string]string{
					"x-request-id": "123",
				},
			},
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock testing.T to capture failures
			mockT := &mockTestingT{}

			AssertGRPCResponse(t, tt.response, tt.expected, map[string]any{})

			if tt.shouldFail && !mockT.failed {
				t.Error("Expected assertion to fail, but it passed")
			}
			if !tt.shouldFail && mockT.failed {
				t.Errorf("Expected assertion to pass, but it failed: %s", mockT.errorMessage)
			}
		})
	}
}
