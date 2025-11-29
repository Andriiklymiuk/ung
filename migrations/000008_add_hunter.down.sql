-- Rollback Job Hunter feature

DROP INDEX IF EXISTS idx_applications_status;
DROP INDEX IF EXISTS idx_applications_job;
DROP INDEX IF EXISTS idx_jobs_posted_at;
DROP INDEX IF EXISTS idx_jobs_match_score;
DROP INDEX IF EXISTS idx_jobs_source;

DROP TABLE IF EXISTS applications;
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS profiles;
