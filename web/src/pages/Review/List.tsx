import { useEffect, useState } from 'react';
import { Card, Table, Tag, Input, Select, Button, Space } from 'antd';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import dayjs from 'dayjs';

const { Search } = Input;
const { Option } = Select;

const ReviewList = () => {
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState([]);
  const [pagination, setPagination] = useState({ page: 1, pageSize: 20, total: 0 });
  const [filters, setFilters] = useState({ status: '', author: '' });
  const navigate = useNavigate();

  useEffect(() => {
    fetchReviews();
  }, [pagination.page, filters]);

  const fetchReviews = async () => {
    setLoading(true);
    try {
      const params = {
        page: pagination.page,
        page_size: pagination.pageSize,
        ...filters,
      };
      const res = await axios.get('/api/reviews', { params });
      setData(res.data.data);
      setPagination(prev => ({ ...prev, total: res.data.pagination.total }));
    } catch (error) {
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', key: 'id', width: 80 },
    { title: 'Repository', dataIndex: ['repository', 'name'], key: 'repository' },
    { title: 'MR Title', dataIndex: 'mr_title', key: 'mr_title', ellipsis: true },
    { title: 'Author', dataIndex: 'mr_author', key: 'author' },
    { 
      title: 'Status', 
      dataIndex: 'status', 
      key: 'status',
      render: (status: string) => {
        const colors: any = { completed: 'success', failed: 'error', processing: 'processing', pending: 'default' };
        return <Tag color={colors[status]}>{status}</Tag>;
      }
    },
    { title: 'Score', dataIndex: 'score', key: 'score', render: (v: number) => v > 0 ? v : '-' },
    { title: 'Issues', dataIndex: 'issues_found', key: 'issues' },
    { title: 'Created', dataIndex: 'created_at', key: 'created_at', render: (d: string) => dayjs(d).format('YYYY-MM-DD HH:mm') },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Card title="Code Reviews">
        <Space style={{ marginBottom: 16 }}>
          <Select value={filters.status} onChange={(v) => setFilters({...filters, status: v})} style={{ width: 150 }}>
            <Option value="">All Status</Option>
            <Option value="completed">Completed</Option>
            <Option value="processing">Processing</Option>
            <Option value="pending">Pending</Option>
            <Option value="failed">Failed</Option>
          </Select>
          <Search placeholder="Search author" onSearch={(v) => setFilters({...filters, author: v})} style={{ width: 200 }} />
        </Space>
        <Table
          loading={loading}
          dataSource={data}
          columns={columns}
          rowKey="id"
          pagination={{
            current: pagination.page,
            pageSize: pagination.pageSize,
            total: pagination.total,
            onChange: (page) => setPagination(prev => ({ ...prev, page })),
          }}
          onRow={(record: any) => ({
            onClick: () => navigate(`/reviews/${record.id}`),
            style: { cursor: 'pointer' },
          })}
        />
      </Card>
    </div>
  );
};

export default ReviewList;
