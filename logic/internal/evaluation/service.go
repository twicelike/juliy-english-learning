package evaluation

import (
	"context"
	"time"
)

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

func (s *service) RecordEvidence(ctx context.Context, ev *Evidence) error {
	if ev.CreatedAt.IsZero() {
		ev.CreatedAt = time.Now()
	}
	if ev.OccurredAt.IsZero() {
		ev.OccurredAt = time.Now()
	}
	return s.repo.InsertEvidence(ctx, ev)
}

func (s *service) RecordEvidences(ctx context.Context, evs []*Evidence) error {
	if len(evs) == 0 {
		return nil
	}
	
	now := time.Now()
	for _, ev := range evs {
		if ev.CreatedAt.IsZero() {
			ev.CreatedAt = now
		}
		if ev.OccurredAt.IsZero() {
			ev.OccurredAt = now
		}
	}
	return s.repo.InsertMany(ctx, evs)
}

func (s *service) CalculateSkillAssessment(ctx context.Context, userID string, skill string) (SkillAssessmentResult, error) {
	evidences, err := s.repo.GetEvidencesBySkill(ctx, userID, skill)
	if err != nil {
		return SkillAssessmentResult{}, err
	}

	return EvaluateSkill(evidences, DefaultScoringPolicy, time.Now())
}
