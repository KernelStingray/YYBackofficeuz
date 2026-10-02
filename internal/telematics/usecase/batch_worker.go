package usecase

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/hextec/yolyolakay-backend/internal/telematics/domain"
)

type BatchWorker struct {
	repo       domain.TelematicsRepository
	batchSize  int
	flushTime  time.Duration
	ctx        context.Context
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
}

func NewBatchWorker(repo domain.TelematicsRepository, batchSize int, flushTime time.Duration) *BatchWorker {
	ctx, cancel := context.WithCancel(context.Background())
	bw := &BatchWorker{
		repo:       repo,
		batchSize:  batchSize,
		flushTime:  flushTime,
		ctx:        ctx,
		cancelFunc: cancel,
	}

	bw.wg.Add(1)
	go bw.start()
	return bw
}

// Push is no longer needed since it's handled by Redis RPUSH directly in the UseCase.

func (b *BatchWorker) start() {
	defer b.wg.Done()
	ticker := time.NewTicker(b.flushTime)
	defer ticker.Stop()

	for {
		select {
		case <-b.ctx.Done():
			b.flush(true) // flush remaining on shutdown synchronously
			return
		case <-ticker.C:
			b.flush(false)
		}
	}
}

func (b *BatchWorker) flush(syncRun bool) {
	locationsToSave, err := b.repo.PopLocationsFromQueue(b.ctx, b.batchSize)
	if err != nil {
		slog.Error("[BatchWorker] Error popping from redis queue", "error", err)
		return
	}

	if len(locationsToSave) == 0 {
		return
	}

	saveFn := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := b.repo.SaveLocationsBatch(ctx, locationsToSave); err != nil {
			slog.Error("[BatchWorker] Error saving batch of locations", "count", len(locationsToSave), "error", err)
			// Optional: Push back to queue or dead-letter queue if it fails
		} else {
			slog.Info("[BatchWorker] Successfully flushed %d locations to PostgreSQL", "count", len(locationsToSave))
		}
	}

	// Asynchronously save to PostgreSQL (Cold Data)
	if syncRun {
		saveFn()
	} else {
		b.wg.Add(1)
		go func() {
			defer b.wg.Done()
			saveFn()
		}()
	}
}

func (b *BatchWorker) Stop() {
	b.cancelFunc()
	b.wg.Wait()
}
