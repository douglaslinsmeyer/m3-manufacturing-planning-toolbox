-- Recreate the co_id column on customer_order_lines
ALTER TABLE customer_order_lines ADD COLUMN co_id BIGINT;

-- Note: Not adding foreign key constraint or NOT NULL since customer_orders will be empty
-- Migration 004 made this nullable which is the state we're restoring to

-- Recreate indexes for co_id
CREATE INDEX IF NOT EXISTS idx_co_lines_no_header ON customer_order_lines(order_number) WHERE co_id IS NULL; -- From migration 004
CREATE INDEX IF NOT EXISTS idx_co_lines_co_id ON customer_order_lines(co_id); -- From migration 016

-- Recreate customer_orders table
CREATE TABLE customer_orders (
    id BIGSERIAL PRIMARY KEY,

    -- Core identifiers
    order_number VARCHAR(50) NOT NULL,
    customer_number VARCHAR(50) NOT NULL,
    customer_name TEXT,

    -- Order details
    order_type VARCHAR(10),
    order_date DATE,

    -- Dates
    requested_delivery_date DATE,
    confirmed_delivery_date DATE,

    -- Status
    status VARCHAR(50) NOT NULL,

    -- Financial
    currency VARCHAR(10),
    total_amount DECIMAL(15, 2),

    -- Organizational context
    warehouse VARCHAR(10),
    sales_person VARCHAR(50),

    -- Additional M3 fields
    additional_data JSONB,

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_customer_order_number UNIQUE (order_number)
);

CREATE INDEX idx_co_order_number ON customer_orders(order_number);
CREATE INDEX idx_co_customer ON customer_orders(customer_number);
CREATE INDEX idx_co_status ON customer_orders(status);
CREATE INDEX idx_co_dates ON customer_orders(requested_delivery_date, confirmed_delivery_date);

CREATE TRIGGER update_customer_orders_updated_at BEFORE UPDATE ON customer_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Recreate deliveries table
CREATE TABLE deliveries (
    id BIGSERIAL PRIMARY KEY,
    co_line_id BIGINT REFERENCES customer_order_lines(id) ON DELETE CASCADE,

    -- Core identifiers
    delivery_number VARCHAR(50) NOT NULL,
    order_number VARCHAR(50) NOT NULL,
    line_number VARCHAR(10) NOT NULL,

    -- Delivery details
    delivery_type VARCHAR(10),

    -- Quantities
    delivery_quantity DECIMAL(15, 3) NOT NULL,
    quantity_unit VARCHAR(10),

    -- Dates
    planned_delivery_date DATE,
    confirmed_delivery_date DATE,
    actual_delivery_date DATE,

    -- Status
    status VARCHAR(50) NOT NULL,

    -- Warehouse and shipping
    warehouse VARCHAR(10),
    shipment_number VARCHAR(50),

    -- Additional M3 fields
    additional_data JSONB,

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_deliveries_number ON deliveries(delivery_number);
CREATE INDEX idx_deliveries_order ON deliveries(order_number, line_number);
CREATE INDEX idx_deliveries_co_line ON deliveries(co_line_id);
CREATE INDEX idx_deliveries_dates ON deliveries(planned_delivery_date, confirmed_delivery_date);

CREATE TRIGGER update_deliveries_updated_at BEFORE UPDATE ON deliveries FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Recreate mo_operations table
CREATE TABLE mo_operations (
    id BIGSERIAL PRIMARY KEY,
    mo_id BIGINT NOT NULL REFERENCES manufacturing_orders(id) ON DELETE CASCADE,

    facility VARCHAR(10) NOT NULL,
    mo_number VARCHAR(50) NOT NULL,
    operation_number VARCHAR(10) NOT NULL,

    -- Operation details
    work_center VARCHAR(50),
    operation_description TEXT,

    -- Times
    setup_time DECIMAL(10, 2),
    run_time_per_unit DECIMAL(10, 2),
    total_run_time DECIMAL(10, 2),

    -- Quantities
    completed_quantity DECIMAL(15, 3) DEFAULT 0,
    scrapped_quantity DECIMAL(15, 3) DEFAULT 0,

    -- Status
    status VARCHAR(50),

    -- Dates
    planned_start_date DATE,
    planned_finish_date DATE,
    actual_start_date DATE,
    actual_finish_date DATE,

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_mo_operation UNIQUE (facility, mo_number, operation_number)
);

CREATE INDEX idx_mo_operations_mo ON mo_operations(mo_id);
CREATE INDEX idx_mo_operations_work_center ON mo_operations(work_center);

CREATE TRIGGER update_mo_operations_updated_at BEFORE UPDATE ON mo_operations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Recreate mo_materials table
CREATE TABLE mo_materials (
    id BIGSERIAL PRIMARY KEY,
    mo_id BIGINT NOT NULL REFERENCES manufacturing_orders(id) ON DELETE CASCADE,

    facility VARCHAR(10) NOT NULL,
    mo_number VARCHAR(50) NOT NULL,

    -- Material details
    component_number VARCHAR(50) NOT NULL,
    component_description TEXT,

    -- Quantities
    required_quantity DECIMAL(15, 3) NOT NULL,
    allocated_quantity DECIMAL(15, 3) DEFAULT 0,
    issued_quantity DECIMAL(15, 3) DEFAULT 0,
    quantity_unit VARCHAR(10),

    -- Dates
    planned_issue_date DATE,
    actual_issue_date DATE,

    -- Warehouse
    warehouse VARCHAR(10),
    location VARCHAR(20),

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_mo_materials_mo ON mo_materials(mo_id);
CREATE INDEX idx_mo_materials_component ON mo_materials(component_number);

CREATE TRIGGER update_mo_materials_updated_at BEFORE UPDATE ON mo_materials FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Recreate snapshot_metadata table
CREATE TABLE snapshot_metadata (
    id BIGSERIAL PRIMARY KEY,

    environment VARCHAR(10) NOT NULL,
    snapshot_type VARCHAR(50) NOT NULL,

    status VARCHAR(50) NOT NULL,
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,

    records_processed INTEGER DEFAULT 0,
    errors_count INTEGER DEFAULT 0,
    error_details TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_snapshot_metadata_env ON snapshot_metadata(environment);
CREATE INDEX idx_snapshot_metadata_status ON snapshot_metadata(status);
