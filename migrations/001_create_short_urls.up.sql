CREATE TABLE IF NOT EXISTS short_urls (
    id BIGSERIAL PRIMARY KEY,
    long_url TEXT NOT NULL,
    short_url VARCHAR(20) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    CONSTRAINT chk_long_url_length CHECK (length(long_url) <= 2048),
    CONSTRAINT chk_short_url_length CHECK (length(short_url) >= 4)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_short_urls_short_url ON short_urls(short_url);
CREATE INDEX IF NOT EXISTS idx_short_urls_created_at ON short_urls(created_at);
