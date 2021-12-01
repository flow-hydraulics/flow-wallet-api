package transactions

import "go.uber.org/ratelimit"

type ServiceOption func(*Service)

func WithTxRatelimiter(limiter ratelimit.Limiter) ServiceOption {
	return func(svc *Service) {
		svc.txRateLimiter = limiter
	}
}
