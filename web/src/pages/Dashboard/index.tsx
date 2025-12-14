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
      legend: { data: ["Reviews", "Avg Score", "Critical Issues"], bottom: 0 },
      xAxis: {
        type: "category",
        data: trends.map((t) => dayjs(t.date).format("MM-DD")),
      },
      yAxis: [
        { type: "value", name: "Count" },
        { type: "value", name: "Score", max: 100 },
      ],
      series: [
        {
          name: "Reviews",
          type: "bar",
          data: trends.map((t) => t.review_count),
        },
        {
          name: "Avg Score",
          type: "line",
          yAxisIndex: 1,
          data: trends.map((t) => t.average_score),
        },
        {
          name: "Critical Issues",
          type: "line",
          data: trends.map((t) => t.critical_issues),
          itemStyle: { color: "#f5222d" },
        },
      ],
    };
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
          emphasis: {
            itemStyle: {
              shadowBlur: 10,
              shadowOffsetX: 0,
              shadowColor: "rgba(0, 0, 0, 0.5)",
            },
          },
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

  const formatTokens = (tokens: number) => {
    if (tokens >= 1000000) return `${(tokens / 1000000).toFixed(1)}M`;
    if (tokens >= 1000) return `${(tokens / 1000).toFixed(1)}k`;
    return tokens.toString();
  };

  const getTokenTrendChartOption = () => {
    if (!tokenUsage?.daily_trend || tokenUsage.daily_trend.length === 0)
      return {};
    const trend = tokenUsage.daily_trend;
    return {
      title: { text: "Token Usage Trend (30 Days)", left: "center" },
      tooltip: { trigger: "axis" },
      legend: { data: ["Tokens", "Reviews"], bottom: 0 },
      xAxis: {
        type: "category",
        data: trend.map((t) => dayjs(t.date).format("MM-DD")),
      },
      yAxis: [
        { type: "value", name: "Tokens" },
        { type: "value", name: "Reviews" },
      ],
      series: [
        {
          name: "Tokens",
          type: "bar",
          data: trend.map((t) => t.total_tokens),
          itemStyle: { color: "#722ed1" },
        },
        {
          name: "Reviews",
          type: "line",
          yAxisIndex: 1,
          data: trend.map((t) => t.review_count),
          itemStyle: { color: "#1890ff" },
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
      render: getStatusTag,
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

      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Total Reviews"
              value={stats?.total_reviews || 0}
              prefix={<CodeOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Completed"
              value={stats?.completed_reviews || 0}
              valueStyle={{ color: "#3f8600" }}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Pending"
              value={stats?.pending_reviews || 0}
              valueStyle={{ color: "#1890ff" }}
              prefix={<ClockCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Failed"
              value={stats?.failed_reviews || 0}
              valueStyle={{ color: "#cf1322" }}
              prefix={<CloseCircleOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="Average Score"
              value={stats?.average_score || 0}
              precision={1}
              suffix="/ 100"
              valueStyle={{ color: getScoreColor(stats?.average_score || 0) }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="Total Issues Found"
              value={stats?.total_issues_found || 0}
              prefix={<BugOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="Critical Issues"
              value={stats?.critical_issues || 0}
              valueStyle={{ color: "#cf1322" }}
              prefix={<SecurityScanOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="Security Issues"
              value={stats?.security_issues || 0}
              prefix={<SecurityScanOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="Performance Issues"
              value={stats?.performance_issues || 0}
              prefix={<ThunderboltOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="Quality Issues"
              value={stats?.quality_issues || 0}
              prefix={<CodeOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={16}>
          <Card title="Review Trends">
            {trends.length > 0 ? (
              <ReactECharts
                option={getTrendChartOption()}
                style={{ height: 400 }}
              />
            ) : (
              <div style={{ textAlign: "center", padding: "100px 0" }}>
                No data available
              </div>
            )}
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card title="Issue Distribution">
            {stats &&
            stats.critical_issues +
              stats.high_issues +
              stats.medium_issues +
              stats.low_issues >
              0 ? (
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

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24}>
          <Card title="Issue Category Distribution">
            {stats &&
            stats.security_issues +
              stats.performance_issues +
              stats.quality_issues >
              0 ? (
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

      {/* Token Usage Section */}
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Total API Calls"
              value={tokenUsage?.summary?.total_calls || 0}
              prefix={<ApiOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Total Tokens"
              value={formatTokens(tokenUsage?.summary?.total_tokens || 0)}
              valueStyle={{ color: "#722ed1" }}
              prefix={<ThunderboltOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Tooltip
              title={`Prompt: ${formatTokens(
                tokenUsage?.summary?.prompt_tokens || 0
              )} / Completion: ${formatTokens(
                tokenUsage?.summary?.completion_tokens || 0
              )}`}
            >
              <Statistic
                title="Token Breakdown"
                value={`${formatTokens(
                  tokenUsage?.summary?.prompt_tokens || 0
                )} / ${formatTokens(
                  tokenUsage?.summary?.completion_tokens || 0
                )}`}
                valueStyle={{ fontSize: 18 }}
              />
            </Tooltip>
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Success Rate"
              value={tokenUsage?.summary?.success_rate || 0}
              precision={1}
              suffix="%"
              valueStyle={{
                color:
                  (tokenUsage?.summary?.success_rate || 0) >= 95
                    ? "#52c41a"
                    : (tokenUsage?.summary?.success_rate || 0) >= 80
                    ? "#faad14"
                    : "#ff4d4f",
              }}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={16}>
          <Card title="Token Usage Trend">
            {tokenUsage?.daily_trend && tokenUsage.daily_trend.length > 0 ? (
              <ReactECharts
                option={getTokenTrendChartOption()}
                style={{ height: 350 }}
              />
            ) : (
              <div style={{ textAlign: "center", padding: "100px 0" }}>
                No token usage data available
              </div>
            )}
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card title="Top Repositories by Token Usage">
            {tokenUsage?.top_repositories &&
            tokenUsage.top_repositories.length > 0 ? (
              <div>
                {tokenUsage.top_repositories.map((repo, index) => (
                  <div
                    key={repo.repository_id}
                    style={{
                      padding: "8px 0",
                      borderBottom:
                        index < tokenUsage.top_repositories.length - 1
                          ? "1px solid #f0f0f0"
                          : "none",
                    }}
                  >
                    <div
                      style={{
                        display: "flex",
                        justifyContent: "space-between",
                        marginBottom: 4,
                      }}
                    >
                      <span style={{ fontWeight: 500 }}>
                        {repo.repository_name ||
                          `Repository #${repo.repository_id}`}
                      </span>
                      <span style={{ color: "#722ed1" }}>
                        {formatTokens(repo.total_tokens)}
                      </span>
                    </div>
                    <Progress
                      percent={Math.round(
                        (repo.total_tokens /
                          (tokenUsage.top_repositories[0]?.total_tokens || 1)) *
                          100
                      )}
                      size="small"
                      showInfo={false}
                      strokeColor="#722ed1"
                    />
                    <div
                      style={{
                        fontSize: 12,
                        color: "#888",
                        marginTop: 2,
                      }}
                    >
                      {repo.review_count} reviews · ~
                      {formatTokens(repo.avg_tokens)}
                      /review
                    </div>
                  </div>
                ))}
              </div>
            ) : (
              <div style={{ textAlign: "center", padding: "50px 0" }}>
                No repository data available
              </div>
            )}
          </Card>
        </Col>
      </Row>

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
