-- Add anomaly detector configuration settings

-- UnlinkedConcentration Detector
INSERT INTO system_settings (environment, setting_key, setting_value, setting_type, description, category, constraints)
VALUES
  ('TRN', 'anomaly_unlinked_concentration_enabled', 'true', 'boolean', 'Enable unlinked concentration anomaly detection', 'anomaly_detection', '{}'),
  ('PRD', 'anomaly_unlinked_concentration_enabled', 'true', 'boolean', 'Enable unlinked concentration anomaly detection', 'anomaly_detection', '{}'),

  ('TRN', 'anomaly_unlinked_concentration_warning_threshold', '{"global": 10.0}', 'json', 'Warning threshold: % of unlinked MOPs for single product', 'anomaly_detection', '{"hierarchical": true, "unit": "%", "min": 1, "max": 100}'),
  ('PRD', 'anomaly_unlinked_concentration_warning_threshold', '{"global": 10.0}', 'json', 'Warning threshold: % of unlinked MOPs for single product', 'anomaly_detection', '{"hierarchical": true, "unit": "%", "min": 1, "max": 100}'),

  ('TRN', 'anomaly_unlinked_concentration_critical_threshold', '{"global": 50.0}', 'json', 'Critical threshold: % of unlinked MOPs for single product', 'anomaly_detection', '{"hierarchical": true, "unit": "%", "min": 1, "max": 100}'),
  ('PRD', 'anomaly_unlinked_concentration_critical_threshold', '{"global": 50.0}', 'json', 'Critical threshold: % of unlinked MOPs for single product', 'anomaly_detection', '{"hierarchical": true, "unit": "%", "min": 1, "max": 100}'),

  ('TRN', 'anomaly_unlinked_concentration_min_affected_count', '100', 'integer', 'Minimum affected records to trigger alert', 'anomaly_detection', '{"min": 1, "max": 100000}'),
  ('PRD', 'anomaly_unlinked_concentration_min_affected_count', '100', 'integer', 'Minimum affected records to trigger alert', 'anomaly_detection', '{"min": 1, "max": 100000}');

-- DateClustering Detector
INSERT INTO system_settings (environment, setting_key, setting_value, setting_type, description, category, constraints)
VALUES
  ('TRN', 'anomaly_date_clustering_enabled', 'true', 'boolean', 'Enable date clustering anomaly detection', 'anomaly_detection', '{}'),
  ('PRD', 'anomaly_date_clustering_enabled', 'true', 'boolean', 'Enable date clustering anomaly detection', 'anomaly_detection', '{}'),

  ('TRN', 'anomaly_date_clustering_warning_threshold', '{"global": 80.0}', 'json', 'Warning threshold: % of MOPs on single date', 'anomaly_detection', '{"hierarchical": true, "unit": "%", "min": 1, "max": 100}'),
  ('PRD', 'anomaly_date_clustering_warning_threshold', '{"global": 80.0}', 'json', 'Warning threshold: % of MOPs on single date', 'anomaly_detection', '{"hierarchical": true, "unit": "%", "min": 1, "max": 100}'),

  ('TRN', 'anomaly_date_clustering_critical_threshold', '{"global": 95.0}', 'json', 'Critical threshold: % of MOPs on single date', 'anomaly_detection', '{"hierarchical": true, "unit": "%", "min": 1, "max": 100}'),
  ('PRD', 'anomaly_date_clustering_critical_threshold', '{"global": 95.0}', 'json', 'Critical threshold: % of MOPs on single date', 'anomaly_detection', '{"hierarchical": true, "unit": "%", "min": 1, "max": 100}'),

  ('TRN', 'anomaly_date_clustering_min_affected_count', '100', 'integer', 'Minimum affected records to trigger alert', 'anomaly_detection', '{"min": 1, "max": 100000}'),
  ('PRD', 'anomaly_date_clustering_min_affected_count', '100', 'integer', 'Minimum affected records to trigger alert', 'anomaly_detection', '{"min": 1, "max": 100000}');

