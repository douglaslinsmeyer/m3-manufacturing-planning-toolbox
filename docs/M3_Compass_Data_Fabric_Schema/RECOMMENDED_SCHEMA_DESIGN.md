# Recommended Database Schema Design

## Overview

This document provides a **complete, production-ready database schema** for storing M3 customer order, manufacturing, and delivery data harvested from the Data Fabric.

---

## Design Principles

1. **Hybrid Approach**: Core fields as columns + flexible attributes in JSONB
2. **Normalized Structure**: Proper foreign keys and relationships
3. **Performance Optimized**: Strategic indexes for common queries
4. **Change Tracking**: Support for incremental loads
5. **Audit Trail**: Full history of modifications
6. **Flexible Attributes**: Handle M3's variable attribute model

---

## Complete Schema DDL

### 1. Customer Order Header

```sql
CREATE TABLE customer_order_headers (
    -- Internal ID
    id BIGSERIAL PRIMARY KEY,

    -- M3 Keys
    company_number INTEGER NOT NULL,
    order_number VARCHAR(20) NOT NULL,

    -- Order Information
    division VARCHAR(10),
    order_type VARCHAR(10),
    facility VARCHAR(10),
    warehouse VARCHAR(10),

    -- Status
    highest_status VARCHAR(10),
    lowest_status VARCHAR(10),

    -- Customers (multi-customer support)
    customer_number VARCHAR(20),
    delivery_customer VARCHAR(20),
    payer VARCHAR(20),
    invoice_recipient VARCHAR(20),

    -- Dates
    order_date DATE,
    customer_po_date DATE,
    requested_delivery_date DATE,
    requested_delivery_time TIME,
    earliest_delivery_date DATE,
    first_delivery_date DATE,
    last_delivery_date DATE,

    -- Financial
    order_currency VARCHAR(10),
    exchange_rate NUMERIC(12,6),
    gross_amount NUMERIC(15,2),
    net_amount NUMERIC(15,2),
    total_cost NUMERIC(15,2),

    -- Terms
    payment_terms VARCHAR(10),
    delivery_method VARCHAR(10),
    delivery_terms VARCHAR(10),

    -- References
    salesperson VARCHAR(20),
    our_reference VARCHAR(30),
    customer_po_number VARCHAR(30),
    quotation_number VARCHAR(20),

    -- Project
    project_number VARCHAR(20),
    project_element VARCHAR(20),

    -- Business Chain (organizational hierarchy)
    business_chain JSONB,  -- {"level1": "REGION", "level2": "DISTRICT", ...}

    -- User-Defined Fields as JSONB
    user_fields JSONB,  -- {"UCA1": "value", "UDN1": 123.45, "UID1": "2024-01-15"}

    -- M3 Metadata
    m3_entry_date DATE,
    m3_change_date DATE,
    m3_changed_by VARCHAR(20),
    m3_change_number INTEGER,
    m3_timestamp BIGINT,

    -- Sync Metadata
    sync_timestamp TIMESTAMP DEFAULT NOW(),
    is_deleted BOOLEAN DEFAULT FALSE,

    -- Constraints
    UNIQUE(company_number, order_number)
);

-- Indexes
CREATE INDEX idx_co_header_cuno ON customer_order_headers(customer_number);
CREATE INDEX idx_co_header_order_date ON customer_order_headers(order_date);
CREATE INDEX idx_co_header_status ON customer_order_headers(highest_status);
CREATE INDEX idx_co_header_change_date ON customer_order_headers(m3_change_date);
CREATE INDEX idx_co_header_user_fields ON customer_order_headers USING GIN(user_fields);
```

### 2. Customer Order Lines

```sql
CREATE TABLE customer_order_lines (
    -- Internal ID
    id BIGSERIAL PRIMARY KEY,

    -- M3 Keys
    company_number INTEGER NOT NULL,
    order_number VARCHAR(20) NOT NULL,
    line_number INTEGER NOT NULL,
    line_suffix INTEGER NOT NULL DEFAULT 0,

    -- Foreign Keys
    order_header_id BIGINT REFERENCES customer_order_headers(id),

    -- Line Information
    division VARCHAR(10),
    line_type VARCHAR(10),
    item_number VARCHAR(20),
    item_description VARCHAR(100),
    facility VARCHAR(10),
    warehouse VARCHAR(10),

    -- Status
    line_status VARCHAR(10),

    -- Quantities (basic U/M)
    ordered_qty NUMERIC(15,6),
    remaining_qty NUMERIC(15,6),
    allocated_qty NUMERIC(15,6),
    delivered_qty NUMERIC(15,6),
    invoiced_qty NUMERIC(15,6),
    returned_qty NUMERIC(15,6),

    -- Quantities (alternate U/M)
    ordered_qty_alt NUMERIC(15,6),
    alternate_uom VARCHAR(10),
    conversion_factor NUMERIC(12,6),

    -- Pricing
    sales_price NUMERIC(15,2),
    net_price NUMERIC(15,2),
    line_amount NUMERIC(15,2),
    currency VARCHAR(10),
    cost_price NUMERIC(15,2),

    -- Dates
    requested_delivery_date DATE,
    confirmed_delivery_date DATE,
    planning_date DATE,

    -- Customer Reference
    customer_number VARCHAR(20),
    customer_po_number VARCHAR(30),
    customer_line_number INTEGER,

    -- Reference Orders (for legacy RORC linking)
    reference_order_category INTEGER,
    reference_order_number VARCHAR(20),
    reference_order_line INTEGER,
    reference_line_suffix INTEGER,

    -- Attributes
    attribute_number BIGINT,
    attribute_model VARCHAR(20),

    -- Built-in Attributes (M3 ATV fields)
    builtin_attributes JSONB,  -- {"ATV1": 123.45, "ATV6": "Red", ...}

    -- User-Defined Attributes
    user_fields JSONB,  -- {"UCA1": "value", "UDN1": 1000, ...}

    -- Discounts (store in JSONB for flexibility)
    discounts JSONB,  -- {"DIP1": 5.0, "DIA1": 100.00, "DIC1": 1, ...}

    -- Project
    project_number VARCHAR(20),
    project_element VARCHAR(20),

    -- Configuration
    configuration_number BIGINT,
    main_product VARCHAR(20),

    -- M3 Metadata
    m3_entry_date DATE,
    m3_change_date DATE,
    m3_changed_by VARCHAR(20),
    m3_timestamp BIGINT,

    -- Sync Metadata
    sync_timestamp TIMESTAMP DEFAULT NOW(),
    is_deleted BOOLEAN DEFAULT FALSE,

    -- Constraints
    UNIQUE(company_number, order_number, line_number, line_suffix)
);

-- Indexes
CREATE INDEX idx_co_line_order ON customer_order_lines(order_number);
CREATE INDEX idx_co_line_item ON customer_order_lines(item_number);
CREATE INDEX idx_co_line_status ON customer_order_lines(line_status);
CREATE INDEX idx_co_line_change_date ON customer_order_lines(m3_change_date);
CREATE INDEX idx_co_line_reference ON customer_order_lines(
    reference_order_category, reference_order_number, reference_order_line
);
CREATE INDEX idx_co_line_builtin_attrs ON customer_order_lines USING GIN(builtin_attributes);
CREATE INDEX idx_co_line_user_fields ON customer_order_lines USING GIN(user_fields);
CREATE INDEX idx_co_line_fk_header ON customer_order_lines(order_header_id);
```

