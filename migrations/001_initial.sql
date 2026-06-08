-- ============================================================
--  DevFolio — Initial Schema + Seed Data
--  Run: psql $DATABASE_URL -f migrations/001_initial.sql
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ── Admins ────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS admins (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ── Projects ──────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS projects (
    id               UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    title            VARCHAR(255) NOT NULL,
    description      TEXT         NOT NULL DEFAULT '',
    long_description TEXT         NOT NULL DEFAULT '',
    tags             JSONB        NOT NULL DEFAULT '[]',
    github_url       VARCHAR(500) NOT NULL DEFAULT '',
    live_url         VARCHAR(500) NOT NULL DEFAULT '',
    image_url        VARCHAR(500) NOT NULL DEFAULT '',
    featured         BOOLEAN      NOT NULL DEFAULT FALSE,
    sort_order       INTEGER      NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_projects_featured ON projects(featured);
CREATE INDEX IF NOT EXISTS idx_projects_sort     ON projects(sort_order);

-- ── Skills ────────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS skills (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(100) NOT NULL,
    category    VARCHAR(100) NOT NULL,
    proficiency INTEGER      NOT NULL DEFAULT 80 CHECK (proficiency BETWEEN 1 AND 100),
    icon_url    VARCHAR(500) NOT NULL DEFAULT '',
    sort_order  INTEGER      NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_skills_category ON skills(category);

-- ── Experience ────────────────────────────────────────────
CREATE TABLE IF NOT EXISTS experiences (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    company     VARCHAR(255) NOT NULL,
    role        VARCHAR(255) NOT NULL,
    description TEXT         NOT NULL DEFAULT '',
    start_date  VARCHAR(50)  NOT NULL DEFAULT '',
    end_date    VARCHAR(50)  NOT NULL DEFAULT '',
    is_current  BOOLEAN      NOT NULL DEFAULT FALSE,
    company_url VARCHAR(500) NOT NULL DEFAULT '',
    sort_order  INTEGER      NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ── Contact Messages ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS messages (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    name       VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL,
    subject    VARCHAR(500) NOT NULL,
    body       TEXT         NOT NULL,
    is_read    BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_messages_is_read ON messages(is_read);

-- ── Seed sample data ──────────────────────────────────────
INSERT INTO projects (title, description, long_description, tags, github_url, live_url, featured, sort_order) VALUES
(
  'LinkLens',
  'A production-grade URL shortener with per-link analytics, Redis caching, and JWT authentication.',
  'Built with Go, Gin, PostgreSQL, and Redis. Features click tracking (device, browser, country), custom short codes, link expiry, and a minimal analytics dashboard.',
  '["Go","Gin","PostgreSQL","Redis","JWT"]',
  'https://github.com/TejdeepKodati/linklens',
  'http://107.22.34.102:8080/',
  TRUE, 1
),
(
  'GoRelay',
  'A distributed webhook delivery engine with goroutine worker pool, exponential backoff retries, and HMAC signing.',
  'Backend-only service solving reliable event delivery. Uses Redis queues + sorted sets for delayed retries, dead-letter queue for permanently failed deliveries, and HMAC-SHA256 for endpoint signature verification.',
  '["Go","Gin","Redis","PostgreSQL","Distributed Systems","Webhooks"]',
  'https://github.com/TejdeepKodati/gorelay',
  '',
  TRUE, 2
)
ON CONFLICT DO NOTHING;

INSERT INTO skills (name, category, proficiency, sort_order) VALUES
('Go',         'Backend',  90, 1),
('Gin',        'Backend',  88, 2),
('REST APIs',  'Backend',  92, 3),
('PostgreSQL', 'Database', 85, 4),
('Redis',      'Database', 83, 5),
('AWS',        'Cloud',    78, 6),
('Docker',     'DevOps',   82, 7),
('Git',        'DevOps',   90, 8)
ON CONFLICT DO NOTHING;

INSERT INTO experiences (company, role, description, start_date, end_date, is_current, sort_order) VALUES
(
  'Independent Open-Source Development',
  'Backend & DevOps Engineer',
  'Building scalable REST APIs with Go and Gin. Designing Redis caching layers and PostgreSQL schemas for high-throughput systems.',
  'Jan 2024', '', TRUE, 1
)
ON CONFLICT DO NOTHING;