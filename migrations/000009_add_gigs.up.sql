-- Gigs table - the central work unit connecting hunter → work → invoice
-- Represents active projects/contracts with workflow status

CREATE TABLE IF NOT EXISTS gigs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,                    -- e.g., "Website Redesign" or job title
    client_id INTEGER,                     -- Link to client (if converted from hunter)
    contract_id INTEGER,                   -- Link to contract (optional)
    application_id INTEGER,                -- Link to hunter application (if came from job hunt)

    -- Status workflow: pipeline → negotiating → active → delivered → invoiced → complete
    status TEXT NOT NULL DEFAULT 'pipeline',

    -- Type of work
    gig_type TEXT DEFAULT 'hourly',        -- hourly, fixed, retainer

    -- Financial
    estimated_hours REAL,                  -- For hourly gigs
    estimated_amount REAL,                 -- For fixed price
    hourly_rate REAL,
    currency TEXT DEFAULT 'USD',

    -- Time tracking aggregation
    total_hours_tracked REAL DEFAULT 0,
    last_tracked_at TIMESTAMP,

    -- Billing
    total_invoiced REAL DEFAULT 0,
    last_invoiced_at TIMESTAMP,

    -- Dates
    start_date TIMESTAMP,
    due_date TIMESTAMP,
    completed_at TIMESTAMP,

    -- Metadata
    description TEXT,
    notes TEXT,
    priority INTEGER DEFAULT 0,            -- 0=normal, 1=high, 2=urgent

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (client_id) REFERENCES clients(id),
    FOREIGN KEY (contract_id) REFERENCES contracts(id),
    FOREIGN KEY (application_id) REFERENCES applications(id)
);

-- Work log/notes for gigs - quick notes attached to work
CREATE TABLE IF NOT EXISTS work_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    gig_id INTEGER,
    client_id INTEGER,                     -- Can be attached to client even without gig
    tracking_session_id INTEGER,           -- Link to time entry (optional)

    content TEXT NOT NULL,                 -- The log/note content
    log_type TEXT DEFAULT 'note',          -- note, decision, meeting, blocker, milestone

    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (gig_id) REFERENCES gigs(id),
    FOREIGN KEY (client_id) REFERENCES clients(id),
    FOREIGN KEY (tracking_session_id) REFERENCES tracking_sessions(id)
);

-- Enhanced goals table with more goal types
CREATE TABLE IF NOT EXISTS income_goals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    amount REAL NOT NULL,
    period TEXT NOT NULL,                  -- monthly, quarterly, yearly
    year INTEGER NOT NULL,
    month INTEGER,                         -- for monthly goals
    quarter INTEGER,                       -- for quarterly goals
    description TEXT,
    goal_type TEXT DEFAULT 'income',       -- income, hours, clients, savings
    target_hours REAL,                     -- for hours goals
    target_clients INTEGER,                -- for client goals
    savings_percent REAL,                  -- for savings goals (e.g., 30 for 30%)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Tasks within gigs - actionable items to complete
CREATE TABLE IF NOT EXISTS gig_tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    gig_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    completed BOOLEAN DEFAULT 0,
    completed_at TIMESTAMP,
    due_date TIMESTAMP,
    sort_order INTEGER DEFAULT 0,           -- For drag-and-drop reordering
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (gig_id) REFERENCES gigs(id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_gigs_status ON gigs(status);
CREATE INDEX IF NOT EXISTS idx_gigs_client ON gigs(client_id);
CREATE INDEX IF NOT EXISTS idx_gigs_priority ON gigs(priority);
CREATE INDEX IF NOT EXISTS idx_gigs_due_date ON gigs(due_date);
CREATE INDEX IF NOT EXISTS idx_gig_tasks_gig ON gig_tasks(gig_id);
CREATE INDEX IF NOT EXISTS idx_gig_tasks_completed ON gig_tasks(completed);
CREATE INDEX IF NOT EXISTS idx_work_logs_gig ON work_logs(gig_id);
CREATE INDEX IF NOT EXISTS idx_work_logs_client ON work_logs(client_id);
CREATE INDEX IF NOT EXISTS idx_work_logs_created ON work_logs(created_at);
