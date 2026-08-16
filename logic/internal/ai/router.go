package ai

import (
	"github.com/english/backend/internal/auth"
	"github.com/english/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts AI Tutor routes onto the provided router group.
// It applies the authentication middleware so that only logged-in users can use the AI.
func RegisterRoutes(rg *gin.RouterGroup, h *Handler, authRepo auth.Repository) {
	aiGroup := rg.Group("/ai")
	aiGroup.Use(middleware.RequireAuth(authRepo))
	{
		words := aiGroup.Group("/words")
		{
			words.POST("/explain-sense", h.ExplainSense)
			words.POST("/nuance", h.ExplainNuance)
			words.POST("/compare", h.CompareWords)
			words.POST("/examples", h.GenerateExamples)

		}

		writing := aiGroup.Group("/writing")
		{
			writing.POST("/check", h.CheckGrammar)
			writing.POST("/vocab", h.SuggestWritingVocab)
		}

		quiz := aiGroup.Group("/quiz")
		{
			quiz.POST("/fill-blank", h.GenerateFillBlank)
		}
	}
}
