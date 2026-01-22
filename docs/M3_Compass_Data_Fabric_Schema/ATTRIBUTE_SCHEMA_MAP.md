# Requirement Order Attribute (MOATTR) - Schema Map

## Table Overview
**M3 Table**: MOATTR
**Description**: Requirement Order Attribute File - stores detailed attribute specifications for orders
**Record Count**: 52 fields
**Alternative Name**: "1/(AH)" in M3 documentation

---

## Core Identifiers

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CONO | integer | Company Number | Company identifier (max: 999) |
| ATNR | integer | Attribute Number | Unique attribute instance ID (max: 100000000000000000) |
| ANSQ | integer | Attribute Sequence Number | Sequence within attribute set (max: 9999) |
| ATID | string | Attribute Identity | Attribute identifier/name |

---

## Order Reference (Critical for Linking)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ORCA | string | Order Category | Type of order this attribute belongs to |
| RIDN | string | Order Number | Order/requirement number |
| RIDL | integer | Order Line | Line number in order (max: 999999) |
| RIDI | integer | Delivery Number | Delivery number (max: 99999999999) |
| RIDX | integer | Line Suffix | Line suffix (max: 999) |

**ORCA (Order Category) Values:**
- Typically "3" for Customer Orders
- Links to OOLINE via RIDN=ORNO, RIDL=PONR, RIDX=POSX

---

## Item Information

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ITNO | string | Item Number | Item this attribute applies to |

---

## Attribute Values (Alpha/String)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| AALF | string | From Attribute Value (Alpha) | Range start value (text) |
| AALT | string | To Attribute Value (Alpha) | Range end value (text) |
| ATAV | string | Target Value (Alpha) | Target/specific value (text) |

---

## Attribute Values (Numeric)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ANUF | number | From Attribute Value (Numeric) | Range start value (number) |
| ANUT | number | To Attribute Value (Numeric) | Range end value (number) |
| ATAN | number | Target Value (Numeric) | Target/specific value (number) |

**Usage Pattern:**
- **Range**: Use AALF/AALT or ANUF/ANUT for min/max specifications
- **Exact**: Use ATAV or ATAN for exact match requirements

---

## Attribute Configuration

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ATMO | string | Attribute Model | Attribute model reference |
| AATT | integer | Allocation Attribute | Used for allocation logic (max: 9) |
| MAAT | integer | Main Attribute | Is main/primary attribute (max: 9) |
| PLAT | integer | Planning Attribute | Planning attribute code (max: 99) |
| CATR | integer | Costing Attribute | Costing attribute code (max: 99) |
| DIMS | integer | Dimension | Dimension identifier (max: 9) |

---

## Search and Sequencing

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| SESQ | integer | Search Sequence | Sequence for attribute search (max: 99999) |
| AVSQ | integer | Attribute Value Sequence | Sequence number for value (max: 9999) |
| DSPS | integer | Forced Sequence | Display/processing sequence (max: 99) |

---

## Status and Control

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ATTS | string | Status | Attribute status code |
| ATER | string | Error | Error code if validation failed |
| COBT | integer | Controlling Object | Controlling object type (max: 999) |

---

## Generation and Source

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| AGET | string | Generation Reference | How attribute was generated/created |

---

## Statistics

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| OSAA | integer | Order Statistic Accumulator | Statistics accumulation flag (max: 9) |
| OSAK | integer | Order Statistic Key | Statistics key field (max: 9) |

---

## Formula and Calculation

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| FMID | string | Formula | Formula identifier for calculations |
| FORI | string | Formula Result Identity | Result field from formula |

---

## Reference Attributes

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RANR | integer | Reference Attribute Number | Reference to another attribute |
| RATI | string | Reference Attribute Identity | Identity of referenced attribute |
| RASQ | integer | Reference Attribute Sequence | Sequence of referenced attribute (max: 9999) |
| RAVS | integer | Reference Attribute Value Seq | Value sequence of reference (max: 9999) |

