import { Card, Row, Col, Statistic, Tooltip, Progress } from "antd";
import {
  ThunderboltOutlined,
  ApiOutlined,
  CheckCircleOutlined,
} from "@ant-design/icons";
import ReactECharts from "echarts-for-react";
import dayjs from "dayjs";
import { formatTokens } from "../../../utils/statusConfig";

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

interface TokenUsageSectionProps {
  tokenUsage: TokenUsageData | null;
}

const TokenUsageSection: React.FC<TokenUsageSectionProps> = ({
  tokenUsage,
}) => {
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

  const successRate = tokenUsage?.summary?.success_rate ?? 0;

  return (
    <>
      {/* Token Statistics Cards */}
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Total API Calls"
              value={tokenUsage?.summary?.total_calls ?? 0}
              prefix={<ApiOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="Total Tokens"
              value={formatTokens(tokenUsage?.summary?.total_tokens ?? 0)}
              valueStyle={{ color: "#722ed1" }}
              prefix={<ThunderboltOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Tooltip
              title={`Prompt: ${formatTokens(
                tokenUsage?.summary?.prompt_tokens ?? 0
              )} / Completion: ${formatTokens(
                tokenUsage?.summary?.completion_tokens ?? 0
              )}`}
            >
              <Statistic
                title="Token Breakdown"
                value={`${formatTokens(
                  tokenUsage?.summary?.prompt_tokens ?? 0
                )} / ${formatTokens(
                  tokenUsage?.summary?.completion_tokens ?? 0
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
              value={successRate}
              precision={1}
              suffix="%"
              valueStyle={{
                color:
                  successRate >= 95
                    ? "#52c41a"
                    : successRate >= 80
                    ? "#faad14"
                    : "#ff4d4f",
              }}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
      </Row>

      {/* Token Trend Chart and Top Repositories */}
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
    </>
  );
};

export default TokenUsageSection;
