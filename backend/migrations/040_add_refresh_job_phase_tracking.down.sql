-- Migration 040 Down: Remove refresh job phase tracking tables

DROP TRIGGER IF EXISTS update_refresh_job_detectors_updated_at ON refresh_job_detectors;
DROP TRIGGER IF EXISTS update_refresh_job_phases_updated_at ON refresh_job_phases;

DROP TABLE IF EXISTS refresh_job_detectors;
DROP TABLE IF EXISTS refresh_job_phases;
