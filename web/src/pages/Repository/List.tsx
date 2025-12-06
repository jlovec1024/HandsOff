import type { Repository } from "../../types";
import { useState, useEffect } from "react";
import {
  Table,
  Button,
  Space,
  Modal,
  message,
  Popconfirm,
  Tag,
  Input,
  Tooltip,
} from "antd";
import {
  PlusOutlined,
  ReloadOutlined,
  DeleteOutlined,
  ExclamationCircleOutlined,
  CheckCircleOutlined,
  MinusCircleOutlined,
  WarningOutlined,
  SearchOutlined,
} from "@ant-design/icons";
import { repositoryApi } from "../../api/repository";
import { formatTime } from "../../utils/time";
import ImportModal from "./ImportModal";
import WebhookConfigModal from "./WebhookConfigModal";

const RepositoryList = () => {
  const [repositories, setRepositories] = useState<Repository[]>([]);
  const [filteredRepositories, setFilteredRepositories] = useState<
    Repository[]
  >([]);
  const [searchText, setSearchText] = useState("");
  const [loading, setLoading] = useState(false);
  const [importModalVisible, setImportModalVisible] = useState(false);
  const [webhookConfigModalVisible, setWebhookConfigModalVisible] =
    useState(false);
  const [selectedRepo, setSelectedRepo] = useState<Repository | null>(null);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });

  // Filter repositories based on search text
  useEffect(() => {
    if (searchText.trim() === "") {
      setFilteredRepositories(repositories);
    } else {
      const filtered = repositories.filter(
        (repo) =>
          repo.name.toLowerCase().includes(searchText.toLowerCase()) ||
          repo.full_path.toLowerCase().includes(searchText.toLowerCase())
      );
      setFilteredRepositories(filtered);
    }
  }, [searchText, repositories]);

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

  const handleWebhookConfig = (repo: Repository) => {
    setSelectedRepo(repo);
    setWebhookConfigModalVisible(true);
  };

  const handleWebhookConfigSuccess = () => {
    loadRepositories();
  };

  const handleTestWebhook = async (repo: Repository) => {
    if (!repo.id) return;

    try {
      const response = await repositoryApi.testWebhook(repo.id);
      if (response.data.status === "success") {
        message.success("Webhook 测试成功");
      } else {
        message.error(`Webhook 测试失败: ${response.data.message}`);
      }
      loadRepositories();
    } catch (error: any) {
      message.error(error.response?.data?.error || "测试失败，请检查网络连接");
    }
  };

  const handleRecreateWebhook = (repo: Repository) => {
    Modal.confirm({
      title: "确认重新配置 Webhook？",
      content: "将删除旧的 Webhook 并使用系统配置创建新的 Webhook",
      okText: "确认",
      cancelText: "取消",
      okType: "danger",
      onOk: async () => {
        if (!repo.id) return;

        try {
          await repositoryApi.recreateWebhook(repo.id);
          message.success("Webhook 已重新配置");
          loadRepositories();
        } catch (error: any) {
          message.error(error.response?.data?.error || "配置失败，请稍后重试");
        }
      },
    });
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
      title: "Webhook 状态",
      dataIndex: "webhook_id",
      key: "webhook",
      width: 200,
      render: (_: any, record: Repository) => {
        const { webhook_id, last_webhook_test_status, last_webhook_test_at } =
          record;

        if (!webhook_id) {
          return (
            <div>
              <Tag icon={<MinusCircleOutlined />}>未配置</Tag>
              <Button
                size="small"
                type="link"
                onClick={() => handleWebhookConfig(record)}
                style={{ padding: 0, marginLeft: 8 }}
              >
                配置
              </Button>
            </div>
          );
        }

        if (last_webhook_test_status === "success") {
          return (
            <div>
              <Tag icon={<CheckCircleOutlined />} color="success">
                正常
              </Tag>
              <div style={{ fontSize: 11, color: "#999", marginTop: 2 }}>
                {formatTime(last_webhook_test_at)}
              </div>
            </div>
          );
        }

        if (last_webhook_test_status === "failed") {
          return (
            <div>
              <Tag icon={<ExclamationCircleOutlined />} color="error">
                异常
              </Tag>
              <div style={{ marginTop: 4 }}>
                <Button
                  size="small"
                  type="link"
                  danger
                  onClick={() => handleRecreateWebhook(record)}
                  style={{ padding: 0 }}
                >
                  重新配置
                </Button>
              </div>
            </div>
          );
        }

        // never or null
        return (
          <div>
            <Tag icon={<WarningOutlined />} color="warning">
              未检测
            </Tag>
            <div style={{ marginTop: 4 }}>
              <Button
                size="small"
                type="link"
                onClick={() => handleTestWebhook(record)}
                style={{ padding: 0 }}
              >
                立即测试
              </Button>
            </div>
          </div>
        );
      },
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
      width: 120,
      render: (_: any, record: Repository) => (
        <Space size="small">
          <Tooltip title="查看 Webhook 详情">
            <Button
              size="small"
              type="text"
              onClick={() => handleWebhookConfig(record)}
            >
              详情
            </Button>
          </Tooltip>
          <Popconfirm
            title="确定删除此仓库吗？"
            description="删除后将从GitLab移除Webhook配置"
            onConfirm={() => handleDelete(record.id!)}
            okText="确定"
            cancelText="取消"
          >
            <Button size="small" type="text" danger icon={<DeleteOutlined />}>
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
          alignItems: "center",
        }}
      >
        <h2>仓库管理</h2>
        <Space>
          <Input
            placeholder="搜索仓库名称或路径"
            prefix={<SearchOutlined />}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            style={{ width: 250 }}
            allowClear
          />
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
        dataSource={filteredRepositories}
        rowKey="id"
        loading={loading}
        pagination={{
          current: pagination.current,
          pageSize: pagination.pageSize,
          total: pagination.total,
          showTotal: (total) => `共 ${total} 个仓库`,
          showSizeChanger: true,
          showQuickJumper: true,
          onChange: loadRepositories,
        }}
      />

      <ImportModal
        visible={importModalVisible}
        onCancel={() => setImportModalVisible(false)}
        onSuccess={handleImportSuccess}
      />

      <WebhookConfigModal
        visible={webhookConfigModalVisible}
        repository={selectedRepo}
        onCancel={() => setWebhookConfigModalVisible(false)}
        onSuccess={handleWebhookConfigSuccess}
      />
    </div>
  );
};

export default RepositoryList;
