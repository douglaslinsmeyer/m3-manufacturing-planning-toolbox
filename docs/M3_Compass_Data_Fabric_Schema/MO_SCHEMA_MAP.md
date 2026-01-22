# Manufacturing Order (MWOHED) - Schema Map

## Table Overview
**M3 Table**: MWOHED
**Description**: Manufacturing Order Header - contains all header-level details for manufacturing orders
**Record Count**: 149 fields

---

## Core Identifiers

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CONO | integer | Company Number | Company identifier (max: 999) |
| FACI | string | Facility | Manufacturing facility code |
| MFNO | string | MO Number | Manufacturing order number (unique) |
| PRNO | string | Product Number | Product being manufactured |
| ITNO | string | Item Number | Item number (same as product for make items) |
| VANO | string | Product Variant | Product variant code |
| DIVI | string | Division | Division code within company |

---

## Status Fields

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| WHST | string | MO Status | Overall manufacturing order status |
| SLDT | integer | Status Change Date | Date status last changed (YYYYMMDD) |
| WHHS | string | Highest Operation Status | Highest status of all operations |
| HSDT | integer | Operation Status Change Date | Date operation status changed (YYYYMMDD) |
| WMST | string | Material Status | Status of material availability |
| MOHS | string | Hold Status | MO hold/release status |

**Common WHST (Status) Values:**
- 10 = Planned
- 15 = Released
- 20 = Started
- 25 = Part Reported
- 30 = Finished
- 90 = Closed

---

## Order Type and Origin

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ORTY | string | Order Type | Manufacturing order type code |
| GETP | integer | Origin | How order was created (max: 9) |

**GETP (Origin) Values:**
- 1 = Manual
- 2 = MRP Generated
- 3 = From Customer Order
- 4 = From MOP (Proposal)
- 5 = Reorder Point

---

## Quantities (Basic U/M)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| OROQ | number | Original Ordered Qty | Initial quantity ordered |
| ORQT | number | Ordered Quantity | Current ordered quantity (basic U/M) |
| RVQT | number | Received Quantity | Quantity reported to stock |
| MAQT | number | Manufactured Quantity | Quantity completed |
| BAQT | number | Yield Quantity | Expected yield after scrap |
| NBEQ | number | Balanced Quantity | Order balanced quantity |

---

## Quantities (Alternate U/M)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| OROA | number | Original Ordered Qty (Alt) | Original in alternate U/M |
| ORQA | number | Ordered Qty (Alt U/M) | Current order in alternate U/M |
| RVQA | number | Received Qty (Alt U/M) | Received in alternate U/M |
| MAQA | number | Manufactured Qty (Alt) | Manufactured in alternate U/M |
| CAQA | number | Approved Qty (Alt U/M) | Approved quantity alternate |

---

## Unit of Measure

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| MAUN | string | Manufacturing U/M | Unit of measure for manufacturing |
| COFA | number | Conversion Factor | Convert basic to alternate U/M |
| DMCF | integer | Conversion Method | Method for conversion (max: 9) |

---

## Dates (Planning)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| FSTD | integer | Original Start Date | First scheduled start (YYYYMMDD) |
| FFID | integer | Original Finish Date | First scheduled finish (YYYYMMDD) |
| STDT | integer | Start Date | Current scheduled start date |
| FIDT | integer | Finish Date | Current scheduled finish date |
| MSTI | integer | Start Time | Scheduled start time (HHMM) |
| MFTI | integer | Finish Time | Scheduled finish time (HHMM) |

---

## Dates (Actual)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RSDT | integer | Actual Start Date | Date production started (YYYYMMDD) |
| REFD | integer | Actual Finish Date | Date production finished (YYYYMMDD) |
| RPDT | integer | Reporting Date | Latest reporting date (YYYYMMDD) |

---

