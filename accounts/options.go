package accounts

import "golang.org/x/time/rate"

type ServiceOption func(*ServiceImpl)

func WithTxRatelimiter(limiter *rate.Limiter) ServiceOption {
	return func(svc *ServiceImpl) {
		svc.txRateLimiter = limiter
	}
}
