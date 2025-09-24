package handlers

import (
	"fmt"
	"net/http"

	"quivra-backend/services"

	"github.com/gin-gonic/gin"
)

type QuestionHandler struct {
	questionService *services.QuestionService
}

func NewQuestionHandler(questionService *services.QuestionService) *QuestionHandler {
	return &QuestionHandler{questionService: questionService}
}

// CreateQuestion 問題作成
func (qh *QuestionHandler) CreateQuestion(c *gin.Context) {
	var req struct {
		Question   string `json:"question" binding:"required"`
		Answer     string `json:"answer" binding:"required"`
		Category   string `json:"category"`
		Difficulty string `json:"difficulty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// デフォルト値の設定
	if req.Category == "" {
		req.Category = "general"
	}
	if req.Difficulty == "" {
		req.Difficulty = "medium"
	}

	question, err := qh.questionService.CreateQuestion(req.Question, req.Answer, req.Category, req.Difficulty)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      question.ID,
		"message": "問題が作成されました",
	})
}

// GetQuestions 問題一覧取得
func (qh *QuestionHandler) GetQuestions(c *gin.Context) {
	category := c.Query("category")
	difficulty := c.Query("difficulty")

	questions, err := qh.questionService.GetQuestions(category, difficulty)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"questions": questions})
}

// GetQuestion 問題取得
func (qh *QuestionHandler) GetQuestion(c *gin.Context) {
	questionID := c.Param("id")
	if questionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "question ID is required"})
		return
	}

	// IDをintに変換
	var id int
	if _, err := fmt.Sscanf(questionID, "%d", &id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid question ID"})
		return
	}

	question, err := qh.questionService.GetQuestion(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, question)
}
