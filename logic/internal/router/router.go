package router

import (
	"context"
	"fmt"

	"github.com/english/backend/internal/ai"
	aiclient "github.com/english/backend/internal/ai/client"
	"github.com/english/backend/internal/auth"
	"github.com/english/backend/internal/cache"
	"github.com/english/backend/internal/config"
	"github.com/english/backend/internal/dictionary"
	"github.com/english/backend/internal/enricher"
	"github.com/english/backend/internal/gamification"
	"github.com/english/backend/internal/handler"
	"github.com/english/backend/internal/learning"
	"github.com/english/backend/internal/middleware"
	"github.com/english/backend/internal/profile"
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
		dictService := dictionary.NewService(dictRepo)
		dictHandler := dictionary.NewHandler(dictService)

		// Translation Repo
		translationRepo := translation.NewRepository(mongoClient.Database(cfg.MongoDB))
		_ = translationRepo.EnsureIndexes(ctx)

		// AI Tutor & Clients
		var rawAI ai.AIClient
		var errAI error

		if cfg.GroqAPIKey != "" {
			rawAI, errAI = aiclient.NewGroqClient(cfg.GroqAPIKey, "llama-3.3-70b-versatile")
		} else if cfg.HFToken != "" {
			rawAI, errAI = aiclient.NewHuggingFaceClient(cfg.HFToken, cfg.HFModel)
		} else if cfg.OpenRouterAPIKey != "" {
			rawAI, errAI = aiclient.NewOpenRouterClient(cfg.OpenRouterAPIKey, cfg.OpenRouterModel)
		} else if cfg.OllamaEndpoint != "" {
			rawAI, errAI = aiclient.NewOllamaClient(cfg.OllamaEndpoint, cfg.OllamaModel)
		} else if cfg.GeminiAPIKey != "" {
			rawAI, errAI = aiclient.NewGeminiClient(context.Background(), cfg.GeminiAPIKey, "gemini-2.0-flash")
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
		authService := auth.NewService(authRepo, cfg.JWTSecret, cfg.JWTExpMinutes, cfg.RefreshExpDays)
		authHandler := auth.NewHandler(authService, cfg.Env, cfg.RefreshExpDays)

		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/register", authHandler.Register)
			authRoutes.POST("/login", authHandler.Login)
			authRoutes.POST("/refresh", authHandler.Refresh)
			authRoutes.POST("/logout", authHandler.Logout)
		}

		// User
		userRepo := user.NewRepository(mongoClient.Database(cfg.MongoDB))
		_ = userRepo.EnsureIndexes(ctx)
		userService := user.NewService(authRepo, userRepo)
		userHandler := user.NewHandler(userService)

		userRoutes := v1.Group("/users")
		userRoutes.Use(middleware.RequireAuth(cfg.JWTSecret))
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
		wordSetRoutes.Use(middleware.RequireAuth(cfg.JWTSecret))
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

		if rawAI != nil && errAI == nil {
			aiCache := ai.NewMemoryCache()
			aiLimiter := ai.NewMemoryRateLimiter(cfg.AIRateLimit)
			aiService := ai.NewService(aiClientWrapped, aiCache, dictRepo, aiLimiter)
			aiHandler := ai.NewHandler(aiService, aiLimiter)
			ai.RegisterRoutes(v1, aiHandler, cfg.JWTSecret)
		}

		// Learning
		
		exerciseContentRepo := learning.NewExerciseContentRepository(mongoClient.Database(cfg.MongoDB))
		_ = exerciseContentRepo.EnsureIndexes(ctx)
		
		var exSvc learning.ExerciseContentService
		if rawAI != nil && errAI == nil {
			exSvc = learning.NewExerciseContentService(exerciseContentRepo, dictRepo, rawAI)
		}

		
		sessionEnrich := enricher.NewSessionEnricher(translationService, wordSetService)
		learnService := learning.NewService(learnRepo, dictRepo, userRepo, learning.NewSM2Scheduler(), gamiService, exSvc, sessionEnrich)
		learnHandler := learning.NewHandler(learnService)

		learnRoutes := v1.Group("/learning")
		learnRoutes.Use(middleware.RequireAuth(cfg.JWTSecret))
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

		writingService := writing.NewService(writingRepo, writingEvaluator, gamiService)
		writingHandler := writing.NewHandler(writingService)

		writingRoutes := v1.Group("/writing")
		writingRoutes.Use(middleware.RequireAuth(cfg.JWTSecret))
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
		profileService := profile.NewService(profileRepo, learnService, rawAI, gamiService) // rawAI could be nil, handled gracefully
		profileHandler := profile.NewHandler(profileService)

		profileRoutes := v1.Group("/profile")
		profileRoutes.Use(middleware.RequireAuth(cfg.JWTSecret))
		{
			profileRoutes.GET("", profileHandler.GetDashboard)
			profileRoutes.GET("/placement-test", profileHandler.GetPlacementTest)
			profileRoutes.POST("/placement-test", profileHandler.SubmitPlacementTest)
		}

		// Gamification Routes
		gamiHandler := gamification.NewHandler(gamiService)
		gamiRoutes := v1.Group("/gamification")
		gamiRoutes.Use(middleware.RequireAuth(cfg.JWTSecret))
		{
			gamiRoutes.GET("/chart", gamiHandler.GetChart)
		}

		// Unified Dashboard Routes
		dashRoutes := v1.Group("/dashboard")
		dashRoutes.Use(middleware.RequireAuth(cfg.JWTSecret))
		{
			dashRoutes.GET("", profileHandler.GetDashboard)
		}
	}

	return r
}
