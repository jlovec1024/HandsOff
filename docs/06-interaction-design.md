# 页面交互逻辑设计

## 📋 文档概述

本文档详细说明前端页面的交互逻辑、状态管理、数据流、表单验证规则等内容。

---

## 1. 状态管理架构

### 1.1 技术方案

使用 **Zustand** 进行全局状态管理：

```typescript
// src/stores/auth.ts
import create from 'zustand';

interface AuthState {
  user: User | null;
  token: string | null;
  isAuthenticated: boolean;
  login: (credentials: LoginDto) => Promise<void>;
  logout: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  token: localStorage.getItem('token'),
  isAuthenticated: !!localStorage.getItem('token'),
  login: async (credentials) => {
    const res = await authApi.login(credentials);
    localStorage.setItem('token', res.data.token);
    set({ user: res.data.user, token: res.data.token, isAuthenticated: true });
  },
  logout: () => {
    localStorage.removeItem('token');
    set({ user: null, token: null, isAuthenticated: false });
  },
}));
```

### 1.2 Store 模块划分

| Store文件 | 职责 | 主要状态 |
|-----------|------|---------|
| `auth.ts` | 用户认证 | user, token, isAuthenticated |
| `platform.ts` | Git平台 | platforms, selectedPlatform |
| `repository.ts` | 代码仓库 | repositories, selectedRepo |
| `group.ts` | 仓库组 | groups, selectedGroup |
| `llm.ts` | LLM配置 | providers, models |
| `notification.ts` | 通知渠道 | channels |
| `review.ts` | Review记录 | reviews, currentReview |
| `autofix.ts` | 自动修复 | tasks, logs, branches |

---

## 2. 数据流设计

### 2.1 典型数据流（以创建Git平台为例）

```
┌────────────┐        ┌─────────────┐        ┌─────────────┐
│   用户操作  │───────▶│  表单组件    │───────▶│  Store Action│
│  (点击保存) │        │ (Form.onFinish)│      │ (createPlatform)│
└────────────┘        └─────────────┘        └─────────────┘
                                                    │
                                                    ▼
                                            ┌─────────────┐
                                            │   API 调用   │
                                            │ (platformApi.create)│
                                            └─────────────┘
                                                    │
                                                    ▼
                                            ┌─────────────┐
                                            │  后端处理    │
                                            └─────────────┘
                                                    │
                                                    ▼
                                            ┌─────────────┐
                                            │  更新Store   │───────▶ 触发组件重新渲染
                                            │ (set state)  │
                                            └─────────────┘
```

### 2.2 WebSocket实时数据流（自动修复日志）

```
┌────────────┐        ┌─────────────┐        ┌─────────────┐
│  后端Worker │───────▶│  WebSocket   │───────▶│  useWebSocket│
│  (执行修复) │  推送   │  服务器      │  接收   │    Hook      │
└────────────┘        └─────────────┘        └─────────────┘
                                                    │
                                                    ▼
                                            ┌─────────────┐
                                            │  更新日志列表 │
                                            │ (setLogs)    │
                                            └─────────────┘
                                                    │
                                                    ▼
                                            ┌─────────────┐
                                            │  LogViewer   │
                                            │  组件渲染    │
                                            └─────────────┘
```

**WebSocket Hook实现:**

```typescript
// src/hooks/useWebSocket.ts
import { useEffect, useState } from 'react';

interface LogMessage {
  timestamp: string;
  level: string;
  message: string;
}

export const useWebSocket = (url: string) => {
  const [logs, setLogs] = useState<LogMessage[]>([]);
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    const ws = new WebSocket(url);

    ws.onopen = () => setConnected(true);
    ws.onmessage = (event) => {
      const log: LogMessage = JSON.parse(event.data);
      setLogs((prev) => [...prev, log]);
    };
    ws.onerror = () => setConnected(false);
    ws.onclose = () => setConnected(false);

    return () => ws.close();
  }, [url]);

  return { logs, connected };
};
```

---

## 3. 核心交互流程

### 3.1 用户登录流程

```
1. 用户访问 /login
   ↓
2. 输入用户名和密码
   ↓
3. 点击"登录"按钮
   ↓
4. 触发 useAuthStore.login(credentials)
   ↓
5. 调用 POST /api/v1/auth/login
   ↓
6. 后端验证成功，返回 JWT Token
   ↓
7. Store 保存 token 到 localStorage
   ↓
8. 重定向到 / (Dashboard)
   ↓
9. App.tsx 中 ProtectedRoute 检查 token
   ↓
10. 显示主界面
```

