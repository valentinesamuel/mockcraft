CREATE TABLE IF NOT EXISTS jobs (
    id VARCHAR(36) PRIMARY KEY,
    status VARCHAR(20) NOT NULL,
    created_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL,
        updated_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL,
        error TEXT,
        output_path TEXT,
        schema_path TEXT,
        progress INTEGER DEFAULT 0,
        total_steps INTEGER DEFAULT 0,
        metadata JSONB
);

CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs (status);

CREATE INDEX IF NOT EXISTS idx_jobs_created_at ON jobs (created_at);