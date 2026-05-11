CREATE TABLE badge_definitions (
    id      TEXT PRIMARY KEY,
    glyph   TEXT NOT NULL,
    name    TEXT NOT NULL,
    name_ru TEXT NOT NULL,
    description TEXT NOT NULL,
    color   TEXT NOT NULL DEFAULT '#E09832',
    rarity  TEXT NOT NULL DEFAULT 'common' CHECK (rarity IN ('common', 'rare', 'epic', 'legendary')),
    sort_order INT NOT NULL DEFAULT 0
);

CREATE TABLE user_badges (
    user_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    badge_id  TEXT NOT NULL REFERENCES badge_definitions(id) ON DELETE CASCADE,
    earned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, badge_id)
);

CREATE INDEX idx_user_badges_user ON user_badges(user_id);

-- заглушка
INSERT INTO badge_definitions (id, glyph, name, name_ru, description, color, rarity, sort_order) VALUES
('first-honey',     '🍯', 'First Honey',     'Первый мёд',       'Первый пост получил хотя бы один голос',         '#E09832', 'common',    1),
('bee-keeper',      '🐝', 'Bee Keeper',       'Пчеловод',         'Создал своё сообщество',                         '#EBCB8B', 'common',    2),
('pollinator',      '🌻', 'Pollinator',       'Опылитель',        'Подписался на 5 сообществ',                      '#A3BE8C', 'common',    3),
('hive-mind',       '🧠', 'Hive Mind',        'Улей-разум',       'Написал 10 комментариев',                        '#88C0D0', 'common',    4),
('honey-flow',      '💧', 'Honey Flow',       'Медоток',          'Написал 10 постов',                              '#81A1C1', 'common',    5),
('queen-bee',       '👑', 'Queen Bee',        'Королева улья',    'Набрал 100 подписчиков в сообществе',            '#B48EAD', 'rare',      6),
('golden-comb',     '✨', 'Golden Comb',      'Золотые соты',     'Пост набрал 50 голосов',                         '#F5C55A', 'rare',      7),
('nectar-hunter',   '🎯', 'Nectar Hunter',    'Охотник за нектаром', 'Получил 10 звёзд на сообществе',              '#D08770', 'rare',      8),
('royal-jelly',     '💎', 'Royal Jelly',      'Маточное молочко', 'Репутация достигла 500',                         '#5E81AC', 'epic',      9),
('swarm-leader',    '🦅', 'Swarm Leader',     'Вожак роя',        'Сообщество набрало 1000 подписчиков',            '#BF616A', 'epic',     10),
('honey-badger',    '🦡', 'Honey Badger',     'Медоед',           'Репутация достигла 1000',                        '#A3BE8C', 'legendary', 11),
('mythical-bloom',  '🌸', 'Mythical Bloom',   'Мифический цвет',  'Пост набрал 500 голосов',                        '#B48EAD', 'legendary', 12);