### 3. Order Line Attributes

```sql
CREATE TABLE order_line_attributes (
    -- Internal ID
    id BIGSERIAL PRIMARY KEY,

    -- M3 Keys
    company_number INTEGER NOT NULL,
    attribute_number BIGINT NOT NULL,
    attribute_sequence INTEGER NOT NULL,

    -- Foreign Keys
    order_line_id BIGINT REFERENCES customer_order_lines(id),

    -- Order Reference
    order_category VARCHAR(10),
    order_number VARCHAR(20),
    order_line INTEGER,
    line_suffix INTEGER,

    -- Attribute Definition
    attribute_id VARCHAR(20),
    attribute_model VARCHAR(20),
    item_number VARCHAR(20),

    -- Values (flexible - store all, query based on type)
    target_value_text VARCHAR(100),
    target_value_numeric NUMERIC(15,6),
    from_value_text VARCHAR(100),
    to_value_text VARCHAR(100),
    from_value_numeric NUMERIC(15,6),
    to_value_numeric NUMERIC(15,6),

    -- Configuration
    is_main_attribute BOOLEAN,
    search_sequence INTEGER,
    planning_attribute INTEGER,
    costing_attribute INTEGER,

    -- Status
    attribute_status VARCHAR(10),
    error_code VARCHAR(10),

    -- M3 Metadata
    m3_change_date DATE,
    m3_timestamp BIGINT,

    -- Sync Metadata
    sync_timestamp TIMESTAMP DEFAULT NOW(),
    is_deleted BOOLEAN DEFAULT FALSE,

    -- Constraints
    UNIQUE(company_number, attribute_number, attribute_sequence)
);

-- Indexes
CREATE INDEX idx_attr_order_ref ON order_line_attributes(
    order_category, order_number, order_line, line_suffix
);
CREATE INDEX idx_attr_atid ON order_line_attributes(attribute_id);
CREATE INDEX idx_attr_atnr ON order_line_attributes(attribute_number);
CREATE INDEX idx_attr_values ON order_line_attributes(target_value_text, target_value_numeric);
CREATE INDEX idx_attr_fk_line ON order_line_attributes(order_line_id);
```

### 4. Manufacturing Orders

