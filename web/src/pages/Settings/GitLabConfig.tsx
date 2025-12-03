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
        access_token: values.access_token || "***masked***",
        is_active: true,
      };

      await platformApi.updateConfig(data);
      message.success("GitLab配置已保存");
      loadConfig();
    } catch (error) {
      console.error("Failed to save config:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleTest = async () => {
    try {
      await form.validateFields();
      const values = form.getFieldsValue();

      setTesting(true);

      const testData: GitPlatformConfig = {
        platform_type: "gitlab",
        base_url: values.base_url,
        access_token: values.access_token || config?.access_token || "",
        is_active: true,
      };

      const response = await platformApi.testConnection();

      if (response.data.success) {
        message.success(response.data.message);
        await loadConfig();
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
