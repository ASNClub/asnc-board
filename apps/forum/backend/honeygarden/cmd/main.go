package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rs/zerolog"

	adapterhttp "honeygarden/internal/adapter/http"
	"honeygarden/internal/adapter/meilisearch"
	natspkg "honeygarden/internal/adapter/nats"
	"honeygarden/internal/adapter/oauthclients"
	"honeygarden/internal/adapter/objectstore"
	"honeygarden/internal/adapter/postgres"
	wakapiadapter "honeygarden/internal/adapter/wakapi"
	"honeygarden/internal/config"
	"honeygarden/internal/domain"
	"honeygarden/internal/metrics"
	"honeygarden/internal/port"
	"honeygarden/internal/service"
	"honeygarden/internal/worker"
)

func main() {
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg := config.Load()

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("postgres: connect failed")
	}
	defer pool.Close()

	if err = pool.Ping(ctx); err != nil {
		log.Fatal().Err(err).Msg("postgres: ping failed")
	}

	secretBox, err := postgres.NewSecretBox(cfg.OAuthStateKey)
	if err != nil {
		log.Fatal().Err(err).Msg("app secret: init failed")
	}

	nc, err := nats.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatal().Err(err).Msg("nats: connect failed")
	}
	defer nc.Drain()

	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatal().Err(err).Msg("nats: jetstream init failed")
	}

	msClient, err := meilisearch.NewClient(cfg.MeilisearchURL, cfg.MeilisearchAPIKey)
	if err != nil {
		log.Fatal().Err(err).Msg("meilisearch: init failed")
	}

	store, err := objectstore.NewMinio(ctx, objectstore.Config{
		Endpoint:  cfg.S3Endpoint,
		Bucket:    cfg.S3Bucket,
		AccessKey: cfg.S3AccessKey,
		SecretKey: cfg.S3SecretKey,
		UseSSL:    cfg.S3UseSSL,
		Region:    cfg.S3Region,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("object store: init failed")
	}

	userRepo := postgres.NewUserRepo(pool)
	userFollowRepo := postgres.NewUserFollowRepo(pool)
	friendshipRepo := postgres.NewFriendshipRepo(pool)
	userBlockRepo := postgres.NewUserBlockRepo(pool)
	gitAccountRepo := postgres.NewGitAccountRepo(pool, secretBox)
	pinnedRepoRepo := postgres.NewPinnedRepoRepo(pool)

	communityRepo := postgres.NewCommunityRepo(pool)
	communityFollowRepo := postgres.NewCommunityFollowRepo(pool)
	starRepo := postgres.NewStarRepo(pool)
	banRepo := postgres.NewBanRepo(pool)
	modRepo := postgres.NewModeratorRepo(pool)
	communityAccessRepo := postgres.NewCommunityAccessRepo(pool)

	postRepo := postgres.NewPostRepo(pool)
	commentRepo := postgres.NewCommentRepo(pool)
	mediaRepo := postgres.NewMediaRepo(pool)
	userResolverRepo := postgres.NewUserResolverRepo(pool)

	notifRepo := postgres.NewNotificationRepo(pool)
	notifPrefRepo := postgres.NewNotifPrefRepo(pool)
	entityLookup := postgres.NewEntityLookupRepo(pool)
	bookmarkRepo := postgres.NewBookmarkRepo(pool)
	feedbackRepo := postgres.NewFeedbackRepo(pool)
	roadmapRepo := postgres.NewRoadmapRepo(pool)
	bannedWordRepo := postgres.NewBannedWordRepo(pool)

	activityRepo := postgres.NewActivityRepo(pool)
	sourceRepo := postgres.NewSourceRepo(pool)
	feedRepo := postgres.NewFeedRepo(pool)

	publisher := natspkg.NewPublisher(js)

	oauthProviders := map[domain.GitProvider]port.OAuthProvider{}
	gitClients := map[domain.GitProvider]port.GitProviderClient{}

	if cfg.GitHubClientID != "" {
		gh := &oauthclients.GitHub{
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			RedirectURL:  cfg.OAuthCallback + "/integrations/github/callback",
		}
		oauthProviders[domain.GitProviderGitHub] = gh
		gitClients[domain.GitProviderGitHub] = gh
	}
	if cfg.GitLabClientID != "" {
		gl := &oauthclients.GitLab{
			ClientID:     cfg.GitLabClientID,
			ClientSecret: cfg.GitLabClientSecret,
			RedirectURL:  cfg.OAuthCallback + "/integrations/gitlab/callback",
			BaseURL:      cfg.GitLabBaseURL,
		}
		oauthProviders[domain.GitProviderGitLab] = gl
		gitClients[domain.GitProviderGitLab] = gl
	}
	if cfg.CodebergClientID != "" {
		cb := &oauthclients.Codeberg{
			ClientID:     cfg.CodebergClientID,
			ClientSecret: cfg.CodebergClientSecret,
			RedirectURL:  cfg.OAuthCallback + "/integrations/codeberg/callback",
		}
		oauthProviders[domain.GitProviderCodeberg] = cb
		gitClients[domain.GitProviderCodeberg] = cb
	}

	userSvc := service.NewUserService(
		userRepo, userFollowRepo, friendshipRepo, userBlockRepo,
		pinnedRepoRepo, bannedWordRepo,
		publisher, log,
	)
	integrationSvc := service.NewIntegrationService(
		gitAccountRepo, pinnedRepoRepo,
		oauthProviders, gitClients,
		[]byte(cfg.OAuthStateKey),
		log,
	)
	communitySvc := service.NewCommunityService(
		communityRepo, communityFollowRepo, starRepo, banRepo, modRepo, userRepo, bannedWordRepo, publisher, log,
	)
	postSvc := service.NewPostService(
		postRepo, commentRepo, mediaRepo, communityAccessRepo, bookmarkRepo, userBlockRepo,
		publisher, userResolverRepo, userResolverRepo, sourceRepo, log,
	)
	notifSvc := service.NewNotificationService(notifRepo, notifPrefRepo, entityLookup, userRepo, postRepo, commentRepo, communityRepo, log)
	feedSvc := service.NewFeedService(sourceRepo, feedRepo, log)
	searchSvc := service.NewSearchService(msClient, log)

	onlineSvc := service.NewOnlineService()

	chatRepo := postgres.NewChatRepo(pool)
	chatSvc := service.NewChatService(chatRepo, userResolverRepo)
	feedbackSvc := service.NewFeedbackService(feedbackRepo, userResolverRepo)
	roadmapSvc := service.NewRoadmapService(roadmapRepo)
	bannedWordSvc := service.NewBannedWordService(bannedWordRepo)

	badgeRepo := postgres.NewBadgeRepo(pool)
	badgeStats := postgres.NewBadgeStatsRepo(pool)
	badgeSvc := service.NewBadgeService(badgeRepo, badgeStats, log)

	wakapiRepo := postgres.NewWakapiRepo(pool, secretBox)
	wakapiClient := wakapiadapter.New()
	wakapiSvc := service.NewWakapiService(wakapiRepo, wakapiClient, log)

	subscriber := natspkg.NewSubscriber(js, notifSvc, searchSvc, badgeSvc, log)
	stopSub, err := subscriber.Start(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("nats: subscriber start failed")
	}
	defer stopSub()

	rssWorker := worker.NewRSSWorker(sourceRepo, postRepo, cfg.RSSInterval, cfg.RSSOTimeout, log)
	stopRSS := rssWorker.Start(ctx)
	defer stopRSS()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(adapterhttp.Metrics())
	r.Use(adapterhttp.Logger(log))
	r.Use(adapterhttp.JWTMiddleware(cfg))
	r.Use(adapterhttp.LastSeenMiddleware(userRepo))

	auth := adapterhttp.RequireAuth(userSvc)
	optAuth := adapterhttp.OptionalAuth(userSvc)
	admin := adapterhttp.RequireAdmin(cfg.AdminAuthIDs)

	adapterhttp.NewUserHandler(userSvc, postSvc, activityRepo, cfg.AdminAuthIDs).Register(r, auth, optAuth)
	adapterhttp.NewCommunityHandler(communitySvc).Register(r, auth, optAuth)
	adapterhttp.NewPostHandler(postSvc).Register(r, auth, optAuth)
	adapterhttp.NewNotificationHandler(notifSvc).Register(r, auth)
	adapterhttp.NewFeedHandler(feedSvc, postSvc).Register(r, auth, optAuth, admin)
	adapterhttp.NewSearchHandler(searchSvc, userSvc).Register(r)
	adapterhttp.NewUploadHandler(store, cfg.S3PublicBase).Register(r, auth)
	adapterhttp.NewBadgeHandler(badgeSvc, userSvc).Register(r)
	adapterhttp.NewOnlineHandler(onlineSvc, userSvc).Register(r, auth)
	adapterhttp.NewChatHandler(chatSvc, cfg.FrontendURL, log).Register(r, auth)
	adapterhttp.NewWakapiHandler(wakapiSvc, userSvc).Register(r, auth)
	adapterhttp.NewIntegrationHandler(integrationSvc, cfg.FrontendURL).Register(r, auth)
	adapterhttp.NewFeedbackHandler(feedbackSvc).Register(r, auth, optAuth, admin)
	adapterhttp.NewRoadmapHandler(roadmapSvc).Register(r, auth, admin)
	adapterhttp.NewBannedWordHandler(bannedWordSvc).Register(r, auth, admin)
	adapterhttp.NewAdminHandler(userSvc, postSvc, communitySvc).Register(r, auth, admin)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/metrics", adapterhttp.MetricsHandler())

	metrics.NewBusinessMetricsPoller(pool, onlineSvc, log).Start(ctx, 30*time.Second)

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Info().Str("addr", srv.Addr).Msg("server started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server shutdown error")
	}
	log.Info().Msg("stopped")
}
