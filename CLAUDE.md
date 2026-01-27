# M3 Manufacturing Planning Tools - Development Notes

## M3 Data Fabric (Compass SQL) Important Findings

### Date Field Handling
- **CRITICAL**: All M3 date fields are stored as **INTEGER in YYYYMMDD format** (e.g., `20260303` for March 3, 2026)
- This applies to ALL date fields across all tables:
  - `STDT`, `FIDT`, `RSDT`, `REFD` (Manufacturing Orders)
  - `PLDT`, `RELD`, `MSTI`, `MFTI` (Planned Orders)
  - `DWDT`, `CODT`, `PLDT`, `FDED`, `LDED` (Customer Order Lines)
  - `LMDT` (Change date metadata - all tables)
- **PostgreSQL Schema**: Store as `INTEGER`, NOT `DATE` type
- **Conversion**: Use `TO_DATE(field::TEXT, 'YYYYMMDD')` when displaying as dates

### Deleted Field Type
- **CORRECTED**: The M3 Data Fabric documentation is **incorrect** about deleted columns
- Documentation claims: `deleted` is BOOLEAN
- **Actual format**: `deleted` is a **STRING with value `"false"` or `"true"`**
- **NOT** `'0'/'1'` as previously noted
- **Query usage**: `WHERE deleted = 'false'` (use string literal)

### Customer Order Line Status (ORST)
The ORST field indicates the status of a customer order line using a **two-digit status code**.

#### Standard Status Values
- `05` = Quotation
- `10` = Preliminary
- `22` = Reserved
- `33` = Allocated (location and lot number selected)
- `44` = Picking list printed
- `66` = Delivered
- `77` = Invoiced
- `99` = Flagged as completed, without delivery

#### Two-Position Status Logic
When status is **higher than 20**, each digit has meaning:
- **First digit**: How far a partial quantity has progressed in the **earliest stage** of order flow
- **Second digit**: How far a partial quantity has progressed in the **latest stage** of order flow

#### Individual Digit Meanings
- `2` = Quantity remains to be allocated (shown in Remaining quantity field)
- `3` = Allocated quantity exists
- `4` = Picking list for the quantity is printed
- `6` = Delivered quantity exists
- `7` = Invoiced quantity exists
- `9` = Quantity manually flagged as completed when picking list is reported

#### Examples
- **Status 33**: Only contains allocated quantity
- **Status 77**: Only contains invoiced quantity
- **Status 37**: Contains both allocated (earliest stage) AND invoiced (latest stage) quantities
- **Status 26**: Has reserved quantity (earliest) and delivered quantity (latest)

#### Query Filtering
For planning purposes, we typically want "open" order lines that are:
- Not yet delivered: `ORST < '66'`
- Or more restrictively, not yet picked: `ORST < '44'`

**Current implementation**: `ORST >= '20' AND ORST < '30'`
- Loads only **Reserved** lines (status 22 and partial statuses 23-29)
- Excludes Quotations (05) and Preliminary (10) - not confirmed for production
- Excludes Allocated (33) and beyond - already in fulfillment process

### Manufacturing Order Status (WHST)
The WHST field indicates the status of a manufacturing order using a **two-digit status code**.

#### Standard Status Values
- `10` = Preliminary
- `20` = Definitive (released)
- `40` = Component availability checked
- `50` = Work center scheduled
- `60` = Order started
- `80` = Order completed, but not fully reported
- `90` = Order ready for cost follow-up

#### Status Rules
- Status **may not be lowered**, except status 90 which can be lowered to 80 under certain circumstances
- Status progression is generally one-way (forward only)

#### Second Digit Meanings
When the **second digit is non-zero**, it indicates special conditions:
- `1` = New entry in progress
- `2` = Change in progress
- `4` = Deletion in progress
- `5` = Hold MO, short term
- `6` = Hold MO, long term
- `9` = (Only with status `20`) Parent/child MO relationship in progress - parent order created while child orders being finalized

#### Examples
- **Status 20**: Definitive/Released MO
- **Status 25**: Released MO on short-term hold
- **Status 29**: Released MO with parent/child relationship in progress
- **Status 60**: Order started (production in progress)
- **Status 90**: Completed and ready for cost accounting

**Field Reference**: WHST (WEWHST)