### 3.2 创建Git平台配置流程

```
1. 用户进入 /platforms
   ↓
2. 点击"新建平台"按钮
   ↓
3. Modal 弹出表单
   ↓
4. 用户选择平台类型 (GitLab/GitHub/Gitea)
   ↓
5. 根据类型动态显示不同字段
   - GitLab: GitLab URL, Private Token
   - GitHub: GitHub URL, Personal Access Token
   - Gitea: Gitea URL, Access Token
   ↓
6. 用户填写表单
   ↓
7. 点击"测试连接"按钮 (可选)
   ↓
8. 调用 POST /api/v1/platforms/test-connection
   ↓
9. 显示测试结果 (成功/失败)
   ↓
10. 点击"保存"按钮
   ↓
11. 表单验证 (必填项、URL格式等)
   ↓
12. 调用 POST /api/v1/platforms
   ↓
13. 后端验证并保存到数据库
   ↓
14. 返回成功响应
   ↓
15. Store 更新平台列表
   ↓
16. 关闭 Modal，刷新列表
   ↓
17. 显示成功提示消息
```

### 3.3 导入代码仓库流程

```
1. 用户进入 /repositories
   ↓
2. 点击"导入仓库"按钮
   ↓
3. Modal 弹出导入向导
   ↓
4. 步骤1: 选择Git平台
   - 调用 GET /api/v1/platforms 获取平台列表
   - 显示平台下拉框
   ↓
5. 用户选择平台 (例如: GitLab A)
   ↓
6. 步骤2: 获取仓库列表
   - 调用 POST /api/v1/repositories/fetch-from-platform
   - 后端调用GitLab API获取仓库
   - 显示仓库列表 (Table with Checkbox)
   ↓
7. 用户选择要导入的仓库 (多选)
   ↓
8. 步骤3: 配置Webhook
   - 显示自定义回调URL输入框
   - 默认值: http://your-server.com/api/v1/webhooks/receive
   - 用户可修改
   ↓
9. 用户点击"导入"按钮
   ↓
10. 调用 POST /api/v1/repositories/batch-import
    - 请求体: { platform_id, repository_ids[], webhook_url }
   ↓
11. 后端处理:
    - 保存仓库到数据库
    - 为每个仓库创建 Webhook 配置
    - 调用 GitLab API 创建 Webhook
   ↓
12. 返回成功/失败结果
   ↓
13. 显示导入结果摘要
    - 成功: 10个
    - 失败: 2个 (显示失败原因)
   ↓
14. 关闭 Modal，刷新仓库列表
```

### 3.4 触发自动修复流程

```
1. 用户在 Review 详情页面 (/reviews/:id)
   ↓
2. 查看修复建议列表
   ↓
3. 点击某个建议的"修复"按钮
   ↓
4. Modal 弹出确认对话框
   - 显示建议详情
   - 显示将要创建的分支名 (例如: fix/issue-123)
   ↓
5. 用户点击"确认修复"
   ↓
6. 调用 POST /api/v1/auto-fix/tasks
   - 请求体: { suggestion_id, branch_name }
   ↓
7. 后端创建任务记录
   - 状态: pending
   - 返回 task_id
   ↓
8. 前端跳转到 /auto-fix/:taskId
   ↓
9. 页面建立 WebSocket 连接
   - ws://server/ws/fix-logs/:taskId
   ↓
10. 后端 Worker 异步执行修复:
    - clone 仓库
    - 创建分支
    - 调用 Snow-CLI 执行修复
    - commit 修改
    - push 分支
    - 每一步推送日志到 WebSocket
   ↓
11. 前端实时显示日志
   - [INFO] 正在克隆仓库...
   - [INFO] 创建分支 fix/issue-123...
   - [INFO] 执行 Snow-CLI...
   - [INFO] 提交修改...
   - [SUCCESS] 修复完成!
   ↓
12. 修复完成后:
   - 更新任务状态为 completed
   - 显示"查看分支"按钮
   - 显示"合并分支"按钮
```

### 3.5 配置仓库组提示词流程

```
1. 用户进入 /groups/:id/edit
   ↓
2. 切换到"提示词配置"Tab
   ↓
3. 显示 Monaco Editor
   - 默认加载系统提示词模板
   - 如果仓库组已有自定义提示词，加载自定义内容
   ↓
4. 用户编辑提示词
   ↓
5. 点击"预览"按钮 (可选)
   - Modal 显示渲染后的提示词
   ↓
6. 点击"保存"按钮
   ↓
7. 调用 PUT /api/v1/groups/:id/prompt
   - 请求体: { prompt_template }
   ↓
8. 后端保存到 prompt_templates 表
   ↓
9. 返回成功响应
   ↓
10. 显示成功提示消息
```

