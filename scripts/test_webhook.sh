#!/bin/bash

# Week 4 Task 1: Webhook接收测试脚本

echo "================================"
echo "HandsOff Webhook测试"
echo "================================"
echo ""

# 配置
API_URL="http://localhost:8080/api/webhook"
WEBHOOK_SECRET="test-secret-token"

# GitLab MR Webhook Payload示例（精简版）
PAYLOAD='{
  "object_kind": "merge_request",
  "event_type": "merge_request",
  "user": {
    "id": 1,
    "name": "Test User",
    "username": "testuser",
    "email": "test@example.com"
  },
  "project": {
    "id": 123,
    "name": "test-repo",
    "web_url": "https://gitlab.com/test/repo",
    "path_with_namespace": "test/repo",
    "default_branch": "main"
  },
  "object_attributes": {
    "id": 456,
    "iid": 42,
    "target_branch": "main",
    "source_branch": "feature/test",
    "title": "Test MR",
    "description": "This is a test merge request",
    "state": "opened",
    "action": "open",
    "url": "https://gitlab.com/test/repo/-/merge_requests/42",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z",
    "last_commit": {
      "id": "abc123",
      "message": "Test commit",
      "timestamp": "2024-01-01T11:55:00Z",
      "url": "https://gitlab.com/test/repo/-/commit/abc123",
      "author": {
        "name": "Test User",
        "email": "test@example.com"
      }
    }
  }
}'

echo "测试1: 发送GitLab MR Webhook（带Secret）"
echo "----------------------------------------"
curl -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "X-Gitlab-Token: $WEBHOOK_SECRET" \
  -H "X-Gitlab-Event: Merge Request Hook" \
  -d "$PAYLOAD" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s | jq .

echo ""
echo "测试2: 发送无效签名的Webhook"
echo "----------------------------------------"
curl -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "X-Gitlab-Token: invalid-token" \
  -H "X-Gitlab-Event: Merge Request Hook" \
  -d "$PAYLOAD" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s | jq .

echo ""
echo "测试3: 发送非MR事件（Push Event）"
echo "----------------------------------------"
PUSH_PAYLOAD='{"object_kind":"push","event_type":"push"}'
curl -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -H "X-Gitlab-Token: $WEBHOOK_SECRET" \
  -H "X-Gitlab-Event: Push Hook" \
  -d "$PUSH_PAYLOAD" \
  -w "\nHTTP Status: %{http_code}\n" \
  -s | jq .

echo ""
echo "================================"
echo "测试完成"
echo "================================"
echo ""
echo "提示："
echo "1. 确保API服务器运行在 http://localhost:8080"
echo "2. 确保Redis服务运行中（Asynq队列需要）"
echo "3. 确保数据库中已导入仓库（platform_repo_id=123）"
echo "4. 确保仓库已配置LLM模型"
echo ""
