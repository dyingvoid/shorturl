CREATE TABLE IF NOT EXISTS click_events (
    short_code  String,
    user_id     String,
    clicked_at  DateTime,
    ip          String,
    user_agent  String,
    referer     String
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(clicked_at)
ORDER BY (short_code, clicked_at);