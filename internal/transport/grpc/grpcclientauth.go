package grpc

// contextKey is a custom type for context keys
type grpcClientContextKey string

// JWTTokenKey is used for JWT token context propagation
const ClientJWTTokenKey grpcClientContextKey = "jwtToken"

// jwtContextKey matches starops-agent middleware exactly
type grpcClientJWTContextKey struct{}

// Use exact same variable name and type as starops-agent middleware
var ClientJWTTokenKeyStruct = grpcClientJWTContextKey{}
