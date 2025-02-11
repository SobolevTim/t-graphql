-- Создание таблицы для постов
CREATE TABLE posts (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    author TEXT NOT NULL,
    allow_comments BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Создание таблицы для комментариев
CREATE TABLE comments (
    id UUID PRIMARY KEY,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    author TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Создание функции для отправки уведомлений при добавлении комментария
CREATE OR REPLACE FUNCTION notify_comment() RETURNS TRIGGER AS $$
DECLARE
    payload JSON;
BEGIN
    payload = json_build_object(
        'id', NEW.id,
        'post_id', NEW.post_id,
        'parent_id', NEW.parent_id,
        'content', NEW.content,
        'author', NEW.author,
        'created_at', NEW.created_at
    );

    PERFORM pg_notify('comments_' || NEW.post_id::text, payload::text);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Создание триггера для вызова функции notify_comment при вставке комментария
CREATE TRIGGER comment_inserted
AFTER INSERT ON comments
FOR EACH ROW
EXECUTE FUNCTION notify_comment();