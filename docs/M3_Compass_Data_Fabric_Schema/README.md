# M3 Data Fabric Schema Maps

Comprehensive documentation for M3 supply chain tables from the Infor Data Fabric.

---

## Quick Start

### New to These Schema Maps?

**Start here**: [SCHEMA_MAP_SUMMARY.md](./SCHEMA_MAP_SUMMARY.md)
- Complete overview of all tables
- Relationship diagrams
- Common query patterns
- Quick reference guide

### Building a Data Warehouse?

**Go to**: [RECOMMENDED_SCHEMA_DESIGN.md](./RECOMMENDED_SCHEMA_DESIGN.md)
- Production-ready database schema (DDL)
- Complete ETL strategy
- Optimal harvesting queries
- Performance tuning guide
- JSONB structure recommendations

### Looking for a Specific Table?

Choose from these detailed schema maps:

#### Order Management
- **[CO_HEADER_SCHEMA_MAP.md](./CO_HEADER_SCHEMA_MAP.md)** - OOHEAD (163 fields)
  - Order header with customer, dates, terms
  - User-defined fields and business chains

- **[CO_LINE_SCHEMA_MAP.md](./CO_LINE_SCHEMA_MAP.md)** - OOLINE (303 fields)
  - Order line details, quantities, pricing
  - Built-in attributes (ATV) and user fields (UCA/UDN)
  - Reference order links

- **[ATTRIBUTE_SCHEMA_MAP.md](./ATTRIBUTE_SCHEMA_MAP.md)** - MOATTR (52 fields)
  - Detailed attribute specifications
  - Text and numeric value ranges
  - Attribute model integration

#### Manufacturing
- **[MO_SCHEMA_MAP.md](./MO_SCHEMA_MAP.md)** - MWOHED (149 fields)
  - Manufacturing order header
  - Multi-level hierarchy tracking
  - Status, quantities, dates

- **[MOP_SCHEMA_MAP.md](./MOP_SCHEMA_MAP.md)** - MMOPLP (91 fields)
  - Planned manufacturing orders
  - MRP action messages
  - Material shortage indicators

#### Supply Chain Linking ⭐
- **[PREALLOCATION_SCHEMA_MAP.md](./PREALLOCATION_SCHEMA_MAP.md)** - MPREAL (44 fields)
  - **CRITICAL**: Links customer orders to MOs/MOPs/POs
  - Shows allocated quantities
  - Enables supply chain visibility

#### Delivery & Fulfillment
- **[DELIVERY_HEADER_SCHEMA_MAP.md](./DELIVERY_HEADER_SCHEMA_MAP.md)** - ODHEAD (87 fields)
  - Delivery header with invoice info
  - Shipment and wave grouping

- **[DELIVERY_LINE_SCHEMA_MAP.md](./DELIVERY_LINE_SCHEMA_MAP.md)** - ODLINE (102 fields)
  - Delivery line quantities
  - Delivered vs invoiced tracking

---

## What's in Each Schema Map?

Every schema map includes:

1. **Field Listings**: All fields with human-readable names
2. **Data Types**: Type and maximum values
3. **Business Descriptions**: What each field means
4. **Categories**: Fields grouped by business function
5. **Relationships**: Foreign keys and join patterns
6. **Common Queries**: Ready-to-use SQL examples
7. **Status Values**: Decode M3 status codes
8. **Best Practices**: Tips for using each table

---

## Key Concepts

### The MPREAL Critical Linking Table

**MPREAL** is the most important table for supply chain visibility:

```
Customer Order Line (demand)
         ↓ MPREAL links
Manufacturing Order (supply)
```

Without MPREAL, you cannot definitively answer:
- "Which MO is making this customer order?"
- "How much of this order is covered by production?"
- "What customer orders is this MO fulfilling?"

**Always use MPREAL** for demand-supply queries!

### Linking Methods: RORC vs MPREAL

M3 provides two methods for linking demand to supply:

| Method | Use Case | Pros | Cons |
|--------|----------|------|------|
| **MPREAL** | Any complex scenario | Many-to-many, shows quantities | Requires extra join |
| **RORC fields** | Simple one-to-one | Direct join | One-to-one only, no quantities |

**Recommendation**: Use MPREAL for production systems.

### Attribute Flexibility

M3 attributes come in multiple forms:

1. **Built-in**: ATV1-ATV0 fields (in OOLINE)
2. **User-Defined**: UCA1-UCA0, UDN1-UDN6 (in OOLINE, OOHEAD)
3. **Detailed**: MOATTR table (flexible attribute specifications)

**Storage Strategy**: Use JSONB to handle the variation!

### Change Tracking

