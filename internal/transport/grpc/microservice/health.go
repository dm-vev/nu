package microservice

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// testHealth tests the gRPC health endpoint
func (m *Service) testHealth() error {
	// Create a gRPC connection with a longer timeout for complex agent initialization
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create gRPC client for health check
	conn, err := grpc.NewClient(
		fmt.Sprintf("localhost:%d", m.port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		fmt.Printf("Debug: Failed to create gRPC connection to localhost:%d: %v\n", m.port, err)
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			// Log the close error but don't fail the whole operation
			fmt.Printf("Warning: failed to close gRPC connection: %v\n", closeErr)
		}
	}()

	// Test the standard gRPC health service
	healthClient := grpc_health_v1.NewHealthClient(conn)
	resp, err := healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{
		Service: "", // Check overall server health
	})

	if err != nil {
		fmt.Printf("Debug: Health check failed for localhost:%d: %v\n", m.port, err)
		return err
	}

	fmt.Printf("Debug: Health check succeeded for localhost:%d, status: %v\n", m.port, resp.Status)
	return nil
}
