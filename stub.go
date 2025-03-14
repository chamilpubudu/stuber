package stuber

import (
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
)

// Stub represents a gRPC service method and its associated data.
type Stub struct {
	ID      uuid.UUID   `json:"id"`      // The unique identifier of the stub.
	Service string      `json:"service"` // The name of the service.
	Method  string      `json:"method"`  // The name of the method.
	Headers InputHeader `json:"headers"` // The headers of the request.
	Input   InputData   `json:"input"`   // The input data of the request.
	Output  Output      `json:"output"`  // The output data of the response.
}

// Key returns the unique identifier of the stub.
func (s Stub) Key() uuid.UUID {
	return s.ID
}

// Left returns the service name of the stub.
func (s Stub) Left() string {
	return s.Service
}

// Right returns the method name of the stub.
func (s Stub) Right() string {
	return s.Method
}

// InputData represents the input data of a gRPC request.
type InputData struct {
	IgnoreArrayOrder bool                   `json:"ignoreArrayOrder,omitempty"` // Whether to ignore the order of arrays in the input data.
	Equals           map[string]interface{} `json:"equals"`                     // The data to match exactly.
	Contains         map[string]interface{} `json:"contains"`                   // The data to match partially.
	Matches          map[string]interface{} `json:"matches"`                    // The data to match using regular expressions.
}

// GetEquals returns the data to match exactly.
func (i InputData) GetEquals() map[string]interface{} {
	return i.Equals
}

// GetContains returns the data to match partially.
func (i InputData) GetContains() map[string]interface{} {
	return i.Contains
}

// GetMatches returns the data to match using regular expressions.
func (i InputData) GetMatches() map[string]interface{} {
	return i.Matches
}

// InputHeader represents the headers of a gRPC request.
type InputHeader struct {
	Equals   map[string]interface{} `json:"equals"`   // The headers to match exactly.
	Contains map[string]interface{} `json:"contains"` // The headers to match partially.
	Matches  map[string]interface{} `json:"matches"`  // The headers to match using regular expressions.
}

// GetEquals returns the headers to match exactly.
func (i InputHeader) GetEquals() map[string]interface{} {
	return i.Equals
}

// GetContains returns the headers to match partially.
func (i InputHeader) GetContains() map[string]interface{} {
	return i.Contains
}

// GetMatches returns the headers to match using regular expressions.
func (i InputHeader) GetMatches() map[string]interface{} {
	return i.Matches
}

// Len returns the total number of headers to match.
func (i InputHeader) Len() int {
	return len(i.Equals) + len(i.Matches) + len(i.Contains)
}

// Output represents the output data of a gRPC response.
type Output struct {
	Headers map[string]string      `json:"headers"`        // The headers of the response.
	Data    interface{}            `json:"data"`           // The data of the response.
	Error   string                 `json:"error"`          // The error message of the response.
	Code    *codes.Code            `json:"code,omitempty"` // The status code of the response.
}
