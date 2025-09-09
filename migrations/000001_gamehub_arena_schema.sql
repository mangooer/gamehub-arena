-- GameHub Arena 数据库初始化脚本
-- 创建时间: 2025-09-08
-- 描述: 分布式实时游戏匹配与对战系统数据库表结构

-- ============================================
-- 1. 用户相关表
-- ============================================

-- 用户基础信息表
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    level INTEGER DEFAULT 1 CHECK (level >= 1 AND level <= 100),
    experience BIGINT DEFAULT 0 CHECK (experience >= 0),
    win_count INTEGER DEFAULT 0 CHECK (win_count >= 0),
    lose_count INTEGER DEFAULT 0 CHECK (lose_count >= 0),
    win_rate DECIMAL(5,4) DEFAULT 0.0000 CHECK (win_rate >= 0 AND win_rate <= 1),
    rank VARCHAR(20) DEFAULT 'Bronze' CHECK (rank IN ('Bronze', 'Silver', 'Gold', 'Platinum', 'Diamond', 'Master', 'Grandmaster')),
    avatar_url VARCHAR(255),
    status VARCHAR(20) DEFAULT 'offline' CHECK (status IN ('online', 'offline', 'in_game', 'away')),
    last_login_at TIMESTAMP,
    last_active_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- 用户统计详细信息表
CREATE TABLE user_stats (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    total_games INTEGER DEFAULT 0 CHECK (total_games >= 0),
    total_playtime BIGINT DEFAULT 0 CHECK (total_playtime >= 0), -- 总游戏时长（秒）
    best_rank VARCHAR(20) DEFAULT 'Bronze',
    current_streak INTEGER DEFAULT 0, -- 当前连胜/连败（正数连胜，负数连败）
    max_win_streak INTEGER DEFAULT 0 CHECK (max_win_streak >= 0),
    max_lose_streak INTEGER DEFAULT 0 CHECK (max_lose_streak >= 0),
    avg_game_duration INTEGER DEFAULT 0, -- 平均游戏时长（秒）
    total_kills INTEGER DEFAULT 0 CHECK (total_kills >= 0),
    total_deaths INTEGER DEFAULT 0 CHECK (total_deaths >= 0),
    total_assists INTEGER DEFAULT 0 CHECK (total_assists >= 0),
    kda_ratio DECIMAL(5,2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(user_id)
);

-- 用户好友关系表
CREATE TABLE user_friends (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'blocked')),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(user_id, friend_id),
    CHECK (user_id != friend_id)
);

-- ============================================
-- 2. 游戏相关表
-- ============================================

-- 游戏房间表
CREATE TABLE game_rooms (
    id BIGSERIAL PRIMARY KEY,
    room_code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    max_players INTEGER NOT NULL CHECK (max_players BETWEEN 2 AND 10),
    current_players INTEGER DEFAULT 0 CHECK (current_players >= 0),
    status VARCHAR(20) DEFAULT 'waiting' CHECK (status IN ('waiting', 'starting', 'in_progress', 'finished', 'cancelled')),
    game_mode VARCHAR(50) NOT NULL CHECK (game_mode IN ('classic', 'ranked', 'casual', 'tournament')),
    map_name VARCHAR(50) DEFAULT 'default_map',
    is_private BOOLEAN DEFAULT FALSE,
    password_hash VARCHAR(255), -- 私人房间密码
    created_by BIGINT NOT NULL REFERENCES users(id),
    started_at TIMESTAMP,
    ended_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP
);

