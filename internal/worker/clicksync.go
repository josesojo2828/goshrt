package worker

import (
	"context"
	"log"
	"time"

	"github.com/jsojo/goshrt/internal/service"
)

type ClickSyncWorker struct {
	svc      *service.Service
	interval time.Duration
	stopCh   chan struct{}
}

func New(svc *service.Service, interval time.Duration) *ClickSyncWorker {
	return &ClickSyncWorker{
		svc:      svc,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (w *ClickSyncWorker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := w.svc.SyncClicks(ctx); err != nil {
				log.Printf("click sync error: %v", err)
			}
		case <-ctx.Done():
			w.flushPending()
			return
		case <-w.stopCh:
			return
		}
	}
}

func (w *ClickSyncWorker) Stop() {
	close(w.stopCh)
}

func (w *ClickSyncWorker) flushPending() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := w.svc.SyncClicks(ctx); err != nil {
		log.Printf("click sync flush error: %v", err)
	}
}
