-- CO QUANTITY MISMATCH DETECTOR
-- Compares production order quantities to CO line remaining quantity (RNQA)
-- Implements M3 putaway logic: MOs don't reduce CO remaining qty until putaway complete

INSERT INTO system_settings (environment, setting_key, setting_value, setting_type, description, category, constraints) VALUES

    -- TRN Environment
    ('TRN', 'detector_co_quantity_mismatch_enabled',
     'true',
     'boolean',
     'Enable detection of CO lines where production order quantities do not match remaining quantity',
     'detection',
     '{}'::jsonb),

    ('TRN', 'detector_co_quantity_mismatch_tolerance_threshold',
     '{"global": 0.01, "overrides": []}',
     'json',
     'Allow quantity variance within ±N units (hierarchical, default 0.01 for decimal precision)',
     'detection',
     '{"min": 0, "max": 100, "unit": "units", "hierarchical": true}'::jsonb),

    ('TRN', 'detector_co_quantity_mismatch_min_quantity_threshold',
     '0',
     'float',
     'Minimum variance threshold to report (filter small mismatches, 0 = report all)',
     'detection',
     '{"min": 0, "max": 10000, "unit": "units"}'::jsonb),

    -- PRD Environment
    ('PRD', 'detector_co_quantity_mismatch_enabled',
     'true',
     'boolean',
     'Enable detection of CO lines where production order quantities do not match remaining quantity',
     'detection',
     '{}'::jsonb),

    ('PRD', 'detector_co_quantity_mismatch_tolerance_threshold',
     '{"global": 0.01, "overrides": []}',
     'json',
     'Allow quantity variance within ±N units (hierarchical, default 0.01 for decimal precision)',
     'detection',
     '{"min": 0, "max": 100, "unit": "units", "hierarchical": true}'::jsonb),

    ('PRD', 'detector_co_quantity_mismatch_min_quantity_threshold',
     '0',
     'float',
     'Minimum variance threshold to report (filter small mismatches, 0 = report all)',
     'detection',
     '{"min": 0, "max": 10000, "unit": "units"}'::jsonb)

ON CONFLICT (environment, setting_key) DO NOTHING;