```sql
CREATE TABLE manufacturing_orders (
    -- Internal ID
    id BIGSERIAL PRIMARY KEY,

    -- M3 Keys
    company_number INTEGER NOT NULL,
    facility VARCHAR(10) NOT NULL,
    mo_number VARCHAR(20) NOT NULL,

    -- Product Information
    product_number VARCHAR(20),
    item_number VARCHAR(20),
    product_variant VARCHAR(20),
    division VARCHAR(10),

    -- Status
    mo_status VARCHAR(10),
    status_change_date DATE,
    highest_operation_status VARCHAR(10),
    material_status VARCHAR(10),
    hold_status VARCHAR(10),

    -- Quantities
    original_ordered_qty NUMERIC(15,6),
    ordered_qty NUMERIC(15,6),
    received_qty NUMERIC(15,6),
    manufactured_qty NUMERIC(15,6),
    yield_qty NUMERIC(15,6),
    manufacturing_uom VARCHAR(10),

    -- Dates
    original_start_date DATE,
    original_finish_date DATE,
    scheduled_start_date DATE,
    scheduled_finish_date DATE,
    actual_start_date DATE,
    actual_finish_date DATE,
    reporting_date DATE,

    -- Planning
    priority INTEGER,
    responsible VARCHAR(20),
    work_center VARCHAR(20),
    production_line VARCHAR(20),
    production_days NUMERIC(6,2),
    lead_time_this_level NUMERIC(6,2),
    times_replanned INTEGER,

    -- Hierarchy (multi-level MO structure)
    product_highest_level VARCHAR(20),
    mo_highest_level VARCHAR(20),
    product_overlying_level VARCHAR(20),
    mo_overlying_level VARCHAR(20),
    level_in_structure INTEGER,
    level_sequence INTEGER,

    -- Reference Orders (for legacy linking)
    reference_order_category INTEGER,
    reference_order_number VARCHAR(20),
    reference_order_line INTEGER,
    reference_line_suffix INTEGER,

    -- Configuration
    attribute_number BIGINT,
    configuration_number BIGINT,

    -- Warehouse
    warehouse VARCHAR(10),
    location VARCHAR(20),
    lot_number VARCHAR(20),

    -- Order Type and Origin
    order_type VARCHAR(10),
    origin_code INTEGER,

    -- Routing
    alternative_routing VARCHAR(20),
    number_of_operations INTEGER,
    number_finished_operations INTEGER,

    -- Structure
    structure_type VARCHAR(10),
    revision_number VARCHAR(20),
    explosion_method INTEGER,

    -- Action Messages
    action_message VARCHAR(10),

    -- Project
    project_number VARCHAR(20),
    project_element VARCHAR(20),

    -- All other fields as JSONB
    additional_fields JSONB,

    -- M3 Metadata
    m3_entry_date DATE,
    m3_change_date DATE,
    m3_changed_by VARCHAR(20),
    m3_timestamp BIGINT,

    -- Sync Metadata
    sync_timestamp TIMESTAMP DEFAULT NOW(),
    is_deleted BOOLEAN DEFAULT FALSE,

    -- Constraints
    UNIQUE(company_number, facility, mo_number)
);

-- Indexes
CREATE INDEX idx_mo_status ON manufacturing_orders(mo_status);
CREATE INDEX idx_mo_item ON manufacturing_orders(item_number);
CREATE INDEX idx_mo_product ON manufacturing_orders(product_number);
CREATE INDEX idx_mo_dates ON manufacturing_orders(scheduled_start_date, scheduled_finish_date);
CREATE INDEX idx_mo_hierarchy ON manufacturing_orders(mo_highest_level, level_in_structure);
CREATE INDEX idx_mo_reference ON manufacturing_orders(
    reference_order_category, reference_order_number, reference_order_line
);
CREATE INDEX idx_mo_change_date ON manufacturing_orders(m3_change_date);
```

### 5. Planned Manufacturing Orders

```sql
CREATE TABLE planned_manufacturing_orders (
    -- Internal ID
    id BIGSERIAL PRIMARY KEY,

    -- M3 Keys
    company_number INTEGER NOT NULL,
    facility VARCHAR(10) NOT NULL,
    planned_order_number BIGINT NOT NULL,
    sub_number INTEGER NOT NULL DEFAULT 0,

    -- Product Information
    product_number VARCHAR(20),
    item_number VARCHAR(20),
    warehouse VARCHAR(10),
    division VARCHAR(10),

    -- Status
    proposal_status VARCHAR(10),
    mo_status VARCHAR(10),
    action_message VARCHAR(10),

    -- Generation
    generation_reference VARCHAR(20),
    number_of_auto_updates INTEGER,

    -- Quantities
    planned_qty NUMERIC(15,6),
    ordered_qty_alt NUMERIC(15,6),
    manufacturing_uom VARCHAR(10),

    -- Dates
    release_date DATE,
    start_date DATE,
    finish_date DATE,
    start_time TIME,
    finish_time TIME,
    planning_date DATE,

    -- Planning
    responsible VARCHAR(20),
    priority VARCHAR(10),
    work_center VARCHAR(20),
    production_line VARCHAR(20),
    production_days NUMERIC(6,2),

    -- Reference Orders (for legacy linking)
    reference_order_category INTEGER,
    reference_order_number VARCHAR(20),
    reference_order_line INTEGER,
    reference_line_suffix INTEGER,
    reference_order_header VARCHAR(20),

    -- Hierarchy
    proposal_overlying_level BIGINT,
    proposal_highest_level BIGINT,
    order_dependent INTEGER,

    -- Material Availability
    material_shortage_indicator INTEGER,
    tool_shortage_indicator INTEGER,

    -- Configuration
    attribute_number BIGINT,
    configuration_number BIGINT,
    main_product VARCHAR(20),

    -- Warning Messages
    warnings JSONB,  -- {"MSG1": "text", "MSG2": "text", ...}

    -- All other fields as JSONB
    additional_fields JSONB,

    -- M3 Metadata
    m3_entry_date DATE,
    m3_change_date DATE,
    m3_changed_by VARCHAR(20),
    m3_timestamp BIGINT,

    -- Sync Metadata
    sync_timestamp TIMESTAMP DEFAULT NOW(),
    is_deleted BOOLEAN DEFAULT FALSE,

    -- Constraints
    UNIQUE(company_number, facility, planned_order_number, sub_number)
);

-- Indexes
CREATE INDEX idx_mop_status ON planned_manufacturing_orders(proposal_status);
CREATE INDEX idx_mop_action ON planned_manufacturing_orders(action_message);
CREATE INDEX idx_mop_item ON planned_manufacturing_orders(item_number);
CREATE INDEX idx_mop_dates ON planned_manufacturing_orders(release_date, finish_date);
CREATE INDEX idx_mop_reference ON planned_manufacturing_orders(
    reference_order_category, reference_order_number, reference_order_line
);
CREATE INDEX idx_mop_hierarchy ON planned_manufacturing_orders(proposal_highest_level);
CREATE INDEX idx_mop_shortage ON planned_manufacturing_orders(material_shortage_indicator);
```

### 6. Pre-Allocations (CRITICAL LINKING TABLE) ⭐

