import { useEffect, useState } from "react";
import {
  Card,
  Descriptions,
  Tag,
  Form,
  Input,
  Button,
  message,
  Space,
} from "antd";
import { SaveOutlined } from "@ant-design/icons";
import { systemApi } from "../../api/system";
import type { SystemWebhookConfig } from "../../types";

const SystemConfig = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [fetching, setFetching] = useState(false);
  const [config, setConfig] = useState<SystemWebhookConfig | null>(null);

  useEffect(() => {
    loadWebhookConfig();
  }, []);

  const loadWebhookConfig = async () => {
    setFetching(true);
    try {
      const response = await systemApi.getWebhookConfig();
      setConfig(response.data);
      form.setFieldsValue(response.data);
    } catch (error) {
      console.error("Failed to load webhook config:", error);
      message.error("获取Webhook配置失败");
    } finally {
      setFetching(false);
    }
  };

  const handleSave = async (values: SystemWebhookConfig) => {
    setLoading(true);
    try {
      await systemApi.updateWebhookConfig(values);
      message.success("Webhook配置已保存");
      setConfig(values);
    } catch (error) {
      console.error("Failed to save webhook config:", error);
      message.error("保存Webhook配置失败");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <Card title="系统信息" style={{ marginBottom: 16 }}>
        <Descriptions column={2}>
          <Descriptions.Item label="版本">1.0.0-mvp</Descriptions.Item>
          <Descriptions.Item label="环境">
            <Tag color="blue">Development</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="API地址">
            {import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api"}
          </Descriptions.Item>
          <Descriptions.Item label="前端版本">React 18</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title="Webhook配置" loading={fetching}>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSave}
          initialValues={{ webhook_callback_url: "" }}
        >
          <Form.Item
            label="Webhook 回调 URL"
            name="webhook_callback_url"
            rules={[
              { required: true, message: "请输入Webhook回调URL" },
              { type: "url", message: "请输入有效的URL" },
            ]}
            extra="此URL将用于所有新添加的仓库。例如: https://your-domain.com/api/webhook"
          >
            <Input placeholder="https://your-domain.com/api/webhook" />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                loading={loading}
                icon={<SaveOutlined />}
              >
                保存配置
              </Button>
              {config && config.webhook_callback_url && (
                <span style={{ color: "#52c41a", marginLeft: 8 }}>
                  ✓ 已配置
                </span>
              )}
            </Space>
          </Form.Item>
        </Form>

        <div
          style={{
            marginTop: 16,
            padding: 12,
            background: "#f5f5f5",
            borderRadius: 4,
          }}
        >
          <strong>💡 使用说明：</strong>
          <ul style={{ marginTop: 8, marginBottom: 0 }}>
            <li>配置后，添加仓库时将自动使用此URL创建Webhook</li>
            <li>修改此配置不会影响已添加的仓库</li>
            <li>如需更新已有仓库的Webhook，请在仓库列表中操作</li>
          </ul>
        </div>
      </Card>
    </div>
  );
};

export default SystemConfig;
