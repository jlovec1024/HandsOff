import { useEffect, useState } from "react";
import { Card, Descriptions, Tag, Table, Typography, Spin, Button } from "antd";
import { useParams, useNavigate } from "react-router-dom";
import { ArrowLeftOutlined } from "@ant-design/icons";
import axios from "axios";
import dayjs from "dayjs";

const { Title, Paragraph } = Typography;

const ReviewDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const [review, setReview] = useState<any>(null);

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

  const severityColors: any = {
    critical: "red",
    high: "orange",
    medium: "gold",
    low: "green",
  };

  const suggestionColumns = [
    { title: "File", dataIndex: "file_path", key: "file", ellipsis: true },
    {
      title: "Lines",
      key: "lines",
      render: (_: any, record: any) =>
        `${record.line_start}-${record.line_end}`,
    },
    {
      title: "Severity",
      dataIndex: "severity",
      key: "severity",
      render: (severity: string) => (
        <Tag color={severityColors[severity]}>{severity}</Tag>
      ),
    },
    { title: "Category", dataIndex: "category", key: "category" },
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
      <Button
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate("/reviews")}
        style={{ marginBottom: 16 }}
      >
        Back to List
      </Button>

      <Card title={<Title level={3}>Review Details #{review.id}</Title>}>
        <Descriptions bordered column={2}>
          <Descriptions.Item label="Repository">
            {review.repository?.name}
          </Descriptions.Item>
          <Descriptions.Item label="Status">
            <Tag
              color={
                review.status === "completed"
                  ? "success"
                  : review.status === "failed"
                  ? "error"
                  : "processing"
              }
            >
              {review.status}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="MR Title" span={2}>
            {review.mr_title}
          </Descriptions.Item>
          <Descriptions.Item label="Author">
            {review.mr_author}
          </Descriptions.Item>
          <Descriptions.Item label="Score">
            <span
              style={{
                fontSize: 18,
                fontWeight: "bold",
                color:
                  review.score >= 80
                    ? "#52c41a"
                    : review.score >= 60
                    ? "#faad14"
                    : "#f5222d",
              }}
            >
              {review.score}/100
            </span>
          </Descriptions.Item>
          <Descriptions.Item label="Source Branch">
            {review.source_branch}
          </Descriptions.Item>
          <Descriptions.Item label="Target Branch">
            {review.target_branch}
          </Descriptions.Item>
          <Descriptions.Item label="Issues Found">
            {review.issues_found}
          </Descriptions.Item>
          <Descriptions.Item label="Critical Issues">
            <Tag color="red">{review.critical_issues_count}</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="Created At">
            {dayjs(review.created_at).format("YYYY-MM-DD HH:mm:ss")}
          </Descriptions.Item>
          <Descriptions.Item label="Reviewed At">
            {review.reviewed_at
              ? dayjs(review.reviewed_at).format("YYYY-MM-DD HH:mm:ss")
              : "-"}
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title="AI Summary" style={{ marginTop: 16 }}>
        <Paragraph>{review.summary || "No summary available"}</Paragraph>
      </Card>

      <Card
        title={`Fix Suggestions (${review.fix_suggestions?.length || 0})`}
        style={{ marginTop: 16 }}
      >
        <Table
          dataSource={review.fix_suggestions || []}
          columns={suggestionColumns}
          rowKey="id"
          expandable={{
            expandedRowRender: (record: any) => (
              <div>
                <p>
                  <strong>Suggestion:</strong> {record.suggestion}
                </p>
                {record.code_snippet && (
                  <pre
                    style={{
                      background: "#f5f5f5",
                      padding: 12,
                      borderRadius: 4,
                    }}
                  >
                    <code>{record.code_snippet}</code>
                  </pre>
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
