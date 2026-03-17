package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

// JobType represents the type of job
type JobType string

const (
	JobTypeVideoProcessing JobType = "video_processing"
)

// Job represents a background job
type Job struct {
	ID        uuid.UUID              `json:"id"`
	Type      JobType                `json:"type"`
	Status    JobStatus              `json:"status"`
	Payload   map[string]interface{} `json:"payload"`
	Error     string                 `json:"error,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Retries   int                    `json:"retries"`
	MaxRetries int                   `json:"max_retries"`
}

// Handler is a function that processes a job
type Handler func(ctx context.Context, job *Job) error

// Queue represents a simple in-memory job queue
type Queue struct {
	jobs     chan *Job
	handlers map[JobType]Handler
	mu       sync.RWMutex
	workers  int
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewQueue creates a new job queue
func NewQueue(workers int, bufferSize int) *Queue {
	ctx, cancel := context.WithCancel(context.Background())

	return &Queue{
		jobs:     make(chan *Job, bufferSize),
		handlers: make(map[JobType]Handler),
		workers:  workers,
		ctx:      ctx,
		cancel:   cancel,
	}
}

// RegisterHandler registers a handler for a job type
func (q *Queue) RegisterHandler(jobType JobType, handler Handler) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handlers[jobType] = handler
}

// Start starts the worker pool
func (q *Queue) Start() {
	log.Info().Int("workers", q.workers).Msg("Starting job queue workers")

	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}
}

// Stop gracefully stops the queue
func (q *Queue) Stop() {
	log.Info().Msg("Stopping job queue...")
	q.cancel()
	q.wg.Wait()
	close(q.jobs)
	log.Info().Msg("Job queue stopped")
}

// Enqueue adds a job to the queue
func (q *Queue) Enqueue(jobType JobType, payload map[string]interface{}) (*Job, error) {
	job := &Job{
		ID:         uuid.New(),
		Type:       jobType,
		Status:     JobStatusPending,
		Payload:    payload,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		MaxRetries: 3,
	}

	select {
	case q.jobs <- job:
		log.Info().
			Str("job_id", job.ID.String()).
			Str("job_type", string(jobType)).
			Msg("Job enqueued")
		return job, nil
	case <-q.ctx.Done():
		return nil, fmt.Errorf("queue is shutting down")
	default:
		return nil, fmt.Errorf("queue is full")
	}
}

// worker processes jobs from the queue
func (q *Queue) worker(id int) {
	defer q.wg.Done()

	log.Info().Int("worker_id", id).Msg("Worker started")

	for {
		select {
		case <-q.ctx.Done():
			log.Info().Int("worker_id", id).Msg("Worker stopping")
			return
		case job, ok := <-q.jobs:
			if !ok {
				log.Info().Int("worker_id", id).Msg("Worker channel closed")
				return
			}

			q.processJob(id, job)
		}
	}
}

// processJob processes a single job
func (q *Queue) processJob(workerID int, job *Job) {
	log.Info().
		Int("worker_id", workerID).
		Str("job_id", job.ID.String()).
		Str("job_type", string(job.Type)).
		Msg("Processing job")

	// Update job status
	job.Status = JobStatusProcessing
	job.UpdatedAt = time.Now()

	// Get handler
	q.mu.RLock()
	handler, exists := q.handlers[job.Type]
	q.mu.RUnlock()

	if !exists {
		log.Error().
			Str("job_id", job.ID.String()).
			Str("job_type", string(job.Type)).
			Msg("No handler registered for job type")
		job.Status = JobStatusFailed
		job.Error = "no handler registered"
		job.UpdatedAt = time.Now()
		return
	}

	// Execute handler with timeout
	ctx, cancel := context.WithTimeout(q.ctx, 5*time.Minute)
	defer cancel()

	if err := handler(ctx, job); err != nil {
		log.Error().
			Err(err).
			Int("worker_id", workerID).
			Str("job_id", job.ID.String()).
			Int("retries", job.Retries).
			Msg("Job failed")

		job.Retries++
		job.Error = err.Error()
		job.UpdatedAt = time.Now()

		// Retry if not exceeded max retries
		if job.Retries < job.MaxRetries {
			job.Status = JobStatusPending
			log.Info().
				Str("job_id", job.ID.String()).
				Int("retry", job.Retries).
				Int("max_retries", job.MaxRetries).
				Msg("Retrying job")

			// Re-enqueue with exponential backoff
			delay := time.Duration(job.Retries*job.Retries) * time.Second
			time.AfterFunc(delay, func() {
				select {
				case q.jobs <- job:
					log.Info().Str("job_id", job.ID.String()).Msg("Job re-enqueued")
				default:
					log.Error().Str("job_id", job.ID.String()).Msg("Failed to re-enqueue job")
				}
			})
		} else {
			job.Status = JobStatusFailed
			log.Error().
				Str("job_id", job.ID.String()).
				Int("retries", job.Retries).
				Msg("Job failed after max retries")
		}
		return
	}

	// Job completed successfully
	job.Status = JobStatusCompleted
	job.UpdatedAt = time.Now()

	log.Info().
		Int("worker_id", workerID).
		Str("job_id", job.ID.String()).
		Msg("Job completed successfully")
}
