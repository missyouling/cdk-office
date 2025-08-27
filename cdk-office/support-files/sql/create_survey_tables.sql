-- SurveyJS调查问卷模块数据库表结构
-- 创建时间：2025-01-27
-- 描述：包含问卷管理、响应收集、智能分析和权限控制的完整表结构

-- 1. 问卷主表
CREATE TABLE IF NOT EXISTS surveys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_id VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    json_definition JSONB NOT NULL,
    created_by UUID NOT NULL,
    team_id UUID NOT NULL,
    status VARCHAR(50) DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'closed', 'archived')),
    is_public BOOLEAN DEFAULT false,
    max_responses INTEGER DEFAULT 0,
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    tags TEXT,
    response_count INTEGER DEFAULT 0,
    view_count INTEGER DEFAULT 0,
    share_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. 问卷响应表
CREATE TABLE IF NOT EXISTS survey_responses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_id VARCHAR(100) NOT NULL,
    user_id UUID,
    team_id UUID NOT NULL,
    response_data JSONB NOT NULL,
    time_spent INTEGER DEFAULT 0,
    ip_address INET,
    user_agent VARCHAR(500),
    is_completed BOOLEAN DEFAULT true,
    completed_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. 问卷分析表
CREATE TABLE IF NOT EXISTS survey_analysis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_id VARCHAR(100) NOT NULL,
    analysis_type VARCHAR(50) NOT NULL CHECK (analysis_type IN ('basic', 'ai', 'custom', 'batch_ai')),
    result_data JSONB NOT NULL,
    dify_workflow_id VARCHAR(100),
    run_id VARCHAR(100),
    status VARCHAR(50) DEFAULT 'pending' CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. 问卷权限表
CREATE TABLE IF NOT EXISTS survey_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_id VARCHAR(100) NOT NULL,
    user_id UUID NOT NULL,
    team_id UUID NOT NULL,
    can_view BOOLEAN DEFAULT true,
    can_edit BOOLEAN DEFAULT false,
    can_delete BOOLEAN DEFAULT false,
    can_manage BOOLEAN DEFAULT false,
    can_analyze BOOLEAN DEFAULT false,
    can_export BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(survey_id, user_id)
);

-- 5. 问卷模板表
CREATE TABLE IF NOT EXISTS survey_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    json_template JSONB NOT NULL,
    preview_image VARCHAR(500),
    is_public BOOLEAN DEFAULT false,
    use_count INTEGER DEFAULT 0,
    rating REAL DEFAULT 0 CHECK (rating >= 0 AND rating <= 5),
    created_by UUID NOT NULL,
    team_id UUID,
    tags TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 6. 问卷文件表
CREATE TABLE IF NOT EXISTS survey_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    survey_id VARCHAR(100) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100),
    storage_type VARCHAR(50) NOT NULL CHECK (storage_type IN ('local', 'cloud_db', 's3', 'webdav')),
    file_url VARCHAR(500),
    uploaded_by UUID NOT NULL,
    team_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引以优化查询性能

-- 问卷表索引
CREATE INDEX IF NOT EXISTS idx_surveys_team_id ON surveys(team_id);
CREATE INDEX IF NOT EXISTS idx_surveys_created_by ON surveys(created_by);
CREATE INDEX IF NOT EXISTS idx_surveys_status ON surveys(status);
CREATE INDEX IF NOT EXISTS idx_surveys_is_public ON surveys(is_public);
CREATE INDEX IF NOT EXISTS idx_surveys_created_at ON surveys(created_at);
CREATE INDEX IF NOT EXISTS idx_surveys_tags ON surveys USING GIN(to_tsvector('english', tags));

-- 问卷响应表索引
CREATE INDEX IF NOT EXISTS idx_survey_responses_survey_id ON survey_responses(survey_id);
CREATE INDEX IF NOT EXISTS idx_survey_responses_user_id ON survey_responses(user_id);
CREATE INDEX IF NOT EXISTS idx_survey_responses_team_id ON survey_responses(team_id);
CREATE INDEX IF NOT EXISTS idx_survey_responses_completed_at ON survey_responses(completed_at);
CREATE INDEX IF NOT EXISTS idx_survey_responses_ip_address ON survey_responses(ip_address);