```sql
CREATE TABLE preallocation_links (
    -- Internal ID
    id BIGSERIAL PRIMARY KEY,

    -- M3 Keys
    company_number INTEGER NOT NULL,
    warehouse VARCHAR(10) NOT NULL,
    item_number VARCHAR(20) NOT NULL,

    -- Demand Side (what needs to be fulfilled)
    demand_category VARCHAR(10) NOT NULL,
    demand_order VARCHAR(20) NOT NULL,
    demand_line INTEGER NOT NULL,
    demand_suffix INTEGER NOT NULL DEFAULT 0,
    demand_sub_1 INTEGER,
    demand_sub_2 INTEGER,

    -- Supply Side (what will fulfill it)
    supply_category VARCHAR(10) NOT NULL,
    supply_order VARCHAR(20) NOT NULL,
    supply_line INTEGER NOT NULL DEFAULT 0,
    supply_suffix INTEGER NOT NULL DEFAULT 0,
    supply_sub_1 INTEGER,
    supply_sub_2 INTEGER,

    -- Foreign Keys (to internal tables)
    demand_order_line_id BIGINT REFERENCES customer_order_lines(id),
    supply_mo_id BIGINT REFERENCES manufacturing_orders(id),
    supply_mop_id BIGINT REFERENCES planned_manufacturing_orders(id),

    -- Quantities
    preallocated_qty NUMERIC(15,6),
    preallocated_qty_reserve NUMERIC(15,6),

    -- Configuration
    preallocation_type VARCHAR(10),
    status VARCHAR(10),
    responsible VARCHAR(20),

    -- Supply Chain
    supply_chain_number VARCHAR(20),
    supply_chain_policy VARCHAR(20),

    -- Notification Flags
    notify_on_change BOOLEAN,
    notify_on_delete BOOLEAN,
    notify_demand_change BOOLEAN,
    notify_demand_delete BOOLEAN,
    notify_supply_change BOOLEAN,
    notify_supply_delete BOOLEAN,

    -- Backorder Status
    supply_backorder BOOLEAN,
    demand_backorder BOOLEAN,

    -- Demand Structure Details
    demand_material_seq INTEGER,
    demand_operation_number INTEGER,
    demand_structure_seq BIGINT,

    -- M3 Metadata
    m3_entry_date DATE,
    m3_change_date DATE,
    m3_changed_by VARCHAR(20),
    m3_timestamp BIGINT,

    -- Sync Metadata
    sync_timestamp TIMESTAMP DEFAULT NOW(),
    is_deleted BOOLEAN DEFAULT FALSE,

    -- Constraints
    UNIQUE(company_number, warehouse, item_number,
           demand_category, demand_order, demand_line, demand_suffix,
           supply_category, supply_order, supply_line, supply_suffix)
);

-- Indexes (CRITICAL for performance)
CREATE INDEX idx_prealloc_demand ON preallocation_links(
    demand_category, demand_order, demand_line, demand_suffix
);

CREATE INDEX idx_prealloc_supply ON preallocation_links(
    supply_category, supply_order, supply_line, supply_suffix
);

CREATE INDEX idx_prealloc_item ON preallocation_links(item_number);
CREATE INDEX idx_prealloc_status ON preallocation_links(status);
CREATE INDEX idx_prealloc_fk_demand ON preallocation_links(demand_order_line_id);
CREATE INDEX idx_prealloc_fk_mo ON preallocation_links(supply_mo_id);
CREATE INDEX idx_prealloc_fk_mop ON preallocation_links(supply_mop_id);

-- Specialized index for common CO→MO query
CREATE INDEX idx_prealloc_co_to_mo ON preallocation_links(
    demand_category, demand_order, demand_line, supply_category
) WHERE demand_category = '3' AND supply_category IN ('2', '5');
```

### 7. Delivery Headers

```sql
CREATE TABLE delivery_headers (
    -- Internal ID
    id BIGSERIAL PRIMARY KEY,

    -- M3 Keys
    company_number INTEGER NOT NULL,
    delivery_number BIGINT NOT NULL,

    -- Foreign Keys
    order_header_id BIGINT REFERENCES customer_order_headers(id),

    -- Order Information
    order_number VARCHAR(20) NOT NULL,
    division VARCHAR(10),
    order_type VARCHAR(10),
    facility VARCHAR(10),
    warehouse VARCHAR(10),

    -- Status
    highest_status VARCHAR(10),

    -- Customers
    customer_number VARCHAR(20),
    delivery_customer VARCHAR(20),
    payer VARCHAR(20),
    invoice_recipient VARCHAR(20),

    -- Delivery Information
    planned_delivery_date DATE,
    delivery_time TIME,
    release_date DATE,

    -- Invoice Information
    invoice_number BIGINT,
    invoice_year INTEGER,
    invoice_date DATE,
    accounting_date DATE,

    -- Financial
    currency VARCHAR(10),
    exchange_rate NUMERIC(12,6),
    gross_amount NUMERIC(15,2),
    net_amount NUMERIC(15,2),

    -- Physical
    gross_weight NUMERIC(12,3),
    net_weight NUMERIC(12,3),
    volume NUMERIC(12,3),

    -- Logistics
    shipment_number BIGINT,
    wave_number VARCHAR(20),
    delivery_note_number VARCHAR(30),
    route VARCHAR(20),
    route_departure INTEGER,
    delivery_method VARCHAR(10),
    delivery_terms VARCHAR(10),

    -- Approval
    approved_by VARCHAR(20),
    approval_date DATE,
    approval_required BOOLEAN,

    -- Return Information (for credits)
    corrected_delivery_number BIGINT,
    reference_invoice_number BIGINT,
    reference_year INTEGER,
    debit_credit_code VARCHAR(10),

    -- Project
    project_number VARCHAR(20),
    project_element VARCHAR(20),

    -- M3 Metadata
    m3_entry_date DATE,
    m3_change_date DATE,
    m3_changed_by VARCHAR(20),
    m3_timestamp BIGINT,

    -- Sync Metadata
    sync_timestamp TIMESTAMP DEFAULT NOW(),
    is_deleted BOOLEAN DEFAULT FALSE,

    -- Constraints
    UNIQUE(company_number, delivery_number)
);

-- Indexes
CREATE INDEX idx_delivery_header_order ON delivery_headers(order_number);
CREATE INDEX idx_delivery_header_date ON delivery_headers(planned_delivery_date);
CREATE INDEX idx_delivery_header_invoice ON delivery_headers(invoice_year, invoice_number);
CREATE INDEX idx_delivery_header_shipment ON delivery_headers(shipment_number);
CREATE INDEX idx_delivery_header_wave ON delivery_headers(wave_number);
CREATE INDEX idx_delivery_header_status ON delivery_headers(highest_status);
CREATE INDEX idx_delivery_header_fk_order ON delivery_headers(order_header_id);
```

