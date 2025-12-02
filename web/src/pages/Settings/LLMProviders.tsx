import { useState, useEffect } from "react";
import {
  Table,
  Button,
  Modal,
  Form,
  Input,
  Select,
  message,
  Space,
  Tag,
  Popconfirm,
} from "antd";
import {
  PlusOutlined,
  EditOutlined,
  DeleteOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from "@ant-design/icons";
import { llmApi } from "../../api/llm";
import type { LLMProvider } from "../../types";

const LLMProviders = () => {
  const [providers, setProviders] = useState<LLMProvider[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingProvider, setEditingProvider] = useState<LLMProvider | null>(
    null
  );
  const [form] = Form.useForm();

  useEffect(() => {
    loadProviders();
  }, []);

  const loadProviders = async () => {
    setLoading(true);
    try {
      const response = await llmApi.listProviders();
      setProviders(Array.isArray(response.data) ? response.data : []);
    } catch (error) {
      console.error("Failed to load providers:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingProvider(null);
    form.resetFields();
    form.setFieldsValue({
      is_active: true,
    });
    setModalVisible(true);
  };

  const handleEdit = (provider: LLMProvider) => {
    setEditingProvider(provider);
    form.setFieldsValue({
      ...provider,
      api_key: "", // Don't show masked key
    });
    setModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    try {
      await llmApi.deleteProvider(id);
      message.success("供应商已删除");
      loadProviders();
    } catch (error) {
      console.error("Failed to delete provider:", error);
    }
  };

  const handleTest = async (id: number) => {
    try {
      const response = await llmApi.testProvider(id);
      if (response.data.success) {
        message.success(response.data.message);
        loadProviders();
      }
    } catch (error) {
      console.error("Test failed:", error);
    }
  };

  const handleSubmit = async (values: any) => {
    try {
      const data = {
        ...values,
        api_key: values.api_key || "***masked***",
      };

      if (editingProvider) {
        await llmApi.updateProvider(editingProvider.id!, data);
        message.success("供应商已更新");
      } else {
        await llmApi.createProvider(data);
        message.success("供应商已创建");
      }

      setModalVisible(false);
      loadProviders();
    } catch (error) {
      console.error("Failed to save provider:", error);
    }
  };

  const columns = [
    {
      title: "名称",
      dataIndex: "name",
      key: "name",
    },
    {
      title: "类型",
      dataIndex: "type",
      key: "type",
      render: (type: string) => {
        const colors: Record<string, string> = {
          openai: "blue",
          deepseek: "purple",
          claude: "orange",
          gemini: "green",
        };
        return <Tag color={colors[type] || "default"}>{type}</Tag>;
      },
    },
    {
      title: "Base URL",
      dataIndex: "base_url",
      key: "base_url",
      ellipsis: true,
    },
    {
      title: "状态",
      dataIndex: "is_active",
      key: "is_active",
      render: (active: boolean) => (
        <Tag color={active ? "success" : "default"}>
          {active ? "启用" : "禁用"}
        </Tag>
      ),
    },
    {
      title: "测试状态",
      dataIndex: "last_test_status",
      key: "last_test_status",
      render: (status: string) => {
        if (!status) return "-";
        return status === "success" ? (
          <Tag icon={<CheckCircleOutlined />} color="success">
            成功
          </Tag>
        ) : (
          <Tag icon={<CloseCircleOutlined />} color="error">
            失败
          </Tag>
        );
      },
    },
    {
      title: "操作",
      key: "action",
      width: 200,
      render: (_: any, record: LLMProvider) => (
        <Space size="small">
          <Button size="small" onClick={() => handleTest(record.id!)}>
            测试
          </Button>
          <Button
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定删除此供应商吗？"
            onConfirm={() => handleDelete(record.id!)}
            okText="确定"
            cancelText="取消"
          >
            <Button size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div>
      <div style={{ marginBottom: 16 }}>
        <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
          添加供应商
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={providers}
        rowKey="id"
        loading={loading}
      />

      <Modal
        title={editingProvider ? "编辑供应商" : "添加供应商"}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item
            label="名称"
            name="name"
            rules={[{ required: true, message: "请输入名称" }]}
          >
            <Input placeholder="如：OpenAI, DeepSeek" />
          </Form.Item>

          <Form.Item
            label="类型"
            name="type"
            rules={[{ required: true, message: "请选择类型" }]}
          >
            <Select
              options={[
                { label: "OpenAI", value: "openai" },
                { label: "DeepSeek", value: "deepseek" },
                { label: "Claude", value: "claude" },
                { label: "Google Gemini", value: "gemini" },
                { label: "Ollama", value: "ollama" },
              ]}
            />
          </Form.Item>

          <Form.Item
            label="Base URL"
            name="base_url"
            rules={[
              { required: true, message: "请输入Base URL" },
              { type: "url", message: "请输入有效的URL" },
            ]}
          >
            <Input placeholder="https://api.openai.com/v1" />
          </Form.Item>

          <Form.Item
            label="API Key"
            name="api_key"
            rules={[{ required: !editingProvider, message: "请输入API Key" }]}
          >
            <Input.Password
              placeholder={editingProvider ? "留空则保持不变" : "输入API Key"}
            />
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                保存
              </Button>
              <Button onClick={() => setModalVisible(false)}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default LLMProviders;
