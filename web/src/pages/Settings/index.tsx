import { Tabs } from 'antd';
import { SettingOutlined, GitlabOutlined, ApiOutlined } from '@ant-design/icons';
import GitLabConfig from './GitLabConfig';
import LLMProviders from './LLMProviders';
import SystemConfig from './SystemConfig';

const Settings = () => {
  const items = [
    {
      key: 'gitlab',
      label: (
        <span>
          <GitlabOutlined />
          GitLab配置
        </span>
      ),
      children: <GitLabConfig />,
    },
    {
      key: 'providers',
      label: (
        <span>
          <ApiOutlined />
          LLM供应商
        </span>
      ),
      children: <LLMProviders />,
    },
    {
      key: 'system',
      label: (
        <span>
          <SettingOutlined />
          系统配置
        </span>
      ),
      children: <SystemConfig />,
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      <h2 style={{ marginBottom: 24 }}>系统设置</h2>
      <Tabs defaultActiveKey="gitlab" items={items} />
    </div>
  );
};

export default Settings;
