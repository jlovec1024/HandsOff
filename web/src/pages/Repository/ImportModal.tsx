import { useState, useEffect } from 'react';
import { Modal, Table, Input, message, Spin } from 'antd';
import type { TableColumnsType } from 'antd';
import { repositoryApi } from '../../api/repository';
import type { GitLabRepository } from '../../types';

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
  const [webhookURL, setWebhookURL] = useState('http://your-server.com/api/webhook');
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 20,
    total: 0,
  });

  useEffect(() => {
    if (visible) {
      loadRepositories();
    }
  }, [visible]);

  const loadRepositories = async (page: number = 1) => {
    setLoading(true);
    try {
      const response = await repositoryApi.listFromGitLab(page, pagination.pageSize);
      const { repositories: repos, total_pages } = response.data;
      setRepositories(repos);
      setPagination((prev) => ({
        ...prev,
        current: page,
        total: total_pages * pagination.pageSize,
      }));
    } catch (error) {
      console.error('Failed to load GitLab repositories:', error);
      message.error('获取GitLab仓库列表失败');
    } finally {
      setLoading(false);
    }
  };

  const handleImport = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请至少选择一个仓库');
      return;
    }

    if (!webhookURL) {
      message.warning('请输入Webhook回调URL');
      return;
    }

    setImporting(true);
    try {
      const repositoryIDs = selectedRowKeys.map((key) => Number(key));
      await repositoryApi.batchImport(repositoryIDs, webhookURL);
      message.success(`成功导入 ${selectedRowKeys.length} 个仓库`);
      setSelectedRowKeys([]);
      onSuccess();
    } catch (error) {
      console.error('Failed to import repositories:', error);
    } finally {
      setImporting(false);
    }
  };

  const columns: TableColumnsType<GitLabRepository> = [
    {
      title: '仓库名称',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record: GitLabRepository) => (
        <div>
          <div style={{ fontWeight: 500 }}>{name}</div>
          <div style={{ fontSize: 12, color: '#999' }}>{record.full_path}</div>
        </div>
      ),
    },
    {
      title: '默认分支',
      dataIndex: 'default_branch',
      key: 'default_branch',
      width: 120,
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
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
      <div style={{ marginBottom: 16 }}>
        <div style={{ marginBottom: 8 }}>Webhook回调URL:</div>
        <Input
          value={webhookURL}
          onChange={(e) => setWebhookURL(e.target.value)}
          placeholder="http://your-server.com/api/webhook"
        />
        <div style={{ marginTop: 4, fontSize: 12, color: '#999' }}>
          导入时将自动为每个仓库配置此Webhook URL
        </div>
      </div>

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
            onChange: loadRepositories,
          }}
        />
      </Spin>
    </Modal>
  );
};

export default ImportModal;
