-- Service Status Table Migration
-- 创建服务健康状态表
-- 2025-01-20

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 创建服务状态表
CREATE TABLE IF NOT EXISTS service_statuses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_name VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('healthy', 'unhealthy', 'degraded')),
    response_time BIGINT DEFAULT 0, -- 响应时间（毫秒）
    status_code INTEGER DEFAULT 0,  -- HTTP状态码
    error_message TEXT,             -- 错误信息
    details TEXT,                   -- JSON格式的详细信息
    checked_at TIMESTAMP NOT NULL,  -- 检查时间
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_service_statuses_service_name ON service_statuses (service_name);
CREATE INDEX IF NOT EXISTS idx_service_statuses_checked_at ON service_statuses (checked_at);
CREATE INDEX IF NOT EXISTS idx_service_statuses_status ON service_statuses (status);
CREATE INDEX IF NOT EXISTS idx_service_statuses_service_checked ON service_statuses (service_name, checked_at DESC);

-- 创建复合索引用于查询最新状态
CREATE INDEX IF NOT EXISTS idx_service_statuses_latest ON service_statuses (service_name, checked_at DESC);

-- 添加表注释
COMMENT ON TABLE service_statuses IS '服务健康状态记录表';
COMMENT ON COLUMN service_statuses.id IS '主键ID';
COMMENT ON COLUMN service_statuses.service_name IS '服务名称';
COMMENT ON COLUMN service_statuses.status IS '服务状态：healthy(健康), unhealthy(不健康), degraded(降级)';
COMMENT ON COLUMN service_statuses.response_time IS '响应时间（毫秒）';
COMMENT ON COLUMN service_statuses.status_code IS 'HTTP状态码';
COMMENT ON COLUMN service_statuses.error_message IS '错误信息';
COMMENT ON COLUMN service_statuses.details IS 'JSON格式的详细信息';
COMMENT ON COLUMN service_statuses.checked_at IS '健康检查执行时间';
COMMENT ON COLUMN service_statuses.created_at IS '记录创建时间';
COMMENT ON COLUMN service_statuses.updated_at IS '记录更新时间';

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_service_statuses_updated_at
    BEFORE UPDATE ON service_statuses
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 创建清理旧记录的存储过程
CREATE OR REPLACE FUNCTION cleanup_old_service_statuses(keep_days INTEGER DEFAULT 30)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM service_statuses 
    WHERE checked_at < CURRENT_TIMESTAMP - INTERVAL '1 day' * keep_days;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- 添加存储过程注释
COMMENT ON FUNCTION cleanup_old_service_statuses(INTEGER) IS '清理指定天数之前的服务状态记录';

-- 创建获取最新服务状态的视图
CREATE OR REPLACE VIEW latest_service_statuses AS
SELECT DISTINCT ON (service_name) 
    id,
    service_name,
    status,
    response_time,
    status_code,
    error_message,
    details,
    checked_at,
    created_at,
    updated_at
FROM service_statuses 
ORDER BY service_name, checked_at DESC;

COMMENT ON VIEW latest_service_statuses IS '每个服务的最新健康状态视图';

-- 创建服务健康统计视图
CREATE OR REPLACE VIEW service_health_summary AS
SELECT 
    COUNT(*) as total_services,
    SUM(CASE WHEN status = 'healthy' THEN 1 ELSE 0 END) as healthy_count,
    SUM(CASE WHEN status = 'degraded' THEN 1 ELSE 0 END) as degraded_count,
    SUM(CASE WHEN status = 'unhealthy' THEN 1 ELSE 0 END) as unhealthy_count,
    AVG(response_time) as avg_response_time,
    MAX(checked_at) as last_check_time
FROM latest_service_statuses;

COMMENT ON VIEW service_health_summary IS '服务健康状态统计摘要视图';

-- 插入一些初始的服务状态记录（用于测试）
INSERT INTO service_statuses (service_name, status, response_time, status_code, checked_at) VALUES
('postgresql_database', 'healthy', 50, 200, CURRENT_TIMESTAMP),
('redis_cache', 'healthy', 20, 200, CURRENT_TIMESTAMP),
('openai_service', 'healthy', 300, 200, CURRENT_TIMESTAMP),
('baidu_ocr_service', 'degraded', 600, 200, CURRENT_TIMESTAMP),
('wechat_api', 'healthy', 150, 200, CURRENT_TIMESTAMP),
('supabase_storage', 'healthy', 100, 200, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;