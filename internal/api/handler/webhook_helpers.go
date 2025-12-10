package handler

import (
	"github.com/handsoff/handsoff/internal/model"
	"github.com/handsoff/handsoff/internal/webhook"
)

// createWebhookEventRecord creates a webhook event record for tracking
func (h *WebhookHandler) createWebhookEventRecord(repo *model.Repository, mrEvent *webhook.GitLabMergeRequestEvent, rawBody []byte) (*model.WebhookEvent, error) {
	mrIID := mrEvent.GetMRID()
	webhookEvent := &model.WebhookEvent{
		RepositoryID: repo.ID,
		EventType:    model.EventTypeMergeRequest,
		Action:       model.MRAction(mrEvent.ObjectAttributes.Action),
		SourceBranch: mrEvent.GetSourceBranch(),
		TargetBranch: mrEvent.GetTargetBranch(),
		MRIID:        &mrIID,
		CommitSHA:    mrEvent.ObjectAttributes.LastCommit.ID,
		Status:       model.EventStatusPending,
		RawPayload:   string(rawBody),
	}

	if err := h.db.Create(webhookEvent).Error; err != nil {
		h.log.Error("Failed to create webhook event record", "error", err)
		return nil, err
	}

	h.log.Info("Webhook event record created",
		"event_id", webhookEvent.ID,
		"repository_id", repo.ID,
		"event_type", webhookEvent.EventType,
		"mr_iid", mrIID)

	return webhookEvent, nil
}
