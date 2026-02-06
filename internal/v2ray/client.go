package v2ray

import (
	"context"
	"fmt"
	"log"
	"time"

	statsService "github.com/v2fly/v2ray-core/v5/app/stats/command"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client wraps the V2Ray StatsService client
type Client struct {
	conn      *grpc.ClientConn
	statsCli  statsService.StatsServiceClient
	timeout   time.Duration
}

// NewClient creates a new V2Ray gRPC client
func NewClient(addr string, timeout time.Duration) (*Client, error) {
	// We use "WithBlock" to ensure connection is ready, but with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial v2ray api at %s: %w", addr, err)
	}

	return &Client{
		conn:     conn,
		statsCli: statsService.NewStatsServiceClient(conn),
		timeout:  timeout,
	}, nil
}

// Close closes the underlying gRPC connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// GetUserStats returns uplink/downlink for a specific user email
func (c *Client) GetUserStats(email string) (int64, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Sing-box/V2Ray standard: user traffic is usually "user>>>[email]>>>traffic>>>uplink"
	// But using the StatsService QueryStats is safer and more standard.
	// Pattern for user: "user>>>[email]>>>traffic>>>uplink"
	
	uplinkName := fmt.Sprintf("user>>>%s>>>traffic>>>uplink", email)
	downlinkName := fmt.Sprintf("user>>>%s>>>traffic>>>downlink", email)

	// Fetch Uplink
	upResp, err := c.statsCli.GetStats(ctx, &statsService.GetStatsRequest{
		Name:   uplinkName,
		Reset_: false,
	})
	var upVal int64 = 0
	if err == nil && upResp.Stat != nil {
		upVal = upResp.Stat.Value
	}

	// Fetch Downlink
	downResp, err := c.statsCli.GetStats(ctx, &statsService.GetStatsRequest{
		Name:   downlinkName,
		Reset_: false,
	})
	var downVal int64 = 0
	if err == nil && downResp.Stat != nil {
		downVal = downResp.Stat.Value
	}

	return upVal, downVal, nil
}

// QueryAllStats fetches all available stats without filtering
func (c *Client) QueryAllStats() ([]*statsService.Stat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	// Query with empty pattern to get everything
	resp, err := c.statsCli.QueryStats(ctx, &statsService.QueryStatsRequest{
		Pattern: "",
		Reset_:  false,
	})
	if err != nil {
		return nil, err
	}
	return resp.Stat, nil
}