### 8. Delivery Lines

```sql
CREATE TABLE delivery_lines (
    -- Internal ID
    id BIGSERIAL PRIMARY KEY,

    -- M3 Keys
    company_number INTEGER NOT NULL,
    delivery_number BIGINT NOT NULL,
    order_number VARCHAR(20) NOT NULL,
    line_number INTEGER NOT NULL,
    line_suffix INTEGER NOT NULL DEFAULT 0,

    -- Foreign Keys
    delivery_header_id BIGINT REFERENCES delivery_headers(id),
    order_line_id BIGINT REFERENCES customer_order_lines(id),

    -- Line Information
    division VARCHAR(10),
    facility VARCHAR(10),
    warehouse VARCHAR(10),
    line_type VARCHAR(10),
    item_number VARCHAR(20),

    -- Quantities (delivered)
    delivered_qty NUMERIC(15,6),
    delivered_qty_alt NUMERIC(15,6),
    delivered_qty_price_uom NUMERIC(15,6),

    -- Quantities (invoiced)
    invoiced_qty NUMERIC(15,6),
    invoiced_qty_alt NUMERIC(15,6),
    invoiced_qty_price_uom NUMERIC(15,6),

    -- Quantity Variance
    qty_difference NUMERIC(15,6),
    qty_difference_price_uom NUMERIC(15,6),

    -- Returned
    returned_qty NUMERIC(15,6),
    returned_qty_alt NUMERIC(15,6),

    -- Pricing
    sales_price NUMERIC(15,2),
    net_price NUMERIC(15,2),
    line_amount NUMERIC(15,2),

    -- Cost
    unit_cost NUMERIC(15,2),
    issued_cost_amount NUMERIC(15,2),

    -- Discounts
    discounts JSONB,  -- All discount fields

    -- Invoice
    invoice_number BIGINT,
    invoice_year INTEGER,

    -- Supplier (drop ship)
    supplier_number VARCHAR(20),
    internal_supplier VARCHAR(20),

    -- Rebate
    supplier_rebate NUMERIC(15,2),
    pending_rebate_claim NUMERIC(15,2),
    rebate_base_amount NUMERIC(15,2),
    rebate_agreement VARCHAR(20),

    -- Catch Weight
    catch_weight NUMERIC(12,3),
    corrective_catch_weight NUMERIC(12,3),

    -- Project
    project_number VARCHAR(20),
    project_element VARCHAR(20),

    -- M3 Metadata
    m3_entry_date DATE,
    m3_change_date DATE,
    m3_changed_by VARCHAR(20),
    m3_timestamp BIGINT,

    -- Sync Metadata
    sync_timestamp TIMESTAMP DEFAULT NOW(),
    is_deleted BOOLEAN DEFAULT FALSE,

    -- Constraints
    UNIQUE(company_number, delivery_number, order_number, line_number, line_suffix)
);

-- Indexes
CREATE INDEX idx_delivery_line_dlix ON delivery_lines(delivery_number);
CREATE INDEX idx_delivery_line_order ON delivery_lines(order_number, line_number, line_suffix);
CREATE INDEX idx_delivery_line_item ON delivery_lines(item_number);
CREATE INDEX idx_delivery_line_invoice ON delivery_lines(invoice_year, invoice_number);
CREATE INDEX idx_delivery_line_fk_header ON delivery_lines(delivery_header_id);
CREATE INDEX idx_delivery_line_fk_order ON delivery_lines(order_line_id);
```

---

## Foreign Key Relationships

```sql
-- Add foreign keys (run after all tables are created)

-- Order Lines → Order Headers
ALTER TABLE customer_order_lines
ADD CONSTRAINT fk_co_line_header
FOREIGN KEY (order_header_id) REFERENCES customer_order_headers(id)
ON DELETE CASCADE;

-- Attributes → Order Lines
ALTER TABLE order_line_attributes
ADD CONSTRAINT fk_attr_order_line
FOREIGN KEY (order_line_id) REFERENCES customer_order_lines(id)
ON DELETE CASCADE;

-- Pre-Allocations → Order Lines (demand)
ALTER TABLE preallocation_links
ADD CONSTRAINT fk_prealloc_demand
FOREIGN KEY (demand_order_line_id) REFERENCES customer_order_lines(id)
ON DELETE SET NULL;

-- Pre-Allocations → MOs (supply)
ALTER TABLE preallocation_links
ADD CONSTRAINT fk_prealloc_mo
FOREIGN KEY (supply_mo_id) REFERENCES manufacturing_orders(id)
ON DELETE SET NULL;

-- Pre-Allocations → MOPs (supply)
ALTER TABLE preallocation_links
ADD CONSTRAINT fk_prealloc_mop
FOREIGN KEY (supply_mop_id) REFERENCES planned_manufacturing_orders(id)
ON DELETE SET NULL;

-- Delivery Headers → Order Headers
ALTER TABLE delivery_headers
ADD CONSTRAINT fk_delivery_header_order
FOREIGN KEY (order_header_id) REFERENCES customer_order_headers(id)
ON DELETE CASCADE;

-- Delivery Lines → Delivery Headers
ALTER TABLE delivery_lines
ADD CONSTRAINT fk_delivery_line_header
FOREIGN KEY (delivery_header_id) REFERENCES delivery_headers(id)
ON DELETE CASCADE;

-- Delivery Lines → Order Lines
ALTER TABLE delivery_lines
ADD CONSTRAINT fk_delivery_line_order
FOREIGN KEY (order_line_id) REFERENCES customer_order_lines(id)
ON DELETE SET NULL;
```

