import { Card, Row, Col } from "antd";
import ReactECharts from "echarts-for-react";

interface DashboardStats {
  critical_issues: number;
  high_issues: number;
  medium_issues: number;
  low_issues: number;
  security_issues: number;
  performance_issues: number;
  quality_issues: number;
}

interface IssueDistributionChartProps {
  stats: DashboardStats | null;
}

const IssueDistributionChart: React.FC<IssueDistributionChartProps> = ({
  stats,
}) => {
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

  return (
    <>
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} lg={16}>
          {/* This is a placeholder - ReviewTrendChart is rendered separately */}
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
    </>
  );
};

export default IssueDistributionChart;
