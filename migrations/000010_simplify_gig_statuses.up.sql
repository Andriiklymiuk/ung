-- Simplify gig statuses: todo → in_progress → sent → done
-- Add optional project field for personal project association

-- Add project field to gigs
ALTER TABLE gigs ADD COLUMN project TEXT;

-- Update existing statuses to new simplified values
-- Map: pipeline/negotiating → todo, active → in_progress, delivered/invoiced → sent, complete → done
UPDATE gigs SET status = 'todo' WHERE status IN ('pipeline', 'negotiating');
UPDATE gigs SET status = 'in_progress' WHERE status = 'active';
UPDATE gigs SET status = 'sent' WHERE status IN ('delivered', 'invoiced');
UPDATE gigs SET status = 'done' WHERE status = 'complete';

-- on_hold and cancelled remain as-is (edge cases)
-- Note: SQLite doesn't support CHECK constraints modification, so we document valid statuses:
-- Valid statuses: todo, in_progress, sent, done, on_hold, cancelled

-- Create index on project for filtering
CREATE INDEX IF NOT EXISTS idx_gigs_project ON gigs(project);
