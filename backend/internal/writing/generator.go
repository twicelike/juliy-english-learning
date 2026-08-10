package writing

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/english/backend/internal/ai"
	"github.com/rs/zerolog/log"
)

type GeneratorService struct {
	aiClient ai.AIClient
}

func NewGeneratorService(aiClient ai.AIClient) *GeneratorService {
	return &GeneratorService{
		aiClient: aiClient,
	}
}

// GenerateMissions calls the AI to generate a batch of missions for a specific topic, scenario, and difficulty.
func (s *GeneratorService) GenerateMissions(ctx context.Context, topic, scenario, difficulty string, count int) ([]WritingMission, error) {
	prompt := s.buildPrompt(topic, scenario, difficulty, count)

	var parsedMissions []WritingMission
	if err := s.aiClient.GenerateJSON(ctx, prompt, nil, &parsedMissions); err != nil {
		log.Error().Err(err).Msg("Failed to generate and parse AI missions JSON")
		return nil, fmt.Errorf("failed to generate AI response: %w", err)
	}

	var validMissions []WritingMission
	for _, m := range parsedMissions {
		// Validate
		if len(m.Vietnamese) == 0 || len(m.TargetVocabulary) == 0 {
			continue // skip invalid
		}
		
		// Set static fields
		m.Topic = topic
		m.Scenario = scenario
		m.Difficulty = difficulty
		m.CreatedBy = "AI"
		m.CreatedAt = time.Now()

		// Generate Hash to prevent duplicates
		m.Hash = generateHash(m.Vietnamese)
		
		validMissions = append(validMissions, m)
	}

	return validMissions, nil
}

// EvaluateMission calls the AI to evaluate a user's submitted translation for a mission.
func (s *GeneratorService) EvaluateMission(ctx context.Context, mission *WritingMission, userContent string) (*EvaluationResponseDTO, error) {
	prompt := s.buildEvaluationPrompt(mission, userContent)

	var resp EvaluationResponseDTO
	if err := s.aiClient.GenerateJSON(ctx, prompt, nil, &resp); err != nil {
		log.Error().Err(err).Msg("Failed to generate and parse AI evaluation JSON")
		return nil, fmt.Errorf("failed to evaluate mission: %w", err)
	}

	return &resp, nil
}

func (s *GeneratorService) buildPrompt(topic, scenario, difficulty string, count int) string {
	var difficultyReqs string
	switch difficulty {
	case DifficultyEasy:
		difficultyReqs = "- 1 câu đơn giản.\n- Cấu trúc ngữ pháp cơ bản (hiện tại đơn, quá khứ đơn).\n- Từ vựng phổ biến."
	case DifficultyMedium:
		difficultyReqs = "- 2 câu.\n- Yêu cầu sử dụng nhiều thì khác nhau.\n- Yêu cầu sử dụng liên từ (and, but, although, because...)."
	case DifficultyHard:
		difficultyReqs = "- 2 đến 3 câu.\n- Cấu trúc ngữ pháp phức tạp (câu điều kiện, mệnh đề quan hệ, đảo ngữ...).\n- Nhiều cách diễn đạt đa dạng."
	}

	return fmt.Sprintf(`Bạn là một chuyên gia tạo bài tập tiếng Anh. Nhiệm vụ của bạn là tạo ra một mảng chứa %d đề bài (Writing Missions) cho người học tiếng Anh.
Chủ đề (Topic): %s
Ngữ cảnh (Scenario): %s
Độ khó (Difficulty): %s

Yêu cầu về độ khó này:
%s

MỖI ĐỀ BÀI PHẢI TUÂN THỦ ĐÚNG JSON SCHEMA SAU ĐÂY:
{
  "vietnamese": ["Câu tiếng Việt 1", "Câu tiếng Việt 2..."],
  "target_vocabulary": ["word1", "word2", "word3"],
  "grammar_focus": ["Tên chủ điểm ngữ pháp 1", "Tên chủ điểm 2"]
}

LUẬT QUAN TRỌNG:
1. Bạn CHỈ ĐƯỢC PHÉP TRẢ VỀ JSON ARRAY. KHÔNG kèm theo bất kỳ văn bản giải thích nào khác. KHÔNG bọc trong markdown code block nếu không cần thiết.
2. Nội dung tiếng Việt phải tự nhiên, thực tế và phù hợp với ngữ cảnh %s.
3. Target vocabulary phải gồm 2-4 từ vựng thực sự quan trọng và liên quan chặt chẽ đến ngữ cảnh.
`, count, topic, scenario, difficulty, difficultyReqs, scenario)
}

func (s *GeneratorService) buildEvaluationPrompt(mission *WritingMission, userContent string) string {
	vietnameseText := strings.Join(mission.Vietnamese, " ")
	vocabText := strings.Join(mission.TargetVocabulary, ", ")
	grammarText := strings.Join(mission.GrammarFocus, ", ")

	return fmt.Sprintf(`Bạn là một giáo viên tiếng Anh chuyên nghiệp. Hãy chấm điểm đoạn văn tiếng Anh do học viên viết dựa trên yêu cầu dịch từ tiếng Việt.

YÊU CẦU:
- Đề bài (Tiếng Việt): "%s"
- Ngữ cảnh: %s (%s)
- Từ vựng mục tiêu cần dùng: [%s]
- Ngữ pháp mục tiêu: [%s]

BÀI LÀM CỦA HỌC VIÊN:
"%s"

NHIỆM VỤ:
Chấm điểm khắt khe trên thang 100 và trả về ĐÚNG định dạng JSON sau (Viết thật NGẮN GỌN, mỗi nhận xét tối đa 1 câu):
{
  "score": 0,
  "breakdown": {
    "grammar": { "score": 0, "feedback": "1 câu ngắn..." },
    "vocabulary": { "score": 0, "feedback": "1 câu ngắn..." },
    "naturalness": { "score": 0, "feedback": "1 câu ngắn..." },
    "completeness": { "score": 0, "feedback": "1 câu ngắn..." }
  },
  "corrected_version": "Bản sửa hoàn chỉnh",
  "detailed_feedback": "Nhận xét tổng quan tối đa 2 câu."
}

LƯU Ý QUAN TRỌNG: 
- KHÔNG trả về bất kỳ text nào khác ngoài JSON object.
- Tổng điểm (score) nên là trung bình cộng của 4 tiêu chí thành phần.
- CỰC KỲ QUAN TRỌNG: Câu nhận xét phải thật ngắn gọn để không bị cắt bớt JSON!
`, vietnameseText, mission.Scenario, mission.Topic, vocabText, grammarText, userContent)
}

func generateHash(vietnamese []string) string {
	// Normalize and join
	var normalized []string
	for _, s := range vietnamese {
		normalized = append(normalized, strings.TrimSpace(s))
	}
	joined := strings.Join(normalized, " ")
	
	// Create SHA256 hash
	h := sha256.New()
	h.Write([]byte(joined))
	return hex.EncodeToString(h.Sum(nil))
}
