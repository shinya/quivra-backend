package services

import (
	"database/sql"
	"fmt"

	"quivra-backend/database"
	"quivra-backend/models"
)

type QuestionService struct {
	db *database.DB
}

func NewQuestionService(db *database.DB) *QuestionService {
	return &QuestionService{db: db}
}

// CreateQuestion 問題を作成
func (qs *QuestionService) CreateQuestion(question, answer, category, difficulty string) (*models.Question, error) {
	query := `INSERT INTO questions (question, answer, category, difficulty) VALUES (?, ?, ?, ?)`
	result, err := qs.db.Exec(query, question, answer, category, difficulty)
	if err != nil {
		return nil, fmt.Errorf("failed to create question: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get question ID: %w", err)
	}

	return &models.Question{
		ID:         int(id),
		Question:   question,
		Answer:     answer,
		Category:   category,
		Difficulty: difficulty,
	}, nil
}

// GetQuestions 問題一覧を取得
func (qs *QuestionService) GetQuestions(category, difficulty string) ([]models.Question, error) {
	query := `SELECT id, question, answer, category, difficulty, created_at FROM questions WHERE 1=1`
	args := []interface{}{}

	if category != "" {
		query += ` AND category = ?`
		args = append(args, category)
	}

	if difficulty != "" {
		query += ` AND difficulty = ?`
		args = append(args, difficulty)
	}

	query += ` ORDER BY created_at DESC`

	rows, err := qs.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query questions: %w", err)
	}
	defer rows.Close()

	var questions []models.Question
	for rows.Next() {
		var question models.Question
		err := rows.Scan(&question.ID, &question.Question, &question.Answer, &question.Category, &question.Difficulty, &question.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan question: %w", err)
		}
		questions = append(questions, question)
	}

	return questions, nil
}

// GetQuestion 問題を取得
func (qs *QuestionService) GetQuestion(id int) (*models.Question, error) {
	var question models.Question
	query := `SELECT id, question, answer, category, difficulty, created_at FROM questions WHERE id = ?`
	err := qs.db.QueryRow(query, id).Scan(&question.ID, &question.Question, &question.Answer, &question.Category, &question.Difficulty, &question.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("question not found")
		}
		return nil, fmt.Errorf("failed to get question: %w", err)
	}

	return &question, nil
}

// GetRandomQuestion ランダムな問題を取得
func (qs *QuestionService) GetRandomQuestion(category, difficulty string) (*models.Question, error) {
	query := `SELECT id, question, answer, category, difficulty, created_at FROM questions WHERE 1=1`
	args := []interface{}{}

	if category != "" {
		query += ` AND category = ?`
		args = append(args, category)
	}

	if difficulty != "" {
		query += ` AND difficulty = ?`
		args = append(args, difficulty)
	}

	query += ` ORDER BY RAND() LIMIT 1`

	var question models.Question
	err := qs.db.QueryRow(query, args...).Scan(&question.ID, &question.Question, &question.Answer, &question.Category, &question.Difficulty, &question.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no questions found")
		}
		return nil, fmt.Errorf("failed to get random question: %w", err)
	}

	return &question, nil
}
