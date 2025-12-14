import { useEffect, useState } from "react";
import { Card, Table, Tag, Spin, message, Row, Col } from "antd";
import ReactECharts from "echarts-for-react";
import dayjs from "dayjs";
import axios from "axios";
import { useNavigate } from "react-router-dom";
import "./styles.css";
import {
  StatisticsCards,
  ReviewTrendChart,
  TokenUsageSection,
} from "./components";
import { renderStatusTag, getScoreColor } from "../../utils/statusConfig";

interface DashboardStats {
  total_reviews: number;
  completed_reviews: number;
  pending_reviews: number;
  failed_reviews: number;
  average_score: number;
  total_issues_found: number;
  critical_issues: number;
  high_issues: number;
  medium_issues: number;
  low_issues: number;
  security_issues: number;
  performance_issues: number;
  quality_issues: number;
}

interface ReviewRecord {
  id: number;
  repository: {
    name: string;
  };
  mr_title: string;
  mr_author: string;
  status: string;
  score: number;
  issues_found: number;
  critical_issues_count: number;
  created_at: string;
}

interface TrendData {
  date: string;
  review_count: number;
  average_score: number;
  total_issues: number;
  critical_issues: number;
}

interface TokenUsageData {
  summary: {
    total_calls: number;
    successful_calls: number;
    failed_calls: number;
    total_tokens: number;
    prompt_tokens: number;
    completion_tokens: number;
    avg_duration_ms: number;
    success_rate: number;
  };
  top_repositories: Array<{
    repository_id: number;
    repository_name: string;
    total_tokens: number;
    review_count: number;
    avg_tokens: number;
  }>;
  daily_trend: Array<{
    date: string;
    total_tokens: number;
    review_count: number;
    avg_duration_ms: number;
    success_rate: number;
  }>;
}

const Dashboard = () => {
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [recentReviews, setRecentReviews] = useState<ReviewRecord[]>([]);
  const [trends, setTrends] = useState<TrendData[]>([]);
  const [tokenUsage, setTokenUsage] = useState<TokenUsageData | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const fetchDashboardData = async () => {
    setLoading(true);
    try {
      const [statsRes, recentRes, trendsRes, tokenRes] = await Promise.all([
        axios.get("/api/dashboard/statistics"),
        axios.get("/api/dashboard/recent?limit=10"),
        axios.get("/api/dashboard/trends?days=30"),
        axios.get("/api/dashboard/token-usage?days=30"),
      ]);

      setStats(statsRes.data);
      setRecentReviews(Array.isArray(recentRes.data) ? recentRes.data : []);
      setTrends(Array.isArray(trendsRes.data) ? trendsRes.data : []);
      setTokenUsage(tokenRes.data);
    } catch (error) {
      message.error("Failed to load dashboard data");
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const getIssueDistributionOption = () => {
    if (!stats) return {};
    return {
      title: { text: "Issue Distribution", left: "center" },
      tooltip: { trigger: "item" },
      legend: { orient: "vertical", left: "left" },
      series: [
        {
          name: "Issues",
          type: "pie",
          radius: "50%",
          data: [
            {
              value: stats.critical_issues,
              name: "Critical",
              itemStyle: { color: "#f5222d" },
            },
            {
              value: stats.high_issues,
              name: "High",
              itemStyle: { color: "#fa8c16" },
            },
            {
              value: stats.medium_issues,
              name: "Medium",
              itemStyle: { color: "#faad14" },
            },
            {
              value: stats.low_issues,
              name: "Low",
              itemStyle: { color: "#52c41a" },
            },
          ],
        },
      ],
    };
  };

  const getCategoryDistributionOption = () => {
    if (!stats) return {};
    return {
      title: { text: "Issue Category", left: "center" },
      tooltip: { trigger: "item" },
      series: [
        {
          name: "Category",
          type: "pie",
          radius: ["40%", "70%"],
          data: [
            { value: stats.security_issues, name: "Security" },
            { value: stats.performance_issues, name: "Performance" },
            { value: stats.quality_issues, name: "Quality" },
          ],
        },
      ],
    };
  };

  const columns = [
    {
      title: "Repository",
      dataIndex: ["repository", "name"],
      key: "repository",
      render: (text: string) => <strong>{text}</strong>,
    },
    {
      title: "MR Title",
      dataIndex: "mr_title",
      key: "mr_title",
      ellipsis: true,
    },
    {
      title: "Author",
      dataIndex: "mr_author",
      key: "mr_author",
    },
    {
      title: "Status",
      dataIndex: "status",
      key: "status",
      render: (status: string) => renderStatusTag(status),
    },
    {
      title: "Score",
      dataIndex: "score",
      key: "score",
      render: (score: number) => (
        <span style={{ color: getScoreColor(score), fontWeight: "bold" }}>
          {score > 0 ? score : "-"}
        </span>
      ),
    },
    {
      title: "Issues",
      dataIndex: "issues_found",
      key: "issues",
      render: (issues: number, record: ReviewRecord) => (
        <span>
          {issues}{" "}
          {record.critical_issues_count > 0 && (
            <Tag color="red">{record.critical_issues_count} Critical</Tag>
          )}
        </span>
      ),
    },
    {
      title: "Created At",
      dataIndex: "created_at",
      key: "created_at",
      render: (date: string) => dayjs(date).format("YYYY-MM-DD HH:mm"),
    },
  ];

  const hasIssues =
    stats &&
    stats.critical_issues +
      stats.high_issues +
      stats.medium_issues +
      stats.low_issues >
      0;

  const hasCategoryData =
    stats &&
    stats.security_issues + stats.performance_issues + stats.quality_issues > 0;

  if (loading) {
    return (
      <div style={{ textAlign: "center", padding: "100px 0" }}>
        <Spin size="large" tip="Loading dashboard..." />
      </div>
    );
  }

  return (
    <div className="dashboard-container">
      <h1>Dashboard</h1>

      {/* Statistics Cards Component */}
      <StatisticsCards stats={stats} />

      {/* Review Trends and Issue Distribution */}
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={16}>
          <ReviewTrendChart trends={trends} />
        </Col>
        <Col xs={24} lg={8}>
          <Card title="Issue Distribution">
            {hasIssues ? (
              <ReactECharts
                option={getIssueDistributionOption()}
                style={{ height: 400 }}
              />
            ) : (
              <div style={{ textAlign: "center", padding: "100px 0" }}>
                No issues found
              </div>
            )}
          </Card>
        </Col>
      </Row>

      {/* Issue Category Distribution */}
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24}>
          <Card title="Issue Category Distribution">
            {hasCategoryData ? (
              <ReactECharts
                option={getCategoryDistributionOption()}
                style={{ height: 300 }}
              />
            ) : (
              <div style={{ textAlign: "center", padding: "50px 0" }}>
                No category data available
              </div>
            )}
          </Card>
        </Col>
      </Row>

      {/* Token Usage Section Component */}
      <TokenUsageSection tokenUsage={tokenUsage} />

      {/* Recent Reviews Table */}
      <Card title="Recent Reviews" style={{ marginTop: 16 }}>
        <Table
          columns={columns}
          dataSource={recentReviews}
          rowKey="id"
          pagination={false}
          onRow={(record) => ({
            onClick: () => navigate(`/reviews/${record.id}`),
            style: { cursor: "pointer" },
          })}
        />
      </Card>
    </div>
  );
};

export default Dashboard;
