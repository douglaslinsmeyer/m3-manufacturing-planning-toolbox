-- Migration 017 Down: Remove Issue Detection System

DROP TABLE IF EXISTS detected_issues;
DROP TABLE IF EXISTS issue_detection_jobs;