---

## JSONB Field Structures

### Customer Order Header - user_fields
```json
{
  "alpha": {
    "UCA1": "CustomValue1",
    "UCA2": "Region-North",
    "UCA3": null
  },
  "numeric": {
    "UDN1": 1500.50,
    "UDN2": 2000.00,
    "UDN3": null
  },
  "dates": {
    "UID1": "2024-01-15",
    "UID2": "2024-02-20",
    "UID3": null
  },
  "text": {
    "UCT1": "Special instructions for this order..."
  }
}
```

### Customer Order Line - builtin_attributes
```json
{
  "numeric": {
    "ATV1": 150.5,
    "ATV2": 200.0,
    "ATV3": null,
    "ATV4": null,
    "ATV5": null
  },
  "string": {
    "ATV6": "Red",
    "ATV7": "Large",
    "ATV8": "Matte",
    "ATV9": null,
    "ATV0": null
  }
}
```

### Customer Order Line - discounts
```json
{
  "discounts": [
    {
      "sequence": 1,
      "status": 1,
      "percentage": 5.0,
      "amount": 100.00,
      "statistics_id": "DISC-TYPE-1"
    },
    {
      "sequence": 2,
      "status": 1,
      "percentage": 2.5,
      "amount": 50.00,
      "statistics_id": "DISC-TYPE-2"
    }
  ],
  "total_percentage": 7.5,
  "total_amount": 150.00
}
```

### Planned MO - warnings
```json
{
  "messages": [
    {"code": "MSG1", "text": "Material shortage detected for ITEM-123"},
    {"code": "MSG2", "text": "Capacity overload at work center WC-10"},
    {"code": "MSG3", "text": null},
    {"code": "MSG4", "text": null}
  ]
}
```

---

## ETL Process Flow

### 1. Extract Phase

```python
# Pseudo-code for extraction
for table in [OOHEAD, OOLINE, MOATTR, MWOHED, MMOPLP, MPREAL, ODHEAD, ODLINE]:
    last_sync = get_last_sync_timestamp(table)

    query = f"""
        SELECT * FROM {table}
        WHERE deleted = 'false'
          AND LMDT >= {last_sync.date}
        ORDER BY LMDT, LMTS
        LIMIT {batch_size}
    """

    records = execute_datafabric_query(query)
    transform_and_load(records, table)
    update_sync_timestamp(table, records[-1].LMDT, records[-1].LMTS)
```

### 2. Transform Phase

```python
# Transform M3 record to target schema
def transform_co_line(m3_record):
    return {
        "company_number": m3_record.CONO,
        "order_number": m3_record.ORNO,
        "line_number": m3_record.PONR,
        "line_suffix": m3_record.POSX,
        "item_number": m3_record.ITNO,
        "ordered_qty": m3_record.ORQT,
        # ... core fields ...

        # Aggregate attributes into JSONB
        "builtin_attributes": {
            "numeric": {
                "ATV1": m3_record.ATV1,
                "ATV2": m3_record.ATV2,
                # ...
            },
            "string": {
                "ATV6": m3_record.ATV6,
                # ...
            }
        },

        "user_fields": {
            "alpha": {f"UCA{i}": getattr(m3_record, f"UCA{i}") for i in range(1,11)},
            "numeric": {f"UDN{i}": getattr(m3_record, f"UDN{i}") for i in range(1,7)},
            "dates": {f"UID{i}": getattr(m3_record, f"UID{i}") for i in range(1,4)}
        },

        # Metadata
        "m3_change_date": parse_m3_date(m3_record.LMDT),
        "m3_timestamp": m3_record.LMTS,
        "sync_timestamp": datetime.now()
    }
```

### 3. Load Phase with Foreign Key Population

```python
# Load with FK resolution
def load_co_line(transformed_record):
    # Find parent order header
    order_header = db.query(customer_order_headers).filter_by(
        company_number=transformed_record["company_number"],
        order_number=transformed_record["order_number"]
    ).first()

    if order_header:
        transformed_record["order_header_id"] = order_header.id

    # Upsert
    db.merge(customer_order_lines(**transformed_record))
    db.commit()

def load_preallocation(transformed_record):
    # Resolve demand FK
    if transformed_record["demand_category"] == '3':
        demand_line = db.query(customer_order_lines).filter_by(
            order_number=transformed_record["demand_order"],
            line_number=transformed_record["demand_line"],
            line_suffix=transformed_record["demand_suffix"]
        ).first()
        if demand_line:
            transformed_record["demand_order_line_id"] = demand_line.id

    # Resolve supply FK
    if transformed_record["supply_category"] == '2':
        mo = db.query(manufacturing_orders).filter_by(
            mo_number=transformed_record["supply_order"]
        ).first()
        if mo:
            transformed_record["supply_mo_id"] = mo.id

    elif transformed_record["supply_category"] == '5':
        mop = db.query(planned_manufacturing_orders).filter_by(
            planned_order_number=int(transformed_record["supply_order"])
        ).first()
        if mop:
            transformed_record["supply_mop_id"] = mop.id

    db.merge(preallocation_links(**transformed_record))
    db.commit()
```

