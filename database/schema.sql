-- PostgreSQL Schema for Error Logging System
-- Create database: CREATE DATABASE error_logs;

-- Extension for UUID generation
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Main errors table
CREATE TABLE errors (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    level VARCHAR(20) NOT NULL DEFAULT 'error', -- error, warning, info, debug
    message TEXT NOT NULL,
    stack_trace TEXT,
    context JSONB DEFAULT '{}',
    source VARCHAR(50) NOT NULL, -- frontend, backend, api
    environment VARCHAR(50) DEFAULT 'production', -- production, development, staging
    user_agent TEXT,
    ip_address INET,
    url TEXT,
    fingerprint VARCHAR(64), -- for grouping similar errors
    resolved BOOLEAN DEFAULT FALSE,
    count INTEGER DEFAULT 1, -- how many times this error occurred
    first_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_seen TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- API keys table for authentication
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_hash VARCHAR(64) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    permissions JSONB DEFAULT '["read"]',
    project_id UUID,
    active BOOLEAN DEFAULT TRUE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used TIMESTAMP WITH TIME ZONE
);

-- Projects table (for multi-project support)
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Alert rules table
CREATE TABLE alert_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    condition TEXT NOT NULL,
    threshold INTEGER NOT NULL,
    time_window VARCHAR(20) DEFAULT '5m',
    enabled BOOLEAN DEFAULT TRUE,
    notifications JSONB DEFAULT '[]',
    last_triggered TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Incidents table
CREATE TABLE incidents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(200) NOT NULL,
    severity VARCHAR(20) DEFAULT 'medium', -- low, medium, high, critical
    status VARCHAR(20) DEFAULT 'open', -- open, investigating, resolved, closed
    description TEXT,
    assigned_to UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Team members table
CREATE TABLE team_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    role VARCHAR(20) DEFAULT 'viewer', -- owner, admin, developer, viewer
    status VARCHAR(20) DEFAULT 'active', -- active, invited, suspended
    last_active TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_errors_timestamp ON errors(timestamp DESC);
CREATE INDEX idx_errors_level ON errors(level);
CREATE INDEX idx_errors_source ON errors(source);
CREATE INDEX idx_errors_fingerprint ON errors(fingerprint);
CREATE INDEX idx_errors_resolved ON errors(resolved);
CREATE INDEX idx_errors_environment ON errors(environment);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_errors_updated_at BEFORE UPDATE
    ON errors FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert sample project and API key
INSERT INTO projects (name, slug) VALUES ('Default Project', 'default');

-- Generate a sample API key (in production, this should be generated securely)
INSERT INTO api_keys (key_hash, name, project_id) 
SELECT 
    encode(sha256('test-api-key'::bytea), 'hex'),
    'Development Key',
    id
FROM projects WHERE slug = 'default';
