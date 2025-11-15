package internal

// GRPCRequestSpec defines a gRPC request specification
type GRPCRequestSpec struct {
	Service  string            `yaml:"service"`            // Service name, e.g., "UserService"
	Method   string            `yaml:"method"`             // Method name, e.g., "GetUser"
	Message  map[string]any    `yaml:"message"`            // Request message as map
	Metadata map[string]string `yaml:"metadata,omitempty"` // gRPC metadata (headers)
}

// GRPCResponseSpec defines expected gRPC response
type GRPCResponseSpec struct {
	Code       string              `yaml:"code"`                 // gRPC status code: "OK", "NOT_FOUND", etc.
	Message    string              `yaml:"message,omitempty"`    // Expected response as JSON for assertion
	Metadata   map[string]string   `yaml:"metadata,omitempty"`   // Expected metadata
	Assertions []ResponseAssertion `yaml:"assertions,omitempty"` // Enhanced assertions
}