#### Query Filtering
For planning purposes:
- Active MOs in planning/production: `WHST >= '20' AND WHST < '90'`
- Released but not started: `WHST >= '20' AND WHST < '60'`
- Current implementation uses: `WHST <= '20'` (preliminary and released only)

### Manufacturing Order Proposal Status (MO-P Status Fields)
MO-Ps (Planned Manufacturing Orders) have **two status fields**: STAT and PSTS (ROPSTS).

#### STAT Field (MMSTAT) - General Status
Basic record status:
- `10` = Preliminary
- `20` = Definite
- `90` = Blocked/expired

#### PSTS Field (ROPSTS/WEPSTS) - Planned Order Status
Detailed planning status:
- `00` = Database error (see error conditions below)
- `<05` = Planned order, no material explosion
- `10` = Planned order
- `15` = Firmed quantity planned order (ref: parameter 030 in Planning Policy)
- `20` = Firmed planned order
- `30` = Release date is within lead time (triggers action message A1 or C3)
- `59` = Change in progress
- `60` = Released (converted to MO)
- `90` = Flagged for deletion

#### Database Error Conditions (Status 00)
Status 00 can be caused by:
- The product is missing
- The product is not definite (status not 20)
- The product has no item or warehouse information
- The product is master planned and contains a customer-order-unique configuration
- The product is a JIT item
- Alternate operation is missing in the routing for a product for which an alternate ID for the routing is entered manually when the order is created

#### Query Filtering
For active planned orders:
- **Current implementation**: `PSTS = '20'` (firmed planned orders only)
- Excludes unfirmed planned orders (10) - not yet committed
- Excludes released (60) and beyond - already converted to MOs
- Excludes deleted (90) and errors (00)

### MPREAL Pre-Allocation Table (Critical for Linking)
The MPREAL table links production orders to customer orders via pre-allocation records:

#### For Planned Manufacturing Orders (MOPs):
```sql
LEFT JOIN MPREAL mpreal
  ON mop.PLPN = CAST(mpreal.ARDN AS BIGINT)  -- ARDN is string, PLPN is bigint
  AND mpreal.AOCA = '100'  -- Acquisition Order Category = Planned MO
  AND mpreal.DOCA = '311'  -- Demand Order Category = Customer Order
  AND mpreal.deleted = 'false'
```
- `ARDN` = Planned Order Number (requires CAST to BIGINT for join)
- `DRDN` = Linked Customer Order Number
- `DRDL` = Linked CO Line Number
- `DRDX` = Linked CO Line Suffix

#### For Manufacturing Orders (MOs):
```sql
LEFT JOIN MPREAL mpreal
  ON mpreal.ARDN = mo.MFNO  -- Both are strings
  AND mpreal.AOCA = '101'  -- Acquisition Order Category = MO
  AND mpreal.DOCA = '311'  -- Demand Order Category = Customer Order
  AND mpreal.deleted = 'false'
```
- `ARDN` = MO Number (string)
- `DRDN` = Linked Customer Order Number
- `DRDL` = Linked CO Line Number

### Field Type Reference

| M3 Field | M3 Type | Compass Returns | PostgreSQL Type | Notes |
|----------|---------|-----------------|-----------------|-------|
| CONO | integer | int/float64 | INTEGER | Company number |
| PLPN | integer | int/float64 | BIGINT | Planned order number (large) |
| MFNO | string | string | VARCHAR(50) | MO number (string in M3) |
| ORNO | string | string | VARCHAR(50) | Customer order number |
| PONR | integer | int/float64 | INTEGER | Order line number |
| POSX | integer | int/float64 | INTEGER | Line suffix |
| ORQT, PPQT, MAQT | number | float64 | DECIMAL(15,6) | Quantities (high precision) |
| STDT, FIDT, PLDT | integer | int/float64 | INTEGER | Dates in YYYYMMDD format |
| LMDT | integer | int/float64 | INTEGER | Change date (YYYYMMDD) |
| LMTS | integer | int64 | BIGINT | Timestamp (large number) |
| ATNR | integer | int64 | BIGINT | Attribute number (large) |
| CFIN | integer | int64 | BIGINT | Configuration number (large) |
| deleted | boolean | **string** | - | Returns "false" or "true" string |

### Data Marshalling Best Practices

