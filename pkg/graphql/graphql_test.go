package graphql

import "testing"

func TestGraphQLRequestFields(t *testing.T) {
	req := GraphQLRequest{
		OperationName: "MyOperation",
	}
	if req.OperationName != "MyOperation" {
		t.Fatalf("expected OperationName MyOperation, got %q", req.OperationName)
	}
}

func TestExtractOperationNameDefault(t *testing.T) {
	name, err := ExtractOperationName([]byte(`query MyQuery { field }`))
	if err != nil {
		t.Fatalf("ExtractOperationName returned error: %v", err)
	}
	if name != "" {
		t.Fatalf("expected empty operation name for default stub, got %q", name)
	}
}