## Planning and Scheduling

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PRIO | integer | Priority | Order priority (1-9, lower = higher) |
| RESP | string | Responsible | Person/planner responsible |
| PLGR | string | Work Center | Primary work center/planning group |
| WCLN | string | Production Line | Manufacturing production line |
| PRDY | number | Production Days | Number of production days required |
| LEAL | number | Lead Time This Level | Lead time for this manufacturing level |
| LTRE | integer | Share of Lead Time | Percentage of total lead time (max: 999) |
| WLDE | integer | Infinite/Finite | Scheduling method (max: 9) |
| SDTB | integer | Same Date for Batches | Batch date handling (max: 9) |
| NURP | integer | Times Replanned | Number of reschedule cycles (max: 999) |

---

## Warehouse and Location

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| WHLO | string | Warehouse | Warehouse for finished goods |
| WHSL | string | Location | Specific storage location |
| BANO | string | Lot Number | Lot number for production |
| NUBA | integer | Number of Lots | Number of production lots (max: 99999) |

---

## Routing and Operations

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| AOID | string | Alternative Routing | Alternative routing ID |
| NUOP | integer | Number of Operations | Total operations in routing (max: 99999) |
| NUFO | integer | Number Finished Operations | Operations completed (max: 99999) |

---

## Material and BOM

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| BDCD | integer | Explosion Method | BOM explosion method (max: 9) |
| SCEX | integer | Subcontracting Exists | Has subcontract operations (max: 9) |
| STRT | string | Structure Type | Product structure type |
| ECVE | string | Revision Number | Engineering change revision |
| ALMA | string | Allow Alternate Material | Allow material substitutions |

---

## Product Structure Hierarchy

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PRHL | string | Product Highest Level | Top-level product in structure |
| MFHL | string | MO Number Highest Level | Top MO in multi-level production |
| PRLO | string | Product Overlying Level | Parent product one level up |
| MFLO | string | MO Number Next Level | Parent MO one level up |
| MSLO | integer | Serial Number Overlying | Serial at parent level (max: 9999) |
| LEVL | integer | Lowest Level | Position in BOM structure (max: 99) |
| LVSQ | integer | Level Sequence | Sequence within level (max: 999) |
| SCOM | integer | Structure Complexity | Complexity score (max: 999) |

---

## Reference Orders (Critical for Linking)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RORC | integer | Reference Order Category | Type of originating order (max: 9) |
| RORN | string | Reference Order Number | Originating order number |
| RORL | integer | Reference Order Line | Line in originating order (max: 999999) |
| RORX | integer | Reference Line Suffix | Line suffix (max: 999) |

**RORC Values:**
- 3 = Customer Order (most common - links to OOLINE)
- 2 = Manufacturing Order (parent MO)
- 4 = Distribution Order
- 5 = MRP Proposal

---

## Configuration

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| CFIN | integer | Configuration Number | Configuration instance (max: 9999999999) |
| ECVS | integer | Simulation Round | Configuration simulation (max: 999) |
| HDPR | string | Main Product | Main product in configuration |

---

## Attributes

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ATNR | integer | Attribute Number | Attribute model instance ID |

---

## Costing

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PCDO | integer | Costing Performed | Has costing been done (max: 9) |
| COTD | number | Cost Other Department | Cost from other departments |
| COSH | number | Costing Percentage | Percentage for partial costing |

---

## Project Management

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PROJ | string | Project Number | Project identifier |
| ELNO | string | Project Element | Project element/WBS code |

---

## Document Control

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| WODP | integer | Documents Printed | MO documents printed flag (max: 9) |
| NUC1 | integer | Put-away Cards Count | Number of put-away cards (max: 99) |
| NUC2 | integer | Material Requisitions Count | Number of pick lists (max: 99) |
| NUC3 | integer | Labor Tickets Count | Number of labor tickets (max: 99) |
| NUC4 | integer | Shop Travelers Count | Number of travelers (max: 99) |
| NUC5 | integer | Routing Cards Count | Number of routing cards (max: 99) |
| NUC6 | integer | Picking Lists Count | Number of picking lists (max: 99) |
| NUC7 | integer | Design Documents Count | Number of drawings (max: 99) |
| NUC8 | integer | Production Lot Cards | Number of lot cards (max: 99) |
| DWNO | string | Drawing Number | Engineering drawing reference |
| TXID | integer | Text Identity | Text reference ID |

