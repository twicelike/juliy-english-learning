package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/english/backend/internal/ai"
	aiclient "github.com/english/backend/internal/ai/client"
	"github.com/english/backend/internal/auth"
	"github.com/english/backend/internal/cache"
	"github.com/english/backend/internal/config"
	"github.com/english/backend/internal/dictionary"
	"github.com/english/backend/internal/enricher"
	"github.com/english/backend/internal/evaluation"
	"github.com/english/backend/internal/events"
	"github.com/english/backend/internal/gamification"
	"github.com/english/backend/internal/handler"
	"github.com/english/backend/internal/learning"
	"github.com/english/backend/internal/mail"
	"github.com/english/backend/internal/match"
	"github.com/english/backend/internal/middleware"
	"github.com/english/backend/internal/profile"
	"github.com/english/backend/internal/speaking"
	"github.com/english/backend/internal/storage"
	"github.com/english/backend/internal/translation"
	"github.com/english/backend/internal/user"
	"github.com/english/backend/internal/wordset"
	"github.com/english/backend/internal/writing"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Setup configures the Gin engine with all middleware and routes registered.
func Setup(ctx context.Context, cfg *config.Config, mongoClient *mongo.Client, cacheSvc cache.Service) *gin.Engine {
	r := gin.New() // use gin.New() (not gin.Default()) so we control all middleware

	// Global middleware — order matters
	r.Use(middleware.RequestID())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.Compression())
	r.Use(middleware.Recovery())
	r.Use(middleware.Metrics())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// Health Endpoints
	healthHandler := handler.NewHealthHandler(mongoClient, cacheSvc, "1.0.0")
	r.GET("/live", healthHandler.Live)
	r.GET("/ready", healthHandler.Ready)
	r.GET("/health", healthHandler.Health)

	// Metrics Endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API v1 group
	v1 := r.Group("/api/v1")
	{

		dictRepo := dictionary.NewRepository(mongoClient.Database("dictionary"))
		_ = dictRepo.EnsureIndexes(ctx)
		dictService := dictionary.NewService(dictRepo)
		dictHandler := dictionary.NewHandler(dictService)

		// Translation Repo
		translationRepo := translation.NewRepository(mongoClient.Database(cfg.MongoDB))
		_ = translationRepo.EnsureIndexes(ctx)

		// AI Tutor & Clients
		var rawAI ai.AIClient
		var errAI error

		if cfg.GroqAPIKey != "" {
			rawAI, errAI = aiclient.NewGroqClient(cfg.GroqAPIKey, "qwen/qwen3.6-27b")
		} else if cfg.HFToken != "" {
			rawAI, errAI = aiclient.NewHuggingFaceClient(cfg.HFToken, cfg.HFModel)
		} else if cfg.OpenRouterAPIKey != "" {
			rawAI, errAI = aiclient.NewOpenRouterClient(cfg.OpenRouterAPIKey, cfg.OpenRouterModel)
		} else if cfg.OllamaEndpoint != "" {
			rawAI, errAI = aiclient.NewOllamaClient(cfg.OllamaEndpoint, cfg.OllamaModel)
		} else if cfg.GeminiAPIKey != "" {
			rawAI, errAI = aiclient.NewGeminiClient(context.Background(), cfg.GeminiAPIKey, "gemini-3.5-flash")
		}

		var translationService translation.Service
		var aiClientWrapped ai.AIClient
		if rawAI != nil && errAI == nil {
			aiClientWrapped = aiclient.WithMetrics(aiclient.WithRetry(rawAI, 2))
			translationService = translation.NewService(aiClientWrapped, translationRepo)

			// Initialize and start Translation Worker
			translationWorker := translation.NewWorker(
				translationRepo,
				translationService,
				dictService,
				cfg.TranslationBatchSize,
				cfg.TranslationPollInterval,
				cfg.TranslationWorkerConcur,
				cfg.TranslationMaxRetries,
			)
			translationWorker.Start(ctx)
		}

		translationHandler := translation.NewHandler(translationRepo)



		words := v1.Group("/words")
		{
			words.GET("/search", dictHandler.Search)
			words.GET("/detail/:word", dictHandler.GetDetail)
			// Apply rate limiter to random word to prevent spam (60 reqs/min)
			words.GET("/random", middleware.IPRateLimiter(60), dictHandler.GetRandom)
			words.GET("/word-of-the-day", dictHandler.GetWordOfTheDay)
		}

		// Auth
		authRepo := auth.NewRepository(mongoClient.Database(cfg.MongoDB))
		// We must ensure indexes
		_ = authRepo.EnsureIndexes(ctx)
		mailerService := mail.NewSMTPMailer(cfg)
		authService := auth.NewService(authRepo, mailerService, cfg.RefreshExpDays, cfg.FrontendURL)
		oauthService := auth.NewGoogleOAuthService(cfg)
		authHandler := auth.NewHandler(authService, oauthService, cfg.Env, cfg.RefreshExpDays, cfg.FrontendURL)

		// Auth Extractors for Rate Limiting
		extractEmailFromJSON := func(c *gin.Context) string {
			ip := c.ClientIP()

			// Try to read email from JSON body without breaking later bindings
			var email string
			if c.Request.Body != nil {
				bodyBytes, err := io.ReadAll(c.Request.Body)
				if err == nil && len(bodyBytes) > 0 {
					// Restore the body for next handlers
					c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

					var payload struct {
						Email string `json:"email"`
					}
					if err := json.Unmarshal(bodyBytes, &payload); err == nil && payload.Email != "" {
						email = payload.Email
					}
				}
			}

			if email != "" {
				return ip + ":" + email
			}
			return ip
		}

		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/register", middleware.IPRateLimiter(5), authHandler.Register)
			authRoutes.POST("/login", middleware.KeyRateLimiter(5, extractEmailFromJSON), authHandler.Login)
			authRoutes.POST("/logout", authHandler.Logout)
			authRoutes.GET("/google", authHandler.GoogleLogin)
			authRoutes.GET("/google/callback", authHandler.GoogleCallback)

			authRoutes.POST("/verify-email", authHandler.VerifyEmail)
			authRoutes.POST("/resend-verification", middleware.KeyRateLimiter(3, extractEmailFromJSON), authHandler.ResendVerification)
			authRoutes.POST("/forgot-password", middleware.KeyRateLimiter(3, extractEmailFromJSON), authHandler.ForgotPassword)
			authRoutes.POST("/reset-password", authHandler.ResetPassword)

			authProtected := authRoutes.Group("")
			authProtected.Use(middleware.RequireAuth(authRepo))
			{
				authProtected.GET("/me", authHandler.GetMe)
				authProtected.POST("/logout-all", authHandler.LogoutAll)
				authProtected.GET("/sessions", authHandler.GetSessions)
				authProtected.DELETE("/sessions/:id", authHandler.RevokeSession)
			}
		}

		adminRoutes := v1.Group("/admin")
		adminRoutes.Use(middleware.RequireAdmin(authRepo))
		{
			adminRoutes.GET("/translation-jobs/stats", translationHandler.GetJobStats)
			adminRoutes.POST("/translation-jobs/retry-failed", translationHandler.RetryFailedJobs)

			adminRoutes.GET("/users", authHandler.AdminListUsers)
			adminRoutes.PATCH("/users/:id", authHandler.AdminUpdateUser)
		}

		// User
		userRepo := user.NewRepository(mongoClient.Database(cfg.MongoDB))
		_ = userRepo.EnsureIndexes(ctx)
		userService := user.NewService(authRepo, userRepo)
		userHandler := user.NewHandler(userService)

		userRoutes := v1.Group("/users")
		userRoutes.Use(middleware.RequireAuth(authRepo))
		{
			userRoutes.GET("/me", userHandler.GetProfile)
			userRoutes.GET("/bookmarks", userHandler.ListBookmarks)
			userRoutes.POST("/bookmarks", userHandler.AddBookmark)
			userRoutes.DELETE("/bookmarks/:word", userHandler.RemoveBookmark)
		}

		// Gamification
		gamiRepo := gamification.NewRepository(mongoClient.Database(cfg.MongoDB))
		_ = gamiRepo.EnsureIndexes(ctx)
		gamiService := gamification.NewService(gamiRepo)

		// Word Set (WS1 & WS2 & WS6)
		wordSetRepo := wordset.NewRepository(mongoClient.Database(cfg.MongoDB))
		if err := wordSetRepo.EnsureIndexes(ctx); err != nil {
			panic(fmt.Errorf("ensure word set indexes: %w", err))
		}

		// Learn Repo (needed for progress checking in WS6)
		learnRepo := learning.NewRepository(mongoClient.Database(cfg.MongoDB))
		_ = learnRepo.EnsureIndexes(ctx)

		wordSetService := wordset.NewService(wordSetRepo, authRepo, dictService, learnRepo, translationService)
		wordSetHandler := wordset.NewHandler(wordSetService)

		wordSetRoutes := v1.Group("/word-sets")
		wordSetRoutes.Use(middleware.RequireAuth(authRepo))
		{
			wordSetRoutes.POST("", wordSetHandler.Create)
			wordSetRoutes.GET("", wordSetHandler.List)
			wordSetRoutes.GET("/explore", wordSetHandler.Explore)
			wordSetRoutes.GET("/topics", wordSetHandler.GetTopics)
			wordSetRoutes.GET("/languages", wordSetHandler.GetSupportedLanguages)
			wordSetRoutes.GET("/:id", wordSetHandler.Get)
			wordSetRoutes.PATCH("/:id", wordSetHandler.Update)
			wordSetRoutes.DELETE("/:id", wordSetHandler.Delete)
			wordSetRoutes.POST("/:id/clone", wordSetHandler.Clone)

			wordSetRoutes.GET("/:id/words", wordSetHandler.GetWords)
			wordSetRoutes.POST("/:id/words", wordSetHandler.AddWord)
			wordSetRoutes.POST("/:id/words/bulk", wordSetHandler.BulkAddWords)
			wordSetRoutes.POST("/:id/words/restore", wordSetHandler.RestoreWords)
			wordSetRoutes.DELETE("/:id/words", wordSetHandler.BatchDeleteWords)
			wordSetRoutes.DELETE("/:id/words/:wordId", wordSetHandler.RemoveWord)
			wordSetRoutes.PATCH("/:id/words/reorder", wordSetHandler.ReorderWords)
			wordSetRoutes.PATCH("/:id/words/:wordId/annotation", wordSetHandler.UpdateAnnotation)
		}

		// Evaluation
		evalRepo := evaluation.NewRepository(mongoClient.Database(cfg.MongoDB))
		_ = evalRepo.EnsureIndexes(ctx)
		evalService := evaluation.NewService(evalRepo)

		// Event Publisher
		publisher := events.NewInProcessPublisher(100)

		if rawAI != nil && errAI == nil {
			aiCache := ai.NewMemoryCache()
			aiLimiter := ai.NewMemoryRateLimiter(cfg.AIRateLimit)
			aiRepo := ai.NewRepository(mongoClient.Database(cfg.MongoDB))
			aiService := ai.NewService(aiClientWrapped, aiCache, aiRepo, dictRepo, aiLimiter)
			aiHandler := ai.NewHandler(aiService, aiLimiter)
			ai.RegisterRoutes(v1, aiHandler, authRepo)

			// Speaking Domain
			speakingRepo := speaking.NewRepository(mongoClient.Database(cfg.MongoDB))
			storageSvc, err := storage.NewLocalStorage("uploads")
			if err != nil {
				panic(fmt.Errorf("failed to initialize local storage: %w", err))
			}
			processorSvc := speaking.NewProcessingService(speakingRepo, aiService, storageSvc, dictRepo, gamiService, evalService, publisher)
			speakingService := speaking.NewService(speakingRepo, aiService, storageSvc, processorSvc)
			recService := speaking.NewRecommendationService(speakingRepo, speakingService)
			speakingHandler := speaking.NewHandler(speakingService, recService)
			speakingRoutes := v1.Group("")
			speakingRoutes.Use(middleware.RequireAuth(authRepo))
			speakingHandler.RegisterRoutes(speakingRoutes)
		}

		// Learning

		exerciseContentRepo := learning.NewExerciseContentRepository(mongoClient.Database(cfg.MongoDB))
		_ = exerciseContentRepo.EnsureIndexes(ctx)

		var exSvc learning.ExerciseContentService
		if rawAI != nil && errAI == nil {
			exSvc = learning.NewExerciseContentService(exerciseContentRepo, dictRepo, rawAI)
		}

		sessionEnrich := enricher.NewSessionEnricher(translationService, wordSetService)
		learnService := learning.NewService(learnRepo, dictRepo, userRepo, learning.NewSM2Scheduler(), gamiService, exSvc, sessionEnrich, evalService, publisher)
		learnHandler := learning.NewHandler(learnService)

		learnRoutes := v1.Group("/learning")
		learnRoutes.Use(middleware.RequireAuth(authRepo))
		{
			learnRoutes.GET("/dashboard", learnHandler.GetDashboard)
			learnRoutes.GET("/session", learnHandler.GetSession)
			learnRoutes.POST("/session", learnHandler.CreateSession)
			learnRoutes.GET("/session/:id", learnHandler.GetSessionByID)
			learnRoutes.POST("/sessions/:id/submit", learnHandler.SubmitSessionResults)
			learnRoutes.PUT("/progress/suspend/:word_id", learnHandler.SuspendWord)
			learnRoutes.POST("/debug/advance-time", learnHandler.DebugAdvanceTime)
		}

		// Writing Missions (Phase 6.1 & 6.2)
		writingRepo := writing.NewRepository(mongoClient.Database(cfg.MongoDB))
		_ = writingRepo.EnsureIndexes(ctx)

		var writingEvaluator *writing.GeneratorService
		if rawAI != nil && errAI == nil {
			writingEvaluator = writing.NewGeneratorService(rawAI)
			// Auto refill job is disabled as per user request.
			// AI is only used for evaluating answers.
		}

		writingService := writing.NewService(writingRepo, writingEvaluator, gamiService, evalService, publisher)
		writingHandler := writing.NewHandler(writingService)

		writingRoutes := v1.Group("/writing")
		writingRoutes.Use(middleware.RequireAuth(authRepo))
		{
			writingRoutes.GET("/dashboard", writingHandler.GetDashboardStats)

			missions := writingRoutes.Group("/missions")
			{
				missions.GET("/new", writingHandler.GetNewMission)
				missions.GET("/review", writingHandler.GetReviewMission)
				missions.POST("/:id/submit", writingHandler.SubmitMission)
			}
		}

		// Profile (Adaptive Learning)
		profileRepo := profile.NewRepository(mongoClient.Database(cfg.MongoDB))
		_ = profileRepo.EnsureIndexes(ctx)
		profileService := profile.NewService(profileRepo, learnService, rawAI, gamiService, evalService) // rawAI could be nil, handled gracefully
		profileHandler := profile.NewHandler(profileService, publisher)

		// Set profile handler to publisher and start worker
		publisher.SetProfileHandler(profileService)
		publisher.StartWorker(ctx)

		profileRoutes := v1.Group("/profile")
		profileRoutes.Use(middleware.RequireAuth(authRepo))
		{
			profileRoutes.GET("", profileHandler.GetDashboard)
			profileRoutes.GET("/placement-test", profileHandler.GetPlacementTest)
			profileRoutes.POST("/placement-test", profileHandler.SubmitPlacementTest)
			profileRoutes.PATCH("/preferences", profileHandler.UpdatePreferences)
		}

		// Also register the manual refresh endpoint under /users/me
		userRoutes.POST("/me/refresh-skills", profileHandler.RequestRefreshSkills())

		// Gamification Routes
		gamiHandler := gamification.NewHandler(gamiService)
		gamiRoutes := v1.Group("/gamification")
		gamiRoutes.Use(middleware.RequireAuth(authRepo))
		{
			gamiRoutes.GET("/chart", gamiHandler.GetChart)
		}

		// Unified Dashboard Routes
		dashRoutes := v1.Group("/dashboard")
		dashRoutes.Use(middleware.RequireAuth(authRepo))
		{
			dashRoutes.GET("", profileHandler.GetDashboard)
		}

		// Match (MVP)
		matchRepo := match.NewRepository(mongoClient.Database(cfg.MongoDB))
		_ = matchRepo.EnsureIndexes(ctx)
		matchService := match.NewService(matchRepo, dictRepo, learnRepo, userService, translationRepo)
		matchHandler := match.NewHandler(matchService)

		matchRoutes := v1.Group("/match")
		matchRoutes.Use(middleware.RequireAuth(authRepo))
		{
			matchRoutes.GET("/recommendations", matchHandler.GetRecommendations)
			matchRoutes.POST("/swipe", matchHandler.Swipe)
		}
	}

	return r
}
