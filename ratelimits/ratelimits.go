package ratelimits

import (
	"context"
	"golang.org/x/time/rate"
)

type ConsumerClientI interface {
	ExpensiveOperation(ctx context.Context) (interface{},error)
}

// Expose operations through a client which enforces rate limits
type ConsumerClient struct{
	Limiter *rate.Limiter
}

func NewConsumerClient(r rate.Limit, b int) ConsumerClientI{
	return &ConsumerClient{
		Limiter: rate.NewLimiter(r, b),
	}
}

func (consumerclient *ConsumerClient) ExpensiveOperation(ctx context.Context) (interface{},error){
	err := consumerclient.Limiter.Wait(ctx)
	if err != nil {
		return nil, err
	}
	return 1,nil
}