---

## 4. 表单验证规则

### 4.1 Git平台配置表单

```typescript
// src/components/PlatformForm.tsx
const rules = {
  name: [
    { required: true, message: '请输入平台名称' },
    { max: 100, message: '名称不能超过100个字符' },
  ],
  type: [
    { required: true, message: '请选择平台类型' },
  ],
  url: [
    { required: true, message: '请输入平台URL' },
    { type: 'url', message: '请输入有效的URL格式' },
  ],
  token: [
    { required: true, message: '请输入访问Token' },
    { min: 20, message: 'Token长度至少20个字符' },
  ],
};
```

### 4.2 仓库组配置表单

```typescript
const rules = {
  name: [
    { required: true, message: '请输入仓库组名称' },
    { max: 100, message: '名称不能超过100个字符' },
  ],
  description: [
    { max: 500, message: '描述不能超过500个字符' },
  ],
  repositories: [
    { required: true, message: '请至少选择一个代码仓库' },
  ],
  llm_model_id: [
    { required: true, message: '请选择LLM模型' },
  ],
  notification_channel_ids: [
    // 可选字段，无验证规则
  ],
};
```

### 4.3 LLM供应商配置表单

```typescript
const rules = {
  name: [
    { required: true, message: '请输入供应商名称' },
  ],
  type: [
    { required: true, message: '请选择供应商类型' },
  ],
  base_url: [
    { required: true, message: '请输入API Base URL' },
    { type: 'url', message: '请输入有效的URL格式' },
  ],
  api_key: [
    { required: true, message: '请输入API Key' },
  ],
};
```

### 4.4 通知渠道配置表单

```typescript
const rules = {
  name: [
    { required: true, message: '请输入渠道名称' },
  ],
  type: [
    { required: true, message: '请选择渠道类型' },
  ],
  webhook_url: [
    { required: true, message: '请输入Webhook URL' },
    { type: 'url', message: '请输入有效的URL格式' },
  ],
  secret: [
    // 飞书需要，其他可选
    { 
      required: (form) => form.type === 'feishu',
      message: '飞书渠道需要提供Secret',
    },
  ],
};
```

---

## 5. 错误处理策略

### 5.1 API请求错误处理

```typescript
// src/utils/request.ts
import axios from 'axios';
import { message } from 'antd';

const instance = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
});

// 请求拦截器
instance.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => Promise.reject(error)
);

// 响应拦截器
instance.interceptors.response.use(
  (response) => {
    const { code, message: msg, data } = response.data;
    if (code !== 0) {
      message.error(msg || '请求失败');
      return Promise.reject(new Error(msg));
    }
    return data;
  },
  (error) => {
    if (error.response) {
      switch (error.response.status) {
        case 401:
          message.error('未授权，请重新登录');
          localStorage.removeItem('token');
          window.location.href = '/login';
          break;
        case 403:
          message.error('没有权限访问');
          break;
        case 404:
          message.error('请求的资源不存在');
          break;
        case 500:
          message.error('服务器错误');
          break;
        default:
          message.error(error.response.data.message || '请求失败');
      }
    } else if (error.request) {
      message.error('网络连接失败');
    } else {
      message.error('请求配置错误');
    }
    return Promise.reject(error);
  }
);

export default instance;
```

### 5.2 表单提交错误处理

```typescript
const handleSubmit = async (values: any) => {
  try {
    setSubmitting(true);
    await platformApi.create(values);
    message.success('创建成功');
    onSuccess();
  } catch (error) {
    // 错误已经在 axios 拦截器中处理
    // 这里不需要再次显示错误消息
  } finally {
    setSubmitting(false);
  }
};
```

### 5.3 WebSocket连接错误处理

```typescript
const useWebSocket = (url: string) => {
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const reconnectTimeoutRef = useRef<number>();

  const connect = () => {
    const ws = new WebSocket(url);

    ws.onopen = () => {
      setConnected(true);
      setError(null);
    };

    ws.onerror = () => {
      setError('WebSocket连接失败');
      setConnected(false);
    };

    ws.onclose = () => {
      setConnected(false);
      // 自动重连
      reconnectTimeoutRef.current = window.setTimeout(() => {
        connect();
      }, 3000);
    };

    return ws;
  };

  useEffect(() => {
    const ws = connect();
    return () => {
      ws.close();
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
    };
  }, [url]);

  return { connected, error };
};
```

