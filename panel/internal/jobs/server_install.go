package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/pelican-dev/panel/internal/models"
)

type ServerInstallPayload struct {
	ServerID uint `json:"server_id"`
}

func (q *Queue) HandleServerInstall(ctx context.Context, t *asynq.Task) error {
	var payload ServerInstallPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	now := time.Now()
	q.DB.Model(&models.Server{}).Where("id = ?", payload.ServerID).Updates(map[string]interface{}{
		"status":       nil,
		"installed_at": &now,
	})
	return nil
}
