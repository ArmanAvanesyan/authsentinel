package grpc

import (
	"github.com/ArmanAvanesyan/authsentinel/pkg/proxy"
)

// ToProxyRequest builds a proxy.Request from gRPC metadata and method.
// TODO: implement; extract headers from metadata, set GRPCService/GRPCMethod from fullMethod.
func ToProxyRequest(fullMethod string, headers map[string]string, body []byte) *proxy.Request {
	svc, method := ExtractMethod(fullMethod)
	return &proxy.Request{
		GRPCService: svc,
		GRPCMethod:  method,
		Headers:     headers,
		Body:        body,
	}
}
