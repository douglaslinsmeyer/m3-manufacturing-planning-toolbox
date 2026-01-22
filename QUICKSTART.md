# Quick Start Guide

## Prerequisites Installed

- Docker & Docker Compose
- Go 1.21+ (for local development)
- Node.js 18+ (for local development)

## Getting Started

### 1. Start All Services with Docker

```bash
cd /Users/douglasl/Projects/manufacturing-planning-tools

# Start PostgreSQL, NATS, Backend, and Frontend
docker-compose up -d

# View logs
docker-compose logs -f

# Check service status
docker-compose ps
```

**Services will be available at:**
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- PostgreSQL: localhost:5432
- NATS: localhost:4222 (client), http://localhost:8222 (management UI)

### 2. Initialize Database

The database migrations will run automatically when the backend starts. To run them manually:

```bash
cd backend
go run cmd/server/main.go migrate
```

### 3. Test the Application

1. Open http://localhost:3000
2. You'll see the login page
3. Select **TRN** or **PRD** environment
4. Click "Sign In with M3"
5. You'll be redirected to Infor M3 SSO
6. Sign in with your M3 credentials
7. After successful login, you'll return to the dashboard

### 4. Switch Between Environments

From the dashboard:
1. Click the **TRN** or **PRD** badge in the header
2. Confirm the switch (this will log you out)
3. Select the new environment at login

## Local Development (Without Docker)

### Backend

```bash
cd backend

# Install dependencies
go mod download

# Run the server
go run cmd/server/main.go

# Or use Makefile
make run
```

### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm start
```

### Database (Local PostgreSQL)

```bash
# Create database
createdb m3_planning

# Run migrations
cd backend
go run cmd/server/main.go migrate
```

### NATS (Local)

```bash
# Using Docker
docker run -d --name nats -p 4222:4222 -p 8222:8222 nats --http_port 8222

# Or install locally with Homebrew (macOS)
brew install nats-server
nats-server --http_port 8222
```

## Stopping Services

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (clears database)
docker-compose down -v
```

## Configuration

The `.env` file is already configured with:
- ✅ TRN OAuth credentials (from Shop Floor app)
- ✅ PRD OAuth credentials (from Shop Floor app)
- ✅ Database connection string
- ✅ NATS connection string
- ✅ Session secrets

**No additional configuration needed for local development!**

## What Works Right Now

### ✅ Authentication
- Login with TRN or PRD environment selection
- OAuth 2.0 flow with Infor M3
- Session management with automatic token refresh
- Environment switching from dashboard
- Logout functionality

### ✅ Frontend Pages
- Login page with environment selector
- Dashboard with environment badge
- Navigation to all data views (placeholders)

### ✅ Backend API
- Health check: `GET /api/health`
- Auth endpoints: `/api/auth/*`
- Protected route middleware
- CORS configuration

### ✅ Database
- PostgreSQL with complete schema
- Three-table architecture (production_orders, manufacturing_orders, planned_manufacturing_orders)
- Customer orders, deliveries, and related tables
- Automatic timestamp updates

### ✅ Infrastructure
- NATS message queue ready for async jobs
- Docker Compose orchestration
- Database migrations

## What's Next

### To Implement

1. **Compass Data Fabric Client** (`backend/internal/compass/`)
   - Submit SQL queries to M3 Data Lake
   - Poll query status
   - Fetch results with pagination

2. **ION API Client** (`backend/internal/ion/`)
   - M3 REST API calls (OIS100MI, MMS001MI, etc.)
   - Token injection from session

3. **NATS Workers** (`backend/internal/workers/`)
   - Snapshot refresh job (bulk data import)
   - Analysis jobs (inconsistency detection)
   - Progress reporting to frontend

4. **Database Layer** (`backend/internal/db/`)
   - CRUD operations for all tables
   - Complex queries for analysis

5. **Analysis Engine** (`backend/internal/analysis/`)
   - Date mismatch detection
   - Missing linkage detection
   - Severity scoring

6. **Frontend Data Views**
   - Production orders table with filters
   - MO/MOP detail pages
   - Customer orders view
   - Inconsistencies dashboard

## Testing Authentication

To verify OAuth is working:

1. Start services: `docker-compose up -d`
2. Visit: http://localhost:3000
3. Select TRN environment
4. Click "Sign In with M3"
5. Check backend logs: `docker-compose logs -f backend`
6. You should see OAuth redirect and token exchange

## Troubleshooting

### Backend won't start
```bash
# Check if port 8080 is available
lsof -i :8080

# Check backend logs
docker-compose logs backend
```

### Frontend won't start
```bash
# Check if port 3000 is available
lsof -i :3000

# Check frontend logs
docker-compose logs frontend
```

### Database connection error
```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check database logs
docker-compose logs postgres

# Verify connection string in .env
cat .env | grep DATABASE_URL
```

### OAuth redirect not working
- Verify `OAUTH_REDIRECT_URI` in `.env` matches `http://localhost:8080/api/auth/callback`
- Check that TRN/PRD credentials are correct
- Ensure backend is accessible at http://localhost:8080

## Useful Commands

```bash
# Restart a specific service
docker-compose restart backend

# View logs for a specific service
docker-compose logs -f backend

# Execute command in container
docker-compose exec backend sh
docker-compose exec postgres psql -U postgres m3_planning

# Rebuild containers after code changes
docker-compose up -d --build

# Check NATS server status
curl http://localhost:8222/varz
```

## Project Status

**Phase 1: Foundation** ✅ Complete
- Project structure
- OAuth authentication
- Database schema
- Frontend scaffold
- Docker environment

**Phase 2: Data Integration** 🔄 Next
- Compass client
- ION API client
- NATS workers
- Data refresh logic

**Phase 3: Analysis** 📋 Planned
- Inconsistency detection
- Timeline visualization
- Reporting

**Phase 4: Polish** 📋 Planned
- UI enhancements
- Error handling
- Performance optimization
- Testing
