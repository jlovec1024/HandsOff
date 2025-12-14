import { useEffect, useState } from "react";
import {
  Card,
  Table,
  Tag,
  Input,
  Select,
  Space,
  Tabs,
  Tooltip,
  Badge,
  Typography,
} from "antd";
import {
  LinkOutlined,
  FireOutlined,
  WarningOutlined,
  ThunderboltOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from "@ant-design/icons";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import dayjs from "dayjs";
import {
  getStatusConfig,
  renderStatusInline,
  getScoreColor,
  formatTokens,
} from "../../utils/statusConfig";

const { Search } = Input;
const { Option } = Select;
const { Text } = Typography;

interface ReviewRecord {
  id: number;
  repository?: { id: number; name: string; full_path: string };
  mr_title: string;
  mr_author: string;
  mr_web_url: string;
  status: string;
  score: number;
  issues_found: number;
  critical_issues_count: number;
  high_issues_count: number;
  total_tokens: number;
  created_at: string;
}

type TabFilter = "all" | "failed" | "critical" | "high_score";

// Tab filter configuration - data structure replaces if/else branches
const TAB_FILTERS: Record<TabFilter, Record<string, string | number>> = {
  all: {},
  failed: { status: "failed" },
  critical: { has_critical: "true" },
  high_score: { min_score: 80 },
};

const ReviewList = () => {
  const [loading, setLoading] = useState(false);
  const [data, setData] = useState<ReviewRecord[]>([]);
  const [pagination, setPagination] = useState({
    page: 1,
    pageSize: 20,
    total: 0,
  });
  const [filters, setFilters] = useState({ status: "", author: "" });
  const [activeTab, setActiveTab] = useState<TabFilter>("all");
  const navigate = useNavigate();

  useEffect(() => {
    fetchReviews();
  }, [pagination.page, filters, activeTab]);

  const fetchReviews = async () => {
    setLoading(true);
    try {
      // Use data structure instead of if/else branches
      const params: Record<string, string | number> = {
        page: pagination.page,
        page_size: pagination.pageSize,
        ...filters,
        ...TAB_FILTERS[activeTab],
      };

      const res = await axios.get("/api/reviews", { params });
      setData(Array.isArray(res.data.data) ? res.data.data : []);
      setPagination((prev) => ({
        ...prev,
        total: res.data.pagination?.total ?? 0,
      }));
    } catch (error) {
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const renderScore = (score: number, record: ReviewRecord) => {
    if (record.status !== "completed" || score === 0) {
      return <Text type="secondary">-</Text>;
    }
    return (
      <Tooltip title={`Quality Score: ${score}/100`}>
        <Badge
          count={score}
          style={{
            backgroundColor: getScoreColor(score),
            fontWeight: "bold",
          }}
          overflowCount={100}
        />
      </Tooltip>
    );
  };

  const renderIssues = (_: number, record: ReviewRecord) => {
    if (record.status !== "completed") {
      return <Text type="secondary">-</Text>;
    }

    const { critical_issues_count, high_issues_count, issues_found } = record;

    return (
      <Space size={4}>
        <Text>{issues_found}</Text>
        {critical_issues_count > 0 && (
          <Tooltip title={`${critical_issues_count} Critical`}>
            <Tag color="red" icon={<FireOutlined />}>
              {critical_issues_count}
            </Tag>
          </Tooltip>
        )}
        {high_issues_count > 0 && (
          <Tooltip title={`${high_issues_count} High`}>
            <Tag color="orange" icon={<WarningOutlined />}>
              {high_issues_count}
            </Tag>
          </Tooltip>
        )}
      </Space>
    );
  };

  const renderTokens = (tokens: number, record: ReviewRecord) => {
    if (record.status !== "completed" || tokens === 0) {
      return <Text type="secondary">-</Text>;
    }
    return (
      <Tooltip title={`${tokens.toLocaleString()} tokens consumed`}>
        <Space size={4}>
          <ThunderboltOutlined style={{ color: "#722ed1" }} />
          <Text>{formatTokens(tokens)}</Text>
        </Space>
      </Tooltip>
    );
  };

  const columns = [
    { title: "ID", dataIndex: "id", key: "id", width: 70 },
    {
      title: "Repository",
      dataIndex: ["repository", "name"],
      key: "repository",
      width: 150,
      ellipsis: true,
      render: (_: string, record: ReviewRecord) => (
        <Tooltip title={record.repository?.full_path}>
          {record.repository?.name || "-"}
        </Tooltip>
      ),
    },
    {
      title: "MR Title",
      dataIndex: "mr_title",
      key: "mr_title",
      ellipsis: true,
      render: (title: string, record: ReviewRecord) => (
        <Space>
          <Text ellipsis style={{ maxWidth: 250 }}>
            {title}
          </Text>
          {record.mr_web_url && (
            <Tooltip title="Open in GitLab">
              <a
                href={record.mr_web_url}
                target="_blank"
                rel="noopener noreferrer"
                onClick={(e) => e.stopPropagation()}
              >
                <LinkOutlined />
              </a>
            </Tooltip>
          )}
        </Space>
      ),
    },
    {
      title: "Author",
      dataIndex: "mr_author",
      key: "author",
      width: 120,
      ellipsis: true,
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      width: 100,
      render: (status: string) => renderStatusInline(status),
    },
    {
      title: "Score",
      dataIndex: "score",
      key: "score",
      width: 80,
      render: renderScore,
    },
    {
      title: "Issues",
      dataIndex: "issues_found",
      key: "issues",
      width: 120,
      render: renderIssues,
    },
    {
      title: "Tokens",
      dataIndex: "total_tokens",
      key: "tokens",
      width: 90,
      render: renderTokens,
    },
    {
      title: "Created",
      dataIndex: "created_at",
      key: "created_at",
      width: 140,
      render: (d: string) => dayjs(d).format("YYYY-MM-DD HH:mm"),
    },
  ];

  const tabItems = [
    { key: "all", label: "All Reviews" },
    {
      key: "failed",
      label: (
        <Space>
          <CloseCircleOutlined />
          Failed
        </Space>
      ),
    },
    {
      key: "critical",
      label: (
        <Space>
          <FireOutlined style={{ color: "#ff4d4f" }} />
          Critical Issues
        </Space>
      ),
    },
    {
      key: "high_score",
      label: (
        <Space>
          <CheckCircleOutlined style={{ color: "#52c41a" }} />
          High Score (≥80)
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <Card
        title="Code Reviews"
        extra={
          <Space>
            <Select
              value={filters.status}
              onChange={(v) => setFilters({ ...filters, status: v })}
              style={{ width: 130 }}
              allowClear
              placeholder="Status"
            >
              <Option value="completed">Completed</Option>
              <Option value="processing">Processing</Option>
              <Option value="pending">Pending</Option>
              <Option value="failed">Failed</Option>
            </Select>
            <Search
              placeholder="Search author"
              onSearch={(v) => setFilters({ ...filters, author: v })}
              style={{ width: 180 }}
              allowClear
            />
          </Space>
        }
      >
        <Tabs
          activeKey={activeTab}
          onChange={(key) => {
            setActiveTab(key as TabFilter);
            setPagination((prev) => ({ ...prev, page: 1 }));
          }}
          items={tabItems}
          style={{ marginBottom: 16 }}
        />
        <Table
          loading={loading}
          dataSource={data}
          columns={columns}
          rowKey="id"
          pagination={{
            current: pagination.page,
            pageSize: pagination.pageSize,
            total: pagination.total,
            showSizeChanger: true,
            showTotal: (total) => `Total ${total} reviews`,
            onChange: (page, pageSize) =>
              setPagination((prev) => ({ ...prev, page, pageSize })),
          }}
          onRow={(record) => ({
            onClick: () => navigate(`/reviews/${record.id}`),
            style: { cursor: "pointer" },
          })}
          size="middle"
        />
      </Card>
    </div>
  );
};

export default ReviewList;
