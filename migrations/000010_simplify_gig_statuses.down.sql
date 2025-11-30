-- Revert simplified statuses back to original
-- Map: todo → pipeline, in_progress → active, sent → delivered, done → complete

UPDATE gigs SET status = 'pipeline' WHERE status = 'todo';
UPDATE gigs SET status = 'active' WHERE status = 'in_progress';
UPDATE gigs SET status = 'delivered' WHERE status = 'sent';
UPDATE gigs SET status = 'complete' WHERE status = 'done';

-- Drop project index
DROP INDEX IF EXISTS idx_gigs_project;

-- Note: SQLite doesn't support DROP COLUMN directly
-- The project column will remain but be unused after downgrade
