-- Job Hunter feature: Profile, Jobs, and Applications

-- Profiles table - stores user's professional profile from CV
CREATE TABLE IF NOT EXISTS profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    title TEXT,
    bio TEXT,
    skills TEXT,              -- JSON array of skills
    experience INTEGER,       -- Years of experience
    rate REAL,               -- Desired hourly rate
    currency TEXT DEFAULT 'USD',
    location TEXT,
    remote BOOLEAN DEFAULT 1,
    languages TEXT,          -- JSON array
    education TEXT,          -- JSON array
    projects TEXT,           -- JSON array of notable projects
    links TEXT,              -- JSON: github, linkedin, portfolio
    pdf_path TEXT,           -- Original CV file path
    pdf_content TEXT,        -- Extracted text from PDF
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Jobs table - stores scraped job postings
CREATE TABLE IF NOT EXISTS jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source TEXT NOT NULL,     -- hackernews, remoteok, upwork, linkedin, manual
    source_id TEXT,           -- External ID from source
    source_url TEXT,          -- Link to original posting
    title TEXT NOT NULL,
    company TEXT,
    description TEXT,
    skills TEXT,              -- JSON array of required skills
    rate_min REAL,
    rate_max REAL,
    rate_type TEXT,           -- hourly, monthly, yearly
    currency TEXT DEFAULT 'USD',
    remote BOOLEAN DEFAULT 1,
    location TEXT,
    job_type TEXT,            -- contract, fulltime, parttime
    match_score REAL,         -- 0-100 match with profile
    posted_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Applications table - tracks job applications
CREATE TABLE IF NOT EXISTS applications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id INTEGER NOT NULL,
    profile_id INTEGER,
    proposal TEXT,            -- Generated proposal text
    proposal_pdf TEXT,        -- Path to PDF proposal
    cover_letter TEXT,
    status TEXT DEFAULT 'draft',  -- draft, applied, viewed, response, interview, offer, rejected, withdrawn
    notes TEXT,
    applied_at TIMESTAMP,
    response_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (job_id) REFERENCES jobs(id),
    FOREIGN KEY (profile_id) REFERENCES profiles(id)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_jobs_source ON jobs(source);
CREATE INDEX IF NOT EXISTS idx_jobs_match_score ON jobs(match_score);
CREATE INDEX IF NOT EXISTS idx_jobs_posted_at ON jobs(posted_at);
CREATE INDEX IF NOT EXISTS idx_applications_job ON applications(job_id);
CREATE INDEX IF NOT EXISTS idx_applications_status ON applications(status);
