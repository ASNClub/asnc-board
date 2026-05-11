-- RSS-посты больше не хранят summary. В ленте/треде показываем только title + cover.
-- На клик — переход на оригинал; обсуждение в комментах.
UPDATE posts SET content = '' WHERE source_id IS NOT NULL AND content <> '';
