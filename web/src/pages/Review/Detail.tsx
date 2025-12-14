import { useEffect, useState } from "react";
import {
  Card,
  Descriptions,
  Tag,
  Table,
  Typography,
  Spin,
  Button,
  Space,
  Row,
  Col,
  Statistic,
  Tooltip,
  Select,
  Alert,
  Divider,
} from "antd";
import { useParams, useNavigate } from "react-router-dom";
import {
  ArrowLeftOutlined,
  LinkOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  ThunderboltOutlined,
  FireOutlined,
  WarningOutlined,
  InfoCircleOutlined,
  MessageOutlined,
  ExportOutlined,
} from "@ant-design/icons";
import axios from "axios";
import dayjs from "dayjs";
import {
  renderStatusTag,
  getScoreColor,
  formatTokens,
  formatDuration,
} from "../../utils/statusConfig";

const { Title, Paragraph, Text } = Typography;
const { Option } = Select;

interface FixSuggestion {
  id: number;
  file_path: string;
  line_start: number;
  line_end: number;
  severity: string;
  category: string;
  description: string;
  suggestion: string;
  code_snippet?: string;
}

interface ReviewData {
  id: number;
  repository?: { id: number; name: string; full_path: string };
  mr_iid: number;
  mr_title: string;
  mr_author: string;
  mr_web_url: string;
  source_branch: string;
  target_branch: string;
  status: string;
  score: number;
  issues_found: number;
  critical_issues_count: number;
  high_issues_count: number;
  medium_issues_count: number;
  low_issues_count: number;
  summary: string;
  fix_suggestions: FixSuggestion[];
  comment_posted: boolean;
  comment_url?: string;
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
  llm_duration_ms: number;
  created_at: string;
  reviewed_at?: string;
  error_message?: string;
}

type SeverityFilter = "all" | "critical" | "high" | "medium" | "low";

const ReviewDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [review, setReview] = useState<ReviewData | null>(null);
  const [severityFilter, setSeverityFilter] = useState<SeverityFilter>("all");

  useEffect(() => {
    fetchReviewDetail();
  }, [id]);

  const fetchReviewDetail = async () => {
    setLoading(true);
    try {
      const res = await axios.get(`/api/reviews/${id}`);
      setReview(res.data);
    } catch (error) {
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const severityColors: Record<string, string> = {
    critical: "red",
    high: "orange",
    medium: "gold",
    low: "green",
  };

  const severityIcons: Record<string, React.ReactNode> = {
    critical: <FireOutlined />,
    high: <WarningOutlined />,
    medium: <InfoCircleOutlined />,
    low: <CheckCircleOutlined />,
  };

  const filteredSuggestions =
    review?.fix_suggestions?.filter(
      (s) => severityFilter === "all" || s.severity === severityFilter
    ) || [];

  const suggestionColumns = [
    {
      title: "File",
      dataIndex: "file_path",
      key: "file",
      ellipsis: true,
      width: 200,
      render: (path: string) => (
        <Tooltip title={path}>
          <Text code style={{ fontSize: 12 }}>
            {path.split("/").pop()}
          </Text>
        </Tooltip>
      ),
    },
    {
      title: "Lines",
      key: "lines",
      width: 80,
      render: (_: unknown, record: FixSuggestion) => (
        <Text type="secondary">
          {record.line_start === record.line_end
            ? `L${record.line_start}`
            : `L${record.line_start}-${record.line_end}`}
        </Text>
      ),
    },
    {
      title: "Severity",
      dataIndex: "severity",
      key: "severity",
      width: 100,
      render: (severity: string) => (
        <Tag color={severityColors[severity]} icon={severityIcons[severity]}>
          {severity}
        </Tag>
      ),
    },
    {
      title: "Category",
      dataIndex: "category",
      key: "category",
      width: 120,
      render: (cat: string) => <Tag>{cat}</Tag>,
    },
    {
      title: "Description",
      dataIndex: "description",
      key: "description",
      ellipsis: true,
    },
  ];

  if (loading) {
    return (
      <div style={{ textAlign: "center", padding: "100px 0" }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!review) {
    return <div style={{ padding: 24 }}>Review not found</div>;
  }

  return (
    <div style={{ padding: 24 }}>
      {/* Header */}
      <Space style={{ marginBottom: 16 }}>
        <Button
          icon={<ArrowLeftOutlined />}
          onClick={() => navigate("/reviews")}
        >
          Back to List
        </Button>
        {review.mr_web_url && (
          <Button
            type="primary"
            icon={<ExportOutlined />}
            href={review.mr_web_url}
            target="_blank"
          >
            View in GitLab
          </Button>
        )}
      </Space>

      {/* Error Alert for Failed Reviews */}
      {review.status === "failed" && review.error_message && (
        <Alert
          type="error"
          message="Review Failed"
          description={review.error_message}
          showIcon
          style={{ marginBottom: 16 }}
        />
      )}

      {/* Main Info Card */}
      <Card
        title={
          <Space>
            <Title level={4} style={{ margin: 0 }}>
              Review #{review.id}
            </Title>
            {renderStatusTag(review.status)}
          </Space>
        }
        extra={
          review.comment_posted && (
            <Tooltip
              title={
                review.comment_url
                  ? "Comment posted to GitLab"
                  : "Comment posted"
              }
            >
              <Tag color="blue" icon={<MessageOutlined />}>
                {review.comment_url ? (
                  <a
                    href={review.comment_url}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    Comment Posted <LinkOutlined />
                  </a>
                ) : (
                  "Comment Posted"
                )}
              </Tag>
            </Tooltip>
          )
        }
      >
        <Descriptions bordered column={{ xs: 1, sm: 2, md: 2 }} size="small">
          <Descriptions.Item label="Repository">
            <Space>
              <Text strong>{review.repository?.name}</Text>
              <Text type="secondary">{review.repository?.full_path}</Text>
            </Space>
          </Descriptions.Item>
          <Descriptions.Item label="MR">
            <Space>
              <Text>!{review.mr_iid}</Text>
              {review.mr_web_url && (
                <a
                  href={review.mr_web_url}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  <LinkOutlined />
                </a>
              )}
            </Space>
          </Descriptions.Item>
          <Descriptions.Item label="Title" span={2}>
            <Text strong>{review.mr_title}</Text>
          </Descriptions.Item>
          <Descriptions.Item label="Author">
            {review.mr_author}
          </Descriptions.Item>
          <Descriptions.Item label="Score">
            {review.status === "completed" && review.score > 0 ? (
              <Text
                style={{
                  fontSize: 20,
                  fontWeight: "bold",
                  color: getScoreColor(review.score),
                }}
              >
                {review.score}/100
              </Text>
            ) : (
              <Text type="secondary">-</Text>
            )}
          </Descriptions.Item>
          <Descriptions.Item label="Source Branch">
            <Tag>{review.source_branch}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Target Branch">
            <Tag color="blue">{review.target_branch}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Created">
            {dayjs(review.created_at).format("YYYY-MM-DD HH:mm:ss")}
          </Descriptions.Item>
          <Descriptions.Item label="Reviewed">
            {review.reviewed_at
              ? dayjs(review.reviewed_at).format("YYYY-MM-DD HH:mm:ss")
              : "-"}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* Statistics Row */}
      <Row gutter={16} style={{ marginTop: 16 }}>
        {/* Issues Summary */}
        <Col xs={24} md={12}>
          <Card size="small" title="Issues Found">
            <Row gutter={16}>
              <Col span={6}>
                <Statistic
                  title={
                    <Space>
                      <FireOutlined style={{ color: "#ff4d4f" }} />
                      Critical
                    </Space>
                  }
                  value={review.critical_issues_count}
                  valueStyle={{
                    color:
                      review.critical_issues_count > 0 ? "#ff4d4f" : undefined,
                  }}
                />
              </Col>
              <Col span={6}>
                <Statistic
                  title={
                    <Space>
                      <WarningOutlined style={{ color: "#fa8c16" }} />
                      High
                    </Space>
                  }
                  value={review.high_issues_count}
                  valueStyle={{
                    color: review.high_issues_count > 0 ? "#fa8c16" : undefined,
                  }}
                />
              </Col>
              <Col span={6}>
                <Statistic
                  title={
                    <Space>
                      <InfoCircleOutlined style={{ color: "#faad14" }} />
                      Medium
                    </Space>
                  }
                  value={review.medium_issues_count || 0}
                />
              </Col>
              <Col span={6}>
                <Statistic
                  title={
                    <Space>
                      <CheckCircleOutlined style={{ color: "#52c41a" }} />
                      Low
                    </Space>
                  }
                  value={review.low_issues_count || 0}
                />
              </Col>
            </Row>
          </Card>
        </Col>

        {/* Token Usage */}
        <Col xs={24} md={12}>
          <Card
            size="small"
            title={
              <Space>
                <ThunderboltOutlined style={{ color: "#722ed1" }} />
                Token Usage
              </Space>
            }
          >
            <Row gutter={16}>
              <Col span={6}>
                <Statistic
                  title="Prompt"
                  value={formatTokens(review.prompt_tokens || 0)}
                  valueStyle={{ color: "#1890ff" }}
                />
              </Col>
              <Col span={6}>
                <Statistic
                  title="Completion"
                  value={formatTokens(review.completion_tokens || 0)}
                  valueStyle={{ color: "#52c41a" }}
                />
              </Col>
              <Col span={6}>
                <Statistic
                  title="Total"
                  value={formatTokens(review.total_tokens || 0)}
                  valueStyle={{ color: "#722ed1", fontWeight: "bold" }}
                />
              </Col>
              <Col span={6}>
                <Statistic
                  title="Duration"
                  value={formatDuration(review.llm_duration_ms || 0)}
                  valueStyle={{ color: "#faad14" }}
                />
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>

      {/* AI Summary */}
      <Card title="AI Summary" style={{ marginTop: 16 }}>
        <Paragraph style={{ whiteSpace: "pre-wrap" }}>
          {review.summary || "No summary available"}
        </Paragraph>
      </Card>

      {/* Fix Suggestions */}
      <Card
        title={
          <Space>
            <span>Fix Suggestions</span>
            <Tag>{review.fix_suggestions?.length || 0}</Tag>
          </Space>
        }
        extra={
          <Select
            value={severityFilter}
            onChange={setSeverityFilter}
            style={{ width: 140 }}
          >
            <Option value="all">All Severities</Option>
            <Option value="critical">
              <Space>
                <FireOutlined style={{ color: "#ff4d4f" }} />
                Critical
              </Space>
            </Option>
            <Option value="high">
              <Space>
                <WarningOutlined style={{ color: "#fa8c16" }} />
                High
              </Space>
            </Option>
            <Option value="medium">
              <Space>
                <InfoCircleOutlined style={{ color: "#faad14" }} />
                Medium
              </Space>
            </Option>
            <Option value="low">
              <Space>
                <CheckCircleOutlined style={{ color: "#52c41a" }} />
                Low
              </Space>
            </Option>
          </Select>
        }
        style={{ marginTop: 16 }}
      >
        <Table
          dataSource={filteredSuggestions}
          columns={suggestionColumns}
          rowKey="id"
          size="small"
          pagination={{
            pageSize: 10,
            showTotal: (total) => `${total} suggestions`,
          }}
          expandable={{
            expandedRowRender: (record: FixSuggestion) => (
              <div style={{ padding: "8px 0" }}>
                <Text strong>File: </Text>
                <Text code>{record.file_path}</Text>
                <Divider style={{ margin: "8px 0" }} />
                <Text strong>Suggestion: </Text>
                <Paragraph style={{ marginTop: 4 }}>
                  {record.suggestion}
                </Paragraph>
                {record.code_snippet && (
                  <>
                    <Text strong>Code Example: </Text>
                    <pre
                      style={{
                        background: "#f6f8fa",
                        padding: 12,
                        borderRadius: 6,
                        marginTop: 4,
                        overflow: "auto",
                        fontSize: 13,
                        border: "1px solid #e1e4e8",
                      }}
                    >
                      <code>{record.code_snippet}</code>
                    </pre>
                  </>
                )}
              </div>
            ),
          }}
        />
      </Card>
    </div>
  );
};

export default ReviewDetail;
