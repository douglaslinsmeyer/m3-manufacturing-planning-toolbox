-- M3 Companies cache (environment-specific)
CREATE TABLE IF NOT EXISTS m3_companies (
    id SERIAL PRIMARY KEY,
    environment VARCHAR(10) NOT NULL,
    company_number VARCHAR(10) NOT NULL,
    company_name VARCHAR(100),
    currency VARCHAR(10),
    database_name VARCHAR(50),
    cached_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_m3_company UNIQUE(environment, company_number)
);

CREATE INDEX idx_m3_companies_env ON m3_companies(environment);
CREATE INDEX idx_m3_companies_cached_at ON m3_companies(cached_at);

-- M3 Divisions cache (environment-specific)
CREATE TABLE IF NOT EXISTS m3_divisions (
    id SERIAL PRIMARY KEY,
    environment VARCHAR(10) NOT NULL,
    company_number VARCHAR(10) NOT NULL,
    division VARCHAR(10) NOT NULL,
    division_name VARCHAR(100),
    facility VARCHAR(10),
    warehouse VARCHAR(10),
    cached_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_m3_division UNIQUE(environment, company_number, division)
);

CREATE INDEX idx_m3_divisions_env_company ON m3_divisions(environment, company_number);
CREATE INDEX idx_m3_divisions_cached_at ON m3_divisions(cached_at);

-- M3 Facilities cache (environment-specific)
CREATE TABLE IF NOT EXISTS m3_facilities (
    id SERIAL PRIMARY KEY,
    environment VARCHAR(10) NOT NULL,
    company_number VARCHAR(10) NOT NULL,
    facility VARCHAR(10) NOT NULL,
    facility_name VARCHAR(100),
    division VARCHAR(10),
    warehouse VARCHAR(10),
    cached_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_m3_facility UNIQUE(environment, company_number, facility)
);

CREATE INDEX idx_m3_facilities_env_company ON m3_facilities(environment, company_number);
CREATE INDEX idx_m3_facilities_cached_at ON m3_facilities(cached_at);

-- M3 Warehouses cache (environment-specific)
CREATE TABLE IF NOT EXISTS m3_warehouses (
    id SERIAL PRIMARY KEY,
    environment VARCHAR(10) NOT NULL,
    company_number VARCHAR(10) NOT NULL,
    warehouse VARCHAR(10) NOT NULL,
    warehouse_name VARCHAR(100),
    division VARCHAR(10),
    facility VARCHAR(10),
    cached_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_m3_warehouse UNIQUE(environment, company_number, warehouse)
);

CREATE INDEX idx_m3_warehouses_env_company_div_faci ON m3_warehouses(environment, company_number, division, facility);
CREATE INDEX idx_m3_warehouses_cached_at ON m3_warehouses(cached_at);
