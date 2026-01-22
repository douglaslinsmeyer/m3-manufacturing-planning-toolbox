# Architecture Overview

## System Components

### Backend (Go)
- **Framework**: Standard library with Gorilla Mux for routing
- **Authentication**: OAuth 2.0 with Infor M3/ION
- **Session Management**: Cookie-based sessions with encrypted storage
- **Database**: PostgreSQL with native driver
- **Message Queue**: NATS for async processing
- **API Clients**:
  - Compass Data Fabric client for bulk SQL queries
  - ION API client for transactional operations

### Frontend (React + TypeScript)
- **Framework**: React 18 with TypeScript
- **Routing**: React Router v6
- **State Management**: Context API for auth, local state for data
- **API Client**: Axios with credentials support
- **Styling**: CSS modules

### Database (PostgreSQL)
Three-table architecture for production orders:

#### 1. `production_orders` (Unified Analysis View)
Lightweight table containing common fields for all production orders:
- Order number, type (MO/MOP)
- Item number, facility, warehouse
- Planned start/finish dates
- Quantity and status
- References to detail tables (mo_id, mop_id)

**Purpose**: Fast sequential analysis, unified timeline view

#### 2. `manufacturing_orders` (Full MO Details)
Complete MO information with related tables:
- `mo_operations`: Operations, work centers, times
- `mo_materials`: Material requirements and allocations
- All M3 MO-specific fields

**Purpose**: Deep dive into released MO details

#### 3. `planned_manufacturing_orders` (Full MOP Details)
Complete MOP information:
- Planning parameters (lot size, safety stock, reorder point)
- Demand order references (CO, forecast)
- Buyer/planner information
- All M3 MOP-specific fields

**Purpose**: Deep dive into planned order details

### Message Queue (NATS)
Used for asynchronous operations:

#### Subjects:
- `snapshot.refresh.{environment}` - Data refresh jobs
- `snapshot.status.{job_id}` - Job status updates
- `analysis.run.{type}` - Analysis tasks
- `analysis.results.{job_id}` - Analysis results

#### Use Cases:
1. **Bulk Data Refresh**: Query tens of thousands of records from M3
2. **Data Aggregation**: Build analysis datasets
3. **Inconsistency Detection**: Run date comparison algorithms
4. **Progress Reporting**: Real-time status updates to frontend

## Data Flow

### Authentication Flow
1. User selects environment (TRN/PRD) on login page
2. Frontend calls `/api/auth/login` with environment
3. Backend generates OAuth authorization URL
4. User redirects to Infor SSO
5. User authenticates with M3 credentials
6. Infor SSO redirects to `/api/auth/callback` with code
7. Backend exchanges code for access/refresh tokens
8. Tokens stored in encrypted session cookie
9. Frontend redirects to dashboard

### Data Refresh Flow
1. User clicks "Refresh Data" on dashboard
2. Frontend calls `/api/snapshot/refresh`
3. Backend publishes message to NATS: `snapshot.refresh.{environment}`
4. Worker subscribes to NATS and processes:
   - Query Compass Data Fabric for MOs, MOPs, COs, deliveries
   - Parse results (tens of thousands of records)
   - Store in PostgreSQL
   - Publish status updates to `snapshot.status.{job_id}`
5. Frontend polls `/api/snapshot/status` for progress
6. On completion, dashboard reloads summary data

### Analysis Flow
1. Frontend requests inconsistencies: `/api/analysis/inconsistencies`
2. Backend checks if analysis is cached
3. If not cached, publishes to NATS: `analysis.run.inconsistencies`
4. Worker performs analysis:
   - Join production_orders with customer_orders and deliveries
   - Detect date mismatches (planned dates vs delivery dates)
   - Identify missing links
   - Calculate severity scores
5. Results stored in database and returned to frontend

## API Integration

### Compass Data Fabric
**Purpose**: Bulk data extraction via SQL queries

**Workflow**:
1. Submit SQL query → GET query ID
2. Poll query status → Wait for completion
3. Fetch results with pagination

**Key Tables** (M3 Data Lake):
- Manufacturing orders
- Planned manufacturing orders
- Customer orders
- Order lines
- Deliveries

**Critical Note**: The `deleted` column is STRING `'false'`/`'true'`, NOT boolean

### ION M3 REST API
**Purpose**: Transactional queries, real-time data

**Key Programs**:
- `OIS100MI`: Customer Order Management
- `MMS001MI`: Manufacturing Order Management
- `PPS200MI`: Manufacturing Order Proposals
- `MWS150`: Active Supply Chain (links MOs to deliveries)
- `MWS410MI`: Delivery information

## User Context

Users select organizational context after login:
- **Company**: M3 company code
- **Division**: Business division
- **Facility**: Manufacturing facility
- **Warehouse**: Warehouse location

Context stored in session and used to filter all queries.

## Inconsistency Detection

### Date Mismatch Detection
Compare:
- Production order planned start date
- Production order planned finish date
- CO line requested delivery date
- CO line confirmed delivery date

**Rules**:
- High severity: MO finish > 7 days after confirmed delivery
- Medium severity: MO finish > 3 days after confirmed delivery
- Low severity: MO finish > confirmed delivery

### Missing Linkages
- Production orders without reference to CO/delivery
- CO lines without corresponding production orders
- Deliveries without linked MOs

### Quantity Mismatches
- MO quantity != CO line quantity
- Delivered quantity > ordered quantity

## Scalability Considerations

### Data Volume
- Expected: 10,000+ production orders
- Expected: 50,000+ CO lines
- Expected: 100,000+ deliveries

### Performance Strategies
1. **Unified Production Orders View**: Fast sequential scans
2. **NATS Async Processing**: Non-blocking refresh operations
3. **Indexed Queries**: Optimized for date range and facility filters
4. **Pagination**: Limit result sets to 100-1000 records per page
5. **Caching**: Cache analysis results for 5-15 minutes

### NATS Benefits
- Lightweight (minimal resource overhead)
- Fast (millions of messages/sec)
- Simple pub/sub patterns
- Built-in request/reply for RPC
- No data persistence required (jobs are transient)

## Security

### Authentication
- OAuth 2.0 with Infor M3
- Client credentials per environment (TRN/PRD)
- Token refresh with 5-minute buffer

### Session Management
- HTTP-only cookies
- Secure flag in production
- Session duration: 24 hours
- CSRF protection via SameSite cookies

### API Security
- CORS restricted to frontend origin
- Credentials required for all protected routes
- Token validation on every request

## Deployment

### Development
```bash
docker-compose up
```

### Production Considerations
- Use environment variables for secrets
- Enable HTTPS/TLS
- Configure PostgreSQL connection pooling
- Set up NATS clustering for HA
- Implement monitoring and logging
- Use reverse proxy (Nginx/Traefik)

## Future Enhancements

1. **Real-time Updates**: WebSocket for live status updates
2. **Advanced Filtering**: Complex query builder UI
3. **Batch Operations**: Bulk MO updates via M3 API
4. **Reporting**: Export inconsistencies to Excel/PDF
5. **Notifications**: Email/Slack alerts for critical issues
6. **Historical Trends**: Track inconsistencies over time