---

## Materialized Views for Performance

Create materialized views for common queries:

### Supply Chain Summary View

```sql
CREATE MATERIALIZED VIEW mv_supply_chain_summary AS
SELECT
  oh.order_number,
  oh.customer_number,
  oh.order_date,
  ol.line_number,
  ol.item_number,
  ol.ordered_qty,
  ol.line_status,

  -- Pre-allocation summary
  COUNT(DISTINCT pa.id) FILTER (WHERE pa.supply_category = '2') as mo_count,
  COUNT(DISTINCT pa.id) FILTER (WHERE pa.supply_category = '5') as mop_count,
  SUM(pa.preallocated_qty) FILTER (WHERE pa.supply_category = '2') as mo_allocated_qty,
  SUM(pa.preallocated_qty) FILTER (WHERE pa.supply_category = '5') as mop_allocated_qty,

  -- Delivery summary
  COUNT(DISTINCT dl.id) as delivery_count,
  SUM(dl.delivered_qty) as total_delivered_qty,
  SUM(dl.invoiced_qty) as total_invoiced_qty,

  -- Coverage calculation
  ol.ordered_qty - COALESCE(SUM(dl.delivered_qty), 0) as remaining_to_deliver

FROM customer_order_headers oh
JOIN customer_order_lines ol ON ol.order_header_id = oh.id
LEFT JOIN preallocation_links pa ON pa.demand_order_line_id = ol.id
LEFT JOIN delivery_lines dl ON dl.order_line_id = ol.id
WHERE oh.is_deleted = FALSE
  AND ol.is_deleted = FALSE
GROUP BY oh.id, ol.id;

CREATE UNIQUE INDEX idx_mv_supply_chain ON mv_supply_chain_summary(order_number, line_number);
CREATE INDEX idx_mv_supply_chain_item ON mv_supply_chain_summary(item_number);
CREATE INDEX idx_mv_supply_chain_status ON mv_supply_chain_summary(line_status);
```

### Refresh Strategy
```sql
-- Refresh after each sync batch
REFRESH MATERIALIZED VIEW CONCURRENTLY mv_supply_chain_summary;
```

---

## Data Retention Policy

```sql
-- Archive old closed orders
CREATE TABLE customer_order_lines_archive (
    LIKE customer_order_lines INCLUDING ALL
);

-- Move to archive
INSERT INTO customer_order_lines_archive
SELECT * FROM customer_order_lines
WHERE line_status = '90'  -- Closed
  AND m3_change_date < CURRENT_DATE - INTERVAL '2 years';

DELETE FROM customer_order_lines
WHERE line_status = '90'
  AND m3_change_date < CURRENT_DATE - INTERVAL '2 years';
```

---

## Summary Statistics Table

Track sync operations:

```sql
CREATE TABLE sync_statistics (
    id BIGSERIAL PRIMARY KEY,
    table_name VARCHAR(50) NOT NULL,
    sync_start TIMESTAMP NOT NULL,
    sync_end TIMESTAMP,
    records_processed INTEGER,
    records_inserted INTEGER,
    records_updated INTEGER,
    records_deleted INTEGER,
    last_m3_change_date DATE,
    last_m3_timestamp BIGINT,
    status VARCHAR(20),
    error_message TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_sync_stats_table ON sync_statistics(table_name, sync_start DESC);
```

---

## Schema Advantages

### 1. Performance
- Strategic indexes on join columns
- JSONB for variable attributes (indexed with GIN)
- Materialized views for complex queries
- Partitioning by date for large tables (optional)

### 2. Flexibility
- JSONB handles M3's variable attribute model
- Easy to add new custom fields
- Supports tenant-specific customizations

### 3. Data Integrity
- Foreign keys ensure referential integrity
- Unique constraints prevent duplicates
- NOT NULL on critical fields

### 4. Maintainability
- Human-readable column names
- Clear relationships via FKs
- Self-documenting structure

### 5. Scalability
- Supports partitioning by date or customer
- Archive strategy for historical data
- Efficient incremental updates

---

## Storage Estimates

For a typical M3 installation:

| Records | OOHEAD | OOLINE | MOATTR | MWOHED | MMOPLP | MPREAL | ODHEAD | ODLINE | Total |
|---------|--------|--------|--------|--------|--------|--------|--------|--------|-------|
| 100K Orders | 800 MB | 1.2 GB | 100 MB | 600 MB | 400 MB | 50 MB | 300 MB | 400 MB | ~4 GB |
| 1M Orders | 8 GB | 12 GB | 1 GB | 6 GB | 4 GB | 500 MB | 3 GB | 4 GB | ~40 GB |

**With indexes**: Add 30-50% overhead
**With history**: Add 50-100% for 2-year retention

---

## Database Platform Recommendations

### PostgreSQL (Recommended)
- ✅ Excellent JSONB support with GIN indexes
- ✅ Advanced query features (CTEs, window functions)
- ✅ Mature partitioning
- ✅ Cost-effective

### SQL Server
- ✅ Good JSON support
- ✅ Enterprise features
- ✅ BI tool integration
- ❌ Higher licensing costs

### MySQL
- ⚠️ Limited JSON query capabilities
- ⚠️ Weaker analytical functions
- ✅ Simple to operate

---

## Next Steps

