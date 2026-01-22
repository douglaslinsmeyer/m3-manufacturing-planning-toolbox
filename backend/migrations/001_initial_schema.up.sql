-- M3 Manufacturing Planning Tools - Initial Schema

-- ========================================
-- Production Orders (Unified View)
-- ========================================
-- Lightweight table with common fields for sequencing and analysis
-- Links to detail tables for full MO or MOP information

CREATE TABLE production_orders (
    id BIGSERIAL PRIMARY KEY,
    order_number VARCHAR(50) NOT NULL,
    order_type VARCHAR(10) NOT NULL CHECK (order_type IN ('MO', 'MOP')),

    -- Basic order info
    item_number VARCHAR(50) NOT NULL,
    item_description TEXT,
    facility VARCHAR(10) NOT NULL,
    warehouse VARCHAR(10),

    -- Dates (critical for analysis)
    planned_start_date DATE NOT NULL,
    planned_finish_date DATE NOT NULL,

    -- Quantity
    ordered_quantity DECIMAL(15, 3) NOT NULL,
    quantity_unit VARCHAR(10),

    -- Status
    status VARCHAR(50) NOT NULL,

    -- References to detail tables
    mo_id BIGINT,
    mop_id BIGINT,

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Indexes for common queries
    CONSTRAINT unique_order_number UNIQUE (order_number, order_type)
);

CREATE INDEX idx_production_orders_dates ON production_orders(planned_start_date, planned_finish_date);
CREATE INDEX idx_production_orders_facility ON production_orders(facility);
CREATE INDEX idx_production_orders_item ON production_orders(item_number);
CREATE INDEX idx_production_orders_type ON production_orders(order_type);
CREATE INDEX idx_production_orders_status ON production_orders(status);

-- ========================================
-- Manufacturing Orders (Full MO Details)
-- ========================================

CREATE TABLE manufacturing_orders (
    id BIGSERIAL PRIMARY KEY,

    -- Core identifiers
    facility VARCHAR(10) NOT NULL,
    mo_number VARCHAR(50) NOT NULL,
    product_number VARCHAR(50) NOT NULL,

    -- Item information
    item_number VARCHAR(50) NOT NULL,
    item_description TEXT,
    product_structure VARCHAR(50),

    -- Quantities
    ordered_quantity DECIMAL(15, 3) NOT NULL,
    manufactured_quantity DECIMAL(15, 3) DEFAULT 0,
    scrapped_quantity DECIMAL(15, 3) DEFAULT 0,
    quantity_unit VARCHAR(10),

    -- Dates and times
    planned_start_date DATE,
    planned_finish_date DATE,
    actual_start_date DATE,
    actual_finish_date DATE,

    -- Status and workflow
    status VARCHAR(50) NOT NULL,
    order_type VARCHAR(10),
    priority VARCHAR(10),

    -- Organizational context
    warehouse VARCHAR(10),
    responsible_person VARCHAR(50),
    planner VARCHAR(50),

    -- Financial
    price_unit DECIMAL(15, 2),
    planned_cost DECIMAL(15, 2),
    actual_cost DECIMAL(15, 2),

    -- References
    reference_order_category VARCHAR(10),
    reference_order_number VARCHAR(50),
    reference_order_line VARCHAR(10),

    -- Additional M3 fields (JSON for flexibility)
    additional_data JSONB,

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_mo_number UNIQUE (facility, mo_number)
);

CREATE INDEX idx_mo_facility_number ON manufacturing_orders(facility, mo_number);
CREATE INDEX idx_mo_item ON manufacturing_orders(item_number);
CREATE INDEX idx_mo_status ON manufacturing_orders(status);
CREATE INDEX idx_mo_dates ON manufacturing_orders(planned_start_date, planned_finish_date);
CREATE INDEX idx_mo_reference ON manufacturing_orders(reference_order_category, reference_order_number);

-- ========================================
-- MO Operations
-- ========================================

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

-- ========================================
-- MO Materials
-- ========================================

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

-- ========================================
-- Planned Manufacturing Orders (Full MOP Details)
-- ========================================

