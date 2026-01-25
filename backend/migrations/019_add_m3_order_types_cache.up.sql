-- M3 Manufacturing Order Types cache (environment and company-specific)
-- Caches order type descriptions for manufacturing orders (ORTY from MWOHED/MMOPLP)
CREATE TABLE IF NOT EXISTS m3_manufacturing_order_types (
    id SERIAL PRIMARY KEY,
    environment VARCHAR(10) NOT NULL,
    company_number VARCHAR(10) NOT NULL,
    order_type VARCHAR(10) NOT NULL,
    order_type_description VARCHAR(100),
    language_code VARCHAR(10) DEFAULT 'GB',
    cached_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_m3_mfg_order_type UNIQUE(environment, company_number, order_type, language_code)
);

CREATE INDEX idx_m3_mfg_order_types_env_company ON m3_manufacturing_order_types(environment, company_number);
CREATE INDEX idx_m3_mfg_order_types_cached_at ON m3_manufacturing_order_types(cached_at);
CREATE INDEX idx_m3_mfg_order_types_lookup ON m3_manufacturing_order_types(environment, company_number, order_type);

COMMENT ON TABLE m3_manufacturing_order_types IS 'Cached M3 manufacturing order type descriptions (ORTY from MWOHED/MMOPLP)';
COMMENT ON COLUMN m3_manufacturing_order_types.order_type IS 'Manufacturing order type code (ORTY) - e.g., 301, 302, etc.';

-- M3 Customer Order Types cache (environment and company-specific)
-- Caches order type descriptions for customer orders (ORTY from OOLINE/OOHEAD)
CREATE TABLE IF NOT EXISTS m3_customer_order_types (
    id SERIAL PRIMARY KEY,
    environment VARCHAR(10) NOT NULL,
    company_number VARCHAR(10) NOT NULL,
    order_type VARCHAR(10) NOT NULL,
    order_type_description VARCHAR(100),
    language_code VARCHAR(10) DEFAULT 'GB',
    cached_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_m3_co_order_type UNIQUE(environment, company_number, order_type, language_code)
);

CREATE INDEX idx_m3_co_order_types_env_company ON m3_customer_order_types(environment, company_number);
CREATE INDEX idx_m3_co_order_types_cached_at ON m3_customer_order_types(cached_at);
CREATE INDEX idx_m3_co_order_types_lookup ON m3_customer_order_types(environment, company_number, order_type);

COMMENT ON TABLE m3_customer_order_types IS 'Cached M3 customer order type descriptions (ORTY from OOLINE/OOHEAD)';
COMMENT ON COLUMN m3_customer_order_types.order_type IS 'Customer order type code (ORTY) - e.g., 201, 202, etc.';