---

## 6. 加载状态管理

### 6.1 页面级加载状态

```typescript
// src/pages/Platforms/index.tsx
const [loading, setLoading] = useState(true);

useEffect(() => {
  const fetchPlatforms = async () => {
    try {
      setLoading(true);
      const data = await platformApi.list();
      setPlatforms(data);
    } catch (error) {
      // 错误已处理
    } finally {
      setLoading(false);
    }
  };

  fetchPlatforms();
}, []);

return (
  <Spin spinning={loading}>
    <Table dataSource={platforms} />
  </Spin>
);
```

### 6.2 按钮级加载状态

```typescript
const [testingConnection, setTestingConnection] = useState(false);

const handleTestConnection = async () => {
  try {
    setTestingConnection(true);
    const result = await platformApi.testConnection(formValues);
    if (result.connected) {
      message.success('连接成功');
    } else {
      message.error(`连接失败: ${result.message}`);
    }
  } catch (error) {
    // 错误已处理
  } finally {
    setTestingConnection(false);
  }
};

return (
  <Button
    onClick={handleTestConnection}
    loading={testingConnection}
  >
    测试连接
  </Button>
);
```

---

## 7. 权限控制

### 7.1 页面级权限

```typescript
// src/components/ProtectedRoute.tsx
import { Navigate } from 'react-router-dom';
import { useAuthStore } from '@/stores/auth';

export const ProtectedRoute = ({ children, requiredRole }: any) => {
  const { isAuthenticated, user } = useAuthStore();

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  if (requiredRole && user?.role !== requiredRole) {
    return <Navigate to="/403" replace />;
  }

  return children;
};
```

### 7.2 操作级权限

```typescript
// src/components/OperationButtons.tsx
const { user } = useAuthStore();

return (
  <>
    {user?.role === 'admin' && (
      <Button onClick={handleDelete}>删除</Button>
    )}
  </>
);
```

---

## 8. 实时通信设计

### 8.1 自动修复日志推送

**WebSocket URL:**
```
ws://server/ws/fix-logs/:taskId
```

**消息格式:**
```json
{
  "timestamp": "2025-01-30T12:34:56Z",
  "level": "info",
  "message": "正在克隆仓库..."
}
```

**日志级别:**
- `info`: 普通信息
- `warning`: 警告信息
- `error`: 错误信息
- `success`: 成功信息

### 8.2 前端日志展示

```typescript
// src/components/LogViewer.tsx
import { useWebSocket } from '@/hooks/useWebSocket';

const LogViewer = ({ taskId }: { taskId: string }) => {
  const { logs, connected } = useWebSocket(
    `ws://${window.location.host}/ws/fix-logs/${taskId}`
  );

  return (
    <div>
      <div>
        状态: {connected ? '已连接' : '未连接'}
      </div>
      <div>
        {logs.map((log, index) => (
          <div key={index} className={`log-${log.level}`}>
            [{log.timestamp}] [{log.level}] {log.message}
          </div>
        ))}
      </div>
    </div>
  );
};
```

---

## 9. 性能优化策略

### 9.1 列表虚拟滚动

```typescript
import { List } from 'react-virtualized';

<List
  width={800}
  height={600}
  rowCount={logs.length}
  rowHeight={30}
  rowRenderer={({ index, key, style }) => (
    <div key={key} style={style}>
      {logs[index].message}
    </div>
  )}
/>
```

### 9.2 表格分页

```typescript
const [pagination, setPagination] = useState({
  current: 1,
  pageSize: 20,
  total: 0,
});

const handleTableChange = (newPagination: any) => {
  setPagination(newPagination);
  fetchData(newPagination.current, newPagination.pageSize);
};

<Table
  dataSource={data}
  pagination={pagination}
  onChange={handleTableChange}
/>
```

### 9.3 防抖与节流

```typescript
import { debounce } from 'lodash';

const handleSearch = debounce((value: string) => {
  fetchData({ keyword: value });
}, 300);

<Input.Search
  onChange={(e) => handleSearch(e.target.value)}
  placeholder="搜索..."
/>
```

---

## 10. 用户体验优化

### 10.1 操作确认

```typescript
import { Modal } from 'antd';

