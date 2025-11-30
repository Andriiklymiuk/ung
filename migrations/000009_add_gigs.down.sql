-- Remove gigs, gig_tasks, and work_logs tables

DROP INDEX IF EXISTS idx_gigs_status;
DROP INDEX IF EXISTS idx_gigs_client;
DROP INDEX IF EXISTS idx_gigs_priority;
DROP INDEX IF EXISTS idx_gigs_due_date;
DROP INDEX IF EXISTS idx_gig_tasks_gig;
DROP INDEX IF EXISTS idx_gig_tasks_completed;
DROP INDEX IF EXISTS idx_work_logs_gig;
DROP INDEX IF EXISTS idx_work_logs_client;
DROP INDEX IF EXISTS idx_work_logs_created;

DROP TABLE IF EXISTS gig_tasks;
DROP TABLE IF EXISTS work_logs;
DROP TABLE IF EXISTS gigs;
DROP TABLE IF EXISTS income_goals;
