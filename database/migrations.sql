-- Quivra Database Schema

-- 1. rooms テーブル
CREATE TABLE IF NOT EXISTS rooms (
    id VARCHAR(10) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status ENUM('waiting', 'playing', 'finished') DEFAULT 'waiting'
);

-- 2. players テーブル
CREATE TABLE IF NOT EXISTS players (
    id VARCHAR(36) PRIMARY KEY,
    room_id VARCHAR(10) NOT NULL,
    name VARCHAR(50) NOT NULL,
    score INT DEFAULT 0,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE
);

-- 3. questions テーブル
CREATE TABLE IF NOT EXISTS questions (
    id INT AUTO_INCREMENT PRIMARY KEY,
    question TEXT NOT NULL,
    answer VARCHAR(255) NOT NULL,
    category VARCHAR(50) DEFAULT 'general',
    difficulty ENUM('easy', 'medium', 'hard') DEFAULT 'medium',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. game_sessions テーブル
CREATE TABLE IF NOT EXISTS game_sessions (
    id VARCHAR(36) PRIMARY KEY,
    room_id VARCHAR(10) NOT NULL,
    question_id INT,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP NULL,
    status ENUM('waiting', 'question', 'buzzed', 'answered', 'finished') DEFAULT 'waiting',
    buzzed_player_id VARCHAR(36) NULL,
    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
    FOREIGN KEY (question_id) REFERENCES questions(id) ON DELETE SET NULL,
    FOREIGN KEY (buzzed_player_id) REFERENCES players(id) ON DELETE SET NULL
);

-- インデックスの作成（存在しない場合のみ作成）
CREATE INDEX idx_players_room_id ON players(room_id);
CREATE INDEX idx_questions_category ON questions(category);
CREATE INDEX idx_questions_difficulty ON questions(difficulty);
CREATE INDEX idx_game_sessions_room_id ON game_sessions(room_id);
CREATE INDEX idx_game_sessions_status ON game_sessions(status);
