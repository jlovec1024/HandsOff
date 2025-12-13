import { useState, useEffect } from "react";
import { Modal, Table, message, Spin, Alert, Input } from "antd";
import { SearchOutlined } from "@ant-design/icons";
import type { TableColumnsType } from "antd";
import { repositoryApi } from "../../api/repository";
import { systemApi } from "../../api/system";
import type { GitLabRepository } from "../../types";

interface ImportModalProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
}

const ImportModal = ({ visible, onCancel, onSuccess }: ImportModalProps) => {
  const [repositories, setRepositories] = useState<GitLabRepository[]>([]);
  const [loading, setLoading] = useState(false);
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [importing, setImporting] = useState(false);
  const [systemWebhookUrl, setSystemWebhookUrl] = useState("");
  const [searchText, setSearchText] = useState("");
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });

  // 加载仓库列表：接受 search 参数，避免闭包问题
  const loadRepositories = async (page: number = 1, search: string = "") => {
    setLoading(true);
    try {
      const response = await repositoryApi.listFromGitLab(
        page,
        pagination.pageSize,
        search
      );
      const { repositories: repos, total_pages } = response.data;
      setRepositories(Array.isArray(repos) ? repos : []);
      setPagination((prev) => ({
        ...prev,
        current: page,
        total: total_pages * pagination.pageSize,
      }));
    } catch (error) {
      console.error("Failed to load GitLab repositories:", error);
      message.error("获取GitLab仓库列表失败");
    } finally {
      setLoading(false);
    }
  };

  // Modal 打开时：重置状态并加载全量数据
  useEffect(() => {
    if (visible) {
      setSearchText("");
      loadSystemWebhookUrl();
      loadRepositories(1, "");
    }
  }, [visible]);

  // 搜索防抖：输入变化后 500ms 触发，清空时立即恢复
  useEffect(() => {
    if (!visible) return;

    const timer = setTimeout(
      () => {
        loadRepositories(1, searchText);
      },
      searchText ? 500 : 0
    ); // 有输入时防抖 500ms，清空时立即加载

    return () => clearTimeout(timer);
  }, [searchText]);

  const loadSystemWebhookUrl = async () => {
    try {
      const response = await systemApi.getWebhookConfig();
      setSystemWebhookUrl(response.data.webhook_callback_url);
    } catch (error) {
      console.error("Failed to load system webhook URL:", error);
      message.warning("未配置系统 Webhook URL，请先在系统设置中配置");
    }
  };

  const handleImport = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning("请至少选择一个仓库");
      return;
    }

    if (!systemWebhookUrl) {
      message.error("系统 Webhook URL 未配置，请先在系统设置中配置");
      return;
    }

    setImporting(true);
    try {
      const repositoryIDs = selectedRowKeys.map((key) => Number(key));
      // Pass empty string, backend will use system config
      await repositoryApi.batchImport(repositoryIDs, "");
      message.success(`成功导入 ${selectedRowKeys.length} 个仓库`);
      setSelectedRowKeys([]);
      onSuccess();
    } catch (error) {
      console.error("Failed to import repositories:", error);
      message.error("导入仓库失败");
    } finally {
      setImporting(false);
    }
  };

  const columns: TableColumnsType<GitLabRepository> = [
    {
      title: "仓库名称",
      dataIndex: "name",
      key: "name",
      render: (name: string, record: GitLabRepository) => (
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
      title: "描述",
      dataIndex: "description",
      key: "description",
      ellipsis: true,
    },
  ];

  const rowSelection = {
    selectedRowKeys,
    onChange: setSelectedRowKeys,
  };

  return (
    <Modal
      title="从GitLab导入仓库"
      open={visible}
      onCancel={onCancel}
      onOk={handleImport}
      okText={`导入 (${selectedRowKeys.length})`}
      cancelText="取消"
      width={800}
      confirmLoading={importing}
    >
      <Input
        placeholder="搜索仓库名称、路径或描述"
        prefix={<SearchOutlined />}
        value={searchText}
        onChange={(e) => setSearchText(e.target.value)}
        style={{ marginBottom: 16 }}
        allowClear
      />
      {systemWebhookUrl ? (
        <Alert
          message="系统 Webhook 配置"
          description={
            <div>
              <div style={{ marginBottom: 4 }}>将使用系统默认 Webhook URL:</div>
              <code
                style={{
                  fontSize: 12,
                  padding: "2px 6px",
                  background: "#f5f5f5",
                  borderRadius: 3,
                }}
              >
                {systemWebhookUrl}
              </code>
              <a
                href="/settings"
                style={{ marginLeft: 12, fontSize: 12 }}
                onClick={(e) => {
                  e.preventDefault();
                  window.open("/settings", "_blank");
                }}
              >
                在系统设置中修改
              </a>
            </div>
          }
          type="info"
          showIcon
          style={{ marginBottom: 16 }}
        />
      ) : (
        <Alert
          message="未配置系统 Webhook URL"
          description={
            <div>
              请先在{" "}
              <a
                href="/settings"
                onClick={(e) => {
                  e.preventDefault();
                  window.open("/settings", "_blank");
                }}
              >
                系统设置
              </a>{" "}
              中配置 Webhook URL
            </div>
          }
          type="warning"
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}

      <Spin spinning={loading}>
        <Table
          rowSelection={rowSelection}
          columns={columns}
          dataSource={repositories}
          rowKey="id"
          pagination={{
            current: pagination.current,
            pageSize: pagination.pageSize,
            total: pagination.total,
            onChange: (page) => loadRepositories(page, searchText),
          }}
        />
      </Spin>
    </Modal>
  );
};

export default ImportModal;