CREATE TABLE planned_manufacturing_orders (
    id BIGSERIAL PRIMARY KEY,

    -- Core identifiers
    mop_number VARCHAR(50) NOT NULL,
    facility VARCHAR(10) NOT NULL,

    -- Item information
    item_number VARCHAR(50) NOT NULL,
    item_description TEXT,

    -- Quantities
    planned_quantity DECIMAL(15, 3) NOT NULL,
    quantity_unit VARCHAR(10),

    -- Dates
    planned_order_date DATE,
    planned_start_date DATE,
    planned_finish_date DATE,
    requirement_date DATE,

    -- Status and workflow
    status VARCHAR(50) NOT NULL,
    proposal_status VARCHAR(50),

    -- Demand information
    demand_order_category VARCHAR(10),
    demand_order_number VARCHAR(50),
    demand_order_line VARCHAR(10),

    -- Planning parameters
    order_policy VARCHAR(10),
    lot_size DECIMAL(15, 3),
    safety_stock DECIMAL(15, 3),
    reorder_point DECIMAL(15, 3),

    -- Organizational context
    warehouse VARCHAR(10),
    buyer VARCHAR(50),
    planner VARCHAR(50),

    -- Additional M3 fields (JSON for flexibility)
    additional_data JSONB,

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_mop_number UNIQUE (mop_number)
);

CREATE INDEX idx_mop_number ON planned_manufacturing_orders(mop_number);
CREATE INDEX idx_mop_item ON planned_manufacturing_orders(item_number);
CREATE INDEX idx_mop_facility ON planned_manufacturing_orders(facility);
CREATE INDEX idx_mop_status ON planned_manufacturing_orders(status);
CREATE INDEX idx_mop_dates ON planned_manufacturing_orders(planned_start_date, planned_finish_date);
CREATE INDEX idx_mop_demand ON planned_manufacturing_orders(demand_order_category, demand_order_number);

-- ========================================
-- Customer Orders
-- ========================================

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

-- ========================================
-- Customer Order Lines
-- ========================================

CREATE TABLE customer_order_lines (
    id BIGSERIAL PRIMARY KEY,
    co_id BIGINT NOT NULL REFERENCES customer_orders(id) ON DELETE CASCADE,

    -- Core identifiers
    order_number VARCHAR(50) NOT NULL,
    line_number VARCHAR(10) NOT NULL,
    line_suffix VARCHAR(10),

    -- Item details
    item_number VARCHAR(50) NOT NULL,
    item_description TEXT,

    -- Quantities
    ordered_quantity DECIMAL(15, 3) NOT NULL,
    delivered_quantity DECIMAL(15, 3) DEFAULT 0,
    quantity_unit VARCHAR(10),

    -- Dates
    requested_delivery_date DATE,
    confirmed_delivery_date DATE,
    actual_delivery_date DATE,

    -- Status
    status VARCHAR(50) NOT NULL,
    line_type VARCHAR(10),

    -- Pricing
    unit_price DECIMAL(15, 4),
    line_amount DECIMAL(15, 2),

    -- Warehouse
    warehouse VARCHAR(10),

    -- Additional M3 fields
    additional_data JSONB,

    -- Metadata
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT unique_order_line UNIQUE (order_number, line_number, line_suffix)
);

CREATE INDEX idx_co_lines_order ON customer_order_lines(co_id);
CREATE INDEX idx_co_lines_item ON customer_order_lines(item_number);
CREATE INDEX idx_co_lines_status ON customer_order_lines(status);
CREATE INDEX idx_co_lines_dates ON customer_order_lines(requested_delivery_date, confirmed_delivery_date);

-- ========================================
-- Deliveries
-- ========================================

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

-- ========================================
-- Snapshot Metadata
-- ========================================
-- Tracks refresh jobs and data freshness

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

-- ========================================
-- Update Triggers
-- ========================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply trigger to all tables with updated_at
CREATE TRIGGER update_production_orders_updated_at BEFORE UPDATE ON production_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_manufacturing_orders_updated_at BEFORE UPDATE ON manufacturing_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_mo_operations_updated_at BEFORE UPDATE ON mo_operations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_mo_materials_updated_at BEFORE UPDATE ON mo_materials FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_planned_manufacturing_orders_updated_at BEFORE UPDATE ON planned_manufacturing_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_customer_orders_updated_at BEFORE UPDATE ON customer_orders FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_customer_order_lines_updated_at BEFORE UPDATE ON customer_order_lines FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_deliveries_updated_at BEFORE UPDATE ON deliveries FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
