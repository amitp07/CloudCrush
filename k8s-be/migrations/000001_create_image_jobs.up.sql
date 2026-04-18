CREATE TABLE IF NOT EXISTS image_jobs (
    id UUID PRIMARY KEY,
    filename TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    original_path TEXT,
    compressed_path TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);