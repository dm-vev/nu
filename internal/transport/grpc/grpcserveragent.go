package grpc

import (
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"

	"nu/internal/agent"
	pb "nu/internal/transport/grpc/pb"
)

// AgentServer implements the gRPC AgentService.
type AgentServer struct {
	pb.UnimplementedAgentServiceServer
	agent        *agent.Agent
	server       *grpc.Server
	listener     net.Listener
	healthServer *health.Server
}

// NewAgentServer creates a gRPC server wrapping the provided agent.
func NewAgentServer(agent *agent.Agent) *AgentServer {
	return &AgentServer{
		agent:        agent,
		healthServer: health.NewServer(),
	}
}
