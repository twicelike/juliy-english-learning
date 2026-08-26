package events

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// EventType is the type of the event.
type EventType string

const (
	EventTypeProfileRefreshRequested EventType = "profile.refresh_requested"
)

// Event represents a generic event in the system.
type Event struct {
	Type      EventType
	UserID    string
	Reason    string
	CreatedAt time.Time
}

// Publisher defines how to publish an event.
type Publisher interface {
	Publish(ctx context.Context, ev Event) error
}

// ProfileRefreshHandler defines the interface for handling profile refreshes.
type ProfileRefreshHandler interface {
	RefreshSkills(ctx context.Context, userID string) error
}

// InProcessPublisher is a lightweight channel-based event publisher.
// It coalesces (debounces) frequent refresh requests for the same user.
type InProcessPublisher struct {
	ch             chan Event
	wg             sync.WaitGroup
	profileHandler ProfileRefreshHandler
	stopCh         chan struct{}
}

// NewInProcessPublisher creates a new InProcessPublisher.
func NewInProcessPublisher(bufferSize int) *InProcessPublisher {
	return &InProcessPublisher{
		ch:     make(chan Event, bufferSize),
		stopCh: make(chan struct{}),
	}
}

// SetProfileHandler sets the handler for profile refresh events.
func (p *InProcessPublisher) SetProfileHandler(h ProfileRefreshHandler) {
	p.profileHandler = h
}

// Publish sends an event to the channel.
func (p *InProcessPublisher) Publish(ctx context.Context, ev Event) error {
	if ev.CreatedAt.IsZero() {
		ev.CreatedAt = time.Now()
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.stopCh:
		return fmt.Errorf("publisher is stopped")
	case p.ch <- ev:
		return nil
	default:
		// Channel is full, log warning
		log.Warn().Str("event", string(ev.Type)).Str("user", ev.UserID).Msg("event publisher channel is full, dropping event")
		return fmt.Errorf("event publisher channel is full")
	}
}

// StartWorker begins the background worker to process events.
func (p *InProcessPublisher) StartWorker(ctx context.Context) {
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()

		// Map to store pending requests for deduplication
		// Key: UserID, Value: Event
		pending := make(map[string]Event)
		ticker := time.NewTicker(2 * time.Second) // Process batch every 2 seconds
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				p.processPending(pending) // Flush before exit
				return
			case <-p.stopCh:
				p.processPending(pending) // Flush before exit
				return
			case ev := <-p.ch:
				if ev.Type == EventTypeProfileRefreshRequested {
					pending[ev.UserID] = ev
				}
			case <-ticker.C:
				p.processPending(pending)
			}
		}
	}()
}

// processPending processes all pending refresh events and clears the map.
func (p *InProcessPublisher) processPending(pending map[string]Event) {
	if len(pending) == 0 {
		return
	}

	for userID, ev := range pending {
		// Use a detached context for the actual processing since we're in background
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		err := p.profileHandler.RefreshSkills(ctx, userID)
		if err != nil {
			log.Error().Err(err).Str("user", userID).Str("reason", ev.Reason).Msg("failed to process profile refresh event")
		} else {
			log.Info().Str("user", userID).Str("reason", ev.Reason).Msg("profile refresh processed successfully")
		}
		cancel()

		delete(pending, userID)
	}
}

// Stop stops the background worker.
func (p *InProcessPublisher) Stop() {
	close(p.stopCh)
	p.wg.Wait()
}
