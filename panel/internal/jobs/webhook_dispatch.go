package jobs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

type WebhookDispatchPayload struct {
	WebhookID uint   `json:"webhook_id"`
	Event     string `json:"event"`
	Data      string `json:"data"`
}

func (q *Queue) HandleWebhookDispatch(ctx context.Context, t *asynq.Task) error {
	var payload WebhookDispatchPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	return nil
}
