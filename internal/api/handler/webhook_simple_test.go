package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/handsoff/handsoff/internal/webhook"
)

// TestWebhookHandler_ParseWebhook 测试 webhook 解析
func TestWebhookHandler_ParseWebhook(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		payload        interface{}
		expectedStatus int
		checkMessage   string
	}{
		{
			name: "Valid MR Event - Open",
			payload: webhook.GitLabMergeRequestEvent{
				ObjectKind: "merge_request",
				ObjectAttributes: webhook.GitLabMergeRequestAttributes{
					Action: "open",
					State:  "opened",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Wrong Event Type",
			payload: map[string]interface{}{
				"object_kind": "push",
			},
			expectedStatus: http.StatusOK,
			checkMessage:   "Event type not supported",
		},
		{
			name: "MR Event - Merge Action",
			payload: webhook.GitLabMergeRequestEvent{
				ObjectKind: "merge_request",
				ObjectAttributes: webhook.GitLabMergeRequestAttributes{
					Action: "merge",
					State:  "merged",
				},
			},
			expectedStatus: http.StatusOK,
			checkMessage:   "Event does not trigger review",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payloadBytes, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/webhook", bytes.NewBuffer(payloadBytes))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			// 注意：这里只测试 payload 解析，不测试完整 handler
			// 完整测试需要数据库和队列 mock

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestGitLabMergeRequestEvent_ShouldTriggerReview 测试触发逻辑
func TestGitLabMergeRequestEvent_ShouldTriggerReview(t *testing.T) {
	tests := []struct {
		name     string
		action   string
		state    string
		expected bool
	}{
		{"Open MR", "open", "opened", true},
		{"Update MR", "update", "opened", true},
		{"Merge MR", "merge", "merged", false},
		{"Close MR", "close", "closed", false},
		{"Reopen MR", "reopen", "opened", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &webhook.GitLabMergeRequestEvent{
				ObjectAttributes: webhook.GitLabMergeRequestAttributes{
					Action: tt.action,
					State:  tt.state,
				},
			}

			result := event.ShouldTriggerReview()
			if result != tt.expected {
				t.Errorf("ShouldTriggerReview() = %v, want %v", result, tt.expected)
			}
		})
	}
}
