package learning

import (
	"testing"
	"time"
)

func TestSM2Scheduler(t *testing.T) {
	s := NewSM2Scheduler()

	// Test Case 1: Initial state (EF=2.5, rep=0) and Quality=4 (Good)
	p1 := &LearningProgress{
		EaseFactor: 2.5,
		Repetition: 0,
		Interval:   0,
	}

	next1 := s.CalculateNext(p1, QualityGood)
	if next1.Repetition != 1 {
		t.Errorf("expected repetition 1, got %d", next1.Repetition)
	}
	if next1.Interval != 1 {
		t.Errorf("expected interval 1, got %d", next1.Interval)
	}
	// EF = 2.5 + (0.1 - (5-4)*(0.08 + (5-4)*0.02)) = 2.5 + (0.1 - 1*0.1) = 2.5
	if next1.EaseFactor != 2.5 {
		t.Errorf("expected EF 2.5, got %f", next1.EaseFactor)
	}

	// Test Case 2: Repetition=1 and Quality=4
	next2 := s.CalculateNext(next1, QualityGood)
	if next2.Repetition != 2 {
		t.Errorf("expected repetition 2, got %d", next2.Repetition)
	}
	if next2.Interval != 6 {
		t.Errorf("expected interval 6, got %d", next2.Interval)
	}

	// Test Case 3: Repetition=2 and Quality=4
	next3 := s.CalculateNext(next2, QualityGood)
	if next3.Repetition != 3 {
		t.Errorf("expected repetition 3, got %d", next3.Repetition)
	}
	// Interval = 6 * 2.5 = 15
	if next3.Interval != 15 {
		t.Errorf("expected interval 15, got %d", next3.Interval)
	}

	// Test Case 4: Quality=1 (Again) -> Repetition resets, Interval=1
	next4 := s.CalculateNext(next3, QualityAgain)
	if next4.Repetition != 0 {
		t.Errorf("expected repetition 0, got %d", next4.Repetition)
	}
	if next4.Interval != 1 {
		t.Errorf("expected interval 1, got %d", next4.Interval)
	}
	// EF should decrease
	// Let's just check it's < 2.5
	if next4.EaseFactor >= 2.5 {
		t.Errorf("expected EF to decrease, got %f", next4.EaseFactor)
	}

	// Test Case 5: EF floor is 1.3
	pLowEF := &LearningProgress{
		EaseFactor: 1.3,
		Repetition: 2,
		Interval:   6,
	}
	nextLow := s.CalculateNext(pLowEF, QualityAgain)
	if nextLow.EaseFactor < 1.3 {
		t.Errorf("expected EF floor 1.3, got %f", nextLow.EaseFactor)
	}
}

func TestIsDueAndLearned(t *testing.T) {
	now := time.Now()

	due := &LearningProgress{NextReviewDate: now.Add(-time.Hour)}
	if !IsDue(due) {
		t.Errorf("expected due to be true")
	}

	notDue := &LearningProgress{NextReviewDate: now.Add(time.Hour)}
	if IsDue(notDue) {
		t.Errorf("expected due to be false")
	}

	learned := &LearningProgress{Interval: 21}
	if !IsLearned(learned) {
		t.Errorf("expected learned to be true")
	}

	notLearned := &LearningProgress{Interval: 20}
	if IsLearned(notLearned) {
		t.Errorf("expected learned to be false")
	}
}
