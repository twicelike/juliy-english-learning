package ai

import (
	"github.com/english/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts AI Tutor routes onto the provided router group.
// All routes require authentication and are subject to rate limiting enforced
// inside each handler.
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, jwtSecret string) {
	ai := rg.Group("/ai")
	ai.Use(middleware.RequireAuth(jwtSecret))
	{
		words := ai.Group("/words")
		{
			words.POST("/explain", h.ExplainWord)
			words.POST("/nuance", h.ExplainNuance)
			words.POST("/compare", h.CompareWords)
			words.POST("/examples", h.GenerateExamples)
			words.POST("/mnemonic", h.CreateMnemonic)
		}

		writing := ai.Group("/writing")
		{
			writing.POST("/check", h.CheckGrammar)
			writing.POST("/vocab", h.SuggestWritingVocab)
		}

		quiz := ai.Group("/quiz")
		{
			quiz.POST("/fill-blank", h.GenerateFillBlank)
		}
	}
}