-- 问卷分析表索引
CREATE INDEX IF NOT EXISTS idx_survey_analysis_survey_id ON survey_analysis(survey_id);
CREATE INDEX IF NOT EXISTS idx_survey_analysis_type ON survey_analysis(analysis_type);
CREATE INDEX IF NOT EXISTS idx_survey_analysis_status ON survey_analysis(status);
CREATE INDEX IF NOT EXISTS idx_survey_analysis_workflow_id ON survey_analysis(dify_workflow_id);
CREATE INDEX IF NOT EXISTS idx_survey_analysis_run_id ON survey_analysis(run_id);

-- 问卷权限表索引
CREATE INDEX IF NOT EXISTS idx_survey_permissions_survey_id ON survey_permissions(survey_id);
CREATE INDEX IF NOT EXISTS idx_survey_permissions_user_id ON survey_permissions(user_id);
CREATE INDEX IF NOT EXISTS idx_survey_permissions_team_id ON survey_permissions(team_id);

-- 问卷模板表索引
CREATE INDEX IF NOT EXISTS idx_survey_templates_category ON survey_templates(category);
CREATE INDEX IF NOT EXISTS idx_survey_templates_is_public ON survey_templates(is_public);
CREATE INDEX IF NOT EXISTS idx_survey_templates_created_by ON survey_templates(created_by);
CREATE INDEX IF NOT EXISTS idx_survey_templates_team_id ON survey_templates(team_id);
CREATE INDEX IF NOT EXISTS idx_survey_templates_use_count ON survey_templates(use_count);
CREATE INDEX IF NOT EXISTS idx_survey_templates_rating ON survey_templates(rating);

-- 问卷文件表索引
CREATE INDEX IF NOT EXISTS idx_survey_files_survey_id ON survey_files(survey_id);
CREATE INDEX IF NOT EXISTS idx_survey_files_uploaded_by ON survey_files(uploaded_by);
CREATE INDEX IF NOT EXISTS idx_survey_files_team_id ON survey_files(team_id);
CREATE INDEX IF NOT EXISTS idx_survey_files_storage_type ON survey_files(storage_type);

-- 创建外键约束（如果需要的话，取决于实际的用户表结构）
-- 注意：这些约束需要根据实际的表结构调整

-- ALTER TABLE surveys ADD CONSTRAINT fk_surveys_created_by 
--     FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE;

-- ALTER TABLE surveys ADD CONSTRAINT fk_surveys_team_id 
--     FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;

-- ALTER TABLE survey_responses ADD CONSTRAINT fk_survey_responses_user_id 
--     FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

-- ALTER TABLE survey_responses ADD CONSTRAINT fk_survey_responses_team_id 
--     FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;

-- ALTER TABLE survey_permissions ADD CONSTRAINT fk_survey_permissions_user_id 
--     FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- ALTER TABLE survey_permissions ADD CONSTRAINT fk_survey_permissions_team_id 
--     FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;

-- 创建触发器以自动更新updated_at字段
CREATE OR REPLACE FUNCTION update_survey_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为各表添加更新时间触发器
CREATE TRIGGER update_surveys_updated_at BEFORE UPDATE ON surveys
    FOR EACH ROW EXECUTE FUNCTION update_survey_updated_at_column();

CREATE TRIGGER update_survey_responses_updated_at BEFORE UPDATE ON survey_responses
    FOR EACH ROW EXECUTE FUNCTION update_survey_updated_at_column();

CREATE TRIGGER update_survey_analysis_updated_at BEFORE UPDATE ON survey_analysis
    FOR EACH ROW EXECUTE FUNCTION update_survey_updated_at_column();

CREATE TRIGGER update_survey_permissions_updated_at BEFORE UPDATE ON survey_permissions
    FOR EACH ROW EXECUTE FUNCTION update_survey_updated_at_column();

CREATE TRIGGER update_survey_templates_updated_at BEFORE UPDATE ON survey_templates
    FOR EACH ROW EXECUTE FUNCTION update_survey_updated_at_column();

CREATE TRIGGER update_survey_files_updated_at BEFORE UPDATE ON survey_files
    FOR EACH ROW EXECUTE FUNCTION update_survey_updated_at_column();

