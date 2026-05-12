package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/pelican-dev/panel/internal/models"
)

type BackupProcessPayload struct {
	BackupID uint `json:"backup_id"`
}

func (q *Queue) HandleBackupProcess(ctx context.Context, t *asynq.Task) error {
	var payload BackupProcessPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	now := time.Now()
	success := true
	q.DB.Model(&models.Backup{}).Where("id = ?", payload.BackupID).Updates(map[string]interface{}{
		"is_successful": &success,
		"completed_at":  &now,
	})
	return nil
}
