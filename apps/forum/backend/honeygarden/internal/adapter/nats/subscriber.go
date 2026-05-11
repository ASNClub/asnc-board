package nats

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog"
	"honeygarden/internal/service"
)

const streamName = "EVENTS"

type Subscriber struct {
	js      jetstream.JetStream
	notif   *service.NotificationService
	search  *service.SearchService
	badges  *service.BadgeService
	log     zerolog.Logger
}

func NewSubscriber(js jetstream.JetStream, notif *service.NotificationService, search *service.SearchService, badges *service.BadgeService, log zerolog.Logger) *Subscriber {
	return &Subscriber{js: js, notif: notif, search: search, badges: badges, log: log}
}

// Start создаёт JetStream стрим и запускает два consumerа
func (s *Subscriber) Start(ctx context.Context) (stop func(), err error) {
	_, err = s.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name: streamName,
		Subjects: []string{
			"user.*",
			"community.*",
			"post.*",
			"comment.*",
			"friendship.*",
			"mention.*",
		},
		Storage:  jetstream.FileStorage,
		MaxAge:   7 * 24 * time.Hour,
		Replicas: 1,
	})
	if err != nil {
		return nil, err
	}

	notifConsumer, err := s.js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Name:    "notification-service",
		Durable: "notification-service",
		FilterSubjects: []string{
			"user.followed",
			"community.followed",
			"community.starred",
			"comment.created",
			"friendship.requested",
			"friendship.accepted",
			"post.voted",
			"mention.created",
		},
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return nil, err
	}

	searchConsumer, err := s.js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Name:    "search-service",
		Durable: "search-service",
		FilterSubjects: []string{
			"post.created",
			"community.created",
		},
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return nil, err
	}

	ccNotif, err := notifConsumer.Consume(func(msg jetstream.Msg) {
		if err := s.notif.HandleEvent(ctx, msg.Subject(), msg.Data()); err != nil {
			if nakErr := msg.Nak(); nakErr != nil {
				s.log.Warn().Err(nakErr).Str("subject", msg.Subject()).Msg("notif: nak failed")
			}
			return
		}
		if err := msg.Ack(); err != nil {
			s.log.Warn().Err(err).Str("subject", msg.Subject()).Msg("notif: ack failed")
		}
	})
	if err != nil {
		return nil, err
	}

	badgeConsumer, err := s.js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Name:    "badge-service",
		Durable: "badge-service",
		FilterSubjects: []string{
			"post.created",
			"post.voted",
			"comment.created",
			"community.created",
			"community.followed",
			"community.starred",
			"user.followed",
		},
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return nil, err
	}

	ccBadge, err := badgeConsumer.Consume(func(msg jetstream.Msg) {
		var p map[string]any
		if err := json.Unmarshal(msg.Data(), &p); err != nil {
			_ = msg.Ack()
			return
		}
		var userID uuid.UUID
		for _, key := range []string{"author_id", "user_id", "follower_id", "owner_id"} {
			if raw, ok := p[key].(string); ok {
				if id, err := uuid.Parse(raw); err == nil {
					userID = id
					break
				}
			}
		}
		if userID != uuid.Nil {
			s.badges.CheckAndAward(ctx, userID)
		}
		_ = msg.Ack()
	})
	if err != nil {
		ccNotif.Stop()
		return nil, err
	}

	ccSearch, err := searchConsumer.Consume(func(msg jetstream.Msg) {
		if err := s.search.HandleEvent(ctx, msg.Subject(), msg.Data()); err != nil {
			if nakErr := msg.Nak(); nakErr != nil {
				s.log.Warn().Err(nakErr).Str("subject", msg.Subject()).Msg("search: nak failed")
			}
			return
		}
		if err := msg.Ack(); err != nil {
			s.log.Warn().Err(err).Str("subject", msg.Subject()).Msg("search: ack failed")
		}
	})
	if err != nil {
		ccNotif.Stop()
		return nil, err
	}

	s.log.Info().Msg("nats subscribers started")
	return func() {
		ccNotif.Stop()
		ccBadge.Stop()
		ccSearch.Stop()
	}, nil
}
