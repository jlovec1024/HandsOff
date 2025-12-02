import { useEffect, useState } from "react";
import { Card, Row, Col, Statistic, Table, Tag, Spin, message } from "antd";
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  BugOutlined,
  SecurityScanOutlined,
  ThunderboltOutlined,
  CodeOutlined,
} from "@ant-design/icons";
import ReactECharts from "echarts-for-react";
import dayjs from "dayjs";
import axios from "axios";
import { useNavigate } from "react-router-dom";
import "./styles.css";

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

const Dashboard = () => {
  const [loading, setLoading] = useState(true);
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [recentReviews, setRecentReviews] = useState<ReviewRecord[]>([]);
  const [trends, setTrends] = useState<TrendData[]>([]);
  const navigate = useNavigate();

  useEffect(() => {
    fetchDashboardData();
  }, []);

  const fetchDashboardData = async () => {
    setLoading(true);
    try {
      const [statsRes, recentRes, trendsRes] = await Promise.all([
        axios.get("/api/dashboard/statistics"),
        axios.get("/api/dashboard/recent?limit=10"),
        axios.get("/api/dashboard/trends?days=30"),
      ]);

      setStats(statsRes.data);
      setRecentReviews(Array.isArray(recentRes.data) ? recentRes.data : []);
      setTrends(Array.isArray(trendsRes.data) ? trendsRes.data : []);
    } catch (error) {
      message.error("Failed to load dashboard data");
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const getStatusTag = (status: string) => {
    const statusMap: Record<string, { color: string; text: string }> = {
      completed: { color: "success", text: "Completed" },
      processing: { color: "processing", text: "Processing" },
      pending: { color: "default", text: "Pending" },
      failed: { color: "error", text: "Failed" },
    };
    const config = statusMap[status] || { color: "default", text: status };
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  const getScoreColor = (score: number) => {
    if (score >= 80) return "#52c41a";
    if (score >= 60) return "#faad14";
    return "#f5222d";
  };

  const getTrendChartOption = () => {
    if (!Array.isArray(trends) || trends.length === 0) return {};
    return {
      title: { text: "Review Trends (30 Days)", left: "center" },
      tooltip: { trigger: "axis" },
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