-- 插入一些示例模板数据
INSERT INTO survey_templates (name, description, category, json_template, is_public, created_by, tags) VALUES
('员工满意度调查', '评估员工对工作环境、薪酬福利、管理制度等方面的满意度', '人事管理', 
'{"title": "员工满意度调查", "pages": [{"elements": [{"type": "radiogroup", "name": "satisfaction", "title": "您对目前的工作满意度如何？", "choices": ["非常满意", "满意", "一般", "不满意", "非常不满意"]}, {"type": "comment", "name": "suggestions", "title": "请提供您的改进建议："}]}]}', 
true, '00000000-0000-0000-0000-000000000000', '人事,满意度,员工'),

('产品反馈问卷', '收集用户对产品功能、界面、性能等方面的反馈意见', '市场调研',
'{"title": "产品反馈问卷", "pages": [{"elements": [{"type": "rating", "name": "overall_rating", "title": "请为产品整体体验打分：", "rateMin": 1, "rateMax": 5}, {"type": "checkbox", "name": "features", "title": "您最喜欢的功能有哪些？", "choices": ["界面设计", "功能完整性", "操作便捷性", "响应速度", "稳定性"]}, {"type": "comment", "name": "feedback", "title": "其他意见和建议："}]}]}',
true, '00000000-0000-0000-0000-000000000000', '产品,反馈,用户体验'),

('培训效果评估', '评估培训课程的内容质量、讲师水平和学习效果', '教育培训',
'{"title": "培训效果评估", "pages": [{"elements": [{"type": "radiogroup", "name": "content_quality", "title": "培训内容的实用性如何？", "choices": ["非常实用", "比较实用", "一般", "不太实用", "完全不实用"]}, {"type": "radiogroup", "name": "instructor_rating", "title": "对讲师的评价：", "choices": ["优秀", "良好", "一般", "较差", "很差"]}, {"type": "text", "name": "learning_outcome", "title": "请简述您的主要收获："}]}]}',
true, '00000000-0000-0000-0000-000000000000', '培训,教育,评估'),

('客户服务评价', '评价客户服务质量，包括响应速度、问题解决能力等', '服务质量',
'{"title": "客户服务评价", "pages": [{"elements": [{"type": "rating", "name": "service_rating", "title": "请为本次服务打分：", "rateMin": 1, "rateMax": 5}, {"type": "radiogroup", "name": "response_time", "title": "服务响应速度如何？", "choices": ["非常快", "比较快", "一般", "比较慢", "很慢"]}, {"type": "boolean", "name": "recommend", "title": "您是否会向朋友推荐我们的服务？"}, {"type": "comment", "name": "additional_feedback", "title": "其他意见或建议："}]}]}',
true, '00000000-0000-0000-0000-000000000000', '客户服务,质量评价,满意度');

-- 创建视图以便查询
CREATE OR REPLACE VIEW survey_stats AS
SELECT 
    s.survey_id,
    s.title,
    s.status,
    s.is_public,
    s.response_count,
    s.view_count,
    s.created_at,
    COUNT(sr.id) as actual_responses,
    AVG(sr.time_spent) as avg_time_spent,
    MAX(sr.completed_at) as last_response_at
FROM surveys s
LEFT JOIN survey_responses sr ON s.survey_id = sr.survey_id
GROUP BY s.survey_id, s.title, s.status, s.is_public, s.response_count, s.view_count, s.created_at;

COMMENT ON TABLE surveys IS '问卷主表，存储问卷基本信息和SurveyJS定义';
COMMENT ON TABLE survey_responses IS '问卷响应表，存储用户提交的问卷答案';
COMMENT ON TABLE survey_analysis IS '问卷分析表，存储AI分析结果';
COMMENT ON TABLE survey_permissions IS '问卷权限表，控制用户对问卷的操作权限';
COMMENT ON TABLE survey_templates IS '问卷模板表，存储可重用的问卷模板';
COMMENT ON TABLE survey_files IS '问卷文件表，存储问卷相关的附件';

COMMENT ON COLUMN surveys.json_definition IS 'SurveyJS JSON格式的问卷定义';
COMMENT ON COLUMN surveys.survey_id IS '问卷唯一标识符，用于前端访问';
COMMENT ON COLUMN survey_responses.response_data IS '用户提交的问卷答案，JSON格式';
COMMENT ON COLUMN survey_analysis.result_data IS 'AI分析结果数据，JSON格式';