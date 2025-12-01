import { Card, Descriptions, Tag } from 'antd';

const SystemConfig = () => {
  return (
    <div>
      <Card title="系统信息" style={{ marginBottom: 16 }}>
        <Descriptions column={2}>
          <Descriptions.Item label="版本">1.0.0-mvp</Descriptions.Item>
          <Descriptions.Item label="环境">
            <Tag color="blue">Development</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="API地址">
            {import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api'}
          </Descriptions.Item>
          <Descriptions.Item label="前端版本">React 18</Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title="Webhook配置">
        <Descriptions column={1}>
          <Descriptions.Item label="Webhook URL">
            <code>http://your-server.com/api/webhook</code>
          </Descriptions.Item>
          <Descriptions.Item label="说明">
            在GitLab仓库设置中配置此URL以接收Merge Request事件
          </Descriptions.Item>
        </Descriptions>
      </Card>

      <Card title="提示词模板" style={{ marginTop: 16 }}>
        <p style={{ color: '#666' }}>
          提示词模板功能将在后续版本中提供，目前使用默认模板。
        </p>
      </Card>
    </div>
  );
};

export default SystemConfig;
