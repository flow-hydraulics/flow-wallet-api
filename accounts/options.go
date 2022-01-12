package accounts

import "go.uber.org/ratelimit"

type ServiceOption func(*ServiceImpl)

func WithTxRatelimiter(limiter ratelimit.Limiter) ServiceOption {
	return func(svc *ServiceImpl) {
		svc.txRateLimiter = limiter
	}
}
