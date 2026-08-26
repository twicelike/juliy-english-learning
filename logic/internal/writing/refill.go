package writing

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

type RefillJob struct {
	repo      *Repository
	generator *GeneratorService
	threshold int
	batchSize int
}

func NewRefillJob(repo *Repository, generator *GeneratorService) *RefillJob {
	return &RefillJob{
		repo:      repo,
		generator: generator,
		threshold: 20, // If count < 20, trigger refill
		batchSize: 20, // Generate 20 missions per API call
	}
}

// Start begins a background goroutine that runs the refill process periodically.
func (j *RefillJob) Start(ctx context.Context, interval time.Duration) {
	go func() {
		log.Info().Dur("interval", interval).Msg("Starting Writing Mission Auto Refill Ticker")

		// Run once immediately
		j.runOnce(ctx)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Info().Msg("Writing Mission Auto Refill Ticker stopped")
				return
			case <-ticker.C:
				j.runOnce(ctx)
			}
		}
	}()
}

// runOnce executes the refill process for all supported topics, scenarios, and difficulties.
func (j *RefillJob) runOnce(ctx context.Context) {
	log.Info().Msg("Writing Mission Auto Refill Job running")

	// We don't want to blast the AI API, so we only refill a max number of categories per run,
	// or we add a sleep between generations. Let's just iterate and refill as needed.
	for topic, scenarios := range SupportedScenarios {
		for _, scenario := range scenarios {
			for _, difficulty := range SupportedDifficulties {

				// Check if context is cancelled
				select {
				case <-ctx.Done():
					log.Warn().Msg("Refill job cancelled by context")
					return
				default:
				}

				count, err := j.repo.GetMissionCountByGroup(ctx, topic, scenario, difficulty)
				if err != nil {
					log.Error().Err(err).Str("topic", topic).Str("scenario", scenario).Str("difficulty", difficulty).Msg("Failed to check mission count")
					continue
				}

				if count < int64(j.threshold) {
					log.Info().
						Str("topic", topic).
						Str("scenario", scenario).
						Str("difficulty", difficulty).
						Int64("count", count).
						Msg("Mission count below threshold. Generating new missions...")

					// Call AI to generate missions
					missions, err := j.generator.GenerateMissions(ctx, topic, scenario, difficulty, j.batchSize)
					if err != nil {
						log.Error().Err(err).Msg("Failed to generate missions")
						continue
					}

					if len(missions) > 0 {
						inserted, err := j.repo.InsertMissions(ctx, missions)
						if err != nil {
							log.Error().Err(err).Msg("Failed to insert generated missions")
						} else {
							log.Info().Int("inserted", inserted).Msg("Successfully inserted new missions")
						}
					}

					// Sleep briefly to avoid hitting rate limits
					time.Sleep(5 * time.Second)
				}
			}
		}
	}

	log.Info().Msg("Writing Mission Auto Refill Job completed")
}
