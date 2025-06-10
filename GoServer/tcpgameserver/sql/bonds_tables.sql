-- 羁绊表
CREATE TABLE IF NOT EXISTS bonds (
    id INT AUTO_INCREMENT PRIMARY KEY COMMENT '羁绊ID',
    name VARCHAR(50) NOT NULL COMMENT '羁绊名称',
    level INT NOT NULL COMMENT '羁绊等级',
    damage DECIMAL(5,1) NOT NULL COMMENT '羁绊伤害',
    skill VARCHAR(100) DEFAULT '' COMMENT '羁绊技能',
    description VARCHAR(200) NOT NULL COMMENT '羁绊描述',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='羁绊表';

-- 羁绊关联表
CREATE TABLE IF NOT EXISTS bond_relations (
    id INT AUTO_INCREMENT PRIMARY KEY COMMENT '关联ID',
    bond_id INT NOT NULL COMMENT '羁绊ID',
    card_name VARCHAR(50) NOT NULL COMMENT '卡牌名称',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    FOREIGN KEY (bond_id) REFERENCES bonds(id) ON DELETE CASCADE,
    UNIQUE KEY uk_bond_card (bond_id, card_name),
    KEY idx_bond_id (bond_id),
    KEY idx_card_name (card_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='羁绊关联表';

-- 插入示例羁绊数据
INSERT INTO bonds (name, level, damage, skill, description) VALUES
('火焰之力', 1, 15.0, '燃烧', '火属性卡牌组合，提供额外火焰伤害'),
('冰霜守护', 1, 12.0, '冰冻', '冰属性卡牌组合，具有冰冻效果'),
('雷电风暴', 2, 25.0, '麻痹', '雷属性高级组合，造成大量伤害并麻痹敌人'),
('自然调和', 2, 18.0, '治愈', '自然属性组合，具有治愈和增益效果'),
('暗影之握', 3, 35.0, '诅咒', '暗属性终极组合，具有强大的诅咒效果');

-- 插入示例羁绊关联数据（假设CardDeck表中已有这些卡牌）
INSERT INTO bond_relations (bond_id, card_name) VALUES
-- 火焰之力羁绊
(1, '火球术'),
(1, '烈火剑'),
(1, '火焰护甲'),

-- 冰霜守护羁绊
(2, '冰锥术'),
(2, '霜冻护盾'),
(2, '寒冰箭'),

-- 雷电风暴羁绊
(3, '闪电链'),
(3, '雷击术'),
(3, '电磁护壁'),

-- 自然调和羁绊
(4, '治愈术'),
(4, '自然之怒'),
(4, '生命回复'),

-- 暗影之握羁绊
(5, '暗影球'),
(5, '死亡之触'),
(5, '黑暗护盾');