**Usage**: Links related attributes together (e.g., Color → Size dependency)

---

## Text and Attachment

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| TXID | integer | Text Identity | Text reference for attribute |
| ATCI | integer | Attachment Indicator | Has attachment flag (max: 9) |

---

## Special Flags

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| SHIP | integer | Ship Less Indicator | Allow ship less than spec (max: 9) |
| QLCT | integer | Quality Controlled | Requires quality control (max: 9) |

---

## M3 Audit Fields

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RGDT | integer | Entry Date | Record creation date (YYYYMMDD) |
| RGTM | integer | Entry Time | Record creation time (HHMMSS) |
| LMDT | integer | Change Date | Last modification date (YYYYMMDD) |
| CHNO | integer | Change Number | Sequential change counter (max: 999) |
| CHID | string | Changed By | User who last modified |
| LMTS | integer | Timestamp | Last modification timestamp (microseconds) |

---

## Data Lake Metadata

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| accountingEntity | string | Accounting Entity | Record accounting entity |
| variationNumber | integer | Variation Number | Record modification sequence |
| timestamp | string (datetime) | Modification Timestamp | Record modification time (ISO format) |
| deleted | boolean | Is Deleted | Record deletion flag (STRING: "true"/"false") |
| archived | boolean | Is Archived | Record archive flag |

---

## Field Count by Category

| Category | Count | Description |
|----------|-------|-------------|
| Core Identifiers | 4 | Attribute identification |
| Order Reference | 5 | **Critical for linking to orders** |
| Item | 1 | Item reference |
| Alpha Values | 3 | Text-based attribute values |
| Numeric Values | 3 | Number-based attribute values |
| Configuration | 6 | Attribute model and type |
| Search/Sequence | 3 | Ordering and search |
| Status | 3 | Status and control |
| Formula | 2 | Calculation support |
| References | 4 | Inter-attribute relationships |
| Text/Attachment | 2 | Documentation |
| Special Flags | 2 | Business rules |
| Audit/Metadata | 11 | Change tracking + Data Lake |
| **Total** | **52** | All fields |

---

## Key Fields for Data Modeling

### Primary Key
- CONO + ATNR + ANSQ

### Foreign Keys / Relationships
- **To Order Lines**: ORCA=3, RIDN=ORNO, RIDL=PONR, RIDX=POSX (links to OOLINE)
- **To Attribute Model**: ATMO
- **To Item**: ITNO
- **To Reference Attribute**: RANR (self-referencing)

### Critical for Incremental Load
- LMDT (Change Date)
- LMTS (Timestamp)
- deleted (must use string comparison: 'false')

---

## Attribute Patterns

### 1. Exact Value Match
```
Color attribute:
ATID = "COLOR"
ATAV = "RED"         (exact match required)
ATAN = null
```

### 2. Numeric Range
```
Diameter specification:
ATID = "DIAMETER"
ANUF = 10.5          (minimum)
ANUT = 12.5          (maximum)
ATAN = 11.5          (target)
```

### 3. Text Range
```
Size specification:
ATID = "SIZE"
AALF = "M"           (minimum)
AALT = "XL"          (maximum)
ATAV = "L"           (preferred)
```

### 4. Multiple Attributes per Order Line
```
Order Line: ORNO="CO123", PONR=1
  ├─ Attribute 1: ATID="COLOR", ANSQ=1
  ├─ Attribute 2: ATID="SIZE", ANSQ=2
  └─ Attribute 3: ATID="FINISH", ANSQ=3
```

---

## Common Query Patterns

### Get All Attributes for an Order Line
```sql
SELECT
  ATID as attribute_name,
  ANSQ as sequence,
  ATAV as target_value_text,
  ATAN as target_value_numeric,
  AALF as from_value_text,
  AALT as to_value_text,
  ANUF as from_value_numeric,
  ANUT as to_value_numeric,
  ATTS as status
FROM MOATTR
WHERE deleted = 'false'
  AND ORCA = '3'           -- Customer order
  AND RIDN = 'CO123456'    -- Order number
  AND RIDL = 1             -- Line number
  AND RIDX = 0             -- Line suffix
ORDER BY ANSQ
```