1. **Parser Helper Functions** (see `backend/internal/compass/parser.go`):
   - `getInt()` - Handles float64/int/string conversion
   - `getInt64()` - For large numbers (PLPN, ATNR, CFIN, LMTS)
   - `getFloat()` - For quantities and decimals
   - `getString()` - Safe string extraction
   - **Always** handle nil values safely

2. **Date Field Handling**:
   ```go
   // In Go structs
   STDT sql.NullInt32  // Nullable integer

   // When inserting to DB
   var stdt interface{}
   if mo.STDT.Valid && mo.STDT.Int32 != 0 {
       stdt = mo.STDT.Int32  // Insert as integer
   }

   // In SQL queries for display
   TO_DATE(NULLIF(mo.stdt, 0)::text, 'YYYYMMDD')  -- Convert to date
   ```

3. **JSONB Attributes**:
   - Store flexible/custom M3 fields in JSONB columns
   - Customer Order Lines: Built-in attributes (ATV1-ATV0), User-defined (UCA1-UCA0, UDN1-UDN6, UID1-UID3), Discounts (DIP1-DIP8, DIA1-DIA8)
   - Manufacturing Orders: Planning details, routing, material info
   - Planned Orders: Messages (MSG1-MSG4), planning parameters

### Query Performance Tips

1. **Always filter by context**:
   ```sql
   WHERE mo.CONO = '100'
     AND mo.FACI = 'AZ1'
     AND mo.deleted = 'false'
   ```

2. **Status filtering** (see status sections above for details):
   - Open CO lines: `ORST >= '20' AND ORST < '30'` (reserved only, excludes quotations/preliminary)
   - Manufacturing Orders: `WHST <= '20'` (prelim/released only)
   - MO Proposals: `PSTS = '20'` (firmed planned orders only, excludes unfirmed)

3. **Use indexed fields**:
   - `LMDT` for incremental loading (though full refresh is current strategy)
   - `RORC, RORN, RORL, RORX` for reference order linking

### Data Refresh Strategy

Current implementation uses **full table truncation** and reload:
1. Truncate all M3 snapshot tables
2. Load MOPs with CO links (filter: `PSTS = '20'` via MPREAL)
3. Load MOs with CO links (via MPREAL)
4. Load reserved CO lines (filter: `ORST >= '20' AND ORST < '30'`)
5. Update unified production_orders view

**Status Filter Rationale**:
- **MO-Ps (PSTS = '20')**: Only Firmed planned orders
  - Excludes unfirmed planned orders (10): Not committed
  - Excludes released (60): Already converted to MOs
- **CO Lines (ORST >= '20' AND < '30')**: Only Reserved lines (22, 23-29)
  - Excludes Quotations (05) and Preliminary (10): Not confirmed
  - Excludes Allocated (33) and beyond: Already in fulfillment

This avoids complexity of incremental updates and ensures data consistency.

### Known Issues & Fixes

- **Migration 006**: Fixed LMDT from DATE to INTEGER across all tables
- **Migration 009**: Fixed customer_order_lines date columns (dwdt, codt, pldt, fded, lded) from DATE to INTEGER
- Both migrations follow M3's native INTEGER date format

### Testing Queries

```sql
-- Verify deleted field format
SELECT DISTINCT deleted, COUNT(*) FROM MMOPLP GROUP BY deleted;
-- Returns: deleted='false', count=339920

-- Test MPREAL linking for MOPs
SELECT mop.PLPN, mpreal.DRDN as co_number, mpreal.DRDL as co_line
FROM MMOPLP mop
LEFT JOIN MPREAL mpreal ON mop.PLPN = CAST(mpreal.ARDN AS BIGINT)
  AND mpreal.AOCA = '100' AND mpreal.DOCA = '311'
  AND mpreal.deleted = 'false'
WHERE mop.deleted = 'false' AND mop.CONO = '100' LIMIT 5;

-- Verify date format
SELECT STDT, FIDT FROM MMOPLP WHERE STDT IS NOT NULL LIMIT 1;
-- Returns: STDT=20260303 (integer)
```

## Snapshot Refresh Architecture

### Overview

The snapshot refresh system uses a **Coordinator-Worker pattern** with NATS message queuing to load M3 data in parallel across multiple worker instances.

### Architecture Flow

