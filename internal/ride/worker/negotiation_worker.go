package worker

import (
	"context"
	"log/slog"
	"time"

	"github.com/hextec/yolyolakay-backend/internal/ride/domain"
)

type NegotiationWorker struct {
	useCase domain.RideUseCase
	ticker  *time.Ticker
	done    chan bool
}

func NewNegotiationWorker(useCase domain.RideUseCase) *NegotiationWorker {
	return &NegotiationWorker{
		useCase: useCase,
		done:    make(chan bool),
	}
}

func (w *NegotiationWorker) Start() {
	w.ticker = time.NewTicker(1 * time.Minute)
	slog.Info("NegotiationWorker started: checking for expired negotiations every minute")

	for {
		select {
		case <-w.done:
			slog.Info("NegotiationWorker stopped")
			return
		case <-w.ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := w.useCase.TimeoutPendingNegotiations(ctx); err != nil {
				slog.Error("NegotiationWorker error", "error", err)
			}
			cancel()
		}
	}
}

func (w *NegotiationWorker) Stop() {
	if w.ticker != nil {
		w.ticker.Stop()
	}
	w.done <- true
}
