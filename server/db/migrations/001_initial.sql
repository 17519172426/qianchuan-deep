CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'viewer',
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS qianchuan_accounts (
    id SERIAL PRIMARY KEY,
    account_name VARCHAR(255) NOT NULL,
    advertiser_id BIGINT UNIQUE NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    balance NUMERIC(15,2) DEFAULT 0,
    last_sync_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS uni_ads (
    id SERIAL PRIMARY KEY,
    account_id INTEGER REFERENCES qianchuan_accounts(id),
    qianchuan_ad_id BIGINT,
    name VARCHAR(100) NOT NULL,
    marketing_goal VARCHAR(32) NOT NULL,
    aweme_id BIGINT,
    product_ids JSONB DEFAULT '[]',
    delivery_setting JSONB NOT NULL DEFAULT '{}',
    creative_setting JSONB NOT NULL DEFAULT '{}',
    status VARCHAR(32) NOT NULL DEFAULT 'create',
    metrics_json JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS creatives (
    id SERIAL PRIMARY KEY,
    account_id INTEGER REFERENCES qianchuan_accounts(id),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,
    url TEXT NOT NULL,
    file_size BIGINT DEFAULT 0,
    duration FLOAT DEFAULT 0,
    tags JSONB DEFAULT '[]',
    metrics_json JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS uni_ad_creatives (
    uni_ad_id INTEGER REFERENCES uni_ads(id),
    creative_id INTEGER REFERENCES creatives(id),
    is_blocked BOOLEAN DEFAULT false,
    PRIMARY KEY (uni_ad_id, creative_id)
);

CREATE TABLE IF NOT EXISTS rules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT DEFAULT '',
    account_id INTEGER REFERENCES qianchuan_accounts(id),
    scope_json JSONB DEFAULT '{}',
    condition_json JSONB NOT NULL DEFAULT '{}',
    action_json JSONB NOT NULL DEFAULT '{}',
    schedule VARCHAR(50) DEFAULT '*/5 * * * *',
    cooldown VARCHAR(20) DEFAULT '1h',
    enabled BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS rule_executions (
    id SERIAL PRIMARY KEY,
    rule_id INTEGER REFERENCES rules(id),
    uni_ad_id INTEGER REFERENCES uni_ads(id),
    triggered_at TIMESTAMP WITH TIME ZONE,
    condition_json JSONB DEFAULT '{}',
    action_json JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'pending',
    result_json JSONB DEFAULT '{}',
    executed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS ai_recommendations (
    id SERIAL PRIMARY KEY,
    uni_ad_id INTEGER REFERENCES uni_ads(id),
    type VARCHAR(30) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    metrics_json JSONB DEFAULT '{}',
    confidence FLOAT DEFAULT 0,
    suggested_action JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'pending',
    reviewed_by INTEGER REFERENCES users(id),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_uni_ads_account ON uni_ads(account_id);
CREATE INDEX IF NOT EXISTS idx_uni_ads_status ON uni_ads(status);
CREATE INDEX IF NOT EXISTS idx_rules_account ON rules(account_id);
CREATE INDEX IF NOT EXISTS idx_ai_recs_status ON ai_recommendations(status);
