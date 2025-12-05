import type { Repository } from "../../types";
import { useState, useEffect } from "react";
import { Table, Button, Space, Modal, Select, message, Popconfirm, Tag } from "antd";
import {
  PlusOutlined,
  ReloadOutlined,
  DeleteOutlined,
  SettingOutlined,
} from "@ant-design/icons";
import { repositoryApi } from "../../api/repository";
import ImportModal from "./ImportModal";

const RepositoryList = () => {
  const [repositories, setRepositories] = useState<Repository[]>([]);
  const [providers] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [importModalVisible, setImportModalVisible] = useState(false);
  const [configModalVisible, setConfigModalVisible] = useState(false);
  const [selectedRepo, setSelectedRepo] = useState<Repository | null>(null);
  const [selectedProviderID, setSelectedProviderID] = useState<number | null>(null);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });

  useEffect(() => {
    loadRepositories();
    loadModels();
  }, []);

  const loadRepositories = async (page: number = 1) => {
    setLoading(true);
    try {
      const response = await repositoryApi.list(page, pagination.pageSize);
      const { repositories: repos, total } = response.data;
      setRepositories(Array.isArray(repos) ? repos : []);
      setPagination((prev) => ({ ...prev, current: page, total }));
    } catch (error) {
      console.error("Failed to load repositories:", error);
    } finally {
      setLoading(false);
    }
  };

  const loadModels = async () => {
    try {
    } catch (error) {
      console.error("Failed to load models:", error);
    }
  };

  const handleImport = () => {
    setImportModalVisible(true);
  };

  const handleImportSuccess = () => {
    setImportModalVisible(false);
    loadRepositories();
  };

  const handleConfigLLM = (repo: Repository) => {
    setSelectedRepo(repo);
    setSelectedProviderID(repo.llm_provider_id || null);
    setConfigModalVisible(true);
  };

  const handleSaveLLMConfig = async () => {
    if (!selectedRepo) return;

    try {
      await repositoryApi.updateLLMProvider(selectedRepo.id!, selectedProviderID);
      message.success("LLM配置已更新");
      setConfigModalVisible(false);
      loadRepositories();
    } catch (error) {
      console.error("Failed to update LLM config:", error);
    }
  };

  const handleDelete = async (id: number) => {
    try {
      await repositoryApi.delete(id);
      message.success("仓库已删除");
      loadRepositories();
    } catch (error) {
      console.error("Failed to delete repository:", error);
    }
  };

  const columns = [
    {
      title: "仓库名称",
      dataIndex: "name",
      key: "name",
      render: (name: string, record: Repository) => (
        <div>
          <div style={{ fontWeight: 500 }}>{name}</div>
          <div style={{ fontSize: 12, color: "#999" }}>{record.full_path}</div>
        </div>
      ),
    },
    {
      title: "默认分支",
      dataIndex: "default_branch",
      key: "default_branch",
      width: 120,
    },
    {
      title: "Webhook",
      dataIndex: "webhook_id",
      key: "webhook",
      width: 120,
      render: (webhookID: number | null) =>
        webhookID ? <Tag color="success">已配置</Tag> : <Tag>未配置</Tag>,
    },
    {
      title: "状态",
      dataIndex: "is_active",
      key: "is_active",
      width: 100,
      render: (active: boolean) => (
        <Tag color={active ? "success" : "default"}>
          {active ? "启用" : "禁用"}
        </Tag>
      ),
    },
    {
      title: "操作",
      key: "action",
      width: 180,
      render: (_: any, record: Repository) => (
        <Space size="small">
          <Button
            size="small"
            icon={<SettingOutlined />}
            onClick={() => handleConfigLLM(record)}
          >
            配置
          </Button>
          <Popconfirm
            title="确定删除此仓库吗？"
            description="删除后将从GitLab移除Webhook配置"
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
    <div style={{ padding: 24 }}>
      <div
        style={{
          marginBottom: 16,
          display: "flex",
          justifyContent: "space-between",
        }}
      >
        <h2>仓库管理</h2>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={() => loadRepositories()}>
            刷新
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleImport}>
            导入仓库
          </Button>
        </Space>
      </div>

      <Table
        columns={columns}
        dataSource={repositories}
        rowKey="id"
        loading={loading}
        pagination={{
          current: pagination.current,
          pageSize: pagination.pageSize,
          total: pagination.total,
          onChange: loadRepositories,
        }}
      />

      <ImportModal
        visible={importModalVisible}
        onCancel={() => setImportModalVisible(false)}
        onSuccess={handleImportSuccess}
      />
    </div>
  );
};

export default RepositoryList;
