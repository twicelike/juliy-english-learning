package config

import (
	"fmt"
	"os"
	"strconv"

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

	// AI
	GeminiAPIKey    string
	GroqAPIKey      string
	OllamaEndpoint  string // e.g. http://localhost:11434
	OllamaModel     string // e.g. qwen2.5:7b
	OpenRouterAPIKey string
	OpenRouterModel  string
	HFToken          string
	HFModel          string
	AIRateLimit      int // max requests per hour per user
	AICacheTTLHours int // cache TTL for global AI responses

	// Redis & Cache
	RedisURI              string
	CacheTTLDictionaryMin int
	CacheTTLDashboardMin  int
	CacheTTLLeaderboardMin int
	CacheTTLAIDays        int
}

// Load reads configuration from a .env file (if present) then falls back to OS
// environment variables. Returns an error if any required variable is missing.
func Load() (*Config, error) {
	// .env is optional — production environments typically use real env vars.
	_ = godotenv.Load()

	jwtExp, _ := strconv.Atoi(getEnv("JWT_EXPIRATION_MINUTES", "15"))
	refreshExp, _ := strconv.Atoi(getEnv("REFRESH_TOKEN_EXPIRATION_DAYS", "7"))

	aiRateLimit, _ := strconv.Atoi(getEnv("AI_RATE_LIMIT_PER_HOUR", "20"))
	aiCacheTTL, _ := strconv.Atoi(getEnv("AI_CACHE_TTL_HOURS", "24"))

	ttlDict, _ := strconv.Atoi(getEnv("CACHE_TTL_DICTIONARY_MINUTES", "1440")) // 24h
	ttlDash, _ := strconv.Atoi(getEnv("CACHE_TTL_DASHBOARD_MINUTES", "5"))
	ttlLead, _ := strconv.Atoi(getEnv("CACHE_TTL_LEADERBOARD_MINUTES", "2"))
	ttlAI, _ := strconv.Atoi(getEnv("CACHE_TTL_AI_DAYS", "7"))

	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		MongoURI:        getEnv("MONGO_URI", ""),
		MongoDB:         getEnv("MONGO_DB", ""),
		Env:             getEnv("ENV", "development"),
		CORSOrigins:     getEnv("CORS_ORIGINS", "*"),
		JWTSecret:       getEnv("JWT_SECRET", "super-secret-jwt-key-replace-me-in-prod"),
		JWTExpMinutes:   jwtExp,
		RefreshExpDays:  refreshExp,
		GeminiAPIKey:    getEnv("GEMINI_API_KEY", ""),
		GroqAPIKey:      getEnv("GROQ_API_KEY", ""),
		OllamaEndpoint:  getEnv("OLLAMA_ENDPOINT", ""),
		OllamaModel:     getEnv("OLLAMA_MODEL", ""),
		OpenRouterAPIKey: getEnv("OPENROUTER_API_KEY", ""),
		OpenRouterModel:  getEnv("OPENROUTER_MODEL", ""),
		HFToken:          getEnv("HF_TOKEN", ""),
		HFModel:          getEnv("HF_MODEL", "deepseek-ai/DeepSeek-V4-Flash:novita"),
		AIRateLimit:     aiRateLimit,
		AICacheTTLHours: aiCacheTTL,
		RedisURI:        getEnv("REDIS_URI", "redis://localhost:6379/0"),
		CacheTTLDictionaryMin:  ttlDict,
		CacheTTLDashboardMin:   ttlDash,
		CacheTTLLeaderboardMin: ttlLead,
		CacheTTLAIDays:         ttlAI,
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
