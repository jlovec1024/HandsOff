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

// 表单数据类型：排除后端自动生成的字段，api_key 在编辑模式下可选
type LLMProviderFormValues = Omit<
  LLMProvider,
  | "id"
  | "created_at"
  | "updated_at"
  | "last_tested_at"
  | "last_test_status"
  | "last_test_message"
  | "api_key"
  | "model"
> & {
  api_key?: string; // 编辑模式可选，创建模式必填（由表单验证保证）
  model?: string; // 从下拉框选择的模型
};

const LLMProviders = () => {
  const [providers] = useState<LLMProvider[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingProvider, setEditingProvider] = useState<LLMProvider | null>(
    null
  );
  const [form] = Form.useForm();
  const [availableModels, setAvailableModels] = useState<string[]>([]);
  const [fetchingModels, setFetchingModels] = useState(false);

  useEffect(() => {
    loadProviders();
  }, []);

  const loadProviders = async () => {
    setLoading(true);
    try {
    } catch (error) {
      console.error("Failed to load providers:", error);
      message.error("加载供应商列表失败");
    } finally {
      setLoading(false);
    }
  };

  // 获取可用模型列表
  const handleFetchModels = async () => {
    const baseURL = form.getFieldValue("base_url");
    const apiKey = form.getFieldValue("api_key");

    if (!baseURL || !apiKey) {
      message.warning("请先填写 Base URL 和 API Key");
      return;
    }

    setFetchingModels(true);
    try {
      const response = await llmApi.fetchModels(baseURL, apiKey);
      setAvailableModels(response.data.models);
      message.success(`成功获取 ${response.data.models.length} 个模型`);
    } catch (error: any) {
      console.error("Failed to fetch models:", error);
      message.error(error.response?.data?.error || "获取模型列表失败");
      setAvailableModels([]);
    } finally {
      setFetchingModels(false);
    }
  };

  const handleCreate = () => {
    setEditingProvider(null);
    form.resetFields();
    form.setFieldsValue({
      is_active: true,
    });
    setAvailableModels([]);
    setModalVisible(true);
  };

  const handleEdit = (provider: LLMProvider) => {
    setEditingProvider(provider);
    form.setFieldsValue({
      ...provider,
      api_key: "", // Don't show masked key
    });
    setAvailableModels([]);
    setModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    try {
      await llmApi.deleteProvider(id);
      message.success("供应商已删除");
      loadProviders();
    } catch (error) {
      console.error("Failed to delete provider:", error);
      message.error("删除供应商失败");
    }
  };

  const handleTest = async (id: number) => {
    try {
      await llmApi.testProvider(id);
      message.success("测试成功");
      loadProviders();
    } catch (error) {
      console.error("Test failed:", error);
      message.error("测试供应商失败");
    }
  };

  const handleSubmit = async (values: LLMProviderFormValues) => {
    try {
      // 编辑模式：如果用户没输入新 Key，就删除 api_key 字段（后端保持原值）
      // 创建模式：api_key 必填
      const data = { ...values };
      if (editingProvider && !values.api_key) {
        delete data.api_key;
      }

      if (editingProvider) {
        await llmApi.updateProvider(editingProvider.id!, data);
        message.success("供应商已更新");
      } else {
        // 创建模式：api_key 由表单验证保证存在
        await llmApi.createProvider(data as Omit<LLMProvider, "id">);
        message.success("供应商已创建");
      }

      setModalVisible(false);
      loadProviders();
    } catch (error) {
      console.error("Failed to save provider:", error);
      message.error("保存供应商失败");
    }
  };

  const columns = [
    {
      title: "名称",
      dataIndex: "name",
      key: "name",
    },
    {
      title: "Base URL",
      dataIndex: "base_url",
      key: "base_url",
      ellipsis: true,
    },
    {
      title: "模型",
      dataIndex: "model",
      key: "model",
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
            <Input placeholder="如：OpenAI Official, DeepSeek China" />
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

          <Form.Item
            label="模型"
            name="model"
            rules={[{ required: true, message: "请选择模型" }]}
            extra={
              <Button
                type="link"
                size="small"
                loading={fetchingModels}
                onClick={handleFetchModels}
                style={{ padding: 0, marginTop: 4 }}
              >
                {availableModels.length > 0 ? "重新获取" : "获取可用模型"}
              </Button>
            }
          >
            <Select
              showSearch
              placeholder="请先获取模型列表"
              loading={fetchingModels}
              options={availableModels.map((model) => ({
                label: model,
                value: model,
              }))}
              disabled={availableModels.length === 0}
              filterOption={(input, option) =>
                (option?.label ?? "")
                  .toLowerCase()
                  .includes(input.toLowerCase())
              }
              notFoundContent={
                availableModels.length === 0
                  ? "请先点击上方按钮获取模型列表"
                  : "未找到匹配的模型"
              }
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
