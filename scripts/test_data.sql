-- ============================================
-- HandsOff 测试数据初始化脚本
-- ============================================
-- 用途: 快速创建测试环境所需的配置数据
-- 使用: 在数据库初始化后运行此脚本
-- ============================================

-- ============================================
-- 1. LLM Provider 配置
-- ============================================

-- DeepSeek Provider
INSERT INTO llm_providers (name, type, api_endpoint, api_key, enabled, created_at, updated_at)
VALUES (
    'DeepSeek Test',
    'deepseek',
    'https://api.deepseek.com/v1',
    'ENCRYPTED_API_KEY_HERE',  -- 需要先加密，或通过 API 创建
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- OpenAI Provider (可选)
INSERT INTO llm_providers (name, type, api_endpoint, api_key, enabled, created_at, updated_at)
VALUES (
    'OpenAI Test',
    'openai',
    'https://api.openai.com/v1',
    'ENCRYPTED_API_KEY_HERE',  -- 需要先加密，或通过 API 创建
    0,  -- 默认禁用
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- ============================================
-- 2. LLM Model 配置
-- ============================================

-- DeepSeek Chat Model
INSERT INTO llm_models (provider_id, name, model_name, model_id, max_tokens, temperature, prompt_template, enabled, created_at, updated_at)
VALUES (
    1,  -- 对应上面的 DeepSeek Provider ID
    'DeepSeek Chat',
    'deepseek-chat',
    'deepseek-chat',
    4096,
    0.7,
    NULL,  -- 使用默认提示词模板
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- OpenAI GPT-4 Model (可选)
INSERT INTO llm_models (provider_id, name, model_name, model_id, max_tokens, temperature, prompt_template, enabled, created_at, updated_at)
VALUES (
    2,  -- 对应上面的 OpenAI Provider ID
    'GPT-4',
    'gpt-4',
    'gpt-4',
    4096,
    0.7,
    NULL,
    0,  -- 默认禁用
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- ============================================
-- 3. Git Platform 配置 (GitLab)
-- ============================================

INSERT INTO git_platform_configs (name, platform_type, base_url, access_token, webhook_secret, enabled, created_at, updated_at)
VALUES (
    'GitLab Test Instance',
    'gitlab',
    'https://gitlab.com',  -- 或你的 GitLab 实例地址
    'glpat-xxxxxxxxxxxxxxxxxxxx',  -- 替换为真实 GitLab Access Token
    'test-webhook-secret',  -- 替换为你的 Webhook Secret
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- ============================================
-- 4. Repository 配置
-- ============================================

INSERT INTO repositories (
    platform_id,
    llm_model_id,
    name,
    full_name,
    git_url,
    default_branch,
    platform_project_id,
    auto_review_enabled,
    review_on_mr_open,
    review_on_mr_update,
    min_review_score,
    created_at,
    updated_at
)
VALUES (
    1,  -- 对应上面的 GitLab Platform ID
    1,  -- 对应上面的 DeepSeek Model ID
    'test-project',
    'your-username/test-project',  -- 替换为真实项目路径
    'https://gitlab.com/your-username/test-project.git',
    'main',
    12345,  -- 替换为真实 GitLab Project ID
    1,  -- 启用自动审查
    1,  -- MR 打开时审查
    1,  -- MR 更新时审查
    60, -- 最低质量分数
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- ============================================
-- 5. 验证数据
-- ============================================

-- 查看已创建的数据
SELECT 'LLM Providers:' as table_name;
SELECT id, name, type, api_endpoint FROM llm_providers;

SELECT 'LLM Models:' as table_name;
SELECT id, provider_id, name, model_name, max_tokens, temperature FROM llm_models;

SELECT 'Git Platforms:' as table_name;
SELECT id, name, platform_type, base_url FROM git_platform_configs;

SELECT 'Repositories:' as table_name;
SELECT id, platform_id, llm_model_id, name, platform_project_id, auto_review_enabled FROM repositories;

-- ============================================
-- 使用说明
-- ============================================

/*
1. 修改配置值:
   - llm_providers.api_key: 需要先通过加密工具加密 API Key
   - git_platform_configs.access_token: 替换为真实的 GitLab Access Token
   - git_platform_configs.webhook_secret: 设置 Webhook 密钥
   - repositories.platform_project_id: 替换为真实的 GitLab Project ID
   - repositories.full_name 和 git_url: 替换为真实的项目路径

2. 加密 API Key (推荐使用工具脚本):
   # 方式 1: 使用测试工具
   go run tools/encrypt_apikey/main.go --key "your-api-key" --encryption-key "your-encryption-key"
   
   # 方式 2: 通过 API 创建 (自动加密)
   curl -X POST http://localhost:8080/api/llm/providers \
     -H "Content-Type: application/json" \
     -d '{
       "name": "DeepSeek",
       "type": "deepseek",
       "api_endpoint": "https://api.deepseek.com/v1",
       "api_key": "sk-xxxxxxxx"
     }'

3. 获取 GitLab Project ID:
   - 访问 GitLab 项目页面
   - 在项目名称下方可以看到 "Project ID: 12345"

4. 生成 GitLab Access Token:
   - GitLab 个人设置 → Access Tokens
   - 权限: api, read_api, read_repository, write_repository
   - 复制生成的 token (glpat-xxxxxxxxxxxx)

5. 配置 Webhook:
   - 项目设置 → Webhooks
   - URL: http://your-server:8080/webhook/gitlab
   - Secret Token: 与 webhook_secret 一致
   - 触发事件: Merge request events
*/
