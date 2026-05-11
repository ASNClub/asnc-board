-- Блокировка пользователей. Симметричное скрытие контента.
-- Block = blocker не видит постов/комментов blocked (и наоборот, т.к. фильтр двусторонний).
-- Block также убивает существующий взаимный follow и friendship через ON CASCADE на стороне service.
-- по хорошему в будущем полноценный 3lvl блок сделать

CREATE TABLE user_blocks (
    blocker_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (blocker_id, blocked_id),
    CHECK (blocker_id <> blocked_id)
);

CREATE INDEX idx_user_blocks_blocked ON user_blocks(blocked_id);