```
API Handler (POST /api/data/refresh)
    ↓
    1. Create refresh job in DB
    2. Publish to NATS subject: snapshot.refresh.{ENV}
    ↓
Coordinator Worker (handleRefreshRequest)
    ↓
    Phase 0: Truncate DB tables for environment
    Phase 1: Publish 3 data jobs to NATS
             ├─ snapshot.batch.{ENV}.mops
             ├─ snapshot.batch.{ENV}.mos
             └─ snapshot.batch.{ENV}.cos
    ↓
Data Loader Workers (handleBatchJob) [3 instances]
    ↓
    Execute full data load from Compass SQL
    Publish completion → snapshot.batch.complete.{jobID}
    ↓
Coordinator (waitForDataJobs)
    ↓
    Wait for 3 completions
    Phase 3: Finalize (update production_orders view)
    Phase 4: Detection (run all issue detectors)
    ↓
    Mark job complete
    Publish progress/complete messages
```

### Key Components

**Message Types:**
- `SnapshotRefreshMessage` - Initial refresh request from API
- `DataBatchJobMessage` - Work distribution to data loaders
- `BatchCompletionMessage` - Data load completion signal
- `ProgressUpdate` - Real-time progress updates (SSE)

**NATS Subjects:**
- `snapshot.refresh.{ENV}` - Coordinator job queue
- `snapshot.batch.{ENV}.{dataType}` - Data loader work queue
- `snapshot.batch.complete.{jobID}` - Completion event stream
- `snapshot.progress.{jobID}` - Progress event stream (SSE)

**Worker Roles:**
- **Coordinator** - Orchestrates phases, waits for completions, runs finalize/detection
- **Data Loaders** - Execute parallel Compass SQL queries, insert to DB

### Design Notes

**Terminology: "Batch" vs "Data Type"**
- Current implementation uses "batch" terminology in message names
- Each "batch job" actually loads the FULL dataset for one data type (MOPs/MOs/COs)
- There is NO actual batching (ID ranges, pagination) currently implemented
- The "batch" terminology is reserved for future ID-range batching if needed

**Parallelism:**
- 3 data loads run in parallel (MOPs, MOs, COs)
- Work distributed via NATS queue groups
- 3 worker instances by default (horizontally scalable)
- Coordinator synchronously waits for all 3 to complete

**Progress Tracking:**
- Phase 0 (Truncate): 0-20%
- Phase 1-2 (Data loading): 25-70% (15% per data type completion)
- Phase 3 (Finalize): 75%
- Phase 4 (Detection): 90-100%

**Detection:**
- Currently runs synchronously in coordinator after finalize
- NOT using message queue for async detection
- Runs on single coordinator worker (potential bottleneck)
- Future: Could be made async with dedicated detector workers

### Environment Isolation

- **TRN** and **PRD** use separate NATS subjects
- Database tables have `environment` column for multi-tenancy
- Truncate operations are environment-scoped
- OAuth tokens are environment-specific

## Development Environment

- Backend: Go with PostgreSQL
- Frontend: React with TypeScript
- M3 Integration: Infor Data Fabric (Compass SQL) + M3 API
- Authentication: OAuth token-based

### Running Services

**IMPORTANT**: This project uses Docker Compose for all services.

- **To restart/rebuild backend**: `docker compose up backend --build`
  - **CRITICAL**: When rebuilding backend, also rebuild workers: `docker compose up backend backend-worker --build`
  - The backend and backend-worker containers share the same Go codebase
  - Forgetting to rebuild workers is a common mistake that causes version mismatches
- **To restart/rebuild frontend**: `docker compose up frontend --build`
- **To view logs**: `docker compose logs -f <service_name>`
- **Never use**: `go run`, `pkill`, or direct process management
- All service management must go through Docker Compose
- **Worker scaling**: 3 backend-worker instances run by default (configured via `deploy.replicas: 3`)

### Database Connection & Queries

**IMPORTANT**: Always check `docker-compose.yml` for correct database connection details before querying:

- **Container name**: `postgres` (use `docker compose exec postgres ...`)
- **Database name**: `m3_planning` (NOT `m3planning`)
- **User**: `postgres`
- **Password**: `postgres`
- **Connection string**: `postgresql://postgres:postgres@postgres:5432/m3_planning?sslmode=disable`

**Example database query**:
```bash
docker compose exec postgres psql -U postgres -d m3_planning -c "SELECT * FROM user_profiles LIMIT 1;"
```