### Join Order Lines with Attributes
```sql
SELECT
  ol.ORNO,
  ol.PONR,
  ol.ITNO,
  ol.ITDS,
  ol.ATNR,
  attr.ATID,
  attr.ATAV,
  attr.ATAN
FROM OOLINE ol
LEFT JOIN MOATTR attr
  ON attr.ORCA = '3'
  AND attr.RIDN = ol.ORNO
  AND attr.RIDL = ol.PONR
  AND attr.RIDX = ol.POSX
  AND attr.ATNR = ol.ATNR
  AND attr.deleted = 'false'
WHERE ol.deleted = 'false'
  AND ol.ORNO = 'CO123456'
ORDER BY ol.PONR, attr.ANSQ
```

### Find Orders with Specific Attribute Values
```sql
SELECT DISTINCT
  attr.RIDN as order_number,
  attr.RIDL as line_number,
  ol.ITNO,
  attr.ATID,
  attr.ATAV
FROM MOATTR attr
JOIN OOLINE ol
  ON ol.ORNO = attr.RIDN
  AND ol.PONR = attr.RIDL
  AND ol.POSX = attr.RIDX
  AND ol.deleted = 'false'
WHERE attr.deleted = 'false'
  AND attr.ORCA = '3'
  AND attr.ATID = 'COLOR'
  AND attr.ATAV = 'RED'
ORDER BY attr.RIDN, attr.RIDL
```

### Get Attribute Model Details
```sql
SELECT
  ATNR,
  ATMO as attribute_model,
  COUNT(*) as attribute_count,
  STRING_AGG(ATID, ', ') as attributes
FROM MOATTR
WHERE deleted = 'false'
  AND ORCA = '3'
GROUP BY ATNR, ATMO
ORDER BY attribute_count DESC
```

---

## Attribute Value Logic

### Priority of Value Fields

When reading attribute values, check in this order:

1. **ATAV / ATAN**: Exact target value (most specific)
2. **AALF+AALT / ANUF+ANUT**: Range specification (min/max)
3. **Formula Result (FORI)**: Calculated value

### Data Type Selection

- Use **ATAV, AALF, AALT** for:
  - Colors, sizes, descriptions
  - Codes and identifiers
  - Text-based specifications

- Use **ATAN, ANUF, ANUT** for:
  - Measurements (length, weight, diameter)
  - Quantities
  - Percentages
  - Numeric tolerances

---

## Attribute Relationships

### Parent-Child Attribute References

Attributes can reference each other via RANR/RATI/RASQ:

```
Attribute 1 (Parent):
  ANSQ = 1
  ATID = "COLOR"
  ATAV = "RED"

Attribute 2 (Child - depends on Attribute 1):
  ANSQ = 2
  ATID = "SHADE"
  ATAV = "CRIMSON"
  RANR = [ATNR of parent]
  RASQ = 1              (references parent sequence)
```

**Use Case**: "If Color=Red, then Shade must be Crimson, Scarlet, or Ruby"

---

## Attribute Model Integration

The ATMO field links to an attribute model (MAMOLI - Attribute Model Lines):

```
Attribute Model "PAINT-SPEC"
  ├─ COLOR (required)
  ├─ FINISH (optional)
  ├─ GLOSS-LEVEL (calculated)
  └─ DRY-TIME (range)
```

When ATNR is created, it instantiates all attributes from the model.

---

## Search Sequence (SESQ)

The SESQ field controls the order attributes are evaluated during:
- Inventory allocation
- Lot selection
- Material matching

Lower numbers = higher priority in search.

---

## Planning and Costing Attributes