1. **Create Database**: Use PostgreSQL 14+ for best JSONB support
2. **Run DDL Scripts**: Execute all CREATE TABLE statements
3. **Add Foreign Keys**: Run ALTER TABLE statements
4. **Create Views**: Set up materialized views
5. **Test Queries**: Validate joins and performance
6. **Build ETL**: Implement extraction and transformation logic
7. **Schedule Sync**: Set up incremental sync jobs
8. **Monitor**: Track sync statistics and data quality

---

## Maintenance Queries

### Find Missing Foreign Key Relationships

```sql
-- Pre-allocations with missing demand orders
SELECT pa.*
FROM preallocation_links pa
WHERE pa.demand_category = '3'
  AND pa.demand_order_line_id IS NULL;

-- Pre-allocations with missing supply orders
SELECT pa.*
FROM preallocation_links pa
WHERE pa.supply_category = '2'
  AND pa.supply_mo_id IS NULL;
```

### Rebuild Foreign Keys

```sql
-- Update missing demand FKs
UPDATE preallocation_links pa
SET demand_order_line_id = ol.id
FROM customer_order_lines ol
WHERE pa.demand_category = '3'
  AND pa.demand_order = ol.order_number
  AND pa.demand_line = ol.line_number
  AND pa.demand_suffix = ol.line_suffix
  AND pa.demand_order_line_id IS NULL;

-- Update missing supply MO FKs
UPDATE preallocation_links pa
SET supply_mo_id = mo.id
FROM manufacturing_orders mo
WHERE pa.supply_category = '2'
  AND pa.supply_order = mo.mo_number
  AND pa.supply_mo_id IS NULL;
```

### Data Quality Dashboard

```sql
-- Summary of data coverage
SELECT
  (SELECT COUNT(*) FROM customer_order_headers WHERE is_deleted = FALSE) as total_orders,
  (SELECT COUNT(*) FROM customer_order_lines WHERE is_deleted = FALSE) as total_lines,
  (SELECT COUNT(*) FROM order_line_attributes WHERE is_deleted = FALSE) as total_attributes,
  (SELECT COUNT(*) FROM manufacturing_orders WHERE is_deleted = FALSE) as total_mos,
  (SELECT COUNT(*) FROM planned_manufacturing_orders WHERE is_deleted = FALSE) as total_mops,
  (SELECT COUNT(*) FROM preallocation_links WHERE is_deleted = FALSE) as total_allocations,
  (SELECT COUNT(*) FROM delivery_headers WHERE is_deleted = FALSE) as total_deliveries,
  (SELECT COUNT(*) FROM delivery_lines WHERE is_deleted = FALSE) as total_delivery_lines;
```

---

## Performance Tuning

### Partitioning for Large Tables

```sql
-- Partition order lines by order date (from header)
CREATE TABLE customer_order_lines_partitioned (
    LIKE customer_order_lines INCLUDING ALL
) PARTITION BY RANGE (m3_change_date);

-- Create partitions
CREATE TABLE customer_order_lines_2024 PARTITION OF customer_order_lines_partitioned
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');

CREATE TABLE customer_order_lines_2025 PARTITION OF customer_order_lines_partitioned
    FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');

CREATE TABLE customer_order_lines_2026 PARTITION OF customer_order_lines_partitioned
    FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');
```

### Query Optimization

```sql
-- Use covering indexes for common queries
CREATE INDEX idx_co_line_covering ON customer_order_lines(
    item_number, line_status, m3_change_date
) INCLUDE (ordered_qty, remaining_qty);

-- Partial indexes for active records
CREATE INDEX idx_co_line_active ON customer_order_lines(order_number)
WHERE is_deleted = FALSE AND line_status < '90';

-- Expression index for JSONB queries
CREATE INDEX idx_co_line_color ON customer_order_lines(
    (builtin_attributes->'string'->>'ATV6')
);
```

---

## Complete Example: Load a Single Order

```sql
-- Load order CO123456 with all dependencies
BEGIN;

-- 1. Load header
INSERT INTO customer_order_headers (...)
VALUES (...) ON CONFLICT (...) DO UPDATE ...;

-- 2. Load lines
INSERT INTO customer_order_lines (...)
SELECT ... FROM datafabric_staging.ooline
WHERE ORNO = 'CO123456';

-- 3. Update header FK in lines
UPDATE customer_order_lines ol
SET order_header_id = oh.id
FROM customer_order_headers oh
WHERE oh.order_number = ol.order_number
  AND ol.order_number = 'CO123456';

-- 4. Load attributes
INSERT INTO order_line_attributes (...)
SELECT ... FROM datafabric_staging.moattr
WHERE RIDN = 'CO123456';

-- 5. Load pre-allocations
INSERT INTO preallocation_links (...)
SELECT ... FROM datafabric_staging.mpreal
WHERE DRDN = 'CO123456';

-- 6. Update FKs in pre-allocations
UPDATE preallocation_links pa
SET demand_order_line_id = ol.id
FROM customer_order_lines ol
WHERE pa.demand_order = ol.order_number
  AND pa.demand_line = ol.line_number;

COMMIT;
```

---

## Backup and Recovery

```sql
-- Daily backup key tables
pg_dump -t customer_order_headers \
        -t customer_order_lines \
        -t preallocation_links \
        -t manufacturing_orders \
        > m3_backup_$(date +%Y%m%d).sql

-- Point-in-time recovery using change dates
SELECT * FROM customer_order_lines
WHERE m3_change_date <= '2026-01-15'
  AND (m3_change_date < '2026-01-15' OR m3_timestamp <= 123456789);
```

---

This schema provides a **production-ready foundation** for your M3 data warehouse, balancing normalization, flexibility, and performance.
