package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"honeygarden/internal/domain"
	"honeygarden/internal/port"
)

type WakapiService struct {
	repo   port.WakapiRepository
	client port.WakapiClient
	log    zerolog.Logger
}

func NewWakapiService(repo port.WakapiRepository, client port.WakapiClient, log zerolog.Logger) *WakapiService {
	return &WakapiService{repo: repo, client: client, log: log}
}

type ConnectWakapiInput struct {
	InstanceURL string `json:"instanceUrl" binding:"required"`
	APIKey      string `json:"apiKey"      binding:"required"`
	Username    string `json:"username"    binding:"required"`
}

func (s *WakapiService) Connect(ctx context.Context, userID uuid.UUID, in ConnectWakapiInput) error {
	if _, err := s.client.FetchStats(ctx, in.InstanceURL, in.APIKey, in.Username); err != nil {
		s.log.Warn().Err(err).
			Str("instance_url", in.InstanceURL).
			Str("username", in.Username).
			Msg("wakapi connect: verification failed")
		return fmt.Errorf("wakapi: %w (%v)", domain.ErrInvalidInput, err)
	}

	return s.repo.Save(ctx, domain.WakapiAccount{
		UserID:      userID,
		InstanceURL: in.InstanceURL,
		APIKey:      in.APIKey,
		Username:    in.Username,
	})
}

func (s *WakapiService) Disconnect(ctx context.Context, userID uuid.UUID) error {
	return s.repo.Delete(ctx, userID)
}

func (s *WakapiService) GetStats(ctx context.Context, userID uuid.UUID) (domain.WakapiStats, error) {
	acc, err := s.repo.Get(ctx, userID)
	if err != nil {
		return domain.WakapiStats{}, err
	}
	return s.client.FetchStats(ctx, acc.InstanceURL, acc.APIKey, acc.Username)
}

func (s *WakapiService) IsConnected(ctx context.Context, userID uuid.UUID) bool {
	_, err := s.repo.Get(ctx, userID)
	return err == nil
}
