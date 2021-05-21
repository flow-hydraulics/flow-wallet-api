package scripts

import (
	"context"

	"github.com/onflow/cadence"
	"github.com/onflow/flow-go-sdk/client"
)

// Service defines the API for script HTTP handlers.
type Service struct {
	fc *client.Client
}

// NewService initiates a new scripts service.
func NewService(fc *client.Client) *Service {
	return &Service{fc}
}

func (s *Service) Execute(ctx context.Context, script Script) (cadence.Value, error) {
	value, err := s.fc.ExecuteScriptAtLatestBlock(
		ctx,
		[]byte(script.Code),
		script.MustDecodeArgs(),
	)

	if err != nil {
		return cadence.Void{}, err
	}

	return value, err
}