---

## Picking and Material

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PLPR | integer | Picking List Round | Picking list print round (max: 9999999) |
| WOSQ | integer | Reporting Number | Sequential reporting counter |

---

## Text Lines

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| TXT1 | string | Text Line 1 | Order text line 1 |
| TXT2 | string | Text Line 2 | Order text line 2 |

---

## Scrap and Yield

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ASPC | number | Cumulative Scrap % | Accumulated scrap percentage |
| REND | integer | Manual Completion | Manual completion flag (max: 9) |

---

## Action Messages

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ACTP | string | Action Message | MRP action message code |

**Common ACTP Values:**
- RI = Reschedule In (expedite)
- RO = Reschedule Out (delay)
- CN = Cancel
- RE = Release

---

## Supply Chain Schedule

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| SCHN | integer | Schedule Number | Supply chain schedule reference |
| NNDT | integer | Alternate Planning Date | Alternative date for planning (YYYYMMDD) |

---

## Subnetwork and SWB (Shop Workload Balancing)

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| SUBN | integer | Sub-Network Mark | Part of subnetwork (max: 9) |
| SUBD | integer | Subnetwork Due Date | Subnetwork due date (YYYYMMDD) |
| CLGP | integer | Color Group | Visual grouping (max: 99) |
| NTWP | integer | External Network Priority | Network priority (max: 99) |
| TSDA | integer | SWB Timestamp Date | Shop workload balance date (YYYYMMDD) |
| TSTE | integer | SWB Timestamp Time | Shop workload balance time (HHMMSS) |
| PULD | integer | Pull-up Delayed Orders | Pull forward delayed orders (max: 9) |
| PULN | integer | Pull-up Early Orders | Pull forward early orders (max: 9) |
| PRAP | integer | SWB Processed | Processed by shop workload balance (max: 9) |
| ACHD | integer | SWB Change Date | Last SWB change date (YYYYMMDD) |
| ACHT | integer | SWB Change Time | Last SWB change time (HHMM) |

---

## Rescheduling

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| RIFD | integer | Reschedule In Filter Date | Filter for expedite messages (YYYYMMDD) |
| ROFD | integer | Reschedule Out Filter Date | Filter for delay messages (YYYYMMDD) |
| RPCH | integer | Rescheduling Check | Reschedule check status (max: 9) |

---

## Process and Special Manufacturing

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| PCCO | integer | Process Code | Process manufacturing code (max: 9) |
| MFPC | string | Process | Process identifier |
| PGTP | integer | Production Group Type | Type of production grouping (max: 9) |
| PPMG | integer | Production Lot Controlled | Has lot-controlled operations (max: 9) |

---

## Expiration and Reclassification

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| EXPI | integer | Expiration Date | MO expiration date (YYYYMMDD) |
| MEXP | integer | Manual Expiration Date | Manually set expiration (YYYYMMDD) |
| MREC | integer | Manual Reclassification Date | Manual reclass date (YYYYMMDD) |
| MRCT | integer | Manual Reclassification Time | Manual reclass time (HHMM) |

---

## Redistribution

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| FRED | integer | Redistribution Status | Redistribution flag (max: 9) |

---

## Sorting Sequences

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| SQNY | integer | Sequence Y | Sorting sequence field Y (max: 999) |
| SQNX | integer | Sequence X | Sorting sequence field X (max: 999) |
| SQNZ | integer | Sequence Z | Sorting sequence field Z (max: 999) |

---

## Item Characteristics

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| COLO | string | Color | Product color code |
| SIZE | string | Size | Product size code |
| CHCS | string | Characteristics | Additional characteristics |

---

## Schedule Reservation

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| SLRO | integer | Schedule Reservation Order | Schedule reservation flag (max: 9) |

