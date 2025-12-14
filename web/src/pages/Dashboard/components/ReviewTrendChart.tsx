import { Card } from "antd";
import ReactECharts from "echarts-for-react";
import dayjs from "dayjs";

interface TrendData {
  date: string;
  review_count: number;
  avg_score?: number;
  average_score?: number;
  total_issues: number;
  critical_issues: number;
}

interface ReviewTrendChartProps {
  trends: TrendData[];
}

const ReviewTrendChart: React.FC<ReviewTrendChartProps> = ({ trends }) => {
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
          data: trends.map((t) => t.avg_score ?? t.average_score ?? 0),
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

  return (
    <Card title="Review Trends">
      {trends.length > 0 ? (
        <ReactECharts option={getTrendChartOption()} style={{ height: 400 }} />
      ) : (
        <div style={{ textAlign: "center", padding: "100px 0" }}>
          No data available
        </div>
      )}
    </Card>
  );
};

export default ReviewTrendChart;