const handleDelete = (id: string) => {
  Modal.confirm({
    title: '确认删除',
    content: '删除后无法恢复，确定要删除吗？',
    onOk: async () => {
      await platformApi.delete(id);
      message.success('删除成功');
      fetchData();
    },
  });
};
```

### 10.2 操作反馈

```typescript
// 成功提示
message.success('操作成功');

// 错误提示
message.error('操作失败');

// 警告提示
message.warning('请先选择平台');

// 加载提示
const hide = message.loading('正在处理...', 0);
// ... 处理完成后
hide();
```

### 10.3 空状态处理

```typescript
<Table
  dataSource={data}
  locale={{
    emptyText: (
      <Empty
        description="暂无数据"
        image={Empty.PRESENTED_IMAGE_SIMPLE}
      >
        <Button type="primary" onClick={handleCreate}>
          新建平台
        </Button>
      </Empty>
    ),
  }}
/>
```

---

## 11. 典型交互场景完整示例

### 11.1 创建仓库组并配置提示词

```typescript
// src/pages/Groups/CreateGroup.tsx
import { useState } from 'react';
import { Form, Input, Select, Button, Steps, message } from 'antd';
import { useNavigate } from 'react-router-dom';
import MonacoEditor from '@monaco-editor/react';

const CreateGroup = () => {
  const [current, setCurrent] = useState(0);
  const [groupId, setGroupId] = useState<string>();
  const [form] = Form.useForm();
  const navigate = useNavigate();

  // 步骤1: 创建基本信息
  const handleCreateBasicInfo = async (values: any) => {
    const res = await groupApi.create(values);
    setGroupId(res.id);
    setCurrent(1);
  };

  // 步骤2: 添加仓库
  const handleAddRepositories = async (values: any) => {
    await groupApi.addRepositories(groupId!, values.repository_ids);
    setCurrent(2);
  };

  // 步骤3: 配置提示词
  const handleConfigurePrompt = async (values: any) => {
    await groupApi.updatePrompt(groupId!, values.prompt_template);
    message.success('创建成功');
    navigate('/groups');
  };

  return (
    <Steps current={current}>
      <Steps.Step title="基本信息" />
      <Steps.Step title="选择仓库" />
      <Steps.Step title="配置提示词" />
    </Steps>

    {current === 0 && (
      <Form form={form} onFinish={handleCreateBasicInfo}>
        <Form.Item name="name" label="仓库组名称" rules={[{ required: true }]}>
          <Input />
        </Form.Item>
        <Form.Item name="description" label="描述">
          <Input.TextArea />
        </Form.Item>
        <Button type="primary" htmlType="submit">
          下一步
        </Button>
      </Form>
    )}

    {current === 1 && (
      <Form form={form} onFinish={handleAddRepositories}>
        <Form.Item name="repository_ids" label="选择仓库" rules={[{ required: true }]}>
          <Select mode="multiple">
            {/* 仓库列表 */}
          </Select>
        </Form.Item>
        <Button onClick={() => setCurrent(0)}>上一步</Button>
        <Button type="primary" htmlType="submit">
          下一步
        </Button>
      </Form>
    )}

    {current === 2 && (
      <Form form={form} onFinish={handleConfigurePrompt}>
        <Form.Item name="prompt_template" label="提示词模板">
          <MonacoEditor
            height="400px"
            language="markdown"
            theme="vs-dark"
            defaultValue={DEFAULT_PROMPT_TEMPLATE}
          />
        </Form.Item>
        <Button onClick={() => setCurrent(1)}>上一步</Button>
        <Button type="primary" htmlType="submit">
          完成
        </Button>
      </Form>
    )}
  );
};
```

---

## 12. 总结

本文档详细说明了前端页面的交互逻辑，包括：

1. ✅ **状态管理**: Zustand Store 模块化设计
2. ✅ **数据流**: 组件 → Store → API → 后端
3. ✅ **核心流程**: 登录、创建平台、导入仓库、触发修复等
4. ✅ **表单验证**: 完整的验证规则定义
5. ✅ **错误处理**: 统一的错误拦截和提示
6. ✅ **加载状态**: 页面级和按钮级加载状态
7. ✅ **权限控制**: 页面级和操作级权限
8. ✅ **实时通信**: WebSocket 日志推送
9. ✅ **性能优化**: 虚拟滚动、分页、防抖
10. ✅ **用户体验**: 确认对话框、操作反馈、空状态

配合 [05-page-design.md](./05-page-design.md) 和 [07-api-design.md](./07-api-design.md)，可以完整实现前端功能。

---

**设计版本**: v1.0  
**最后更新**: 2025-01-30
