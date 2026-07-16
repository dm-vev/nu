package server

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"

	pb "nu/internal/transport/grpc/pb"
)

// Health returns the health status of the agent service
func (s *Server) Health(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	// Simple health check - if we can respond, we're healthy
	return &pb.HealthResponse{
		Status:  pb.HealthResponse_SERVING,
		Message: "Agent service is healthy",
	}, nil
}

// Ready returns the readiness status of the agent service
func (s *Server) Ready(ctx context.Context, req *pb.ReadinessRequest) (*pb.ReadinessResponse, error) {
	// Check if agent is properly initialized
	if s.agent == nil {
		return &pb.ReadinessResponse{
			Ready:   false,
			Message: "Agent is not initialized",
		}, nil
	}

	if s.agent.GetName() == "" {
		return &pb.ReadinessResponse{
			Ready:   false,
			Message: "Agent name is not set",
		}, nil
	}

	return &pb.ReadinessResponse{
		Ready:   true,
		Message: "Agent service is ready",
	}, nil
}

// Start starts the gRPC server on the specified port
func (s *Server) Start(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", port, err)
	}

	return s.StartWithListener(listener)
}

// StartWithListener starts the gRPC server with an existing listener
func (s *Server) StartWithListener(listener net.Listener) error {
	s.listener = listener
	s.server = grpc.NewServer()

	// Register the agent service
	pb.RegisterAgentServiceServer(s.server, s)

	// Register the standard gRPC health service
	grpc_health_v1.RegisterHealthServer(s.server, s.healthServer)

	// Set the health status to SERVING for the overall service and agent service
	s.healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	s.healthServer.SetServingStatus("AgentService", grpc_health_v1.HealthCheckResponse_SERVING)

	port := listener.Addr().(*net.TCPAddr).Port
	fmt.Printf("Agent server starting on port %d...\n", port)
	return s.server.Serve(listener)
}

// Stop stops the gRPC server
func (s *Server) Stop() {
	if s.healthServer != nil {
		// Set health status to NOT_SERVING before stopping
		s.healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
		s.healthServer.SetServingStatus("AgentService", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	}

	if s.server != nil {
		s.server.GracefulStop()
	}
}

// GetPort returns the port the server is listening on
func (s *Server) GetPort() int {
	if s.listener != nil {
		return s.listener.Addr().(*net.TCPAddr).Port
	}
	return 0
}