-- 房间玩家关系表
CREATE TABLE room_players (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT NOT NULL REFERENCES game_rooms(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team VARCHAR(10) NOT NULL CHECK (team IN ('team_a', 'team_b', 'spectator')),
    position INTEGER CHECK (position BETWEEN 1 AND 5), -- 队伍内位置
    is_ready BOOLEAN DEFAULT FALSE,
    is_captain BOOLEAN DEFAULT FALSE, -- 是否为队长
    joined_at TIMESTAMP DEFAULT NOW(),
    left_at TIMESTAMP,
    
    UNIQUE(room_id, user_id),
    UNIQUE(room_id, team, position) -- 确保同队同位置唯一
);

-- 游戏记录表
CREATE TABLE game_records (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT NOT NULL REFERENCES game_rooms(id),
    winner_team VARCHAR(10) CHECK (winner_team IN ('team_a', 'team_b', 'draw')),
    duration INTEGER CHECK (duration > 0), -- 游戏时长（秒）
    status VARCHAR(20) DEFAULT 'completed' CHECK (status IN ('completed', 'abandoned', 'cancelled')),
    game_data JSONB, -- 游戏详细数据（JSON格式）
    replay_url VARCHAR(255), -- 回放文件URL
    started_at TIMESTAMP NOT NULL,
    ended_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    
    CHECK (ended_at > started_at)
);

-- 玩家游戏表现记录表
CREATE TABLE player_game_stats (
    id BIGSERIAL PRIMARY KEY,
    game_record_id BIGINT NOT NULL REFERENCES game_records(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team VARCHAR(10) NOT NULL,
    kills INTEGER DEFAULT 0 CHECK (kills >= 0),
    deaths INTEGER DEFAULT 0 CHECK (deaths >= 0),
    assists INTEGER DEFAULT 0 CHECK (assists >= 0),
    damage_dealt BIGINT DEFAULT 0 CHECK (damage_dealt >= 0),
    damage_taken BIGINT DEFAULT 0 CHECK (damage_taken >= 0),
    gold_earned INTEGER DEFAULT 0 CHECK (gold_earned >= 0),
    experience_gained INTEGER DEFAULT 0 CHECK (experience_gained >= 0),
    is_winner BOOLEAN DEFAULT FALSE,
    mvp_score DECIMAL(5,2) DEFAULT 0.00, -- MVP评分
    performance_rating DECIMAL(5,2) DEFAULT 0.00, -- 表现评分
    created_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(game_record_id, user_id)
);

-- ============================================
-- 3. 匹配相关表
-- ============================================

-- 匹配队列表
CREATE TABLE match_queue (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    game_mode VARCHAR(50) NOT NULL,
    rank VARCHAR(20) NOT NULL,
    estimated_wait_time INTEGER, -- 预计等待时间（秒）
    actual_wait_time INTEGER, -- 实际等待时间（秒）
    priority_score DECIMAL(5,2) DEFAULT 0.00, -- 优先级评分
    status VARCHAR(20) DEFAULT 'waiting' CHECK (status IN ('waiting', 'matched', 'cancelled', 'timeout')),
    matched_at TIMESTAMP,
    cancelled_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(user_id, game_mode) -- 用户在同一游戏模式下只能有一个队列记录
);

-- 匹配记录表
CREATE TABLE match_records (
    id BIGSERIAL PRIMARY KEY,
    match_id VARCHAR(50) UNIQUE NOT NULL, -- 匹配ID
    game_mode VARCHAR(50) NOT NULL,
    player_count INTEGER NOT NULL CHECK (player_count > 0),
    avg_rank VARCHAR(20),
    avg_wait_time INTEGER, -- 平均等待时间（秒）
    match_quality DECIMAL(3,2) CHECK (match_quality BETWEEN 0 AND 1), -- 匹配质量评分 0-1
    algorithm_version VARCHAR(20), -- 匹配算法版本
    room_id BIGINT REFERENCES game_rooms(id),
    status VARCHAR(20) DEFAULT 'completed' CHECK (status IN ('completed', 'failed', 'cancelled')),
    created_at TIMESTAMP DEFAULT NOW()
);

-- 匹配玩家记录表
CREATE TABLE match_players (
    id BIGSERIAL PRIMARY KEY,
    match_record_id BIGINT NOT NULL REFERENCES match_records(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    queue_time INTEGER NOT NULL, -- 排队时间（秒）
    rank VARCHAR(20) NOT NULL,
    win_rate DECIMAL(5,4),
    team_assignment VARCHAR(10), -- 分配的队伍
    created_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(match_record_id, user_id)
);

-- ============================================
-- 4. 排行榜相关表
-- ============================================

-- 排行榜记录表（支持多种排行榜类型）
CREATE TABLE leaderboards (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    leaderboard_type VARCHAR(50) NOT NULL CHECK (leaderboard_type IN ('global', 'seasonal', 'weekly', 'monthly', 'mode_specific')),
    game_mode VARCHAR(50), -- 特定模式排行榜
    rank_position INTEGER NOT NULL CHECK (rank_position > 0),
    score BIGINT NOT NULL DEFAULT 0,
    tier VARCHAR(20) NOT NULL,
    points INTEGER DEFAULT 0, -- 排位积分
    season VARCHAR(20), -- 赛季标识
    region VARCHAR(10) DEFAULT 'global', -- 地区
    last_updated TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    
    UNIQUE(leaderboard_type, game_mode, season, region, user_id)
);

-- 排行榜历史记录表
CREATE TABLE leaderboard_history (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    leaderboard_type VARCHAR(50) NOT NULL,
    old_rank INTEGER,
    new_rank INTEGER,
    old_score BIGINT,
    new_score BIGINT,
    change_reason VARCHAR(100), -- 变化原因
    game_record_id BIGINT REFERENCES game_records(id),
    created_at TIMESTAMP DEFAULT NOW()
);

-- ============================================
-- 5. 系统相关表
-- ============================================

-- 系统配置表
CREATE TABLE system_configs (
    id BIGSERIAL PRIMARY KEY,
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT NOT NULL,
    config_type VARCHAR(20) DEFAULT 'string' CHECK (config_type IN ('string', 'integer', 'float', 'boolean', 'json')),
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 操作日志表
CREATE TABLE operation_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    operation_type VARCHAR(50) NOT NULL,
    operation_desc TEXT,
    ip_address INET,
    user_agent TEXT,
    request_data JSONB,
    response_data JSONB,
    status VARCHAR(20) DEFAULT 'success' CHECK (status IN ('success', 'failed', 'error')),
    execution_time INTEGER, -- 执行时间（毫秒）
    created_at TIMESTAMP DEFAULT NOW()
);

-- 游戏事件日志表（用于分析和调试）
CREATE TABLE game_events (
    id BIGSERIAL PRIMARY KEY,
    room_id BIGINT REFERENCES game_rooms(id),
    user_id BIGINT REFERENCES users(id),
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB NOT NULL,
    timestamp_ms BIGINT NOT NULL, -- 游戏内时间戳（毫秒）
    created_at TIMESTAMP DEFAULT NOW()
);

-- ============================================
-- 6. 创建索引
-- ============================================

-- 用户表索引
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_rank ON users(rank);
CREATE INDEX idx_users_win_rate ON users(win_rate DESC);
CREATE INDEX idx_users_level ON users(level DESC);
CREATE INDEX idx_users_deleted_at ON users(deleted_at);
CREATE INDEX idx_users_last_active ON users(last_active_at DESC);

-- 用户统计表索引
CREATE INDEX idx_user_stats_user_id ON user_stats(user_id);
CREATE INDEX idx_user_stats_total_games ON user_stats(total_games DESC);
CREATE INDEX idx_user_stats_kda ON user_stats(kda_ratio DESC);

-- 用户好友表索引
CREATE INDEX idx_user_friends_user_id ON user_friends(user_id);
CREATE INDEX idx_user_friends_friend_id ON user_friends(friend_id);
CREATE INDEX idx_user_friends_status ON user_friends(status);

-- 游戏房间表索引
CREATE INDEX idx_game_rooms_room_code ON game_rooms(room_code);
CREATE INDEX idx_game_rooms_status ON game_rooms(status);
CREATE INDEX idx_game_rooms_game_mode ON game_rooms(game_mode);
CREATE INDEX idx_game_rooms_created_by ON game_rooms(created_by);
CREATE INDEX idx_game_rooms_created_at ON game_rooms(created_at DESC);
CREATE INDEX idx_game_rooms_deleted_at ON game_rooms(deleted_at);

-- 房间玩家表索引
CREATE INDEX idx_room_players_room_id ON room_players(room_id);
CREATE INDEX idx_room_players_user_id ON room_players(user_id);
CREATE INDEX idx_room_players_team ON room_players(team);

-- 游戏记录表索引
CREATE INDEX idx_game_records_room_id ON game_records(room_id);
CREATE INDEX idx_game_records_started_at ON game_records(started_at DESC);
CREATE INDEX idx_game_records_duration ON game_records(duration);
CREATE INDEX idx_game_records_status ON game_records(status);

-- 玩家游戏统计表索引
CREATE INDEX idx_player_game_stats_game_record_id ON player_game_stats(game_record_id);
CREATE INDEX idx_player_game_stats_user_id ON player_game_stats(user_id);
CREATE INDEX idx_player_game_stats_mvp_score ON player_game_stats(mvp_score DESC);

-- 匹配队列表索引
CREATE INDEX idx_match_queue_user_id ON match_queue(user_id);
CREATE INDEX idx_match_queue_game_mode ON match_queue(game_mode);
CREATE INDEX idx_match_queue_status ON match_queue(status);
CREATE INDEX idx_match_queue_created_at ON match_queue(created_at);
CREATE INDEX idx_match_queue_rank_mode ON match_queue(rank, game_mode);

-- 匹配记录表索引
CREATE INDEX idx_match_records_match_id ON match_records(match_id);
CREATE INDEX idx_match_records_game_mode ON match_records(game_mode);
CREATE INDEX idx_match_records_created_at ON match_records(created_at DESC);

-- 匹配玩家表索引
CREATE INDEX idx_match_players_match_record_id ON match_players(match_record_id);
CREATE INDEX idx_match_players_user_id ON match_players(user_id);

-- 排行榜表索引
CREATE INDEX idx_leaderboards_type_mode_season ON leaderboards(leaderboard_type, game_mode, season);
CREATE INDEX idx_leaderboards_user_id ON leaderboards(user_id);
CREATE INDEX idx_leaderboards_rank_position ON leaderboards(rank_position);
CREATE INDEX idx_leaderboards_score ON leaderboards(score DESC);
CREATE INDEX idx_leaderboards_points ON leaderboards(points DESC);

-- 排行榜历史表索引
CREATE INDEX idx_leaderboard_history_user_id ON leaderboard_history(user_id);
CREATE INDEX idx_leaderboard_history_created_at ON leaderboard_history(created_at DESC);

-- 系统配置表索引
CREATE INDEX idx_system_configs_key ON system_configs(config_key);
CREATE INDEX idx_system_configs_active ON system_configs(is_active);

-- 操作日志表索引
CREATE INDEX idx_operation_logs_user_id ON operation_logs(user_id);
CREATE INDEX idx_operation_logs_operation_type ON operation_logs(operation_type);
CREATE INDEX idx_operation_logs_created_at ON operation_logs(created_at DESC);
CREATE INDEX idx_operation_logs_status ON operation_logs(status);

-- 游戏事件表索引
CREATE INDEX idx_game_events_room_id ON game_events(room_id);
CREATE INDEX idx_game_events_user_id ON game_events(user_id);
CREATE INDEX idx_game_events_event_type ON game_events(event_type);
CREATE INDEX idx_game_events_created_at ON game_events(created_at DESC);
CREATE INDEX idx_game_events_timestamp ON game_events(timestamp_ms);

-- ============================================
-- 7. 创建触发器和函数
-- ============================================

-- 更新 updated_at 字段的触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- 为需要的表创建触发器
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_stats_updated_at BEFORE UPDATE ON user_stats
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_friends_updated_at BEFORE UPDATE ON user_friends
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_game_rooms_updated_at BEFORE UPDATE ON game_rooms
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_match_queue_updated_at BEFORE UPDATE ON match_queue
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_system_configs_updated_at BEFORE UPDATE ON system_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- 自动更新用户胜率的触发器函数
CREATE OR REPLACE FUNCTION update_user_win_rate()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.win_count + NEW.lose_count > 0 THEN
        NEW.win_rate = NEW.win_count::DECIMAL / (NEW.win_count + NEW.lose_count);
    ELSE
        NEW.win_rate = 0;
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_win_rate BEFORE UPDATE OF win_count, lose_count ON users
    FOR EACH ROW EXECUTE FUNCTION update_user_win_rate();

-- ============================================
-- 8. 插入初始数据
-- ============================================

-- 插入系统配置
INSERT INTO system_configs (config_key, config_value, config_type, description) VALUES
('match_timeout_seconds', '300', 'integer', '匹配超时时间（秒）'),
('max_queue_size', '10000', 'integer', '最大队列大小'),
('game_duration_limit', '3600', 'integer', '游戏时长限制（秒）'),
('rank_points_win', '25', 'integer', '胜利获得积分'),
('rank_points_lose', '20', 'integer', '失败扣除积分'),
('leaderboard_update_interval', '300', 'integer', '排行榜更新间隔（秒）'),
('maintenance_mode', 'false', 'boolean', '维护模式开关'),
('max_friends_count', '100', 'integer', '最大好友数量'),
('room_idle_timeout', '1800', 'integer', '房间空闲超时时间（秒）'),
('anti_cheat_enabled', 'true', 'boolean', '反作弊系统开关');

-- 创建管理员用户（密码需要在应用层加密）
INSERT INTO users (username, email, password_hash, level, rank, status) VALUES
('admin', 'admin@gamehub-arena.com', '$2a$10$placeholder_hash', 100, 'Grandmaster', 'online'),
('system', 'system@gamehub-arena.com', '$2a$10$placeholder_hash', 1, 'Bronze', 'offline');

COMMENT ON TABLE users IS '用户基础信息表';
COMMENT ON TABLE user_stats IS '用户统计详细信息表';
COMMENT ON TABLE user_friends IS '用户好友关系表';
COMMENT ON TABLE game_rooms IS '游戏房间表';
COMMENT ON TABLE room_players IS '房间玩家关系表';
COMMENT ON TABLE game_records IS '游戏记录表';
COMMENT ON TABLE player_game_stats IS '玩家游戏表现记录表';
COMMENT ON TABLE match_queue IS '匹配队列表';
COMMENT ON TABLE match_records IS '匹配记录表';
COMMENT ON TABLE match_players IS '匹配玩家记录表';
COMMENT ON TABLE leaderboards IS '排行榜记录表';
COMMENT ON TABLE leaderboard_history IS '排行榜历史记录表';
COMMENT ON TABLE system_configs IS '系统配置表';
COMMENT ON TABLE operation_logs IS '操作日志表';
COMMENT ON TABLE game_events IS '游戏事件日志表';

-- 数据库初始化完成
SELECT 'GameHub Arena database schema initialized successfully!' AS status;
