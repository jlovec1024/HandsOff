import { useState, useEffect } from "react";
import {
  Table,
  Button,
  Modal,
  Form,
  Input,
  InputNumber,
  Select,
  message,
  Space,
  Tag,
  Popconfirm,
} from "antd";
import { PlusOutlined, EditOutlined, DeleteOutlined } from "@ant-design/icons";
import { llmApi } from "../../api/llm";
import type { LLMModel, LLMProvider } from "../../types";

const LLMModels = () => {
  const [models, setModels] = useState<LLMModel[]>([]);
  const [providers, setProviders] = useState<LLMProvider[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [editingModel, setEditingModel] = useState<LLMModel | null>(null);
  const [form] = Form.useForm();

  useEffect(() => {
    loadProviders();
    loadModels();
  }, []);

  const loadProviders = async () => {
    try {
      const response = await llmApi.listProviders();
      setProviders(Array.isArray(response.data) ? response.data : []);
    } catch (error) {
      console.error("Failed to load providers:", error);
    }
  };

  const loadModels = async () => {
    setLoading(true);
    try {
      const response = await llmApi.listModels();
      setModels(Array.isArray(response.data) ? response.data : []);
    } catch (error) {
      console.error("Failed to load models:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setEditingModel(null);
    form.resetFields();
    form.setFieldsValue({
      is_active: true,
      max_tokens: 4096,
      temperature: 0.7,
    });
    setModalVisible(true);
  };

  const handleEdit = (model: LLMModel) => {
    setEditingModel(model);
    form.setFieldsValue(model);
    setModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    try {
      await llmApi.deleteModel(id);
      message.success("模型已删除");
      loadModels();
    } catch (error) {
      console.error("Failed to delete model:", error);
    }
  };

  const handleSubmit = async (values: any) => {
    try {
      if (editingModel) {
        await llmApi.updateModel(editingModel.id!, values);
        message.success("模型已更新");
      } else {
        await llmApi.createModel(values);
        message.success("模型已创建");
      }

      setModalVisible(false);
      loadModels();
    } catch (error) {
      console.error("Failed to save model:", error);
    }
  };

  const columns = [
    {
      title: "显示名称",
      dataIndex: "display_name",
      key: "display_name",
    },
    {
      title: "模型名称",
      dataIndex: "model_name",
      key: "model_name",
    },
    {
      title: "供应商",
      dataIndex: ["provider", "name"],
      key: "provider",
    },
    {
      title: "最大Tokens",
      dataIndex: "max_tokens",
      key: "max_tokens",
    },
    {
      title: "Temperature",
      dataIndex: "temperature",
      key: "temperature",
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
      title: "操作",
      key: "action",
      width: 150,
      render: (_: any, record: LLMModel) => (
        <Space size="small">
          <Button
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEdit(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确定删除此模型吗？"
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
          添加模型
        </Button>
      </div>

      <Table
        columns={columns}
        dataSource={models}
        rowKey="id"
        loading={loading}
      />

      <Modal
        title={editingModel ? "编辑模型" : "添加模型"}
        open={modalVisible}
        onCancel={() => setModalVisible(false)}
        footer={null}
        width={600}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{
            is_active: true,
            max_tokens: 4096,
            temperature: 0.7,
          }}
        >
          <Form.Item
            label="供应商"
            name="provider_id"
            rules={[{ required: true, message: "请选择供应商" }]}
          >
            <Select
              placeholder="选择供应商"
              options={providers.map((p) => ({
                label: p.name,
                value: p.id,
              }))}
            />
          </Form.Item>

          <Form.Item
            label="模型名称"
            name="model_name"
            rules={[{ required: true, message: "请输入模型名称" }]}
            extra="API调用时使用的模型标识，如：gpt-4, deepseek-chat"
          >
            <Input placeholder="gpt-4" />
          </Form.Item>

          <Form.Item
            label="显示名称"
            name="display_name"
            rules={[{ required: true, message: "请输入显示名称" }]}
          >
            <Input placeholder="GPT-4" />
          </Form.Item>

          <Form.Item label="描述" name="description">
            <Input.TextArea rows={3} placeholder="可选的模型描述" />
          </Form.Item>

          <Form.Item
            label="最大Tokens"
            name="max_tokens"
            rules={[{ required: true, message: "请输入最大Tokens" }]}
          >
            <InputNumber min={1} max={100000} style={{ width: "100%" }} />
          </Form.Item>

          <Form.Item
            label="Temperature"
            name="temperature"
            rules={[{ required: true, message: "请输入Temperature" }]}
          >
            <InputNumber min={0} max={2} step={0.1} style={{ width: "100%" }} />
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

export default LLMModels;
