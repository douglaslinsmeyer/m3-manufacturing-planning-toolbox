# M3 Manufacturing Planning Tools

A web application for analyzing M3 CloudSuite manufacturing orders and planned manufacturing orders to identify inconsistencies, errors, and planning issues.

## Overview

This tool provides snapshot-based analysis of M3 data, identifying:
- Production orders (MOs and MOPs) planned outside reasonable timeframes from delivery dates
- Inconsistencies between production planning and CO delivery commitments
- Planning errors and anomalies requiring resolution

## Architecture

- **Backend**: Go REST API with OAuth 2.0 authentication
- **Frontend**: React with TypeScript
- **Database**: PostgreSQL with unified production order view
- **Data Sources**:
  - Compass Data Fabric (SQL queries for bulk data)
  - ION M3 REST API (transactional queries)

### Data Model

Three-table architecture for production orders:
- **`production_orders`**: Unified view for analysis (common fields only)
- **`manufacturing_orders`**: Full MO details with operations, materials, actuals
- **`planned_manufacturing_orders`**: Full MOP details with planning parameters

## Prerequisites

- Go 1.21+
- Node.js 18+
- PostgreSQL 15+
- Docker & Docker Compose (optional)

## Environment Configuration

This application uses the **same OAuth credentials as the M3 Shop Floor app**, enabling seamless integration with existing M3 authentication.

The `.env` file has already been created with the correct credentials. If you need to regenerate it:

```bash
cp .env.template .env
# Then update with credentials from Shop Floor app
```

### Environment Switching

Users can switch between TRN and PRD environments:

1. **At Login**: Select TRN or PRD before signing in
2. **From Dashboard**: Click the environment badge (TRN/PRD) in the header to switch
   - Switching environments logs you out and clears all cached data
   - You'll be redirected to login with the new environment

### Configuration Details

#### TRN (Training) Environment
- Testing and development environment
- Pre-configured with Shop Floor app credentials

#### PRD (Production) Environment
- Production environment
- Pre-configured with Shop Floor app credentials

#### Database
- `DATABASE_URL`: PostgreSQL connection string
- Default: `postgresql://postgres:postgres@localhost:5432/m3_planning`

#### NATS Message Queue
- `NATS_URL`: NATS server connection string
- Default: `nats://localhost:4222`

## Quick Start

### Using Docker Compose

#### Production Build (Optimized)

```bash
# Start all services with production frontend (nginx, ~92MB)
docker compose up -d

# Rebuild after code changes
docker compose up --build -d

# View logs
docker compose logs -f

# Stop services
docker compose down
```

#### Development Build (with Hot Reload)

```bash
# Start with development frontend (Vite dev server, ~400MB, supports HMR)
docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d

# View logs
docker compose -f docker-compose.yml -f docker-compose.dev.yml logs -f frontend

# Stop services
docker compose -f docker-compose.yml -f docker-compose.dev.yml down
```

**Frontend Build Comparison:**
- **Production**: nginx serving static files (~92MB, optimized for deployment)
- **Development**: Vite dev server with HMR (~400MB, optimized for development)

The application will be available at:
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080

### Manual Setup

#### Backend

```bash
cd backend
go mod download
go run cmd/server/main.go
```

#### Frontend

```bash
cd frontend
npm install
npm start
```

#### Database

```bash
# Create database
createdb m3_planning

# Run migrations
cd backend
go run cmd/server/main.go migrate
```

## Project Structure

```
manufacturing-planning-tools/
├── backend/               # Go backend API
│   ├── cmd/              # Application entry points
│   ├── internal/         # Internal packages
│   │   ├── api/         # HTTP handlers and routes
│   │   ├── auth/        # OAuth 2.0 authentication
│   │   ├── compass/     # Data Fabric client
│   │   ├── ion/         # ION API client
│   │   ├── db/          # Database models and queries
│   │   ├── analysis/    # Business logic for analysis
│   │   └── config/      # Configuration management
│   └── migrations/      # Database migrations
├── frontend/             # React frontend
│   ├── src/
│   │   ├── components/  # React components
│   │   ├── pages/       # Page components
│   │   ├── services/    # API clients
│   │   └── types/       # TypeScript types
└── docker-compose.yml   # Docker services
```

## Development

### Backend Development

```bash
cd backend

# Run tests
go test ./...

# Run with hot reload (requires air)
air

# Format code
go fmt ./...
```

### Frontend Development

```bash
cd frontend

# Run tests
npm test

# Build for production
npm run build

# Lint
npm run lint
```

## API Documentation

### Authentication Endpoints

- `POST /api/auth/login` - Initiate OAuth login
- `GET /api/auth/callback` - OAuth callback handler
- `POST /api/auth/logout` - Logout and clear session
- `GET /api/auth/status` - Check authentication status

### Data Management

- `POST /api/snapshot/refresh` - Refresh data snapshot
- `GET /api/snapshot/status` - Get snapshot status
- `GET /api/snapshot/summary` - Get snapshot summary

### Analysis Endpoints

- `GET /api/production-orders` - List all production orders (unified MO/MOP view)
- `GET /api/manufacturing-orders/:id` - Get full MO details
- `GET /api/planned-orders/:id` - Get full MOP details
- `GET /api/customer-orders` - List customer orders
- `GET /api/deliveries` - List deliveries
- `GET /api/analysis/inconsistencies` - List detected inconsistencies

## Data Model

### Core Entities

- **Production Orders**: Unified view of MOs and MOPs for analysis
- **Manufacturing Orders (MOs)**: Released production orders with operations
- **Planned Manufacturing Orders (MOPs)**: Proposed production orders
- **Customer Orders (COs)**: Sales orders from customers
- **CO Lines**: Individual line items on customer orders
- **Deliveries**: Delivery schedules linked to CO lines

### Inconsistency Detection

The system analyzes:
1. Production order planned dates vs. CO delivery dates
2. Production order planned dates vs. CO confirmed delivery dates
3. MOP timing relative to demand dates
4. Missing linkages between production orders and deliveries
5. Overlapping or conflicting schedules

## License

Proprietary - Internal Use Only