All tables have:
- **LMDT**: Last modification date (YYYYMMDD)
- **LMTS**: Last modification timestamp (microseconds)
- **deleted**: Deletion flag (STRING "true"/"false" - not boolean!)

**For incremental loads**: Use `WHERE deleted = 'false' AND LMDT >= [date]`

---

## Common Use Cases

### Build Supply Chain Dashboard
1. Read: SCHEMA_MAP_SUMMARY.md (understand relationships)
2. Use: PREALLOCATION_SCHEMA_MAP.md (demand-supply links)
3. Implement: Queries from MO_SCHEMA_MAP.md

### Create Data Warehouse
1. Read: RECOMMENDED_SCHEMA_DESIGN.md (complete DDL)
2. Review: Each individual schema map for field details
3. Customize: Adjust based on your business needs

### Understand M3 Data
1. Start: SCHEMA_MAP_SUMMARY.md (overview)
2. Deep dive: Individual table maps
3. Reference: M3 API programs listed in summary

---

## Table Relationships at a Glance

```
OOHEAD (Order Header)
    │
    ├─→ OOLINE (Lines)
    │      │
    │      ├─→ MOATTR (Attributes)
    │      │
    │      └─→ MPREAL (Links) ⭐
    │             │
    │             ├─→ MWOHED (MOs)
    │             └─→ MMOPLP (MOPs)
    │
    └─→ ODHEAD (Deliveries)
           │
           └─→ ODLINE (Delivery Lines)
```

---

## Data Harvesting Quick Guide

### Optimal Query Order (respects dependencies)

1. OOHEAD - Order headers
2. OOLINE - Order lines
3. MOATTR - Attributes
4. MWOHED - Manufacturing orders
5. MMOPLP - Planned MOs
6. **MPREAL - Allocations** (requires all above)
7. ODHEAD - Deliveries
8. ODLINE - Delivery lines

### Recommended Sync Frequency

- **Real-time**: ODHEAD, ODLINE (10 min)
- **Frequent**: OOHEAD, OOLINE, MWOHED, MPREAL (15 min)
- **Moderate**: MMOPLP (30 min)
- **Periodic**: MOATTR (1 hour)

### Critical Filters

```sql
-- Always filter deleted records (note: STRING comparison!)
WHERE deleted = 'false'

-- Incremental load by change date
AND LMDT >= 20260101

-- Order by change time for consistency
ORDER BY LMDT, LMTS
```

---

## Field Count Summary

| Table | Fields | Complexity | Importance |
|-------|--------|------------|------------|
| OOHEAD | 163 | Medium | High |
| OOLINE | 303 | Very High | Critical |
| MOATTR | 52 | Low | Medium |
| MWOHED | 149 | High | Critical |
| MMOPLP | 91 | Medium | High |
| **MPREAL** | **44** | **Low** | **CRITICAL** ⭐ |
| ODHEAD | 87 | Medium | High |
| ODLINE | 102 | Medium | High |

**Note**: MPREAL has the fewest fields but is the MOST CRITICAL for linking!

---

## Getting Help

### Question: "How do I find which MO is making a customer order?"
**Answer**: See [PREALLOCATION_SCHEMA_MAP.md](./PREALLOCATION_SCHEMA_MAP.md) - query pattern #1

### Question: "What database schema should I use?"
**Answer**: See [RECOMMENDED_SCHEMA_DESIGN.md](./RECOMMENDED_SCHEMA_DESIGN.md) - complete DDL included

### Question: "How do I handle all the different attribute fields?"
**Answer**: See [ATTRIBUTE_SCHEMA_MAP.md](./ATTRIBUTE_SCHEMA_MAP.md) - JSONB strategy recommended

### Question: "What's the difference between RORC and MPREAL?"
**Answer**: See [SCHEMA_MAP_SUMMARY.md](./SCHEMA_MAP_SUMMARY.md#demand-supply-linking-rorc-vs-mpreal) - section on linking methods

### Question: "What fields should I extract for basic reporting?"
**Answer**: See "Recommended Field Selection" in [SCHEMA_MAP_SUMMARY.md](./SCHEMA_MAP_SUMMARY.md#recommended-field-selection)

---

## Contact and Contributions

These schema maps are based on:
- M3 Data Catalog metadata (v1.0.0, generated 2025-08-02)
- Standard M3 configurations
- Best practices from production implementations

**Remember**: Your M3 environment may have:
- Custom fields beyond standard
- Different status code meanings
- Industry-specific configurations
- Modified workflows

Always validate against your specific M3 tenant!

---

## License

These schema maps are provided as reference documentation for the infor-mcp project.

---

**Last Updated**: 2026-01-21
**Version**: 1.3
