import { useState } from "react";
import { Modal, Button, Alert, Descriptions, Tag, Space } from "antd";
import { repositoryApi } from "../../api/repository";
import type { Repository } from "../../types";
import { formatTime } from "../../utils/time";

interface WebhookConfigModalProps {
  visible: boolean;
  repository: Repository | null;
  onCancel: () => void;
  onSuccess: () => void;
}

const WebhookConfigModal = ({
  visible,
  repository,
  onCancel,
  onSuccess,
}: WebhookConfigModalProps) => {
  const [testing, setTesting] = useState(false);
  const [recreating, setRecreating] = useState(false);
  const [testResult, setTestResult] = useState<{
    status: string;
    message: string;
  } | null>(null);

  const handleTest = async () => {
    if (!repository?.id) return;

    setTesting(true);
    setTestResult(null);
    try {
      const response = await repositoryApi.testWebhook(repository.id);
      setTestResult(response.data);
      if (response.data.status === "success") {
        onSuccess(); // Refresh repository list
      }
    } catch (error: any) {
      setTestResult({
        status: "failed",
        message: error.response?.data?.error || "测试失败",
      });
    } finally {
      setTesting(false);
    }
  };

  const handleRecreate = async () => {
    if (!repository?.id) return;

    setRecreating(true);
    try {
      await repositoryApi.recreateWebhook(repository.id);
      Modal.success({
        title: "Webhook 已重新配置",
        content: "已删除旧的 Webhook 并创建新的配置",
      });
      onSuccess(); // Refresh repository list
      onCancel();
    } catch (error: any) {
      Modal.error({
        title: "配置失败",
        content: error.response?.data?.error || "重新配置失败，请稍后重试",
      });
    } finally {
      setRecreating(false);
    }
  };

  const getStatusTag = () => {
    if (!repository?.webhook_id) {
      return <Tag>未配置</Tag>;
    }

    const status = repository.last_webhook_test_status;
    if (status === "success") {
      return <Tag color="success">✓ 正常</Tag>;
    } else if (status === "failed") {
      return <Tag color="error">✗ 异常</Tag>;
    }
    return <Tag color="warning">⚠ 未检测</Tag>;
  };

  return (
    <Modal
      title={`Webhook 配置 - ${repository?.name || ""}`}
      open={visible}
      onCancel={onCancel}
      width={600}
      footer={[
        <Button key="cancel" onClick={onCancel}>
          关闭
        </Button>,
        <Button
          key="test"
          onClick={handleTest}
          loading={testing}
          disabled={!repository?.webhook_id}
        >
          测试连接
        </Button>,
        <Button
          key="recreate"
          type="primary"
          onClick={handleRecreate}
          loading={recreating}
          danger
        >
          重新配置
        </Button>,
      ]}
    >
      {repository && (
        <Space direction="vertical" size="large" style={{ width: "100%" }}>
          {/* Webhook 信息 */}
          <Descriptions column={1} bordered size="small">
            <Descriptions.Item label="仓库路径">
              {repository.full_path}
            </Descriptions.Item>
            <Descriptions.Item label="Webhook URL">
              <code style={{ fontSize: 12 }}>
                {repository.webhook_url || "未配置"}
              </code>
            </Descriptions.Item>
            <Descriptions.Item label="Webhook ID">
              {repository.webhook_id || "-"}
            </Descriptions.Item>
            <Descriptions.Item label="当前状态">
              {getStatusTag()}
            </Descriptions.Item>
            {repository.last_webhook_test_at && (
              <Descriptions.Item label="最后检测时间">
                {formatTime(repository.last_webhook_test_at)}
              </Descriptions.Item>
            )}
            {repository.last_webhook_test_error && (
              <Descriptions.Item label="错误信息">
                <span style={{ color: "#ff4d4f", fontSize: 12 }}>
                  {repository.last_webhook_test_error}
                </span>
              </Descriptions.Item>
            )}
          </Descriptions>

          {/* 测试结果 */}
          {testResult && (
            <Alert
              message={
                testResult.status === "success" ? "测试成功" : "测试失败"
              }
              description={testResult.message}
              type={testResult.status === "success" ? "success" : "error"}
              showIcon
            />
          )}

          {/* 说明 */}
          <Alert
            message="重新配置说明"
            description={
              <div style={{ fontSize: 12 }}>
                <p style={{ marginBottom: 8 }}>
                  点击"重新配置"将会：
                </p>
                <ol style={{ paddingLeft: 20, margin: 0 }}>
                  <li>删除 GitLab 上的旧 Webhook 配置</li>
                  <li>使用当前系统配置的 Webhook URL 创建新的 Webhook</li>
                  <li>更新仓库的 Webhook 状态</li>
                </ol>
                <p style={{ marginTop: 8, marginBottom: 0, color: "#faad14" }}>
                  ⚠️ 此操作适用于 Webhook URL 变更或配置异常的情况
                </p>
              </div>
            }
            type="info"
            showIcon
          />
        </Space>
      )}
    </Modal>
  );
};

export default WebhookConfigModal;
