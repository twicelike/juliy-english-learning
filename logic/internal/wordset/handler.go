package wordset

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetTopics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"topics": PredefinedTopics})
}

func (h *Handler) Explore(c *gin.Context) {
	var req ExploreWordSetsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.svc.ExploreWordSets(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) getUserID(c *gin.Context) (bson.ObjectID, bool) {
	uidStr := c.GetString("user_id")
	if uidStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return bson.NilObjectID, false
	}
	uid, err := bson.ObjectIDFromHex(uidStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return bson.NilObjectID, false
	}
	return uid, true
}

func (h *Handler) getSetID(c *gin.Context) (bson.ObjectID, bool) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wordset id"})
		return bson.NilObjectID, false
	}
	return id, true
}

func (h *Handler) Create(c *gin.Context) {
	uid, ok := h.getUserID(c)
	if !ok {
		return
	}

	var req CreateWordSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	set, err := h.svc.CreateWordSet(c.Request.Context(), uid, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, set)
}

func (h *Handler) List(c *gin.Context) {
	uid, ok := h.getUserID(c)
	if !ok {
		return
	}

	// Parse multiple contains_word IDs (e.g., ?contains_word=id1&contains_word=id2)
	var containsWordIDs []bson.ObjectID
	for _, cw := range c.QueryArray("contains_word") {
		if id, err := bson.ObjectIDFromHex(cw); err == nil {
			containsWordIDs = append(containsWordIDs, id)
		}
	}

	res, err := h.svc.GetWordSetsByUserID(c.Request.Context(), uid, containsWordIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) Get(c *gin.Context) {
	uid, ok := h.getUserID(c)
	if !ok {
		return
	}
	setID, ok := h.getSetID(c)
	if !ok {
		return
	}

	res, err := h.svc.GetWordSetDetail(c.Request.Context(), uid, setID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) Update(c *gin.Context) {
	uid, ok := h.getUserID(c)
	if !ok {
		return
	}
	setID, ok := h.getSetID(c)
	if !ok {
		return
	}

	var req UpdateWordSetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	set, err := h.svc.UpdateWordSet(c.Request.Context(), uid, setID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, set)
}

func (h *Handler) Delete(c *gin.Context) {
	uid, ok := h.getUserID(c)
	if !ok {
		return
	}
	setID, ok := h.getSetID(c)
	if !ok {
		return
	}

	err := h.svc.DeleteWordSet(c.Request.Context(), uid, setID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) GetWords(c *gin.Context) {
	uid, ok := h.getUserID(c)
	if !ok {
		return
	}
	setID, ok := h.getSetID(c)
	if !ok {
		return
	}

	words, err := h.svc.GetWordSetWords(c.Request.Context(), uid, setID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, words)
}

func (h *Handler) AddWord(c *gin.Context) {
	uid, ok := h.getUserID(c)
	if !ok {
		return
	}
	setID, ok := h.getSetID(c)
	if !ok {
		return
	}

	var req AddWordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wordID, err := bson.ObjectIDFromHex(req.WordID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid word ID format"})
		return
	}

	err = h.svc.AddWord(c.Request.Context(), uid, setID, wordID, req.SenseIndex)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) BulkAddWords(c *gin.Context) {
	uid, ok := h.getUserID(c)
	if !ok {
		return
	}
	setID, ok := h.getSetID(c)
	if !ok {
		return
	}

	var req BulkAddWordsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.svc.BulkAddWords(c.Request.Context(), uid, setID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Debug: log what was actually received and inserted
	c.Header("X-Debug-Received", fmt.Sprintf("%d", len(req.Items)))
	c.Header("X-Debug-Added", fmt.Sprintf("%d", res.Added))
	c.JSON(http.StatusOK, res)
}

func (h *Handler) RemoveWord(c *gin.Context) {
	uid, ok := h.getUserID(c)
	if !ok {
		return
	}
	setID, ok := h.getSetID(c)
	if !ok {
		return
	}
	wordIDStr := c.Param("wordId")
	wordID, err := bson.ObjectIDFromHex(wordIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid word id"})
		return
	}

	err = h.svc.RemoveWord(c.Request.Context(), uid, setID, wordID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) ReorderWords(c *gin.Context) {
	uid, ok := h.getUserID(c)
	if !ok {
		return
	}
	setID, ok := h.getSetID(c)
	if !ok {
		return
	}

	var req ReorderWordsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.svc.ReorderWords(c.Request.Context(), uid, setID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *Handler) Clone(c *gin.Context) {
	uid, ok := h.getUserID(c)
	if !ok {
		return
	}
	setID, ok := h.getSetID(c)
	if !ok {
		return
	}

	set, err := h.svc.CloneWordSet(c.Request.Context(), uid, setID)
	if err != nil {
		if err.Error() == "wordset not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == "you already own this word set" {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": set})
}

func (h *Handler) BatchDeleteWords(c *gin.Context) {
	uid, ok := h.getUserID(c)
	if !ok {
		return
	}
	setID, ok := h.getSetID(c)
	if !ok {
		return
	}

	var req BatchDeleteWordsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.svc.BatchDeleteWords(c.Request.Context(), uid, setID, req)
	if err != nil {
		if err.Error() == "wordset not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == fmt.Sprintf("unauthorized: you don't own this wordset") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *Handler) RestoreWords(c *gin.Context) {
	uid, ok := h.getUserID(c)
	if !ok {
		return
	}
	setID, ok := h.getSetID(c)
	if !ok {
		return
	}

	var req RestoreWordsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.svc.RestoreWords(c.Request.Context(), uid, setID, req)
	if err != nil {
		if err.Error() == "wordset not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if err.Error() == fmt.Sprintf("unauthorized: you don't own this wordset") {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, res)
}

