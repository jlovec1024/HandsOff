import { Tabs } from 'antd';
import { SettingOutlined, GitlabOutlined, RobotOutlined, ApiOutlined } from '@ant-design/icons';
import GitLabConfig from './GitLabConfig';
import LLMProviders from './LLMProviders';
import LLMModels from './LLMModels';
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
      key: 'models',
      label: (
        <span>
          <RobotOutlined />
          LLM模型
        </span>
      ),
      children: <LLMModels />,
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
