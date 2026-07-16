package server

import (
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"

	"github.com/dm-vev/nu/agent"
	"github.com/dm-vev/nu/internal/transport/grpc/pb"
)

// Server implements the gRPC AgentService.
type Server struct {
	pb.UnimplementedAgentServiceServer
	agent        *agent.Agent
	server       *grpc.Server
	listener     net.Listener
	healthServer *health.Server
}

// New creates a gRPC server wrapping the provided agent.
func New(agent *agent.Agent) *Server {
	return &Server{
		agent:        agent,
		healthServer: health.NewServer(),
	}
}
