package metrics

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

var (
	UsersTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hg_users_total",
		Help: "Total registered users.",
	})
	CommunitiesTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hg_communities_total",
		Help: "Total communities.",
	})
	PostsTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hg_posts_total",
		Help: "Total posts (including RSS).",
	})
	PostsUserTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hg_user_posts_total",
		Help: "Total user-authored posts (excluding RSS).",
	})
	PostsRSSTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hg_rss_posts_total",
		Help: "Total RSS-imported posts.",
	})
	CommentsTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hg_comments_total",
		Help: "Total comments.",
	})
	ChatMessagesTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hg_chat_messages_total",
		Help: "Total stored chat messages.",
	})
	OnlineUsersGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hg_online_users",
		Help: "Currently online users (heartbeat within TTL).",
	})
	NotificationsUnreadTotal = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "hg_notifications_unread_total",
		Help: "Total unread notifications across all users.",
	})

	PostsCreated = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "hg_posts_created_total",
		Help: "Posts created since process start.",
	}, []string{"kind"})
	CommentsCreated = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hg_comments_created_total",
		Help: "Comments created since process start.",
	})
	PostVotes = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hg_post_votes_total",
		Help: "Post upvotes since process start.",
	})
	CommentVotes = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hg_comment_votes_total",
		Help: "Comment upvotes since process start.",
	})
	ChatSent = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hg_chat_messages_sent_total",
		Help: "Chat messages sent since process start.",
	})
	UsersRegistered = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hg_users_registered_total",
		Help: "New users provisioned since process start.",
	})
	CommunitiesCreated = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hg_communities_created_total",
		Help: "Communities created since process start.",
	})
	BookmarksAdded = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hg_bookmarks_added_total",
		Help: "Bookmarks added since process start.",
	})
	UploadsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "hg_uploads_total",
		Help: "Files uploaded since process start.",
	})
)

func init() {
	prometheus.MustRegister(
		UsersTotal, CommunitiesTotal, PostsTotal, PostsUserTotal, PostsRSSTotal,
		CommentsTotal, ChatMessagesTotal, OnlineUsersGauge, NotificationsUnreadTotal,
		PostsCreated, CommentsCreated, PostVotes, CommentVotes, ChatSent,
		UsersRegistered, CommunitiesCreated, BookmarksAdded, UploadsTotal,
	)
}

type OnlineCounter interface {
	OnlineCount() int
}

type BusinessMetricsPoller struct {
	pool   *pgxpool.Pool
	online OnlineCounter
	log    zerolog.Logger
}

func NewBusinessMetricsPoller(pool *pgxpool.Pool, online OnlineCounter, log zerolog.Logger) *BusinessMetricsPoller {
	return &BusinessMetricsPoller{pool: pool, online: online, log: log}
}

func (p *BusinessMetricsPoller) Start(ctx context.Context, interval time.Duration) {
	go func() {
		p.poll(ctx)
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				p.poll(ctx)
			}
		}
	}()
}

func (p *BusinessMetricsPoller) poll(ctx context.Context) {
	queries := []struct {
		gauge prometheus.Gauge
		sql   string
	}{
		{UsersTotal, `SELECT COUNT(*) FROM users`},
		{CommunitiesTotal, `SELECT COUNT(*) FROM communities`},
		{PostsTotal, `SELECT COUNT(*) FROM posts`},
		{PostsUserTotal, `SELECT COUNT(*) FROM posts WHERE source_id IS NULL`},
		{PostsRSSTotal, `SELECT COUNT(*) FROM posts WHERE source_id IS NOT NULL`},
		{CommentsTotal, `SELECT COUNT(*) FROM comments`},
		{ChatMessagesTotal, `SELECT COUNT(*) FROM chat_messages`},
		{NotificationsUnreadTotal, `SELECT COUNT(*) FROM notifications WHERE read_at IS NULL`},
	}
	for _, q := range queries {
		var n int64
		qctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err := p.pool.QueryRow(qctx, q.sql).Scan(&n)
		cancel()
		if err != nil {
			p.log.Warn().Err(err).Str("sql", q.sql).Msg("business metrics: poll failed")
			continue
		}
		q.gauge.Set(float64(n))
	}
	if p.online != nil {
		OnlineUsersGauge.Set(float64(p.online.OnlineCount()))
	}
}
