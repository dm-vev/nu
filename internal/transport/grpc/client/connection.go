package client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"

	pb "nu/internal/transport/grpc/pb"
)

// Connect establishes a connection to the remote agent service
func (r *Client) Connect() error {
	if r.conn != nil {
		return nil // Already connected
	}

	conn, err := grpc.NewClient(r.url,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", r.url, err)
	}

	r.conn = conn
	r.client = pb.NewAgentServiceClient(conn)

	// Test the connection with standard gRPC health check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	healthClient := grpc_health_v1.NewHealthClient(conn)
	_, err = healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{
		Service: "", // Check the overall server health (empty string means overall server)
	})
	if err != nil {
		if closeErr := r.conn.Close(); closeErr != nil {
			// Log the close error but continue with the original error
			fmt.Printf("Warning: failed to close connection during cleanup: %v\n", closeErr)
		}
		r.conn = nil
		r.client = nil
		return fmt.Errorf("health check failed for %s: %w", r.url, err)
	}

	return nil
}

// Disconnect closes the connection to the remote agent service
func (r *Client) Disconnect() error {
	if r.conn != nil {
		err := r.conn.Close()
		r.conn = nil
		r.client = nil
		return err
	}
	return nil
}

// ensureConnected ensures that the client is connected to the remote service
func (r *Client) ensureConnected() error {
	if r.conn == nil || r.client == nil {
		return r.Connect()
	}
	return nil
}

// IsConnected returns true if the client is connected
func (r *Client) IsConnected() bool {
	return r.conn != nil && r.client != nil
}

// GetURL returns the URL of the remote agent
func (r *Client) GetURL() string {
	return r.url
}