---

## External References

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| EXRN | string | External Reference Number | External order/reference number |
| EXD2 | string | External Reference Desc | Description of external reference |

---

## Component Validation

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| ICOE | integer | Invalid Components Exist | Has invalid/expired components (max: 9) |

---

## Backflush

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| AGBF | integer | Aggregated Backflush | Use aggregated backflush (max: 9) |

---

## Order Locking

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| MOLK | integer | MO Lock | Manufacturing order locked (max: 9) |
| VEOC | integer | Verify Operation Closing | Verify before closing operations (max: 9) |

---

## Harvested Date

| M3 Field | Data Type | Human-Readable Name | Description |
|----------|-----------|---------------------|-------------|
| HVDI | integer | Harvested Date Inherited | Inherit harvest date from parent (max: 9) |

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
| Core Identifiers | 7 | Basic MO identification |
| Status | 6 | Order and operation status |
| Quantities | 11 | All quantity fields |
| Dates/Times | 12 | Planning and actual dates |
| Planning/Scheduling | 10 | Planning parameters |
| Hierarchy | 8 | Multi-level structure relationships |
| Reference Orders | 4 | **Critical for MO→CO linking** |
| Operations/Routing | 3 | Routing information |
| Material/BOM | 6 | Material and structure |
| Configuration | 3 | Product configuration |
| Costing | 3 | Cost information |
| Documents | 11 | Document counts and references |
| SWB | 11 | Shop workload balancing |
| Project | 2 | Project references |
| Audit/Metadata | 11 | Change tracking + Data Lake |
| **Total** | **149** | All fields |

---

## Key Fields for Data Modeling

### Primary Key
- CONO + FACI + MFNO

### Foreign Keys / Relationships
- **To Customer Orders**: RORC=3, RORN=ORNO, RORL=PONR, RORX=POSX
- **To Parent MO**: RORC=2, RORN=Parent MFNO (multi-level)
- **To Item/Product**: ITNO, PRNO
- **To Facility**: FACI
- **To Warehouse**: WHLO
- **To Work Center**: PLGR
- **To Project**: PROJ

### Multi-Level Hierarchy
- **Top Level**: PRHL (product), MFHL (MO number)
- **Parent Level**: PRLO (product), MFLO (MO number)
- **Current Level**: PRNO, MFNO
- **Position**: LEVL (level in structure), LVSQ (sequence)

### Critical for Incremental Load
- LMDT (Change Date)
- LMTS (Timestamp)
- deleted (must use string comparison: 'false')

### Most Commonly Used Fields
1. MFNO, PRNO, ITNO (identification)
2. WHST (status)
3. ORQT, RVQT, MAQT (quantities)
4. STDT, FIDT, RSDT, REFD (dates)
5. RORC, RORN, RORL, RORX (demand links)
6. PRHL, MFHL (top-level reference)
7. LMDT (change tracking)

---

## Status Transition Flow

```
10 (Planned) → 15 (Released) → 20 (Started) → 25 (Part Reported) → 30 (Finished) → 90 (Closed)
```

## Common Query Patterns

### Find MOs for a Customer Order Line
```sql
SELECT * FROM MWOHED
WHERE RORC = 3
  AND RORN = 'CO12345'
  AND RORL = 1
  AND deleted = 'false'
```

### Multi-Level MO Structure
```sql
-- Get full hierarchy
SELECT child.*, parent.*
FROM MWOHED child
LEFT JOIN MWOHED parent
  ON parent.MFNO = child.MFLO
  AND parent.FACI = child.FACI
WHERE child.MFHL = 'MO-TOP-LEVEL'
  AND child.deleted = 'false'
ORDER BY child.LEVL, child.LVSQ
```

### Active MOs by Status
```sql
SELECT * FROM MWOHED
WHERE WHST IN ('15', '20', '25')  -- Released, Started, Part Reported
  AND deleted = 'false'
  AND FACI = 'FAC1'
ORDER BY STDT, PRIO
```
