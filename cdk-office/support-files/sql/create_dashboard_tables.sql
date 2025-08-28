-- Dashboard待办事项和日程模块数据库表结构
-- 创建时间：2025-01-27
-- 描述：包含用户待办事项和日程提醒功能的完整表结构

-- 1. 待办事项表
CREATE TABLE IF NOT EXISTS todo_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    team_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    completed BOOLEAN DEFAULT false,
    due_date TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. 日程事件表
CREATE TABLE IF NOT EXISTS calendar_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    team_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    all_day BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引以优化查询性能

-- 待办事项表索引
CREATE INDEX IF NOT EXISTS idx_todo_items_user_id ON todo_items(user_id);
CREATE INDEX IF NOT EXISTS idx_todo_items_team_id ON todo_items(team_id);
CREATE INDEX IF NOT EXISTS idx_todo_items_completed ON todo_items(completed);
CREATE INDEX IF NOT EXISTS idx_todo_items_due_date ON todo_items(due_date);
CREATE INDEX IF NOT EXISTS idx_todo_items_created_at ON todo_items(created_at);
CREATE INDEX IF NOT EXISTS idx_todo_items_user_team ON todo_items(user_id, team_id);

-- 日程事件表索引
CREATE INDEX IF NOT EXISTS idx_calendar_events_user_id ON calendar_events(user_id);
CREATE INDEX IF NOT EXISTS idx_calendar_events_team_id ON calendar_events(team_id);
CREATE INDEX IF NOT EXISTS idx_calendar_events_start_time ON calendar_events(start_time);
CREATE INDEX IF NOT EXISTS idx_calendar_events_end_time ON calendar_events(end_time);
CREATE INDEX IF NOT EXISTS idx_calendar_events_all_day ON calendar_events(all_day);
CREATE INDEX IF NOT EXISTS idx_calendar_events_user_team ON calendar_events(user_id, team_id);
CREATE INDEX IF NOT EXISTS idx_calendar_events_time_range ON calendar_events(start_time, end_time);

-- 创建复合索引优化特定查询
CREATE INDEX IF NOT EXISTS idx_todo_items_user_completed ON todo_items(user_id, team_id, completed);
CREATE INDEX IF NOT EXISTS idx_calendar_events_upcoming ON calendar_events(user_id, team_id, start_time) WHERE start_time >= CURRENT_TIMESTAMP;

-- 创建外键约束（如果需要的话，取决于实际的用户表结构）
-- 注意：这些约束需要根据实际的表结构调整

-- ALTER TABLE todo_items ADD CONSTRAINT fk_todo_items_user_id 
--     FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- ALTER TABLE todo_items ADD CONSTRAINT fk_todo_items_team_id 
--     FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;

-- ALTER TABLE calendar_events ADD CONSTRAINT fk_calendar_events_user_id 
--     FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- ALTER TABLE calendar_events ADD CONSTRAINT fk_calendar_events_team_id 
--     FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;

-- 创建检查约束确保数据一致性
ALTER TABLE calendar_events ADD CONSTRAINT chk_calendar_events_time_order 
    CHECK (end_time > start_time);

-- 创建触发器以自动更新updated_at字段
CREATE OR REPLACE FUNCTION update_todo_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为待办事项表添加更新时间触发器
CREATE TRIGGER update_todo_items_updated_at BEFORE UPDATE ON todo_items
    FOR EACH ROW EXECUTE FUNCTION update_todo_updated_at_column();

-- 插入一些示例数据（用于测试）
-- 注意：这些数据需要根据实际的用户ID和团队ID调整

-- INSERT INTO todo_items (user_id, team_id, title, completed, due_date) VALUES 
-- ('550e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440001', '完成项目报告', false, CURRENT_TIMESTAMP + INTERVAL '3 days'),
-- ('550e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440001', '参加团队会议', false, CURRENT_TIMESTAMP + INTERVAL '1 day'),
-- ('550e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440001', '审核代码', true, NULL);

-- INSERT INTO calendar_events (user_id, team_id, title, description, start_time, end_time, all_day) VALUES 
-- ('550e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440001', '团队站会', '每日站会讨论进度', CURRENT_TIMESTAMP + INTERVAL '1 day', CURRENT_TIMESTAMP + INTERVAL '1 day 30 minutes', false),
-- ('550e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440001', '项目演示', '向客户演示项目成果', CURRENT_TIMESTAMP + INTERVAL '2 days', CURRENT_TIMESTAMP + INTERVAL '2 days 2 hours', false),
-- ('550e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440001', '公司年会', '年度总结和庆祝', CURRENT_TIMESTAMP + INTERVAL '7 days', CURRENT_TIMESTAMP + INTERVAL '7 days 8 hours', true);

-- 创建视图以便快速查询统计信息
CREATE OR REPLACE VIEW todo_stats AS
SELECT 
    user_id,
    team_id,
    COUNT(*) as total_todos,
    COUNT(CASE WHEN completed = true THEN 1 END) as completed_todos,
    COUNT(CASE WHEN completed = false THEN 1 END) as pending_todos,
    COUNT(CASE WHEN completed = false AND due_date < CURRENT_TIMESTAMP THEN 1 END) as overdue_todos
FROM todo_items
GROUP BY user_id, team_id;

CREATE OR REPLACE VIEW today_events AS
SELECT 
    user_id,
    team_id,
    COUNT(*) as today_events_count
FROM calendar_events
WHERE DATE(start_time) = CURRENT_DATE
GROUP BY user_id, team_id;

-- 创建函数以获取即将到来的事件（用于提醒功能）
CREATE OR REPLACE FUNCTION get_upcoming_events(minutes_ahead INTEGER DEFAULT 15)
RETURNS TABLE(
    event_id UUID,
    user_id UUID,
    team_id UUID,
    title VARCHAR(255),
    start_time TIMESTAMP
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        ce.id,
        ce.user_id,
        ce.team_id,
        ce.title,
        ce.start_time
    FROM calendar_events ce
    WHERE ce.start_time BETWEEN CURRENT_TIMESTAMP AND CURRENT_TIMESTAMP + (minutes_ahead || ' minutes')::INTERVAL
    ORDER BY ce.start_time ASC;
END;
$$ LANGUAGE plpgsql;

-- 说明文档
COMMENT ON TABLE todo_items IS '用户待办事项表，存储用户的个人任务';
COMMENT ON TABLE calendar_events IS '日程事件表，存储用户的日程安排';
COMMENT ON COLUMN todo_items.due_date IS '截止日期，可以为空表示无截止日期';
COMMENT ON COLUMN calendar_events.all_day IS '是否为全天事件';
COMMENT ON FUNCTION get_upcoming_events IS '获取即将到来的事件，用于提醒功能';
COMMENT ON VIEW todo_stats IS '待办事项统计视图';
COMMENT ON VIEW today_events IS '今日事件统计视图';