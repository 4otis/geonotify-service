package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/4otis/geonotify-service/internal/entity"
	"github.com/4otis/geonotify-service/internal/port/repo"
	"github.com/4otis/geonotify-service/pkg/redis"
	"go.uber.org/zap"
)

type WebhookWorker struct {
	logger      *zap.Logger
	webhookRepo repo.WebhookRepo
	redis       *redis.Client
	webhookURL  string
	maxRetries  int
	retryDelay  time.Duration
	stopChan    chan struct{}
}

func NewWebhookWorker(
	logger *zap.Logger,
	webhookRepo repo.WebhookRepo,
	redis *redis.Client,
	webhookURL string,
	maxRetries int,
	retryDelaySeconds int,
) *WebhookWorker {
	return &WebhookWorker{
		logger:      logger,
		webhookRepo: webhookRepo,
		redis:       redis,
		webhookURL:  webhookURL,
		maxRetries:  maxRetries,
		retryDelay:  time.Duration(retryDelaySeconds) * time.Second,
		stopChan:    make(chan struct{}),
	}
}

func (w *WebhookWorker) Start(ctx context.Context) {
	w.logger.Info("Starting webhook worker")

	go w.processQueue(ctx)
	go w.processDB(ctx)
}

func (w *WebhookWorker) Stop() {
	w.logger.Info("Stopping webhook worker")
	close(w.stopChan)
}

func (w *WebhookWorker) processQueue(ctx context.Context) {
	w.logger.Info("Starting queue processor")

	for {
		select {
		case <-w.stopChan:
			return
		case <-ctx.Done():
			return
		default:
			_, data, err := w.redis.BRPop("webhooks:queue", 5*time.Second)
			if err != nil {
				if err != redis.ErrNotFound {
					w.logger.Error("Failed to pop from queue", zap.Error(err))
				}
				continue
			}

			go w.processTask(ctx, data)
		}
	}
}

func (w *WebhookWorker) processDB(ctx context.Context) {
	w.logger.Info("Starting DB processor")

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			webhooks, err := w.webhookRepo.ReadInProgress(ctx, 10)
			if err != nil {
				w.logger.Error("Failed to read in-progress webhooks", zap.Error(err))
				continue
			}

			for _, wh := range webhooks {
				if wh.ScheduledAt.After(time.Now()) {
					continue
				}

				task := map[string]interface{}{
					"webhook_id": wh.ID,
					"check_id":   wh.CheckID,
					"payload":    string(wh.Payload),
				}

				if err := w.redis.LPush("webhooks:queue", task); err != nil {
					w.logger.Error("Failed to push webhook to queue",
						zap.Error(err),
						zap.Int("webhook_id", wh.ID))
				}
			}
		}
	}
}

func (w *WebhookWorker) processTask(ctx context.Context, data []byte) {
	var task map[string]interface{}
	if err := json.Unmarshal(data, &task); err != nil {
		w.logger.Error("Failed to unmarshal task", zap.Error(err))
		return
	}

	webhookID, ok := task["webhook_id"].(float64)
	if !ok {
		w.logger.Error("Invalid webhook_id in task")
		return
	}

	wh, err := w.webhookRepo.Read(ctx, int(webhookID))
	if err != nil {
		w.logger.Error("Failed to get webhook by ID",
			zap.Error(err),
			zap.Int("webhook_id", int(webhookID)))
		return
	}

	if err := w.sendWebhook(ctx, wh); err != nil {
		w.logger.Error("Failed to send webhook",
			zap.Error(err),
			zap.Int("webhook_id", wh.ID))
	}
}

func (w *WebhookWorker) sendWebhook(ctx context.Context, wh *entity.Webhook) error {
	if err := w.webhookRepo.UpdateState(ctx, wh.ID, "processing", wh.RetryCnt); err != nil {
		return fmt.Errorf("failed to update state: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", w.webhookURL, bytes.NewReader(wh.Payload))
	if err != nil {
		return w.handleRetry(ctx, wh, err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return w.handleRetry(ctx, wh, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if err := w.webhookRepo.MarkAsDelivered(ctx, wh.ID); err != nil {
			return fmt.Errorf("failed to mark as delivered: %w", err)
		}
		w.logger.Info("Webhook delivered successfully",
			zap.Int("webhook_id", wh.ID),
			zap.Int("status_code", resp.StatusCode))
		return nil
	}

	return w.handleRetry(ctx, wh, fmt.Errorf("HTTP status: %d", resp.StatusCode))
}

func (w *WebhookWorker) handleRetry(ctx context.Context, wh *entity.Webhook, err error) error {
	if wh.RetryCnt >= w.maxRetries {
		if updateErr := w.webhookRepo.UpdateState(ctx, wh.ID, "failed", wh.RetryCnt); updateErr != nil {
			return fmt.Errorf("failed to mark as failed: %v (original: %w)", updateErr, err)
		}
		w.logger.Error("Webhook failed after max retries",
			zap.Int("webhook_id", wh.ID),
			zap.Int("retry_count", wh.RetryCnt),
			zap.Error(err))
		return fmt.Errorf("max retries exceeded: %w", err)
	}

	newRetryCount := wh.RetryCnt + 1
	if updateErr := w.webhookRepo.UpdateState(ctx, wh.ID, "in progress", newRetryCount); updateErr != nil {
		return fmt.Errorf("failed to update retry count: %v (original: %w)", updateErr, err)
	}

	retryTask := map[string]interface{}{
		"webhook_id": wh.ID,
		"check_id":   wh.CheckID,
		"payload":    string(wh.Payload),
	}

	time.Sleep(w.retryDelay)
	if pushErr := w.redis.LPush("webhooks:queue", retryTask); pushErr != nil {
		w.logger.Error("Failed to schedule retry",
			zap.Error(pushErr),
			zap.Int("webhook_id", wh.ID))
	}

	w.logger.Info("Webhook scheduled for retry",
		zap.Int("webhook_id", wh.ID),
		zap.Int("retry_count", newRetryCount),
		zap.Error(err))

	return err
}
