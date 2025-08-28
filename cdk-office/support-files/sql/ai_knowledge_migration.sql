-- 智能问答和知识库同步功能数据库迁移脚本
-- 创建时间：2025-01-27
-- 描述：包含AI智能问答、文档同步、权限管理的完整表结构和索引

-- ===========================================
-- 1. AI智能问答相关表
-- ===========================================

-- 知识问答表（如果不存在）
CREATE TABLE IF NOT EXISTS knowledge_qa (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    team_id UUID NOT NULL,
    question TEXT NOT NULL,
    answer TEXT NOT NULL,
    confidence DECIMAL(5,4) DEFAULT 0.0,
    message_id VARCHAR(255),
    feedback TEXT,
    ai_provider VARCHAR(50) DEFAULT 'dify',
    context JSONB,
    sources JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- AI服务配置表
CREATE TABLE IF NOT EXISTS ai_service_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name VARCHAR(100) NOT NULL,
    service_type VARCHAR(50) NOT NULL, -- ocr, ai_chat, ai_translate
    provider VARCHAR(50) NOT NULL, -- baidu, tencent, aliyun, openai, dify
    api_endpoint VARCHAR(255),
    api_key VARCHAR(255),
    secret_key VARCHAR(255),
    max_retries INTEGER DEFAULT 3,
    timeout INTEGER DEFAULT 30,
    is_enabled BOOLEAN DEFAULT true,
    is_default BOOLEAN DEFAULT false,
    priority INTEGER DEFAULT 0,
    config_data JSONB, -- 额外配置参数
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ===========================================
-- 2. 文档同步相关表
-- ===========================================

-- Dify文档同步表（如果不存在）
CREATE TABLE IF NOT EXISTS dify_document_syncs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL,
    dify_document_id VARCHAR(255),
    dataset_id VARCHAR(255),
    team_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    content TEXT,
    document_type VARCHAR(50),
    sync_status VARCHAR(50) DEFAULT 'pending', -- pending, processing, synced, failed
    indexing_status VARCHAR(50) DEFAULT 'pending', -- pending, processing, completed, failed
    error_message TEXT,
    metadata JSONB,
    created_by UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 文档内容提取日志表
CREATE TABLE IF NOT EXISTS document_extraction_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL,
    extraction_type VARCHAR(50), -- ocr, pdf_parse, text_extract
    status VARCHAR(50) DEFAULT 'pending', -- pending, processing, completed, failed
    extracted_content TEXT,
    content_summary TEXT,
    file_size BIGINT,
    processing_time INTEGER, -- 处理时间(毫秒)
    error_message TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ===========================================
-- 3. 权限管理相关表（Casbin）
-- ===========================================

-- Casbin规则表
CREATE TABLE IF NOT EXISTS casbin_rules (
    id SERIAL PRIMARY KEY,
    ptype VARCHAR(100) NOT NULL,
    v0 VARCHAR(100),
    v1 VARCHAR(100),
    v2 VARCHAR(100),
    v3 VARCHAR(100),
    v4 VARCHAR(100),
    v5 VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 用户角色管理表
CREATE TABLE IF NOT EXISTS user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    team_id UUID NOT NULL,
    role VARCHAR(50) NOT NULL, -- admin, manager, user, collaborator
    assigned_by UUID,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 权限审计日志表
CREATE TABLE IF NOT EXISTS permission_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    team_id UUID,
    action VARCHAR(100) NOT NULL, -- role_assigned, role_removed, permission_granted, permission_denied
    resource VARCHAR(100),
    permission VARCHAR(100),
    old_value TEXT,
    new_value TEXT,
    operator_id UUID,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ===========================================
-- 4. 服务状态监控表
-- ===========================================

-- 服务状态监控表
CREATE TABLE IF NOT EXISTS service_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL,
    service_type VARCHAR(50) NOT NULL, -- ai, ocr, sms, email
    status VARCHAR(50) DEFAULT 'healthy', -- healthy, degraded, unavailable
    response_time BIGINT, -- 响应时间(毫秒)
    success_rate DECIMAL(5,4) DEFAULT 1.0, -- 成功率 0-1
    error_count INTEGER DEFAULT 0,
    last_error TEXT,
    last_check_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ===========================================
-- 5. 创建索引优化查询性能
-- ===========================================

-- 知识问答表索引
CREATE INDEX IF NOT EXISTS idx_knowledge_qa_user_id ON knowledge_qa(user_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_qa_team_id ON knowledge_qa(team_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_qa_created_at ON knowledge_qa(created_at);
CREATE INDEX IF NOT EXISTS idx_knowledge_qa_message_id ON knowledge_qa(message_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_qa_user_team ON knowledge_qa(user_id, team_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_qa_provider ON knowledge_qa(ai_provider);
CREATE INDEX IF NOT EXISTS idx_knowledge_qa_confidence ON knowledge_qa(confidence);

-- AI服务配置表索引
CREATE INDEX IF NOT EXISTS idx_ai_service_configs_type ON ai_service_configs(service_type);
CREATE INDEX IF NOT EXISTS idx_ai_service_configs_provider ON ai_service_configs(provider);
CREATE INDEX IF NOT EXISTS idx_ai_service_configs_enabled ON ai_service_configs(is_enabled);
CREATE INDEX IF NOT EXISTS idx_ai_service_configs_default ON ai_service_configs(is_default);
CREATE INDEX IF NOT EXISTS idx_ai_service_configs_priority ON ai_service_configs(priority);

-- Dify文档同步表索引
CREATE INDEX IF NOT EXISTS idx_dify_document_syncs_document_id ON dify_document_syncs(document_id);
CREATE INDEX IF NOT EXISTS idx_dify_document_syncs_dify_id ON dify_document_syncs(dify_document_id);
CREATE INDEX IF NOT EXISTS idx_dify_document_syncs_team_id ON dify_document_syncs(team_id);
CREATE INDEX IF NOT EXISTS idx_dify_document_syncs_status ON dify_document_syncs(sync_status);
CREATE INDEX IF NOT EXISTS idx_dify_document_syncs_indexing_status ON dify_document_syncs(indexing_status);
CREATE INDEX IF NOT EXISTS idx_dify_document_syncs_created_at ON dify_document_syncs(created_at);
CREATE INDEX IF NOT EXISTS idx_dify_document_syncs_created_by ON dify_document_syncs(created_by);

-- 文档提取日志表索引
CREATE INDEX IF NOT EXISTS idx_document_extraction_logs_document_id ON document_extraction_logs(document_id);
CREATE INDEX IF NOT EXISTS idx_document_extraction_logs_status ON document_extraction_logs(status);
CREATE INDEX IF NOT EXISTS idx_document_extraction_logs_type ON document_extraction_logs(extraction_type);
CREATE INDEX IF NOT EXISTS idx_document_extraction_logs_created_at ON document_extraction_logs(created_at);

-- Casbin规则表索引
CREATE INDEX IF NOT EXISTS idx_casbin_rules_ptype ON casbin_rules(ptype);
CREATE INDEX IF NOT EXISTS idx_casbin_rules_v0 ON casbin_rules(v0);
CREATE INDEX IF NOT EXISTS idx_casbin_rules_v1 ON casbin_rules(v1);
CREATE INDEX IF NOT EXISTS idx_casbin_rules_v0_v1 ON casbin_rules(v0, v1);

-- 用户角色表索引
CREATE INDEX IF NOT EXISTS idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_team_id ON user_roles(team_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_role ON user_roles(role);
CREATE INDEX IF NOT EXISTS idx_user_roles_active ON user_roles(is_active);
CREATE INDEX IF NOT EXISTS idx_user_roles_user_team ON user_roles(user_id, team_id);
CREATE INDEX IF NOT EXISTS idx_user_roles_expires_at ON user_roles(expires_at);

-- 权限审计日志表索引
CREATE INDEX IF NOT EXISTS idx_permission_audit_logs_user_id ON permission_audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_permission_audit_logs_team_id ON permission_audit_logs(team_id);
CREATE INDEX IF NOT EXISTS idx_permission_audit_logs_action ON permission_audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_permission_audit_logs_created_at ON permission_audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_permission_audit_logs_operator ON permission_audit_logs(operator_id);

-- 服务状态表索引
CREATE INDEX IF NOT EXISTS idx_service_status_service_id ON service_status(service_id);
CREATE INDEX IF NOT EXISTS idx_service_status_type ON service_status(service_type);
CREATE INDEX IF NOT EXISTS idx_service_status_status ON service_status(status);
CREATE INDEX IF NOT EXISTS idx_service_status_last_check ON service_status(last_check_at);

-- ===========================================
-- 6. 创建唯一约束
-- ===========================================

-- AI服务配置唯一约束
CREATE UNIQUE INDEX IF NOT EXISTS idx_ai_service_configs_unique 
ON ai_service_configs(service_type, provider) WHERE is_enabled = true;

-- 用户角色唯一约束
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_roles_unique 
ON user_roles(user_id, team_id, role) WHERE is_active = true;

-- 文档同步唯一约束
CREATE UNIQUE INDEX IF NOT EXISTS idx_dify_document_syncs_unique 
ON dify_document_syncs(document_id) WHERE sync_status = 'synced';

-- ===========================================
-- 7. 创建检查约束
-- ===========================================

-- 知识问答置信度范围约束
ALTER TABLE knowledge_qa ADD CONSTRAINT chk_knowledge_qa_confidence 
CHECK (confidence >= 0.0 AND confidence <= 1.0);

-- AI服务配置超时时间约束
ALTER TABLE ai_service_configs ADD CONSTRAINT chk_ai_service_configs_timeout 
CHECK (timeout > 0 AND timeout <= 300);

-- AI服务配置重试次数约束
ALTER TABLE ai_service_configs ADD CONSTRAINT chk_ai_service_configs_retries 
CHECK (max_retries >= 0 AND max_retries <= 10);

-- 服务状态成功率约束
ALTER TABLE service_status ADD CONSTRAINT chk_service_status_success_rate 
CHECK (success_rate >= 0.0 AND success_rate <= 1.0);

-- 用户角色过期时间约束
ALTER TABLE user_roles ADD CONSTRAINT chk_user_roles_expires_at 
CHECK (expires_at IS NULL OR expires_at > assigned_at);

-- ===========================================
-- 8. 创建触发器函数
-- ===========================================

-- 更新时间戳触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为相关表添加更新时间触发器
CREATE TRIGGER update_knowledge_qa_updated_at BEFORE UPDATE ON knowledge_qa
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ai_service_configs_updated_at BEFORE UPDATE ON ai_service_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_dify_document_syncs_updated_at BEFORE UPDATE ON dify_document_syncs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_roles_updated_at BEFORE UPDATE ON user_roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_service_status_updated_at BEFORE UPDATE ON service_status
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_casbin_rules_updated_at BEFORE UPDATE ON casbin_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ===========================================
-- 9. 创建视图用于快速查询
-- ===========================================

-- 文档同步状态统计视图
CREATE OR REPLACE VIEW document_sync_stats AS
SELECT 
    team_id,
    COUNT(*) as total_syncs,
    COUNT(CASE WHEN sync_status = 'synced' THEN 1 END) as synced_count,
    COUNT(CASE WHEN sync_status = 'failed' THEN 1 END) as failed_count,
    COUNT(CASE WHEN sync_status = 'processing' THEN 1 END) as processing_count,
    COUNT(CASE WHEN sync_status = 'pending' THEN 1 END) as pending_count,
    AVG(CASE WHEN sync_status = 'synced' THEN 
        EXTRACT(EPOCH FROM (updated_at - created_at)) 
    END) as avg_sync_time_seconds
FROM dify_document_syncs
GROUP BY team_id;

-- AI问答统计视图
CREATE OR REPLACE VIEW ai_chat_stats AS
SELECT 
    team_id,
    COUNT(*) as total_chats,
    COUNT(CASE WHEN DATE(created_at) = CURRENT_DATE THEN 1 END) as today_chats,
    AVG(confidence) as avg_confidence,
    COUNT(CASE WHEN feedback IS NOT NULL THEN 1 END) as feedback_count,
    COUNT(DISTINCT user_id) as active_users
FROM knowledge_qa
GROUP BY team_id;

-- 用户权限概览视图
CREATE OR REPLACE VIEW user_permissions_overview AS
SELECT 
    ur.user_id,
    ur.team_id,
    ur.role,
    ur.assigned_at,
    ur.expires_at,
    ur.is_active,
    CASE 
        WHEN ur.expires_at IS NOT NULL AND ur.expires_at < CURRENT_TIMESTAMP THEN 'expired'
        WHEN ur.is_active = false THEN 'inactive'
        ELSE 'active'
    END as status
FROM user_roles ur
WHERE ur.is_active = true;

-- ===========================================
-- 10. 创建存储过程
-- ===========================================

-- 清理过期的权限审计日志
CREATE OR REPLACE FUNCTION cleanup_audit_logs(days_to_keep INTEGER DEFAULT 90)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM permission_audit_logs 
    WHERE created_at < CURRENT_TIMESTAMP - (days_to_keep || ' days')::INTERVAL;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 获取用户在团队中的有效权限
CREATE OR REPLACE FUNCTION get_user_effective_permissions(
    p_user_id UUID,
    p_team_id UUID
)
RETURNS TABLE(
    resource VARCHAR(100),
    action VARCHAR(100),
    granted_by VARCHAR(50)
) AS $$
BEGIN
    RETURN QUERY
    WITH user_roles_in_team AS (
        SELECT role
        FROM user_roles ur
        WHERE ur.user_id = p_user_id 
        AND ur.team_id = p_team_id 
        AND ur.is_active = true
        AND (ur.expires_at IS NULL OR ur.expires_at > CURRENT_TIMESTAMP)
    )
    SELECT 
        SPLIT_PART(cr.v1, '/', 2) as resource,
        cr.v2 as action,
        cr.v0 as granted_by
    FROM casbin_rules cr
    JOIN user_roles_in_team urt ON cr.v0 = urt.role
    WHERE cr.ptype = 'p'
    AND cr.v1 LIKE p_team_id::TEXT || '/%'
    
    UNION
    
    -- 直接分配给用户的权限
    SELECT 
        SPLIT_PART(cr.v1, '/', 2) as resource,
        cr.v2 as action,
        'direct' as granted_by
    FROM casbin_rules cr
    WHERE cr.ptype = 'p'
    AND cr.v0 = p_user_id::TEXT
    AND cr.v1 LIKE p_team_id::TEXT || '/%';
END;
$$ LANGUAGE plpgsql;

-- ===========================================
-- 11. 初始化默认数据
-- ===========================================

-- 插入默认AI服务配置
INSERT INTO ai_service_configs (service_name, service_type, provider, api_endpoint, is_enabled, is_default, priority)
VALUES 
('Dify智能问答', 'ai_chat', 'dify', 'https://api.dify.ai/v1', true, true, 1),
('百度OCR', 'ocr', 'baidu', 'https://aip.baidubce.com/rest/2.0/ocr/v1/general_basic', true, true, 1),
('腾讯云OCR', 'ocr', 'tencent', 'https://ocr.tencentcloudapi.com', true, false, 2)
ON CONFLICT DO NOTHING;

-- 插入默认Casbin权限规则
INSERT INTO casbin_rules (ptype, v0, v1, v2)
VALUES 
-- 超级管理员权限
('p', 'admin', '*/team', 'read'),
('p', 'admin', '*/team', 'write'),
('p', 'admin', '*/team', 'delete'),
('p', 'admin', '*/document', 'read'),
('p', 'admin', '*/document', 'write'),
('p', 'admin', '*/document', 'delete'),
('p', 'admin', '*/ai', 'read'),
('p', 'admin', '*/ai', 'write'),
('p', 'admin', '*/user', 'read'),
('p', 'admin', '*/user', 'write'),

-- 团队管理员权限
('p', 'manager', '*/team', 'read'),
('p', 'manager', '*/team', 'write'),
('p', 'manager', '*/document', 'read'),
('p', 'manager', '*/document', 'write'),
('p', 'manager', '*/ai', 'read'),
('p', 'manager', '*/ai', 'write'),
('p', 'manager', '*/user', 'read'),

-- 普通用户权限
('p', 'user', '*/team', 'read'),
('p', 'user', '*/document', 'read'),
('p', 'user', '*/document', 'write'),
('p', 'user', '*/ai', 'read'),
('p', 'user', '*/ai', 'write'),

-- 协作用户权限
('p', 'collaborator', '*/team', 'read'),
('p', 'collaborator', '*/document', 'read'),
('p', 'collaborator', '*/ai', 'read')
ON CONFLICT DO NOTHING;

-- ===========================================
-- 12. 创建示例数据（仅用于测试环境）
-- ===========================================

-- 注释掉示例数据，生产环境中不应该插入测试数据
/*
-- 插入测试用的知识问答记录
INSERT INTO knowledge_qa (user_id, team_id, question, answer, confidence, message_id, ai_provider) VALUES 
('550e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440001', '什么是CDK-Office？', 'CDK-Office是一个现代化的办公协作平台。', 0.95, 'msg_001', 'dify'),
('550e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440001', '如何上传文档？', '您可以通过拖拽或点击上传按钮来上传文档。', 0.88, 'msg_002', 'dify');

-- 插入测试用的文档同步记录
INSERT INTO dify_document_syncs (document_id, team_id, title, document_type, sync_status, created_by) VALUES 
('doc_001', '550e8400-e29b-41d4-a716-446655440001', '用户手册.pdf', 'pdf', 'synced', '550e8400-e29b-41d4-a716-446655440000'),
('doc_002', '550e8400-e29b-41d4-a716-446655440001', '项目计划.docx', 'docx', 'processing', '550e8400-e29b-41d4-a716-446655440000');
*/

-- ===========================================
-- 13. 添加表注释
-- ===========================================

COMMENT ON TABLE knowledge_qa IS 'AI智能问答记录表';
COMMENT ON TABLE ai_service_configs IS 'AI服务配置表';
COMMENT ON TABLE dify_document_syncs IS 'Dify文档同步记录表';
COMMENT ON TABLE document_extraction_logs IS '文档内容提取日志表';
COMMENT ON TABLE casbin_rules IS 'Casbin权限规则表';
COMMENT ON TABLE user_roles IS '用户角色管理表';
COMMENT ON TABLE permission_audit_logs IS '权限操作审计日志表';
COMMENT ON TABLE service_status IS '服务状态监控表';

COMMENT ON COLUMN knowledge_qa.confidence IS '回答置信度(0-1)';
COMMENT ON COLUMN knowledge_qa.context IS '问答上下文信息JSON';
COMMENT ON COLUMN knowledge_qa.sources IS '答案来源文档信息JSON';
COMMENT ON COLUMN dify_document_syncs.sync_status IS '同步状态: pending, processing, synced, failed';
COMMENT ON COLUMN dify_document_syncs.indexing_status IS '索引状态: pending, processing, completed, failed';
COMMENT ON COLUMN user_roles.expires_at IS '角色过期时间，NULL表示永不过期';
COMMENT ON COLUMN service_status.success_rate IS '服务成功率(0-1)';

-- ===========================================
-- 脚本执行完成
-- ===========================================

-- 显示创建的表信息
SELECT 
    table_name,
    table_comment
FROM information_schema.tables 
WHERE table_schema = 'public' 
    AND table_name IN (
        'knowledge_qa', 
        'ai_service_configs', 
        'dify_document_syncs', 
        'document_extraction_logs',
        'casbin_rules', 
        'user_roles', 
        'permission_audit_logs', 
        'service_status'
    )
ORDER BY table_name;