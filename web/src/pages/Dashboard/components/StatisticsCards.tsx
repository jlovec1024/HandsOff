import { Card, Row, Col, Statistic } from "antd";
import {
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  BugOutlined,
  SecurityScanOutlined,
  ThunderboltOutlined,
  CodeOutlined,
} from "@ant-design/icons";
import { getScoreColor } from "../../../utils/statusConfig";

interface DashboardStats {
  total_reviews: number;
  completed_reviews: number;
  pending_reviews: number;
  failed_reviews: number;
  average_score: number;
  total_issues_found: number;
  critical_issues: number;
  security_issues: number;
  performance_issues: number;
  quality_issues: number;
}

interface StatisticsCardsProps {
  stats: DashboardStats | null;
}

const StatisticsCards: React.FC<StatisticsCardsProps> = ({ stats }) => {
  return (
    <>
      {/* Review Status Cards */}
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Total Reviews"
              value={stats?.total_reviews ?? 0}
              prefix={<CodeOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Completed"
              value={stats?.completed_reviews ?? 0}
              valueStyle={{ color: "#3f8600" }}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Pending"
              value={stats?.pending_reviews ?? 0}
              valueStyle={{ color: "#1890ff" }}
              prefix={<ClockCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Failed"
              value={stats?.failed_reviews ?? 0}
              valueStyle={{ color: "#cf1322" }}
              prefix={<CloseCircleOutlined />}
            />
          </Card>
        </Col>
      </Row>

      {/* Score and Issues Cards */}
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="Average Score"
              value={stats?.average_score ?? 0}
              precision={1}
              suffix="/ 100"
              valueStyle={{ color: getScoreColor(stats?.average_score ?? 0) }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="Total Issues Found"
              value={stats?.total_issues_found ?? 0}
              prefix={<BugOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={8}>
          <Card>
            <Statistic
              title="Critical Issues"
              value={stats?.critical_issues ?? 0}
              valueStyle={{ color: "#cf1322" }}
              prefix={<SecurityScanOutlined />}
            />
          </Card>
        </Col>
      </Row>

      {/* Category Cards */}
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="Security Issues"
              value={stats?.security_issues ?? 0}
              prefix={<SecurityScanOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="Performance Issues"
              value={stats?.performance_issues ?? 0}
              prefix={<ThunderboltOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={8}>
          <Card>
            <Statistic
              title="Quality Issues"
              value={stats?.quality_issues ?? 0}
              prefix={<CodeOutlined />}
            />
          </Card>
        </Col>
      </Row>
    </>
  );
};

export default StatisticsCards;
