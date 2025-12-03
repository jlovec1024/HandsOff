import { useState, useEffect } from "react";
import { Form, Input, Button, Card, message, Space, Tag } from "antd";
import { CheckCircleOutlined, CloseCircleOutlined } from "@ant-design/icons";
import { platformApi } from "../../api/platform";
import type { GitPlatformConfig } from "../../types";

const GitLabConfig = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [testing, setTesting] = useState(false);
  const [config, setConfig] = useState<GitPlatformConfig | null>(null);

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    try {
      const response = await platformApi.getConfig();
      const data = response.data;

      if ("exists" in data && !data.exists) {
        // No config yet
        form.setFieldsValue({
          platform_type: "gitlab",
          is_active: true,
        });
      } else {
        const cfg = data as GitPlatformConfig;
        setConfig(cfg);
        form.setFieldsValue({
          ...cfg,
          access_token: "", // Don't show masked token
        });
      }
    } catch (error) {
      console.error("Failed to load config:", error);
    }
  };

  const handleSave = async (values: any) => {
    setLoading(true);
    try {
      const data: GitPlatformConfig = {
        platform_type: "gitlab",
        base_url: values.base_url,
        // Use undefined instead of empty string to signal "keep existing value"
        access_token: values.access_token || undefined,
        is_active: true,
      };

      const response = await platformApi.updateConfig(data);
      message.success("GitLab配置已保存");

      // Use returned config directly (avoids extra API call)
      if (response.data.config) {
        setConfig(response.data.config);
        form.setFieldsValue({
          ...response.data.config,
          access_token: "", // Don't show masked token
        });
      } else {
        // Fallback: if backend doesn't return config (backward compatibility)
        loadConfig();
      }
    } catch (error) {
      console.error("Failed to save config:", error);
      message.error("保存配置失败");
    } finally {
      setLoading(false);
    }
  };

  const handleTest = async () => {
    try {
      await form.validateFields();
      const values = form.getFieldsValue();

      // Validate that access_token is provided
      if (!values.access_token) {
        message.error("请输入 Access Token 以测试连接");
        return;
      }

      setTesting(true);

      // Send form data for testing (no need to save first)
      const testData = {
        platform_type: "gitlab",
        base_url: values.base_url,
        access_token: values.access_token,
      };

      const response = await platformApi.testConnection(testData);

      if (response.data.success) {
        message.success(response.data.message);
      } else {
        message.error(response.data.message || "连接测试失败");
      }
    } catch (error: any) {
      if (error.errorFields) {
        message.error("请先填写完整的配置信息");
      } else {
        const errorMsg = error?.response?.data?.error || "连接测试失败";
        message.error(errorMsg);
      }
    } finally {
      setTesting(false);
    }
  };

  return (
    <Card>
      <Form
        form={form}
        layout="vertical"
        onFinish={handleSave}
        initialValues={{
          platform_type: "gitlab",
          is_active: true,
        }}
      >
        <Form.Item
          label="GitLab URL"
          name="base_url"
          rules={[
            { required: true, message: "请输入GitLab URL" },
            { type: "url", message: "请输入有效的URL" },
          ]}
        >
          <Input placeholder="https://gitlab.com" />
        </Form.Item>

        <Form.Item
          label="Access Token"
          name="access_token"
          rules={[
            {
              required: !config,
              message: "请输入Personal Access Token",
            },
          ]}
          extra="需要以下权限：api, read_api, read_repository"
        >
          <Input.Password
            placeholder={
              config ? "留空则保持不变" : "输入Personal Access Token"
            }
          />
        </Form.Item>

        {config && config.last_test_status && (
          <Form.Item label="上次测试">
            <Space>
              {config.last_test_status === "success" ? (
                <Tag icon={<CheckCircleOutlined />} color="success">
                  {config.last_test_message}
                </Tag>
              ) : (
                <Tag icon={<CloseCircleOutlined />} color="error">
                  {config.last_test_message}
                </Tag>
              )}
              {config.last_tested_at && (
                <span style={{ color: "#999", fontSize: 12 }}>
                  {new Date(config.last_tested_at).toLocaleString()}
                </span>
              )}
            </Space>
          </Form.Item>
        )}

        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={loading}>
              保存配置
            </Button>
            <Button onClick={handleTest} loading={testing || loading}>
              测试连接
            </Button>
          </Space>
        </Form.Item>
      </Form>
    </Card>
  );
};

export default GitLabConfig;
