-- Drop bulk operation tables
-- Foreign keys will cascade automatically

-- Drop issue results first (has FK to jobs)
DROP TABLE IF EXISTS bulk_operation_issue_results;

-- Drop batch results (has FK to jobs)
DROP TABLE IF EXISTS bulk_operation_batch_results;

-- Drop jobs table
DROP TABLE IF EXISTS bulk_operation_jobs;