-- MOPDemandRatio Detector
INSERT INTO system_settings (environment, setting_key, setting_value, setting_type, description, category, constraints)
VALUES
  ('TRN', 'anomaly_mop_demand_ratio_enabled', 'true', 'boolean', 'Enable MOP-to-demand ratio anomaly detection', 'anomaly_detection', '{}'),
  ('PRD', 'anomaly_mop_demand_ratio_enabled', 'true', 'boolean', 'Enable MOP-to-demand ratio anomaly detection', 'anomaly_detection', '{}'),

  ('TRN', 'anomaly_mop_demand_ratio_warning_mops_per_co_line', '{"global": 10.0}', 'json', 'Warning threshold: unlinked MOPs per CO line', 'anomaly_detection', '{"hierarchical": true, "min": 1, "max": 1000}'),
  ('PRD', 'anomaly_mop_demand_ratio_warning_mops_per_co_line', '{"global": 10.0}', 'json', 'Warning threshold: unlinked MOPs per CO line', 'anomaly_detection', '{"hierarchical": true, "min": 1, "max": 1000}'),

  ('TRN', 'anomaly_mop_demand_ratio_critical_mops_per_co_line', '{"global": 50.0}', 'json', 'Critical threshold: unlinked MOPs per CO line', 'anomaly_detection', '{"hierarchical": true, "min": 1, "max": 1000}'),
  ('PRD', 'anomaly_mop_demand_ratio_critical_mops_per_co_line', '{"global": 50.0}', 'json', 'Critical threshold: unlinked MOPs per CO line', 'anomaly_detection', '{"hierarchical": true, "min": 1, "max": 1000}'),

  ('TRN', 'anomaly_mop_demand_ratio_critical_mops_per_unit_demand', '{"global": 5.0}', 'json', 'Critical threshold: unlinked MOPs per unit demand', 'anomaly_detection', '{"hierarchical": true, "min": 0.1, "max": 100}'),
  ('PRD', 'anomaly_mop_demand_ratio_critical_mops_per_unit_demand', '{"global": 5.0}', 'json', 'Critical threshold: unlinked MOPs per unit demand', 'anomaly_detection', '{"hierarchical": true, "min": 0.1, "max": 100}');

-- AbsoluteVolume Detector
INSERT INTO system_settings (environment, setting_key, setting_value, setting_type, description, category, constraints)
VALUES
  ('TRN', 'anomaly_absolute_volume_enabled', 'true', 'boolean', 'Enable absolute volume anomaly detection', 'anomaly_detection', '{}'),
  ('PRD', 'anomaly_absolute_volume_enabled', 'true', 'boolean', 'Enable absolute volume anomaly detection', 'anomaly_detection', '{}'),

  ('TRN', 'anomaly_absolute_volume_warning_threshold', '{"global": 1000}', 'json', 'Warning threshold: unlinked MOPs for single product/warehouse', 'anomaly_detection', '{"hierarchical": true, "min": 1, "max": 100000}'),
  ('PRD', 'anomaly_absolute_volume_warning_threshold', '{"global": 1000}', 'json', 'Warning threshold: unlinked MOPs for single product/warehouse', 'anomaly_detection', '{"hierarchical": true, "min": 1, "max": 100000}'),

  ('TRN', 'anomaly_absolute_volume_critical_threshold', '{"global": 10000}', 'json', 'Critical threshold: unlinked MOPs for single product/warehouse', 'anomaly_detection', '{"hierarchical": true, "min": 1, "max": 100000}'),
  ('PRD', 'anomaly_absolute_volume_critical_threshold', '{"global": 10000}', 'json', 'Critical threshold: unlinked MOPs for single product/warehouse', 'anomaly_detection', '{"hierarchical": true, "min": 1, "max": 100000}');
