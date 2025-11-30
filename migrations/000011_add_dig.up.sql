-- Dig feature: Idea Analysis and Incubation
-- Multi-step idea validation with AI-powered analysis from different perspectives

-- Dig Sessions table - stores user's ideas and analysis sessions
CREATE TABLE IF NOT EXISTS dig_sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT,                           -- Auto-generated or user-provided title
    raw_idea TEXT NOT NULL,               -- Original idea as typed by user
    refined_idea TEXT,                    -- AI-refined/improved version of the idea
    status TEXT DEFAULT 'pending',        -- pending, analyzing, completed, failed
    overall_score REAL,                   -- 0-100 overall viability score
    recommendation TEXT,                  -- proceed, pivot, abandon, refine

    -- Analysis stage tracking
    current_stage TEXT DEFAULT 'first_principles',  -- Current analysis stage
    stages_completed TEXT,                -- JSON array of completed stages

    -- Timestamps
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Dig Analysis Results table - stores individual analysis from each perspective
CREATE TABLE IF NOT EXISTS dig_analyses (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL,
    perspective TEXT NOT NULL,            -- first_principles, designer, marketing, technical, financial

    -- Analysis content
    summary TEXT,                         -- Brief summary of this perspective's analysis
    strengths TEXT,                       -- JSON array of identified strengths
    weaknesses TEXT,                      -- JSON array of identified weaknesses
    opportunities TEXT,                   -- JSON array of opportunities
    threats TEXT,                         -- JSON array of threats/risks
    recommendations TEXT,                 -- JSON array of recommendations
    score REAL,                           -- 0-100 score from this perspective

    -- Detailed sections (JSON)
    detailed_analysis TEXT,               -- Full detailed analysis as JSON

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES dig_sessions(id) ON DELETE CASCADE
);

-- Dig Execution Plan table - stores implementation roadmap
CREATE TABLE IF NOT EXISTS dig_execution_plans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL,

    -- High-level plan
    summary TEXT,                         -- Executive summary
    mvp_scope TEXT,                       -- What to build first (MVP)
    full_scope TEXT,                      -- Full product vision

    -- Technical decisions
    architecture TEXT,                    -- JSON: recommended architecture decisions
    tech_stack TEXT,                      -- JSON: recommended technologies
    integrations TEXT,                    -- JSON: third-party services needed

    -- Phases
    phases TEXT,                          -- JSON array of implementation phases
    milestones TEXT,                      -- JSON array of key milestones

    -- Resources
    team_requirements TEXT,               -- JSON: what team/skills needed
    estimated_cost TEXT,                  -- JSON: cost breakdown

    -- LLM Ready
    llm_prompt TEXT,                      -- Ready-to-use prompt for LLMs to help build

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES dig_sessions(id) ON DELETE CASCADE
);

-- Dig Marketing Materials table - stores generated marketing content
CREATE TABLE IF NOT EXISTS dig_marketing (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL,

    -- Positioning
    value_proposition TEXT,               -- Core value proposition
    target_audience TEXT,                 -- JSON: target audience segments
    positioning_statement TEXT,           -- Market positioning

    -- Copy
    taglines TEXT,                        -- JSON array of tagline options
    elevator_pitch TEXT,                  -- 30-second pitch
    headlines TEXT,                       -- JSON array of headline options
    descriptions TEXT,                    -- JSON array of description variations

    -- Visual
    color_suggestions TEXT,               -- JSON: suggested brand colors
    imagery_prompts TEXT,                 -- JSON array of image generation prompts
    generated_images TEXT,                -- JSON array of generated image paths/URLs

    -- Channels
    channel_strategy TEXT,                -- JSON: recommended marketing channels
    launch_strategy TEXT,                 -- Launch approach

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES dig_sessions(id) ON DELETE CASCADE
);

-- Dig Revenue Projections table - stores financial projections
CREATE TABLE IF NOT EXISTS dig_revenue_projections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL,

    -- Market
    market_size TEXT,                     -- JSON: TAM, SAM, SOM
    market_growth TEXT,                   -- Market growth rate
    competitors TEXT,                     -- JSON array of competitor analysis

    -- Pricing
    pricing_models TEXT,                  -- JSON array of pricing options
    recommended_price TEXT,               -- Recommended pricing
    pricing_rationale TEXT,               -- Why this pricing

    -- Projections
    year1_revenue TEXT,                   -- JSON: monthly projections year 1
    year2_revenue TEXT,                   -- JSON: monthly projections year 2
    year3_revenue TEXT,                   -- JSON: yearly projection year 3

    -- Metrics
    key_metrics TEXT,                     -- JSON: KPIs to track
    break_even_analysis TEXT,             -- When to expect break-even

    -- Assumptions
    assumptions TEXT,                     -- JSON: assumptions made
    risks TEXT,                           -- JSON: financial risks

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES dig_sessions(id) ON DELETE CASCADE
);

-- Dig Alternative Ideas table - stores pivot/refinement suggestions
CREATE TABLE IF NOT EXISTS dig_alternatives (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id INTEGER NOT NULL,

    alternative_idea TEXT NOT NULL,       -- The suggested alternative/pivot
    rationale TEXT,                       -- Why this alternative
    comparison TEXT,                      -- How it compares to original
    viability_score REAL,                 -- 0-100 score
    effort_level TEXT,                    -- low, medium, high
    potential TEXT,                       -- low, medium, high, very_high

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES dig_sessions(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_dig_sessions_status ON dig_sessions(status);
CREATE INDEX IF NOT EXISTS idx_dig_sessions_created ON dig_sessions(created_at);
CREATE INDEX IF NOT EXISTS idx_dig_analyses_session ON dig_analyses(session_id);
CREATE INDEX IF NOT EXISTS idx_dig_analyses_perspective ON dig_analyses(perspective);
CREATE INDEX IF NOT EXISTS idx_dig_execution_session ON dig_execution_plans(session_id);
CREATE INDEX IF NOT EXISTS idx_dig_marketing_session ON dig_marketing(session_id);
CREATE INDEX IF NOT EXISTS idx_dig_revenue_session ON dig_revenue_projections(session_id);
CREATE INDEX IF NOT EXISTS idx_dig_alternatives_session ON dig_alternatives(session_id);
