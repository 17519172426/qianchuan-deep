package grpc

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/example/qianchuan-saas/grpc/strategy"
)

type Client struct {
	conn *grpc.ClientConn
	Stub pb.StrategyServiceClient
}

func NewClient(addr string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, Stub: pb.NewStrategyServiceClient(conn)}, nil
}

func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Client) EvaluateRules(ctx context.Context, rules []*pb.RuleDef, ads []*pb.AdContext) ([]*pb.RuleAction, error) {
	resp, err := c.Stub.EvaluateRules(ctx, &pb.EvaluateRequest{Rules: rules, Ads: ads})
	if err != nil {
		log.Printf("gRPC EvaluateRules failed: %v", err)
		return nil, err
	}
	return resp.Actions, nil
}

func (c *Client) TestRule(ctx context.Context, rule *pb.RuleDef, ads []*pb.AdContext) (*pb.TestRuleResponse, error) {
	resp, err := c.Stub.TestRule(ctx, &pb.TestRuleRequest{Rule: rule, Ads: ads})
	if err != nil {
		log.Printf("gRPC TestRule failed: %v", err)
		return nil, err
	}
	return resp, nil
}

func (c *Client) DetectAnomalies(ctx context.Context, current []*pb.AdMetrics, history []*pb.MetricsWindow) ([]*pb.Anomaly, error) {
	resp, err := c.Stub.DetectAnomalies(ctx, &pb.AnomalyRequest{Current: current, History: history})
	if err != nil {
		log.Printf("gRPC DetectAnomalies failed: %v", err)
		return nil, err
	}
	return resp.Anomalies, nil
}

func (c *Client) PredictROI(ctx context.Context, adID int64, history []*pb.AdMetrics) (*pb.PredictResponse, error) {
	resp, err := c.Stub.PredictROI(ctx, &pb.PredictRequest{AdId: adID, History_7D: history})
	if err != nil {
		log.Printf("gRPC PredictROI failed: %v", err)
		return nil, err
	}
	return resp, nil
}

func (c *Client) GenerateRecommendations(ctx context.Context, adIDs []int64, current []*pb.AdMetrics, history []*pb.MetricsWindow) ([]*pb.Recommendation, error) {
	resp, err := c.Stub.GenerateRecommendations(ctx, &pb.RecRequest{
		AdIds:          adIDs,
		CurrentMetrics: current,
		History_7D:     history,
	})
	if err != nil {
		log.Printf("gRPC GenerateRecommendations failed: %v", err)
		return nil, err
	}
	return resp.Recommendations, nil
}
