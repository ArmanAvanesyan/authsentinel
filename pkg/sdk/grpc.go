package sdk

import (
	"context"

	"github.com/ArmanAvanesyan/authsentinel/pkg/token"
)

// IdentityFromGRPCContext extracts an identity principal from a gRPC context.
// TODO: implement metadata extraction logic.
func IdentityFromGRPCContext(ctx context.Context) (*token.Principal, error) {
	return nil, nil
}
