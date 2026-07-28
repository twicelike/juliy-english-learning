package router

import (
	"context"

	"github.com/english/backend/internal/ai"
	aiclient "github.com/english/backend/internal/ai/client"
	"github.com/english/backend/internal/auth"
	"github.com/english/backend/internal/config"
	"github.com/english/backend/internal/dictionary"
	"github.com/english/backend/internal/handler"
	"github.com/english/backend/internal/learning"
	"github.com/english/backend/internal/middleware"
	"github.com/english/backend/internal/profile"
	"github.com/english/backend/internal/user"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// Setup configures the Gin engine with all middleware and routes registered.
func Setup(cfg *config.Config, mongoClient *mongo.Client) *gin.Engine {
	r := gin.New() // use gin.New() (not gin.Default()) so we control all middleware

	// Global middleware — order matters
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())

	// API v1 group
	v1 := r.Group("/api/v1")
	{
		healthHandler := handler.NewHealthHandler(mongoClient)
		v1.GET("/health", healthHandler.Check)

		dictRepo := dictionary.NewRepository(mongoClient.Database("dictionary"))
		dictService := dictionary.NewService(dictRepo)
		dictHandler := dictionary.NewHandler(dictService)

		words := v1.Group("/words")
		{
			words.GET("/search", dictHandler.Search)
			words.GET("/detail/:word", dictHandler.GetDetail)
			words.GET("/random", dictHandler.GetRandom)
			words.GET("/word-of-the-day", dictHandler.GetWordOfTheDay)
		}

		// Auth
		authRepo := auth.NewRepository(mongoClient.Database(cfg.MongoDB))
		// We must ensure indexes
		_ = authRepo.EnsureIndexes(context.Background())
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
		_ = userRepo.EnsureIndexes(context.Background())
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

		// Learning
		learnRepo := learning.NewRepository(mongoClient.Database(cfg.MongoDB))
		_ = learnRepo.EnsureIndexes(context.Background())
		learnService := learning.NewService(learnRepo, dictRepo, userRepo, learning.NewSM2Scheduler())
		learnHandler := learning.NewHandler(learnService)

		learnRoutes := v1.Group("/learning")
		learnRoutes.Use(middleware.RequireAuth(cfg.JWTSecret))
		{
			learnRoutes.GET("/dashboard", learnHandler.GetDashboard)
			learnRoutes.GET("/session", learnHandler.GetSession)
			learnRoutes.POST("/progress", learnHandler.SubmitProgress)
		}
		// AI Tutor
		var rawAI ai.AIClient
		var errAI error

		if cfg.OpenRouterAPIKey != "" {
			rawAI, errAI = aiclient.NewOpenRouterClient(cfg.OpenRouterAPIKey, cfg.OpenRouterModel)
		} else if cfg.OllamaEndpoint != "" {
			rawAI, errAI = aiclient.NewOllamaClient(cfg.OllamaEndpoint, cfg.OllamaModel)
		} else if cfg.GroqAPIKey != "" {
			rawAI, errAI = aiclient.NewGroqClient(cfg.GroqAPIKey, "llama3-70b-8192")
		} else if cfg.GeminiAPIKey != "" {
			rawAI, errAI = aiclient.NewGeminiClient(context.Background(), cfg.GeminiAPIKey, "gemini-2.0-flash")
		}

		if rawAI != nil && errAI == nil {
			aiClientWrapped := aiclient.WithMetrics(aiclient.WithRetry(rawAI, 2))
			aiCache := ai.NewMemoryCache()
			aiLimiter := ai.NewMemoryRateLimiter(cfg.AIRateLimit)
			aiService := ai.NewService(aiClientWrapped, aiCache, aiLimiter)
			aiHandler := ai.NewHandler(aiService, aiLimiter)
			ai.RegisterRoutes(v1, aiHandler, cfg.JWTSecret)
		}

		// Profile (Adaptive Learning)
		profileRepo := profile.NewRepository(mongoClient.Database(cfg.MongoDB))
		_ = profileRepo.EnsureIndexes(context.Background())
		profileService := profile.NewService(profileRepo, learnService, rawAI) // rawAI could be nil, handled gracefully
		profileHandler := profile.NewHandler(profileService)

		profileRoutes := v1.Group("/profile")
		profileRoutes.Use(middleware.RequireAuth(cfg.JWTSecret))
		{
			profileRoutes.GET("", profileHandler.GetDashboard)
			profileRoutes.GET("/placement-test", profileHandler.GetPlacementTest)
			profileRoutes.POST("/placement-test", profileHandler.SubmitPlacementTest)
		}
	}

	return r
}
