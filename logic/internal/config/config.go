package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application configuration loaded from environment variables.
type Config struct {
	Port           string
	MongoURI       string
	MongoDB        string
	Env            string
	CORSOrigins    string
	JWTSecret      string
	JWTExpMinutes  int
	RefreshExpDays int

	// OAuth
	GoogleClientID     string
	GoogleClientSecret string
	GoogleCallbackURL  string
	FrontendURL        string

	// SMTP Mailer
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string

	// AI
	GeminiAPIKey     string
	GroqAPIKey       string
	OllamaEndpoint   string
	OllamaModel      string
	OpenRouterAPIKey string
	OpenRouterModel  string
	HFToken          string
	HFModel          string
	AIRateLimit      int // max requests per hour per user
	AICacheTTLHours  int // cache TTL for global AI responses

	// Translation Worker
	TranslationBatchSize    int
	TranslationWorkerConcur int
	TranslationMaxRetries   int
	TranslationPollInterval time.Duration

	// Redis & Cache
	RedisURI               string
	CacheTTLDictionaryMin  int
	CacheTTLDashboardMin   int
	CacheTTLLeaderboardMin int
	CacheTTLAIDays         int
}

// Load reads configuration from a .env file (if present) then falls back to OS
// environment variables. Returns an error if any required variable is missing.
func Load() (*Config, error) {
	// .env is optional — production environments typically use real env vars.
	_ = godotenv.Load(".env", "../.env", "../../.env")

	jwtExp, _ := strconv.Atoi(getEnv("JWT_EXPIRATION_MINUTES", "15"))
	refreshExp, _ := strconv.Atoi(getEnv("REFRESH_TOKEN_EXPIRATION_DAYS", "7"))

	aiRateLimit, _ := strconv.Atoi(getEnv("AI_RATE_LIMIT_PER_HOUR", "20"))
	aiCacheTTL, _ := strconv.Atoi(getEnv("AI_CACHE_TTL_HOURS", "24"))

	transBatch, _ := strconv.Atoi(getEnv("TRANSLATION_BATCH_SIZE", "10"))
	transConc, _ := strconv.Atoi(getEnv("TRANSLATION_WORKER_CONCURRENCY", "1"))
	transRetries, _ := strconv.Atoi(getEnv("TRANSLATION_MAX_RETRIES", "3"))
	transPoll, _ := strconv.Atoi(getEnv("TRANSLATION_POLL_INTERVAL_SEC", "2"))

	ttlDict, _ := strconv.Atoi(getEnv("CACHE_TTL_DICTIONARY_MINUTES", "1440")) // 24h
	ttlDash, _ := strconv.Atoi(getEnv("CACHE_TTL_DASHBOARD_MINUTES", "5"))
	ttlLead, _ := strconv.Atoi(getEnv("CACHE_TTL_LEADERBOARD_MINUTES", "2"))
	ttlAI, _ := strconv.Atoi(getEnv("CACHE_TTL_AI_DAYS", "7"))

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	cfg := &Config{
		Port:                    getEnv("PORT", "8080"),
		MongoURI:                getEnv("MONGO_URI", ""),
		MongoDB:                 getEnv("MONGO_DB", ""),
		Env:                     getEnv("ENV", "development"),
		CORSOrigins:             getEnv("CORS_ORIGINS", "*"),
		JWTSecret:               getEnv("JWT_SECRET", "super-secret-jwt-key-replace-me-in-prod"),
		JWTExpMinutes:           jwtExp,
		RefreshExpDays:          refreshExp,
		GoogleClientID:          getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:      getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleCallbackURL:       getEnv("GOOGLE_CALLBACK_URL", "http://localhost:3000/api/v1/auth/google/callback"),
		FrontendURL:             getEnv("FRONTEND_URL", "http://localhost:3000"),
		SMTPHost:                getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:                smtpPort,
		SMTPUser:                getEnv("SMTP_USER", ""),
		SMTPPassword:            getEnv("SMTP_PASSWORD", ""),
		GeminiAPIKey:            getEnv("GEMINI_API_KEY", ""),
		GroqAPIKey:              getEnv("GROQ_API_KEY", ""),
		OllamaEndpoint:          getEnv("OLLAMA_ENDPOINT", ""),
		OllamaModel:             getEnv("OLLAMA_MODEL", ""),
		OpenRouterAPIKey:        getEnv("OPENROUTER_API_KEY", ""),
		OpenRouterModel:         getEnv("OPENROUTER_MODEL", ""),
		HFToken:                 getEnv("HF_TOKEN", ""),
		HFModel:                 getEnv("HF_MODEL", "meta-llama/Llama-3.3-70B-Instruct"),
		AIRateLimit:             aiRateLimit,
		AICacheTTLHours:         aiCacheTTL,
		TranslationBatchSize:    transBatch,
		TranslationWorkerConcur: transConc,
		TranslationMaxRetries:   transRetries,
		TranslationPollInterval: time.Duration(transPoll) * time.Second,
		RedisURI:                getEnv("REDIS_URI", "redis://localhost:6379/0"),
		CacheTTLDictionaryMin:   ttlDict,
		CacheTTLDashboardMin:    ttlDash,
		CacheTTLLeaderboardMin:  ttlLead,
		CacheTTLAIDays:          ttlAI,
	}

	if cfg.MongoURI == "" {
		return nil, fmt.Errorf("MONGO_URI is required but not set")
	}
	if cfg.MongoDB == "" {
		return nil, fmt.Errorf("MONGO_DB is required but not set")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