### PLAT (Planning Attribute)
- Controls how MRP uses the attribute
- Affects supply/demand matching
- Can trigger specific planning logic

### CATR (Costing Attribute)
- Indicates attribute affects costing
- Used in product costing calculations
- May adjust standard costs

### DIMS (Dimension)
- Represents physical dimension
- Used in volume/space calculations
- Warehouse layout planning

---

## Quality Control Integration

### QLCT (Quality Controlled)
When set, this attribute requires:
- Quality inspection
- Specification validation
- Test results before acceptance

Common for:
- Chemical compositions
- Physical properties
- Performance characteristics

---

## Flexible Shipping (SHIP)

The SHIP indicator controls "ship-less" logic:

```
Specification: 100 units ± 5%
ANUF = 95
ANUT = 105
SHIP = 1      (allow partial shipment within range)
```

If SHIP=1, can ship 97 units and close the line.

---

## Formula-Based Attributes

When FMID is populated:

```
Attribute: TOTAL-WEIGHT
FMID = "CALC-WEIGHT"
FORI = "WT"
```

The formula calculates the value dynamically based on other attributes or order data.

---

## Status Values (ATTS)

Common status codes:
- **10** = Active/Valid
- **20** = Validated
- **30** = Reserved
- **90** = Inactive

Check ATER for validation errors.

---

## Incremental Load Strategy

```sql
SELECT *
FROM MOATTR
WHERE deleted = 'false'
  AND LMDT >= 20260101  -- Change date filter
  AND ORCA = '3'        -- Customer orders only
ORDER BY LMDT, LMTS
```

---

## JSON Storage Recommendation

For flexible schema design, store attributes as JSON:

```json
{
  "attribute_number": 123456789,
  "model": "PAINT-SPEC",
  "attributes": [
    {
      "sequence": 1,
      "id": "COLOR",
      "value_text": "RED",
      "status": "10"
    },
    {
      "sequence": 2,
      "id": "DIAMETER",
      "value_numeric": 11.5,
      "range_from": 10.5,
      "range_to": 12.5,
      "status": "10"
    }
  ]
}
```

---

## Normalized Schema Alternative

For relational database:

```sql
CREATE TABLE order_line_attributes (
    id BIGSERIAL PRIMARY KEY,

    -- Order reference
    order_category VARCHAR(10),
    order_number VARCHAR(20),
    order_line INTEGER,
    line_suffix INTEGER,

    -- Attribute identification
    attribute_number BIGINT,
    attribute_sequence INTEGER,
    attribute_id VARCHAR(20),
    attribute_model VARCHAR(20),

    -- Values (store all, use based on type)
    value_text VARCHAR(100),
    value_numeric NUMERIC(15,6),
    range_from_text VARCHAR(100),
    range_to_text VARCHAR(100),
    range_from_numeric NUMERIC(15,6),
    range_to_numeric NUMERIC(15,6),

    -- Configuration
    is_main_attribute BOOLEAN,
    planning_attribute INTEGER,
    costing_attribute INTEGER,
    search_sequence INTEGER,

    -- Status
    status VARCHAR(10),
    error_code VARCHAR(10),

    -- Reference attributes
    reference_attribute_number BIGINT,

    -- Metadata
    lmdt DATE,
    lmts BIGINT,
    deleted BOOLEAN,

    UNIQUE(attribute_number, attribute_sequence)
);

CREATE INDEX idx_order_attrs ON order_line_attributes(
    order_category, order_number, order_line, line_suffix
);
CREATE INDEX idx_attr_id ON order_line_attributes(attribute_id);
CREATE INDEX idx_attr_value ON order_line_attributes(value_text, value_numeric);
```

---

## Attribute Aggregation Query

Get all attributes for an order in structured format:

```sql
SELECT
  RIDN as order_number,
  RIDL as line_number,
  ATNR as attribute_number,
  ATMO as model,
  JSON_AGG(
    JSON_BUILD_OBJECT(
      'sequence', ANSQ,
      'id', ATID,
      'value_text', ATAV,
      'value_numeric', ATAN,
      'from_text', AALF,
      'to_text', AALT,
      'from_numeric', ANUF,
      'to_numeric', ANUT,
      'status', ATTS
    ) ORDER BY ANSQ
  ) as attributes
FROM MOATTR
WHERE deleted = 'false'
  AND ORCA = '3'
  AND RIDN = 'CO123456'
GROUP BY RIDN, RIDL, ATNR, ATMO
ORDER BY RIDL
```

---

## Performance Considerations

### Recommended Indexes

```sql
-- Primary key
CREATE INDEX idx_moattr_pk ON moattr(CONO, ATNR, ANSQ);

-- Order reference (most important!)
CREATE INDEX idx_moattr_order ON moattr(ORCA, RIDN, RIDL, RIDX);

-- Attribute lookup
CREATE INDEX idx_moattr_atid ON moattr(ATID);

-- Value search (for finding orders by attribute)
CREATE INDEX idx_moattr_values ON moattr(ATID, ATAV, ATAN);

-- Change tracking
CREATE INDEX idx_moattr_lmdt ON moattr(LMDT);

-- Attribute number lookup
CREATE INDEX idx_moattr_atnr ON moattr(ATNR);
```

---

## Common Attribute Examples

### Manufacturing
- **DIAMETER**: Shaft diameter in mm
- **LENGTH**: Part length in cm
- **MATERIAL**: Material grade/type
- **FINISH**: Surface finish specification
- **TOLERANCE**: Manufacturing tolerance

### Food & Beverage
- **FLAVOR**: Product flavor
- **SIZE**: Package size
- **EXPIRY**: Shelf life requirement
- **BATCH**: Batch code specification
- **ORGANIC**: Organic certification

### Apparel
- **COLOR**: Product color
- **SIZE**: Clothing size
- **FABRIC**: Fabric type/composition
- **FIT**: Fit style (slim, regular, loose)
- **SEASON**: Season collection

### Electronics
- **VOLTAGE**: Operating voltage
- **FREQUENCY**: Operating frequency
- **CAPACITY**: Storage/memory capacity
- **INTERFACE**: Connection interface
- **WARRANTY**: Warranty period

---

## Best Practices

1. **Always link via ATNR**: The ATNR field in OOLINE links to MOATTR
2. **Check both text and numeric**: An attribute may use either ATAV or ATAN
3. **Respect sequence (ANSQ)**: Display/process in sequence order
4. **Use attribute models**: Don't create ad-hoc attributes
5. **Validate ranges**: Check both FROM and TO values for ranges
6. **Consider formulas**: Check FMID before manual calculation
7. **Document models**: Maintain attribute model documentation

---

## Integration with Order Lines

When querying order lines with attributes:

```sql
-- Get order with all attribute details
SELECT
  ol.ORNO,
  ol.PONR,
  ol.ITNO,
  ol.ITDS,
  ol.ORQT,
  ol.ATNR,
  -- Aggregate attributes into JSON
  (
    SELECT JSON_AGG(
      JSON_BUILD_OBJECT(
        'id', attr.ATID,
        'sequence', attr.ANSQ,
        'value', COALESCE(attr.ATAV, attr.ATAN::text),
        'from', COALESCE(attr.AALF, attr.ANUF::text),
        'to', COALESCE(attr.AALT, attr.ANUT::text)
      ) ORDER BY attr.ANSQ
    )
    FROM MOATTR attr
    WHERE attr.ORCA = '3'
      AND attr.RIDN = ol.ORNO
      AND attr.RIDL = ol.PONR
      AND attr.RIDX = ol.POSX
      AND attr.deleted = 'false'
  ) as attributes
FROM OOLINE ol
WHERE ol.deleted = 'false'
  AND ol.ORNO = 'CO123456'
ORDER BY ol.PONR
```
