# Project Ascend MVP Implementation Plan
## 18-Week Development Roadmap (Athlete-Focused EMEA Launch)

**Version:** 1.0
**Date:** October 6, 2025
**Status:** Approved - Ready for Implementation
**Related Documents:** [PRD v2.0](./prd.md), [Technical Design Document v1.0](./technical-design.md)

---

## Executive Summary

This implementation plan provides a week-by-week roadmap for building Project Ascend MVP - an athlete-focused weightlifting performance tracking platform launching in EMEA (Q1 2026). The plan covers 18 weeks of development across backend (Go + Gin + GORM), mobile (React Native + TypeScript), and infrastructure (AWS ECS Fargate).

### Key Milestones
- **Week 4:** Authentication and 1RM tracking functional
- **Week 8:** Complete offline-first session logging working
- **Week 11:** Analytics dashboard with ACWR calculation
- **Week 15:** Video upload, processing, and analysis complete
- **Week 18:** Beta launch to 30 athletes, App Store submission

### Team Requirements
- 1x Backend Developer (Go)
- 1x Mobile Developer (React Native)
- 0.5x DevOps Engineer (can be shared resource)

### Technology Stack
- **Backend:** Go 1.22+, Gin 1.10+, GORM 1.25+, PostgreSQL 15+
- **Mobile:** React Native 0.74+, TypeScript 5.3+, WatermelonDB
- **Infrastructure:** AWS ECS Fargate, RDS, S3, Lambda, CloudFront
- **Region:** AWS eu-west-1 (GDPR compliance)

---

## Phase 1: Foundation (Weeks 1-4)

### Week 1: Project Setup & Infrastructure

#### Backend Development
**Objective:** Set up Go project structure and core infrastructure

**Tasks:**
1. Initialize Go 1.22+ project with proper module structure
   ```bash
   mkdir -p ascend-api/{cmd/api,internal/{config,middleware,domain,repository,service,handler,dto,util,worker},pkg/{logger,errors,response},migrations,tests}
   cd ascend-api && go mod init github.com/ascend/api
   ```

2. Install core dependencies
   ```bash
   go get github.com/gin-gonic/gin@v1.10.0
   go get gorm.io/gorm@v1.25.7
   go get gorm.io/driver/postgres@v1.5.7
   go get github.com/jackc/pgx/v5@v5.5.5
   go get github.com/rs/zerolog@v1.32.0
   go get github.com/joho/godotenv@v1.5.1
   ```

3. Create application entry point (`cmd/api/main.go`)
   - Initialize Gin router
   - Configure zerolog structured logging
   - Load environment variables from .env
   - Add graceful shutdown handling

4. Set up base middleware
   - CORS middleware with EMEA-specific origins
   - Request ID middleware (for tracing)
   - Logging middleware (request/response logging)
   - Error recovery middleware
   - Request timeout middleware (30 seconds)

5. Create health check endpoint (`GET /health`)
   - Return 200 OK with service status
   - Include database connectivity check
   - Include Redis connectivity check (if applicable)

**AWS Infrastructure Setup:**
6. Create AWS account (if not exists) and set up billing alerts
7. Create VPC in eu-west-1 with public/private subnets
8. Set up security groups:
   - RDS security group (allow 5432 from ECS)
   - ECS security group (allow 8080 from ALB)
   - ALB security group (allow 443 from internet)

9. Provision RDS PostgreSQL 15 instance
   - Instance type: db.t4g.micro (MVP scale)
   - Multi-AZ deployment
   - Automated backups enabled (30-day retention)
   - Region: eu-west-1 (Ireland)
   - Database name: ascend_db

**Deliverables:**
- [ ] Go project structure created with all folders
- [ ] Basic Gin server running on localhost:8080
- [ ] Health check endpoint responding
- [ ] RDS PostgreSQL instance provisioned and accessible
- [ ] Logging to console with structured JSON

**Time Estimate:** 3 days

---

#### Mobile Development
**Objective:** Initialize React Native project with core navigation

**Tasks:**
1. Initialize React Native 0.74 project with TypeScript
   ```bash
   npx react-native@latest init AscendMobile --template react-native-template-typescript
   cd AscendMobile
   ```

2. Install core dependencies
   ```bash
   npm install @react-navigation/native@^6.1.9 @react-navigation/bottom-tabs@^6.5.11 @react-navigation/stack@^6.3.20
   npm install @reduxjs/toolkit@^2.0.1 react-redux@^9.0.4
   npm install @nozbe/watermelondb@^0.27.1
   npm install react-native-screens react-native-safe-area-context
   npm install axios@^1.6.2 date-fns@^3.0.6
   ```

3. Configure React Navigation
   - Set up NavigationContainer
   - Create Stack and Bottom Tab navigators
   - Define navigation types (TypeScript)
   - Create placeholder screens (Login, Register, Sessions, Analytics, Videos, Profile)

4. Set up Redux Toolkit store
   - Configure store with middleware
   - Create placeholder slices (auth, sessions, analytics, videos, sync)
   - Set up typed hooks (useAppDispatch, useAppSelector)

5. Initialize WatermelonDB
   - Create database schema file
   - Define model classes (Session, ExerciseSet, OneRepMax, Video)
   - Set up database initialization

6. Create app shell
   - Bottom tab navigation with 4 tabs
   - Tab icons and labels
   - Basic screen components (empty states)
   - Add React Native safe area handling

**Deliverables:**
- [ ] React Native app building for iOS and Android
- [ ] Navigation working between placeholder screens
- [ ] Redux store configured
- [ ] WatermelonDB initialized
- [ ] App launching on simulator/emulator

**Time Estimate:** 3 days

---

#### DevOps Setup
**Objective:** Set up CI/CD pipeline and containerization

**Tasks:**
1. Create GitHub repository
   - Initialize with README, .gitignore
   - Set up branch protection rules (require PR reviews)
   - Create branches: main, develop

2. Set up GitHub Actions workflows
   - Backend: Go linting (golangci-lint), testing, Docker build
   - Mobile: ESLint, TypeScript checks, Jest tests
   - Create `.github/workflows/backend-ci.yml` and `mobile-ci.yml`

3. Create Dockerfile for Go API
   - Multi-stage build (builder + runtime)
   - Use alpine:3.19 as base for small image size
   - Copy binary and migrations
   - Set up non-root user
   - Expose port 8080

4. Set up Docker Compose for local development
   - PostgreSQL container
   - Redis container (for future use)
   - API container
   - Network configuration

5. Create environment configuration files
   - `.env.example` for backend
   - `.env.development`, `.env.staging`, `.env.production`

**Deliverables:**
- [ ] GitHub repository created with branch protection
- [ ] CI pipeline running on every PR
- [ ] Dockerfile building successfully
- [ ] Docker Compose stack running locally

**Time Estimate:** 2 days

---

### Week 2: Database Schema & Migrations

#### Backend Development
**Objective:** Design and implement complete database schema

**Tasks:**
1. Create GORM models for core entities
   - `internal/domain/athlete.go`: Athlete entity
   - `internal/domain/session.go`: Session entity
   - `internal/domain/exercise_set.go`: ExerciseSet entity
   - `internal/domain/one_rep_max.go`: OneRepMax entity
   - `internal/domain/video.go`: Video entity

2. Write database migration files
   - Create `migrations/001_create_athletes.sql`
   - Create `migrations/002_create_one_rep_maxes.sql`
   - Create `migrations/003_create_sessions.sql`
   - Create `migrations/004_create_exercise_sets.sql`
   - Create `migrations/005_create_videos.sql`
   - Create `migrations/006_create_refresh_tokens.sql`
   - Create `migrations/007_create_audit_logs.sql`

3. Add indexes for query optimization
   - Index on `athletes.email` (WHERE deleted_at IS NULL)
   - Index on `sessions.athlete_id, date DESC`
   - Index on `exercise_sets.session_id, set_number`
   - Index on `one_rep_maxes.athlete_id, exercise, date DESC`
   - Index on `videos.athlete_id, uploaded_at DESC`

4. Implement repository pattern
   - Create `internal/repository/athlete_repository.go`
   - Create `internal/repository/session_repository.go`
   - Create `internal/repository/one_rep_max_repository.go`
   - Define repository interfaces
   - Implement GORM repository methods (CRUD operations)

5. Write repository unit tests
   - Use testify for assertions
   - Set up test database connection
   - Test create, read, update, delete operations
   - Test query methods with filters

**Deliverables:**
- [ ] All GORM models defined with proper tags
- [ ] Database migrations create all tables
- [ ] Indexes created for performance
- [ ] Repository pattern implemented
- [ ] Repository tests passing (>90% coverage)

**Time Estimate:** 4 days

---

#### Mobile Development
**Objective:** Define WatermelonDB schema and models

**Tasks:**
1. Define WatermelonDB schema
   - Create `src/services/database/schema.ts`
   - Define tables: sessions, exercise_sets, one_rep_maxes, videos
   - Specify column types and constraints
   - Add indexes for frequently queried fields

2. Create WatermelonDB model classes
   - `src/services/database/models/Session.ts`
   - `src/services/database/models/ExerciseSet.ts`
   - `src/services/database/models/OneRepMax.ts`
   - `src/services/database/models/Video.ts`
   - Add decorators (@field, @date, @children, @relation)

3. Set up database initialization
   - Create database adapter (SQLite)
   - Initialize database with schema
   - Add database provider context

4. Create database service layer
   - `src/services/database/sessionService.ts`
   - Implement CRUD operations for sessions
   - Add query methods (fetch by date range, by athlete)
   - Handle relationships (session with sets)

5. Write database service tests
   - Use Jest with React Native Testing Library
   - Test create, read, update, delete operations
   - Test relationships and cascading deletes

**Deliverables:**
- [ ] WatermelonDB schema defined
- [ ] All models created with proper decorators
- [ ] Database initializing successfully
- [ ] Database service layer functional
- [ ] Database tests passing

**Time Estimate:** 3 days

---

### Week 3: Authentication System

#### Backend Development (Go)
**Objective:** Implement JWT-based authentication

**Tasks:**
1. Create authentication service
   - `internal/service/auth_service.go`
   - Implement user registration logic
   - Implement login with password verification
   - Generate JWT access tokens (15-minute expiry)
   - Generate refresh tokens (7-day expiry)

2. Implement password hashing service
   - `internal/service/password_service.go`
   - Use bcrypt with cost factor 12
   - Validate password strength (min 8 chars, uppercase, lowercase, digit, special char)
   - Compare password hashes securely

3. Create token service
   - `internal/service/token_service.go`
   - Generate JWT tokens with claims (athlete_id, email, type)
   - Validate and parse JWT tokens
   - Generate and validate refresh tokens (random 32-byte hex)
   - Store refresh token hashes in database

4. Implement authentication handlers
   - `internal/handler/auth_handler.go`
   - POST /auth/register: Create new athlete account
   - POST /auth/login: Authenticate and return tokens
   - POST /auth/refresh: Refresh access token
   - POST /auth/logout: Revoke refresh token

5. Create JWT authentication middleware
   - `internal/middleware/auth.go`
   - Extract JWT from Authorization header
   - Validate token signature and expiry
   - Extract athlete_id and add to request context
   - Return 401 Unauthorized if invalid

6. Add rate limiting for auth endpoints
   - Use go-redis for rate limit counters
   - Limit login attempts: 5 per 15 minutes per IP
   - Return 429 Too Many Requests when exceeded

7. Write authentication tests
   - Unit tests for password hashing
   - Unit tests for JWT generation/validation
   - Integration tests for auth endpoints
   - Test rate limiting behavior

**Deliverables:**
- [ ] Registration endpoint working (POST /auth/register)
- [ ] Login endpoint returning JWT tokens (POST /auth/login)
- [ ] Refresh token endpoint functional (POST /auth/refresh)
- [ ] JWT middleware protecting routes
- [ ] Rate limiting active on auth endpoints
- [ ] Authentication tests passing (>90% coverage)

**Time Estimate:** 5 days

---

#### Mobile Development
**Objective:** Build authentication UI and state management

**Tasks:**
1. Create authentication Redux slice
   - `src/features/auth/authSlice.ts`
   - Define auth state (isAuthenticated, user, tokens, loading, error)
   - Create async thunks (register, login, refreshToken, logout)
   - Add reducers for auth actions

2. Build Login screen
   - `src/features/auth/screens/LoginScreen.tsx`
   - Email and password input fields
   - Form validation with React Hook Form
   - Error message display
   - "Forgot Password" link (future)
   - Navigation to Register screen

3. Build Register screen
   - `src/features/auth/screens/RegisterScreen.tsx`
   - Input fields: email, password, name, body weight, gender, birth year
   - Password strength indicator
   - Form validation (email format, password requirements)
   - Navigation to Login screen

4. Implement secure token storage
   - Use react-native-keychain for iOS/Android keychain
   - Store access token and refresh token securely
   - Retrieve tokens on app launch
   - Clear tokens on logout

5. Create API client with token refresh
   - `src/services/api/client.ts`
   - Configure Axios instance with base URL
   - Add request interceptor to attach Authorization header
   - Add response interceptor to handle 401 errors
   - Implement automatic token refresh flow

6. Create AuthContext for app-wide auth state
   - Provide isAuthenticated, user, login, logout methods
   - Handle auto-login on app launch (check stored tokens)
   - Navigate to Main or Auth screens based on auth state

**Deliverables:**
- [ ] Login screen functional with validation
- [ ] Register screen functional with all required fields
- [ ] Tokens stored securely in keychain
- [ ] Auth Redux slice managing state correctly
- [ ] Automatic token refresh working
- [ ] Navigation between Auth and Main based on auth state

**Time Estimate:** 5 days

---

### Week 4: Exercise Library & 1RM Management

#### Backend Development
**Objective:** Implement 1RM tracking and exercise library

**Tasks:**
1. Create exercise library seed data
   - `migrations/008_seed_exercises.sql`
   - Insert 60+ exercises:
     - Olympic lifts: Snatch, Clean, Jerk variations (20 exercises)
     - Squats: Back, Front, Overhead, Bulgarian Split (12 exercises)
     - Pulls: Clean Pull, Snatch Pull, Deadlift variations (10 exercises)
     - Presses: Push Press, Strict Press, Bench Press (8 exercises)
     - Accessories: Rows, Pull-ups, Core work (10 exercises)
   - Add exercise category tags (olympic, squat, pull, press, accessory)

2. Implement 1RM repository
   - `internal/repository/one_rep_max_repository.go`
   - Create methods: Create, GetAll, GetByExercise, GetHistory, Update, Delete
   - Query 1RM history sorted by date DESC
   - Get current max (most recent) for each exercise

3. Create 1RM service
   - `internal/service/one_rep_max_service.go`
   - Business logic for 1RM management
   - Calculate Sinclair coefficient (based on body weight, gender, lift total)
   - Calculate Robi coefficient (age-adjusted Sinclair)
   - Detect PR (personal record) when new 1RM exceeds previous

4. Implement 1RM handlers
   - `internal/handler/one_rep_max_handler.go`
   - GET /athlete/1rm: Get all 1RMs for authenticated athlete
   - POST /athlete/1rm: Create new 1RM entry
   - PUT /athlete/1rm/:id: Update existing 1RM
   - DELETE /athlete/1rm/:id: Delete 1RM entry
   - GET /athlete/1rm/history/:exercise: Get 1RM history for specific exercise

5. Add 1RM validation
   - Weight: 20kg - 300kg (configurable in settings)
   - Exercise name must exist in library
   - Date cannot be in the future
   - Sinclair/Robi coefficients calculated automatically

6. Write 1RM tests
   - Unit tests for Sinclair/Robi calculations
   - Integration tests for 1RM endpoints
   - Test PR detection logic

**Deliverables:**
- [ ] Exercise library seeded with 60+ exercises
- [ ] 1RM CRUD endpoints functional
- [ ] Sinclair/Robi coefficients calculated correctly
- [ ] 1RM history endpoint returning sorted data
- [ ] 1RM tests passing (>90% coverage)

**Time Estimate:** 4 days

---

#### Mobile Development
**Objective:** Build 1RM management UI

**Tasks:**
1. Create exercise library browser
   - `src/features/profile/components/ExerciseLibrary.tsx`
   - Display 60+ exercises in categorized sections
   - Add search/filter functionality
   - Use SectionList for grouped display

2. Build 1RM input screen
   - `src/features/profile/screens/Add1RMScreen.tsx`
   - Exercise picker (autocomplete from library)
   - Weight input (kg/lbs based on user preference)
   - Date picker
   - Notes input (optional)
   - Calculate and display Sinclair coefficient

3. Create 1RM list screen
   - `src/features/profile/screens/OneRepMaxesScreen.tsx`
   - Display current 1RMs for key lifts (Snatch, C&J, Back Squat, Front Squat)
   - Show "Add 1RM" button
   - Navigate to Add1RM or Edit1RM screens

4. Build 1RM history chart
   - `src/features/profile/components/1RMHistoryChart.tsx`
   - Use Victory Native for line chart
   - Display 1RM progression over time for selected exercise
   - Add time range selector (30d, 90d, 6m, 1y, all)
   - Highlight PRs with markers

5. Implement local 1RM storage
   - Create WatermelonDB queries for 1RMs
   - Sync 1RMs to backend on create/update/delete
   - Handle offline 1RM creation

6. Add 1RM Redux slice
   - `src/features/profile/oneRepMaxSlice.ts`
   - Manage 1RM state (current maxes, history, loading, error)
   - Create async thunks for CRUD operations

**Deliverables:**
- [ ] Exercise library browsable with search
- [ ] 1RM input screen functional
- [ ] 1RM list displaying current maxes
- [ ] 1RM history chart showing progression
- [ ] 1RMs syncing to backend
- [ ] Sinclair coefficient calculated and displayed

**Time Estimate:** 4 days

---

## Phase 2: Core Logging (Weeks 5-8)

### Week 5: Session Creation UI

#### Mobile Development
**Objective:** Build workout session creation flow

**Tasks:**
1. Create "Start Workout" flow
   - `src/features/workout/screens/CreateSessionScreen.tsx`
   - Prominent "Start Workout" button on Home screen
   - Navigate to session configuration screen
   - Auto-populate today's date and current time

2. Build session configuration screen
   - Date picker (default: today)
   - Session type selector (heavy, moderate, light, technique, testing)
   - Start time input
   - Notes input (optional, max 500 characters)
   - "Begin Logging" button

3. Design exercise selection interface
   - `src/features/workout/components/ExerciseSelector.tsx`
   - Search bar with autocomplete
   - Recent exercises list (last 7 days)
   - Favorite exercises
   - Browse by category (Olympic, Squat, Pull, Press, Accessory)

4. Build set input form
   - `src/features/workout/components/SetInput.tsx`
   - Large touch targets (min 44pt height)
   - Weight input (number pad keyboard)
   - Reps input (number pad keyboard)
   - RPE input (1-10 scale with picker)
   - Notes input (optional)
   - "Add Set" button with haptic feedback

5. Add smart defaults
   - Pre-fill weight from previous session for same exercise
   - Suggest reps based on previous sets
   - Pre-populate RPE if pattern detected

6. Implement percentage-based calculations
   - Add "% of 1RM" toggle
   - Input percentage (e.g., "80%")
   - Auto-calculate weight from current 1RM
   - Update weight when 1RM changes

**Deliverables:**
- [ ] Start Workout button navigating to session creation
- [ ] Session configuration screen functional
- [ ] Exercise selection with search and recent exercises
- [ ] Set input form with validation
- [ ] Smart defaults pre-filling previous weights
- [ ] Percentage calculations working

**Time Estimate:** 5 days

---

#### Backend Development
**Objective:** Create session and set management endpoints

**Tasks:**
1. Implement session repository
   - `internal/repository/session_repository.go`
   - Create methods: Create, GetAll, GetById, Update, Delete
   - Query sessions with pagination (cursor-based)
   - Filter by date range, session type
   - Fetch session with related sets (eager loading)

2. Create session service
   - `internal/service/session_service.go`
   - Business logic for session creation
   - Calculate total volume load (sum of all set volumes)
   - Validate session data (date not in future, valid session type)
   - Handle transactional session + sets creation

3. Implement session handlers
   - `internal/handler/session_handler.go`
   - POST /sessions: Create new session with sets
   - GET /sessions: List sessions with pagination
   - GET /sessions/:id: Get session details with sets
   - PUT /sessions/:id: Update session metadata
   - DELETE /sessions/:id: Delete session (cascade to sets)

4. Add session validation
   - Date: Cannot be more than 1 day in the future
   - Session type: Must be one of (heavy, moderate, light, technique, testing)
   - Overall RPE: 1-10 range
   - Sets: At least 1 set required

5. Write session endpoint tests
   - Integration tests for session CRUD
   - Test transactional behavior (session + sets created atomically)
   - Test pagination and filtering

**Deliverables:**
- [ ] Session creation endpoint functional (POST /sessions)
- [ ] Session list endpoint with pagination (GET /sessions)
- [ ] Session detail endpoint returning sets (GET /sessions/:id)
- [ ] Session update and delete endpoints working
- [ ] Total volume load calculated correctly
- [ ] Session tests passing

**Time Estimate:** 3 days

---

### Week 6: Set-by-Set Logging

#### Mobile Development
**Objective:** Build comprehensive set logging interface

**Tasks:**
1. Build set logging interface
   - `src/features/workout/screens/LogSessionScreen.tsx`
   - Display current exercise name and category
   - Show list of logged sets for current exercise
   - Large "Add Set" button (always visible, bottom of screen)
   - Set number auto-incrementing

2. Implement set editing and deletion
   - Tap set to edit weight/reps/RPE
   - Swipe to delete set
   - Confirmation dialog for deletion
   - Recalculate set numbers after deletion

3. Create session timer display
   - Show elapsed time since session start
   - Display in header (MM:SS format)
   - Persist timer on screen navigation

4. Build session summary view
   - `src/features/workout/components/SessionSummary.tsx`
   - Total volume load (sum of all set volumes)
   - Total duration (start to end time)
   - Set count per exercise
   - Average RPE

5. Add session RPE input
   - Display prompt when athlete taps "Finish Workout"
   - Slider for RPE (1-10)
   - Submit button to complete session

6. Implement optimistic UI updates
   - Create session locally immediately
   - Update UI before API call completes
   - Handle API errors gracefully (retry or manual sync)

**Deliverables:**
- [ ] Set logging interface functional with large touch targets
- [ ] Set editing and deletion working
- [ ] Session timer displayed and accurate
- [ ] Session summary showing totals
- [ ] Overall session RPE input functional
- [ ] Optimistic UI updates for fast UX

**Time Estimate:** 5 days

---

#### Backend Development
**Objective:** Implement exercise set management

**Tasks:**
1. Create exercise set repository
   - `internal/repository/exercise_set_repository.go`
   - Create methods: Create, GetBySession, GetById, Update, Delete
   - Support batch creation (multiple sets at once)

2. Implement set handlers
   - `internal/handler/exercise_set_handler.go`
   - POST /sessions/:sessionId/sets: Add sets to existing session
   - PUT /sets/:id: Update individual set
   - DELETE /sets/:id: Delete set

3. Add set validation
   - Weight: 20-300 kg (configurable)
   - Reps: 1-20
   - RPE: 1-10 (optional)
   - Velocity: 0.1-3.0 m/s (optional)
   - Exercise name must exist in library

4. Implement volume load calculation
   - Calculate volume_load as weight * reps
   - Use PostgreSQL generated column for automatic calculation
   - Update session total_volume_load when sets change

5. Write set endpoint tests
   - Integration tests for set CRUD
   - Test volume load calculation
   - Test session volume update on set changes

**Deliverables:**
- [ ] Set creation endpoint functional (POST /sessions/:sessionId/sets)
- [ ] Set update endpoint working (PUT /sets/:id)
- [ ] Set deletion endpoint functional (DELETE /sets/:id)
- [ ] Volume load calculated and stored correctly
- [ ] Session total volume updated automatically
- [ ] Set tests passing

**Time Estimate:** 3 days

---

### Week 7: Offline Storage & Sync Preparation

#### Mobile Development
**Objective:** Implement offline-first architecture with sync

**Tasks:**
1. Implement local-first write pattern
   - All creates/updates write to WatermelonDB first
   - Mark records as synced=false
   - Display data from local database immediately

2. Build sync queue system
   - `src/services/sync/syncQueue.ts`
   - Queue create/update/delete operations
   - Store operations in local queue table
   - Retry failed syncs with exponential backoff

3. Create sync status indicators
   - Add "Synced" badge on session cards (green checkmark)
   - Add "Pending sync" badge (yellow clock icon)
   - Display sync errors (red X icon with error message)

4. Implement optimistic UI updates
   - Create session → immediately show in list
   - Update set → immediately reflect changes
   - Delete session → immediately remove from list
   - Rollback on sync failure (mark as error, allow retry)

5. Handle offline mode gracefully
   - Detect network connectivity (NetInfo)
   - Disable sync when offline
   - Show "Offline" banner in header
   - Queue all changes for sync when online

6. Build pull-to-refresh for manual sync
   - Add RefreshControl to session list
   - Trigger sync on pull-to-refresh
   - Show sync progress
   - Update UI with synced data

7. Add sync settings
   - Auto-sync toggle (default: on)
   - WiFi-only sync toggle (default: off)
   - Manual sync button

**Deliverables:**
- [ ] All writes to WatermelonDB first (optimistic UI)
- [ ] Sync queue storing pending operations
- [ ] Sync status indicators displayed correctly
- [ ] Offline mode handled gracefully
- [ ] Pull-to-refresh triggering sync
- [ ] Sync settings functional

**Time Estimate:** 5 days

---

#### Backend Development
**Objective:** Build sync endpoints with delta sync

**Tasks:**
1. Create sync endpoints
   - POST /sync/sessions: Upload created/updated sessions
   - GET /sync/changes: Fetch changes since last sync
   - Handle batch sync (multiple sessions at once)

2. Implement delta sync
   - Accept `last_sync_at` timestamp from client
   - Return only records changed since last sync
   - Include deleted records (soft delete with deleted_at)

3. Add sync metadata
   - Add `synced_at` timestamp to sessions, sets, 1RMs, videos
   - Track last sync time per athlete
   - Return sync timestamp in response

4. Handle conflict resolution
   - Implement last-write-wins strategy (MVP)
   - Compare updated_at timestamps
   - Accept newer record
   - Log conflicts for monitoring

5. Optimize sync performance
   - Batch insert sessions and sets
   - Use transactions for consistency
   - Add indexes on synced_at, updated_at

6. Write sync endpoint tests
   - Test delta sync logic
   - Test conflict resolution
   - Test batch sync performance

**Deliverables:**
- [ ] Sync endpoints functional (POST /sync/sessions, GET /sync/changes)
- [ ] Delta sync returning only changed records
- [ ] Conflict resolution implemented (last-write-wins)
- [ ] Sync metadata tracked
- [ ] Sync tests passing

**Time Estimate:** 3 days

---

### Week 8: Session History & Management

#### Mobile Development
**Objective:** Build session list and detail screens

**Tasks:**
1. Build Session List screen
   - `src/features/workout/screens/SessionListScreen.tsx`
   - Use FlatList with virtualization
   - Display 30 sessions initially
   - Load more on scroll (infinite scroll)

2. Create SessionCard component
   - `src/features/workout/components/SessionCard.tsx`
   - Display date, session type, total volume, set count
   - Show sync status (synced/pending/error)
   - Tap to navigate to SessionDetail screen

3. Add date range filtering
   - Filter bar with presets (Last 7 days, 30 days, 90 days, All time)
   - Custom date range picker
   - Update list based on filter

4. Implement session search
   - Search bar at top of list
   - Filter by exercise name
   - Show matching sessions only

5. Build SessionDetail screen
   - `src/features/workout/screens/SessionDetailScreen.tsx`
   - Display all session metadata (date, type, RPE, notes)
   - Show all sets grouped by exercise
   - Display total volume, duration
   - "Edit Session" button
   - "Delete Session" button

6. Add session editing
   - Navigate to edit screen (reuse CreateSessionScreen)
   - Allow editing session type, RPE, notes
   - Allow editing/deleting individual sets
   - Update session and sync

7. Implement session deletion
   - Show confirmation dialog
   - Delete from local database
   - Queue deletion for sync
   - Remove from list immediately

**Deliverables:**
- [ ] Session list displaying recent sessions
- [ ] FlatList virtualization for performance
- [ ] Date range filtering functional
- [ ] Session search by exercise working
- [ ] SessionDetail screen showing full session data
- [ ] Session editing functional
- [ ] Session deletion with confirmation

**Time Estimate:** 5 days

---

#### Backend Development
**Objective:** Enhance session endpoints with filtering and pagination

**Tasks:**
1. Implement session list pagination
   - Use cursor-based pagination (more efficient than offset)
   - Return pagination metadata (nextCursor, hasMore)
   - Default limit: 30 sessions
   - Max limit: 100 sessions

2. Add session filtering
   - Filter by date range (startDate, endDate)
   - Filter by session type
   - Filter by exercise (join with exercise_sets)
   - Combine multiple filters (AND logic)

3. Create session update endpoint
   - PUT /sessions/:id: Update session metadata only
   - Allow updating: sessionType, overallRpe, notes
   - Recalculate total_volume_load if sets changed
   - Return updated session

4. Add session deletion
   - DELETE /sessions/:id: Soft delete session
   - Set deleted_at timestamp
   - Cascade delete to sets (soft delete)
   - Create audit log entry

5. Optimize query performance
   - Add composite indexes on (athlete_id, date)
   - Use EXPLAIN ANALYZE to identify slow queries
   - Optimize joins between sessions and sets

6. Write session management tests
   - Test pagination (cursor-based)
   - Test filtering logic
   - Test update endpoint
   - Test soft delete behavior

**Deliverables:**
- [ ] Session list with cursor-based pagination
- [ ] Date range and session type filtering working
- [ ] Session update endpoint functional
- [ ] Session soft delete working
- [ ] Query performance optimized
- [ ] Tests passing for all session management features

**Time Estimate:** 3 days

---

## Phase 3: Analytics (Weeks 9-11)

### Week 9: Progress Charts & 1RM Tracking

#### Backend Development
**Objective:** Build analytics endpoints for progress tracking

**Tasks:**
1. Create analytics endpoints
   - GET /analytics/progress: 1RM progress over time
   - Query parameters: exercise (required), timeRange (30d, 90d, 6m, 1y, all)
   - Return data points with date and weight

2. Implement 1RM progress calculation
   - Fetch all 1RM entries for specified exercise
   - Filter by time range
   - Sort by date ASC
   - Calculate percent change from first to last

3. Add PR detection logic
   - Identify personal records (highest weight for each exercise)
   - Return PR date and weight
   - Calculate current max (most recent 1RM)

4. Calculate Sinclair/Robi coefficients
   - Use body weight and gender from athlete profile
   - Calculate Sinclair coefficient: total / (A * exp(B * (ln(bw/bw0))^2))
   - Constants: Male (A=0.751945, B=175.508, bw0=173.961), Female (A=0.783497, B=153.655, bw0=153.757)
   - Calculate Robi (age-adjusted Sinclair)

5. Create materialized view for performance
   - `athlete_1rm_progress`: Pre-aggregate 1RM data
   - Refresh daily at midnight UTC
   - Indexed on athlete_id, exercise

6. Write analytics tests
   - Test 1RM progress calculation
   - Test time range filtering
   - Test Sinclair/Robi calculations
   - Test PR detection

**Deliverables:**
- [ ] Progress endpoint functional (GET /analytics/progress)
- [ ] 1RM progress data returned correctly
- [ ] Percent change calculated
- [ ] Sinclair/Robi coefficients accurate
- [ ] Materialized view for performance
- [ ] Analytics tests passing

**Time Estimate:** 4 days

---

#### Mobile Development
**Objective:** Build analytics dashboard with progress charts

**Tasks:**
1. Create Analytics Dashboard screen
   - `src/features/analytics/screens/AnalyticsDashboardScreen.tsx`
   - Display lift progress charts for key lifts
   - Add lift selector (Snatch, Clean & Jerk, Back Squat, Front Squat)
   - Show current max, PR date, Sinclair coefficient

2. Build lift progress chart
   - `src/features/analytics/components/ProgressChart.tsx`
   - Use Victory Native for line chart
   - X-axis: Date, Y-axis: Weight (kg)
   - Plot 1RM data points
   - Add trendline (optional)

3. Add time range selector
   - Segmented control: 30d, 90d, 6m, 1y, All time
   - Update chart when range changes
   - Store selected range in Redux

4. Display PR indicator badges
   - Show "PR" badge on highest data point
   - Display PR weight and date in legend
   - Highlight PR in different color

5. Create chart interactions
   - Tap data point to see exact weight and date
   - Zoom/pan gestures for detailed view
   - Long-press to compare with other lifts

6. Build analytics Redux slice
   - `src/features/analytics/analyticsSlice.ts`
   - Manage progress data, loading state, selected exercise, time range
   - Create async thunks to fetch progress data

**Deliverables:**
- [ ] Analytics Dashboard displaying lift progress
- [ ] Progress chart showing 1RM over time
- [ ] Time range selector functional
- [ ] PR indicators displayed correctly
- [ ] Sinclair coefficient shown
- [ ] Chart interactions working

**Time Estimate:** 5 days

---

### Week 10: Volume Load Analytics

#### Backend Development
**Objective:** Implement volume analytics with trend detection

**Tasks:**
1. Create volume analytics endpoint
   - GET /analytics/volume: Weekly volume load trends
   - Query parameters: timeRange (30d, 90d, 6m, 1y)
   - Return weekly aggregations with session count, avg RPE

2. Implement weekly volume aggregation
   - Group sessions by week (date_trunc('week', date))
   - Sum total_volume_load for each week
   - Count sessions per week
   - Calculate average RPE per week

3. Add volume by category breakdown
   - Join with exercise_sets
   - Group by exercise_category (olympic, squat, pull, press, accessory)
   - Sum volume_load per category
   - Return breakdown with percentages

4. Calculate 4-week moving average
   - Use window function: AVG(weekly_volume) OVER (ORDER BY week ROWS BETWEEN 3 PRECEDING AND CURRENT ROW)
   - Return moving average alongside actual volume

5. Detect volume spikes
   - Compare current week volume to 4-week average
   - Flag if > 20% above average (yellow zone)
   - Flag if > 30% above average (red zone)

6. Create materialized view
   - `athlete_volume_weekly`: Pre-aggregate weekly volume
   - Refresh daily at midnight UTC
   - Indexed on athlete_id, week_start

7. Write volume analytics tests
   - Test weekly aggregation
   - Test category breakdown
   - Test moving average calculation
   - Test volume spike detection

**Deliverables:**
- [ ] Volume endpoint functional (GET /analytics/volume)
- [ ] Weekly volume data returned
- [ ] Volume by category breakdown working
- [ ] 4-week moving average calculated
- [ ] Volume spike detection functional
- [ ] Materialized view for performance
- [ ] Tests passing

**Time Estimate:** 4 days

---

#### Mobile Development
**Objective:** Build volume analytics visualizations

**Tasks:**
1. Build volume load trend chart
   - `src/features/analytics/components/VolumeChart.tsx`
   - Use Victory Native bar chart
   - X-axis: Week, Y-axis: Volume (kg)
   - Display last 12 weeks
   - Add 4-week moving average line

2. Create volume breakdown chart
   - `src/features/analytics/components/VolumeCategoryChart.tsx`
   - Use stacked bar chart or pie chart
   - Show volume by category (Olympic, Squat, Pull, Press, Accessory)
   - Display percentages and absolute values

3. Add color-coded volume zones
   - Green: Within normal range (within 20% of average)
   - Yellow: Elevated volume (20-30% above average)
   - Red: High volume spike (>30% above average)

4. Display session-by-session volume
   - Create table view showing individual sessions
   - Sort by date DESC
   - Display volume and % of weekly total

5. Show weekly totals with trend indicators
   - Display current week volume
   - Show change from previous week (up/down arrow, percentage)
   - Display 4-week average

6. Add volume alerts
   - Show banner when volume spike detected
   - Recommend deload if sustained high volume
   - Dismissible alerts

**Deliverables:**
- [ ] Volume trend chart displaying 12-week history
- [ ] Moving average line overlaid on chart
- [ ] Volume breakdown by category visualized
- [ ] Color-coded volume zones displayed
- [ ] Session-by-session volume table functional
- [ ] Volume alerts displayed when spikes detected

**Time Estimate:** 5 days

---

### Week 11: ACWR & Load Management

#### Backend Development
**Objective:** Implement ACWR calculation and load management

**Tasks:**
1. Create ACWR calculation endpoint
   - GET /analytics/acwr: Calculate Acute:Chronic Workload Ratio
   - Calculate acute load: Sum of last 7 days total volume
   - Calculate chronic load: Average weekly volume of last 28 days

2. Implement ACWR logic
   - ACWR = acute_load / chronic_load
   - Handle edge cases (chronic_load = 0 → return null)
   - Classify into risk zones:
     - Optimal: 0.8 - 1.3 (green)
     - Caution: 1.3 - 1.5 (yellow)
     - High risk: > 1.5 (red)

3. Generate load recommendations
   - If ACWR < 0.8: "Consider increasing training volume"
   - If ACWR 0.8 - 1.3: "Continue current training load"
   - If ACWR 1.3 - 1.5: "Monitor closely, consider reducing volume"
   - If ACWR > 1.5: "High injury risk, consider deload week"

4. Add frequency analytics endpoint
   - GET /analytics/frequency: Training frequency metrics
   - Calculate sessions per week (average)
   - Calculate current training streak (consecutive days)
   - Calculate longest training streak

5. Optimize ACWR query performance
   - Use materialized view `athlete_volume_weekly`
   - Add indexes on week_start
   - Cache ACWR result in Redis (5-minute TTL)

6. Write ACWR tests
   - Test ACWR calculation with various scenarios
   - Test risk zone classification
   - Test recommendation generation
   - Test edge cases (no data, chronic_load = 0)

**Deliverables:**
- [ ] ACWR endpoint functional (GET /analytics/acwr)
- [ ] ACWR calculated correctly
- [ ] Risk zones classified accurately
- [ ] Load recommendations generated
- [ ] Frequency analytics endpoint working
- [ ] Query performance optimized
- [ ] ACWR tests passing

**Time Estimate:** 4 days

---

#### Mobile Development
**Objective:** Build ACWR dashboard and load management UI

**Tasks:**
1. Build ACWR display
   - `src/features/analytics/components/ACWRGauge.tsx`
   - Use circular gauge or dial visualization
   - Display ACWR value in center
   - Color-code gauge based on risk zone (green/yellow/red)

2. Add risk zone indicators
   - Display risk zone label ("Optimal", "Caution", "High Risk")
   - Show color-coded banner
   - Display acute load, chronic load values

3. Create training frequency heatmap
   - `src/features/analytics/components/FrequencyHeatmap.tsx`
   - Calendar heatmap showing session frequency
   - Color intensity based on sessions per day (0, 1, 2+)
   - Display last 90 days

4. Display streak tracking
   - Show current training streak (consecutive days with sessions)
   - Show longest streak achieved
   - Add streak badges/achievements

5. Show weekly volume recommendations
   - Display recommended volume for next week
   - Based on ACWR and recent trends
   - Show as banner or card

6. Add ACWR alerts
   - Show visual alert when ACWR > 1.5
   - Display recommendation (e.g., "Consider deload week")
   - Allow dismissal but persist warning

7. Create load management education
   - Add info icon explaining ACWR
   - Link to resources on load management
   - Explain risk zones and recommendations

**Deliverables:**
- [ ] ACWR gauge displaying current ratio
- [ ] Risk zones color-coded correctly
- [ ] Training frequency heatmap showing 90-day history
- [ ] Streak tracking displayed
- [ ] Volume recommendations shown
- [ ] ACWR alerts displayed when high risk detected
- [ ] Educational content accessible

**Time Estimate:** 5 days

---

## Phase 4: Video Features (Weeks 12-15)

### Week 12: Video Upload Infrastructure

#### Backend Development
**Objective:** Set up S3 infrastructure and presigned URL generation

**Tasks:**
1. Create AWS S3 bucket
   - Bucket name: `ascend-videos-eu-west-1`
   - Region: eu-west-1 (Ireland) for GDPR compliance
   - Enable versioning
   - Set up lifecycle policies (delete incomplete multipart uploads after 7 days)

2. Configure S3 bucket policies
   - Private bucket (no public read access)
   - Allow authenticated uploads via presigned URLs
   - Set up CORS rules (allow mobile app origins)

3. Implement presigned URL generation
   - `internal/service/video_service.go`
   - Generate presigned POST URL with 1-hour expiry
   - Include required fields (key, bucket, policy, signature)
   - Validate file size limit (max 100MB)

4. Create video upload endpoints
   - POST /videos/upload: Request presigned URL
   - Return video_id, upload_url, fields for S3 POST
   - Store video metadata in database (pending status)

5. Implement upload completion endpoint
   - POST /videos/:id/complete: Notify backend upload done
   - Update video status to "processing"
   - Queue video processing job

6. Set up video metadata model
   - Add video table fields: s3_bucket, s3_key, processing_status
   - Track upload timestamp, file size, duration

7. Write video upload tests
   - Test presigned URL generation
   - Test upload completion flow
   - Test metadata storage

**Deliverables:**
- [ ] S3 bucket created with correct policies
- [ ] Presigned URL generation functional
- [ ] Upload request endpoint working (POST /videos/upload)
- [ ] Upload completion endpoint functional (POST /videos/:id/complete)
- [ ] Video metadata stored in database
- [ ] Tests passing

**Time Estimate:** 3 days

---

#### Infrastructure Setup
**Objective:** Configure Lambda for video processing

**Tasks:**
1. Create Lambda function for video processing
   - Function name: `ascend-video-processor`
   - Runtime: Go 1.x or Python 3.12
   - Memory: 3008 MB (maximum for FFmpeg)
   - Timeout: 5 minutes

2. Add FFmpeg Lambda layer
   - Use prebuilt FFmpeg layer (or build custom)
   - Include libx264, libaac codecs
   - Attach layer to Lambda function

3. Configure S3 event trigger
   - Trigger on `ObjectCreated:*` events in videos/ prefix
   - Invoke Lambda function automatically
   - Set up dead letter queue for failed invocations

4. Set up asynq job queue
   - Install go-redis and asynq
   - Configure Redis connection (ElastiCache)
   - Create video processing task handler

5. Create CloudFront distribution
   - Origin: S3 bucket (ascend-videos-eu-west-1)
   - Enable HTTPS only
   - Set up custom domain (videos.ascend.app)
   - Configure signed URLs with 1-hour expiry

6. Write Lambda deployment script
   - Use AWS SAM or Serverless Framework
   - Automate deployment from CI/CD

**Deliverables:**
- [ ] Lambda function deployed with FFmpeg
- [ ] S3 event trigger configured
- [ ] asynq job queue functional
- [ ] CloudFront distribution serving videos
- [ ] Signed URLs working

**Time Estimate:** 4 days

---

#### Mobile Development
**Objective:** Build video upload UI

**Tasks:**
1. Build video picker
   - `src/features/video/components/VideoPicker.tsx`
   - Use react-native-image-picker
   - Request camera roll permissions
   - Allow selecting video from library or recording new

2. Implement direct-to-S3 upload
   - `src/services/video/uploadService.ts`
   - Request presigned URL from backend
   - Upload video to S3 using POST multipart upload
   - Track upload progress (0-100%)

3. Add upload progress indicator
   - Display progress bar during upload
   - Show current progress percentage
   - Allow canceling upload

4. Create video metadata input
   - `src/features/video/screens/UploadVideoScreen.tsx`
   - Exercise name picker
   - Weight input (optional)
   - Link to session (optional)
   - Notes input (max 200 characters)

5. Handle upload failures
   - Retry failed uploads automatically (3 attempts)
   - Queue for manual retry if all attempts fail
   - Display error message with retry button

6. Add WiFi-only upload preference
   - Settings toggle: "Upload on WiFi only"
   - Check network type before upload
   - Queue cellular uploads if WiFi-only enabled

7. Create video upload queue
   - Store pending uploads in local database
   - Process queue when app comes to foreground
   - Show upload queue status in settings

**Deliverables:**
- [ ] Video picker functional (library and camera)
- [ ] Direct-to-S3 upload working
- [ ] Upload progress displayed
- [ ] Video metadata input screen functional
- [ ] Upload failure handling with retry
- [ ] WiFi-only upload preference working
- [ ] Upload queue managing pending uploads

**Time Estimate:** 5 days

---

### Week 13: Video Processing & Compression

#### Backend Development (Lambda)
**Objective:** Implement FFmpeg video compression and thumbnail generation

**Tasks:**
1. Implement FFmpeg video compression
   - `lambda/videoProcessor/handler.go` (or Python)
   - Download video from S3 to /tmp
   - Run FFmpeg compression command:
     ```bash
     ffmpeg -i input.mp4 \
       -vf "scale='min(1920,iw)':'min(1080,ih)':force_original_aspect_ratio=decrease" \
       -c:v libx264 -preset medium -crf 23 \
       -maxrate 4M -bufsize 8M \
       -movflags +faststart \
       -c:a aac -b:a 128k \
       output.mp4
     ```
   - Target: 1080p max, H.264 codec, 2-4 Mbps bitrate

2. Extract video metadata
   - Use FFprobe to get duration, resolution, fps, codec
   - Store metadata in video table

3. Generate thumbnails
   - Extract frame at 2-second mark (or 25% of duration if < 8s)
   - Create 3 sizes:
     - Small: 160x90 (for grid view)
     - Medium: 320x180 (for list view)
     - Large: 640x360 (for detail view)
   - Save as JPEG with 80% quality
   - Upload thumbnails to S3

4. Update video processing status
   - Update database: processing_status = "completed"
   - Set processed_at timestamp
   - Update file_size with compressed size

5. Handle processing failures
   - Catch errors and update status = "failed"
   - Log error details to CloudWatch
   - Retry failed jobs (max 3 attempts)

6. Calculate compression ratio
   - Compare original file size to compressed size
   - Target: 60-70% file size reduction
   - Log compression metrics

7. Write Lambda tests
   - Test FFmpeg compression with sample videos
   - Test thumbnail generation
   - Test error handling

**Deliverables:**
- [ ] FFmpeg compression working (60-70% size reduction)
- [ ] Video metadata extracted and stored
- [ ] 3 thumbnail sizes generated
- [ ] Processing status updated correctly
- [ ] Error handling functional
- [ ] Compression metrics logged
- [ ] Lambda tests passing

**Time Estimate:** 5 days

---

#### Infrastructure Management
**Objective:** Configure CloudFront and monitoring

**Tasks:**
1. Set up CloudFront signed URLs
   - Generate CloudFront key pair
   - Store private key in AWS Secrets Manager
   - Implement signed URL generation in backend

2. Configure CloudFront caching
   - Cache videos for 24 hours
   - Cache thumbnails for 7 days
   - Invalidate cache on video delete

3. Add CloudWatch monitoring
   - Log Lambda execution time
   - Log compression ratio
   - Alert on processing failures (>5% failure rate)
   - Track video processing queue depth

4. Set up S3 storage metrics
   - Monitor total storage usage
   - Alert when approaching 5TB limit
   - Track storage cost per user

**Deliverables:**
- [ ] CloudFront signed URLs working (1-hour expiry)
- [ ] Caching policies configured
- [ ] CloudWatch alarms set up
- [ ] S3 storage metrics tracked

**Time Estimate:** 2 days

---

### Week 14: Video Playback & Library

#### Mobile Development
**Objective:** Build video library and playback UI

**Tasks:**
1. Build Video Library screen
   - `src/features/video/screens/VideoLibraryScreen.tsx`
   - Display videos in thumbnail grid (3 columns)
   - Use FlatList with numColumns={3}
   - Load thumbnails with FastImage for caching

2. Implement video player
   - `src/features/video/screens/VideoPlayerScreen.tsx`
   - Use react-native-video for playback
   - Full-screen video player
   - Custom playback controls

3. Add playback controls
   - `src/features/video/components/VideoControls.tsx`
   - Play/pause button (large touch target)
   - Scrub bar for seeking
   - Current time / total duration display
   - Volume control

4. Implement slow-motion playback
   - Speed selector: 0.25x, 0.5x, 0.75x, 1x
   - Segmented control in player UI
   - Update playback rate dynamically

5. Add frame-by-frame navigation
   - Show frame navigation arrows for 60fps+ videos
   - Step forward/backward by 1 frame
   - Display current frame number

6. Create video filtering
   - Filter by exercise name
   - Filter by date range
   - Filter by quality rating (1-5 stars)
   - Sort by: Date, rating, duration

7. Build video detail view
   - Display video metadata (exercise, weight, date, notes)
   - Show quality rating with stars
   - Add "Edit" and "Delete" buttons

**Deliverables:**
- [ ] Video Library displaying thumbnails in grid
- [ ] Video player functional with custom controls
- [ ] Slow-motion playback working (0.25x - 1x)
- [ ] Frame-by-frame navigation functional for high fps videos
- [ ] Video filtering and sorting working
- [ ] Video detail view displaying metadata

**Time Estimate:** 5 days

---

#### Backend Development
**Objective:** Implement video retrieval and management endpoints

**Tasks:**
1. Create video list endpoint
   - GET /videos: List athlete's videos
   - Support pagination (cursor-based, 30 per page)
   - Filter by exercise, date range, rating
   - Sort by uploaded_at DESC

2. Implement video detail endpoint
   - GET /videos/:id: Get video details
   - Generate signed CloudFront URL for playback (1-hour expiry)
   - Return thumbnail URL, metadata, notes

3. Add video update endpoint
   - PUT /videos/:id: Update video metadata
   - Allow updating: quality_rating, notes, privacy_level
   - Validate rating (1-5)

4. Implement video deletion
   - DELETE /videos/:id: Delete video
   - Delete S3 objects (video + thumbnails)
   - Soft delete database record
   - Create audit log entry

5. Optimize video queries
   - Add indexes on (athlete_id, uploaded_at)
   - Eager load related exercise set if linked
   - Cache video list in Redis (5-minute TTL)

6. Write video endpoint tests
   - Test video list with filtering
   - Test signed URL generation
   - Test video update
   - Test video deletion (S3 + database)

**Deliverables:**
- [ ] Video list endpoint functional (GET /videos)
- [ ] Video detail endpoint returning signed URLs (GET /videos/:id)
- [ ] Video update endpoint working (PUT /videos/:id)
- [ ] Video deletion functional (DELETE /videos/:id)
- [ ] Query performance optimized
- [ ] Tests passing

**Time Estimate:** 3 days

---

### Week 15: Video Analysis Tools

#### Mobile Development
**Objective:** Build video comparison and annotation features

**Tasks:**
1. Build side-by-side comparison screen
   - `src/features/video/screens/VideoComparisonScreen.tsx`
   - Display 2 videos side-by-side (vertical split or horizontal)
   - Select videos from library for comparison
   - Common use cases: Today's lift vs PR, Heavy vs Light technique

2. Implement synchronized playback
   - Play/pause both videos simultaneously
   - Synchronize scrubbing (move both videos together)
   - Sync slow-motion speed

3. Add manual annotation tools
   - `src/features/video/components/VideoAnnotation.tsx`
   - Draw lines and angles on paused video frames
   - Color picker (3-4 colors: red, blue, green, yellow)
   - Line thickness selector
   - Undo/redo functionality

4. Create zoom capability
   - Pinch-to-zoom gestures
   - Pan to move zoomed view
   - Zoom up to 3x for detailed review

5. Add grid overlay option
   - Toggle grid overlay on video
   - Vertical reference line for bar path
   - Horizontal reference lines for position analysis

6. Build video sharing
   - `src/features/video/screens/ShareVideoScreen.tsx`
   - Generate shareable link (backend API call)
   - Copy link to clipboard
   - Set privacy level (private, shareable, public)

7. Add video quality rating
   - Star rating (1-5) on video detail screen
   - Tap to update rating
   - Display average rating in library

8. Create annotation export
   - Save annotated frame as image
   - Export to camera roll or share

**Deliverables:**
- [ ] Side-by-side comparison screen functional
- [ ] Synchronized playback working
- [ ] Annotation tools drawing correctly (lines, angles)
- [ ] Zoom capability functional
- [ ] Grid overlay toggle working
- [ ] Video sharing generating shareable links
- [ ] Quality rating functional
- [ ] Annotated frames exportable

**Time Estimate:** 5 days

---

#### Backend Development
**Objective:** Implement video sharing and privacy controls

**Tasks:**
1. Create video sharing endpoint
   - POST /videos/:id/share: Generate shareable link
   - Create unique share token (32-byte random hex)
   - Store token with expiry (default: 7 days, max: 30 days)
   - Return public share URL

2. Implement public video viewing
   - GET /share/:token: View shared video
   - No authentication required
   - Return video URL, thumbnail, metadata
   - Track view count (optional)

3. Add privacy level controls
   - Video privacy_level: private (default), shareable, public
   - Private: Only athlete can view
   - Shareable: Anyone with link can view
   - Public: Listed in public feed (future feature)

4. Implement share token expiry
   - Cron job to delete expired tokens daily
   - Return 404 if token expired
   - Allow extending expiry

5. Add share analytics (optional)
   - Track view count per share token
   - Log viewer IP and user agent

6. Write video sharing tests
   - Test share token generation
   - Test public video viewing
   - Test privacy level enforcement
   - Test token expiry

**Deliverables:**
- [ ] Video sharing endpoint functional (POST /videos/:id/share)
- [ ] Public viewing working (GET /share/:token)
- [ ] Privacy levels enforced correctly
- [ ] Share token expiry functional
- [ ] Tests passing

**Time Estimate:** 2 days

---

## Phase 5: Polish & Testing (Weeks 16-18)

### Week 16: Bug Fixes & Performance Optimization

#### Backend Development
**Objective:** Optimize performance and fix critical bugs

**Tasks:**
1. Profile database queries
   - Use EXPLAIN ANALYZE on slow queries
   - Identify missing indexes
   - Add composite indexes where needed:
     - `sessions.athlete_id, date DESC`
     - `exercise_sets.session_id, set_number`
     - `videos.athlete_id, uploaded_at DESC`

2. Optimize ACWR calculation
   - Pre-aggregate weekly volume in materialized view
   - Refresh view daily at midnight UTC
   - Cache ACWR result in Redis (5-minute TTL)

3. Implement Redis caching
   - Cache 1RMs (key: `1rm:athlete_id:exercise`, TTL: 1 hour)
   - Cache recent sessions (key: `sessions:athlete_id:30d`, TTL: 5 minutes)
   - Invalidate cache on data changes

4. Add database connection pooling
   - Configure pgx connection pool
   - Max connections: 20 per API instance
   - Min connections: 5
   - Connection lifetime: 30 minutes

5. Fix all critical bugs
   - Review bug tracker (GitHub Issues)
   - Prioritize by severity (critical, high, medium, low)
   - Fix critical and high-priority bugs

6. Conduct load testing
   - Use k6 or Apache JMeter
   - Simulate 1,000 concurrent users
   - Test scenarios: Login, session creation, analytics fetch
   - Measure p95 and p99 response times
   - Target: p95 < 300ms, p99 < 500ms

7. Optimize API response size
   - Use gzip compression for responses >1KB
   - Paginate large responses
   - Remove unnecessary fields from API responses

**Deliverables:**
- [ ] All critical queries have proper indexes
- [ ] ACWR calculation optimized (<100ms)
- [ ] Redis caching implemented for hot paths
- [ ] Connection pooling configured
- [ ] All critical and high bugs fixed
- [ ] Load testing completed (p95 < 300ms)
- [ ] API responses compressed

**Time Estimate:** 5 days

---

#### Mobile Development
**Objective:** Fix UI bugs and optimize app performance

**Tasks:**
1. Fix all UI/UX bugs
   - Review internal testing feedback
   - Fix layout issues on different screen sizes
   - Fix dark mode inconsistencies
   - Resolve navigation bugs

2. Optimize FlatList rendering
   - Use `getItemLayout` for fixed-height items
   - Set `initialNumToRender={10}` and `maxToRenderPerBatch={10}`
   - Add `removeClippedSubviews={true}` for performance
   - Use `React.memo` for list item components

3. Reduce app cold start time
   - Lazy load heavy screens
   - Defer non-critical initialization
   - Optimize bundle size with code splitting
   - Target: <2 seconds cold start

4. Optimize image loading
   - Use react-native-fast-image for caching
   - Preload thumbnails for next page
   - Compress images before upload

5. Add memoization
   - Use `useMemo` for expensive calculations (ACWR, volume totals)
   - Use `useCallback` for event handlers
   - Memoize chart components

6. Test on low-end devices
   - Test on 2-year-old iPhone (iPhone 11)
   - Test on 2-year-old Android (Samsung Galaxy A50)
   - Ensure smooth scrolling and transitions
   - Fix performance bottlenecks

7. Reduce bundle size
   - Analyze bundle with react-native-bundle-visualizer
   - Remove unused dependencies
   - Use smaller alternatives where possible
   - Target: <30MB app size

**Deliverables:**
- [ ] All UI/UX bugs fixed
- [ ] FlatList scrolling smoothly (60fps)
- [ ] App cold start <2 seconds
- [ ] Image loading optimized
- [ ] Expensive calculations memoized
- [ ] App performant on 2-year-old devices
- [ ] Bundle size reduced (<30MB)

**Time Estimate:** 5 days

---

### Week 17: Onboarding, Beta Testing & Localization

#### Mobile Development
**Objective:** Build onboarding flow and German localization

**Tasks:**
1. Build first-time user onboarding
   - `src/features/onboarding/screens/OnboardingScreen.tsx`
   - 3-screen tutorial (swipeable carousel)
   - Screen 1: Track your training
   - Screen 2: Analyze your progress
   - Screen 3: Record technique videos
   - "Get Started" button on final screen

2. Create tutorial tooltips
   - Add tooltips for key features (first use only)
   - Explain session logging flow
   - Explain ACWR risk zones
   - Explain video comparison
   - Use react-native-walkthrough-tooltip

3. Add sample workout
   - Pre-load sample session on first launch
   - Show example of logged session with sets
   - Allow deletion or keeping as template

4. Implement dark mode
   - Create dark theme color palette
   - Add dark mode toggle in settings
   - Persist preference in AsyncStorage
   - Use React Context for theme provider

5. Add German language translations
   - Create translation files (i18n)
   - Translate all UI strings to German
   - Use react-i18next for localization
   - Add language picker in settings (English, German)

6. Test localization
   - Verify date format (DD/MM/YYYY) for EMEA
   - Verify 24-hour time format
   - Verify kg as default unit
   - Test number formatting (comma vs period)

**Deliverables:**
- [ ] Onboarding flow functional (<2 minutes)
- [ ] Tutorial tooltips displayed on first use
- [ ] Sample workout pre-loaded
- [ ] Dark mode toggle working
- [ ] German translations complete
- [ ] Language picker functional
- [ ] Localization tested (dates, times, units)

**Time Estimate:** 4 days

---

#### Backend Development
**Objective:** Implement GDPR compliance features

**Tasks:**
1. Set up Sentry error tracking
   - Install sentry-go SDK
   - Configure Sentry DSN
   - Capture panics and errors
   - Add request context to errors

2. Configure PostHog analytics
   - Use PostHog EU instance (GDPR-compliant)
   - Track key events: session_created, video_uploaded, analytics_viewed
   - Respect opt-out preference

3. Implement GDPR data export
   - GET /athlete/export: Export all athlete data
   - Return JSON with sessions, sets, 1RMs, videos, profile
   - Include video download links (signed URLs)
   - Add CSV export option

4. Add GDPR data deletion
   - DELETE /athlete/account: Permanently delete account
   - 30-day grace period (soft delete)
   - Delete all related data (sessions, sets, videos, 1RMs)
   - Delete S3 videos
   - Create audit log entry

5. Create audit logging
   - Log all data access (login, export, deletion)
   - Store in audit_logs table
   - Include IP address, user agent, timestamp
   - Retain for 2 years

6. Write GDPR compliance tests
   - Test data export completeness
   - Test data deletion cascade
   - Test audit logging

**Deliverables:**
- [ ] Sentry error tracking functional
- [ ] PostHog analytics tracking events
- [ ] Data export endpoint returning complete data (GET /athlete/export)
- [ ] Data deletion endpoint functional (DELETE /athlete/account)
- [ ] Audit logging capturing all critical actions
- [ ] GDPR tests passing

**Time Estimate:** 3 days

---

#### Beta Testing Preparation
**Objective:** Recruit and onboard 30 beta athletes

**Tasks:**
1. Recruit 30 beta athletes
   - UK: 10 athletes (English speakers)
   - Germany: 8 athletes (German localization testing)
   - France: 5 athletes (future French localization)
   - Nordics: 4 athletes (tech-savvy early adopters)
   - Ireland: 3 athletes (EU timezone, English)
   - Post on European weightlifting forums, Instagram, Reddit

2. Distribute beta builds
   - iOS: TestFlight beta
   - Android: Google Play Console closed testing
   - Send invitations to beta testers

3. Set up feedback channels
   - In-app feedback form (link to Google Form or Typeform)
   - Email: beta@ascend.app
   - Weekly check-in surveys

4. Create beta testing guide
   - Document with testing instructions
   - Key features to test
   - How to report bugs
   - Expected behaviors

5. Monitor error rates
   - Set up Sentry dashboard for beta testers
   - Track crash rate (target: <0.5%)
   - Monitor API error rate (target: <0.1%)

**Deliverables:**
- [ ] 30 beta athletes recruited
- [ ] Beta builds distributed (TestFlight + Play Console)
- [ ] Feedback channels set up
- [ ] Beta testing guide created
- [ ] Error monitoring active

**Time Estimate:** 2 days

---

### Week 18: App Store Preparation & Launch

#### App Store Submission
**Objective:** Prepare and submit app to stores

**Tasks:**
1. Create App Store listing
   - App name: Project Ascend
   - Subtitle: Weightlifting Performance Tracker
   - Description (English): Highlight features, benefits, EMEA focus
   - Description (German): Professional translation
   - Keywords: weightlifting, Olympic lifting, performance tracking, ACWR, technique analysis

2. Design app icon
   - Create 1024x1024 app icon
   - Follow iOS and Android design guidelines
   - Represent weightlifting and performance

3. Create screenshots
   - iPhone: 6.7" (iPhone 14 Pro Max)
   - Android: 1080x1920 pixels
   - Show key screens: Session logging, Analytics dashboard, Video analysis
   - Add captions explaining features

4. Record demo video
   - 30-second app preview video
   - Show session logging, ACWR, video comparison
   - Add music and voice-over (optional)

5. Submit iOS app to App Store Review
   - Build release version with Xcode
   - Upload to App Store Connect
   - Complete metadata and screenshots
   - Submit for review (2-3 day review time)

6. Submit Android app to Play Store
   - Build release APK with Android Studio
   - Upload to Play Console
   - Complete store listing
   - Submit for review (few hours to 1 day)

**Deliverables:**
- [ ] App Store and Play Store listings complete
- [ ] App icon designed and uploaded
- [ ] Screenshots created and uploaded
- [ ] Demo video recorded and uploaded
- [ ] iOS app submitted to App Store Review
- [ ] Android app submitted to Play Store

**Time Estimate:** 3 days

---

#### Backend Deployment
**Objective:** Deploy production infrastructure

**Tasks:**
1. Deploy production infrastructure
   - Set up ECS Fargate cluster in eu-west-1
   - Create Application Load Balancer
   - Configure auto-scaling (min 2, max 10 tasks)
   - Set up Route 53 DNS (api.ascend.app)

2. Set up CloudWatch alarms
   - API error rate > 1% (critical)
   - Database CPU > 80% (warning)
   - Video processing failures > 5% (warning)
   - Send alerts to PagerDuty or email

3. Configure database backups
   - Automated daily backups (RDS)
   - 30-day retention period
   - Test restore procedure

4. Implement auto-scaling policies
   - Scale up: CPU > 70% for 2 minutes
   - Scale down: CPU < 30% for 5 minutes
   - Target tracking: 50% CPU utilization

5. Set up SSL certificates
   - Use AWS Certificate Manager (ACM)
   - Configure HTTPS on ALB
   - Redirect HTTP to HTTPS

6. Document API
   - Create Swagger/OpenAPI spec
   - Host API documentation (Redoc or Swagger UI)
   - URL: docs.ascend.app

**Deliverables:**
- [ ] Production infrastructure deployed to AWS
- [ ] Auto-scaling configured
- [ ] CloudWatch alarms set up
- [ ] Database backups enabled
- [ ] SSL certificates configured (HTTPS)
- [ ] API documentation published

**Time Estimate:** 3 days

---

#### Launch Activities
**Objective:** Execute launch plan and monitor metrics

**Tasks:**
1. Monitor beta feedback
   - Review all bug reports
   - Fix showstopper bugs (P0)
   - Prioritize P1 bugs for post-launch

2. Prepare launch announcement
   - Write launch post for European weightlifting forums
   - Create Instagram launch post
   - Prepare Reddit r/weightlifting post
   - Schedule posts for launch day

3. Set up customer support
   - Create support email: support@ascend.app
   - Set up email forwarding
   - Create FAQ page

4. Create privacy policy and terms of service
   - Hire lawyer or use template
   - Ensure GDPR compliance
   - Host on website (ascend.app/privacy, ascend.app/terms)
   - Link from app

5. Launch to public
   - Announce on social media
   - Post on weightlifting forums
   - Send email to waitlist
   - Monitor app store reviews

6. Track launch metrics
   - Day 1 installs (target: 50)
   - Week 1 installs (target: 100)
   - Crash rate (target: <0.5%)
   - Session logging rate (target: 50% of users log at least 1 session)

**Deliverables:**
- [ ] All P0 bugs fixed before launch
- [ ] Launch announcement prepared
- [ ] Customer support set up
- [ ] Privacy policy and terms published
- [ ] App launched publicly
- [ ] Launch metrics tracked

**Time Estimate:** 2 days

---

## Success Metrics

### Week 4 Checkpoint
- [ ] Authentication working end-to-end
- [ ] 1RMs logged and displayed
- [ ] Internal team able to create accounts

### Week 8 Checkpoint
- [ ] 100 sessions logged by internal team
- [ ] Offline-first sync functional
- [ ] Session list and detail screens working

### Week 11 Checkpoint
- [ ] ACWR calculated correctly
- [ ] Analytics dashboard functional
- [ ] Volume trends displayed

### Week 15 Checkpoint
- [ ] 50 videos uploaded by internal team
- [ ] Video playback and comparison working
- [ ] Side-by-side analysis functional

### Week 18 Launch Checkpoint
- [ ] 30 beta athletes onboarded
- [ ] 100 public installs in first week
- [ ] <0.5% crash rate
- [ ] 50% of users log at least 1 session
- [ ] App Store rating >4.0 stars

---

## Risk Mitigation

### Risk: Development timeline slips
**Impact:** High - Delays launch date
**Probability:** Medium
**Mitigation:**
- Defer program templates feature if needed (not in critical path)
- Focus on core logging + analytics + video as MVP
- Add buffer time in Weeks 16-18 for catch-up

### Risk: Video processing costs exceed budget
**Impact:** Medium - Unexpected AWS costs
**Probability:** Low
**Mitigation:**
- Implement aggressive compression (60-70% file size reduction)
- 5GB storage cap per user (approximately 100 videos)
- Monitor Lambda execution costs weekly
- Set up billing alerts at $100, $500, $1000 thresholds

### Risk: Low beta tester engagement
**Impact:** Medium - Lack of feedback to improve MVP
**Probability:** Medium
**Mitigation:**
- Incentivize beta testers with €30 Amazon gift cards
- Weekly check-ins with active testers
- Clear testing instructions and goals
- Active engagement in European weightlifting communities

### Risk: GDPR compliance issues
**Impact:** High - Legal liability in EU
**Probability:** Low
**Mitigation:**
- Legal review of privacy policy and terms
- Implement data export and deletion features
- Use EU-based AWS region (eu-west-1)
- Clear consent mechanisms for data collection

### Risk: App Store rejection
**Impact:** High - Delays public launch
**Probability:** Low
**Mitigation:**
- Follow iOS and Android design guidelines strictly
- Test thoroughly on devices before submission
- Address any review feedback immediately
- Have Plan B: Launch with TestFlight/Play Console beta only initially

---

## Post-Launch Roadmap (Months 4-6)

### Month 4: Iteration & Bug Fixes
- Fix all P1 and P2 bugs reported by users
- Improve onboarding based on user feedback
- Add French language translation
- Optimize performance based on real-world usage

### Month 5: Enhanced Features
- Add program templates feature
- Implement exercise library customization
- Add wellness tracking (optional sleep, soreness)
- Improve video analysis tools based on feedback

### Month 6: Growth & Expansion
- Partner with 5 European weightlifting clubs
- Add social features (follow athletes, PR feed)
- Implement referral program
- Target: 500 athletes across 5 countries

---

## Appendices

### Appendix A: Key Dependencies

**Backend (Go):**
- Gin 1.10+ (web framework)
- GORM 1.25+ (ORM)
- pgx 5.x (PostgreSQL driver)
- go-redis 9.x (Redis client)
- asynq 0.24+ (job queue)
- zerolog 1.32+ (logging)
- golang-jwt 5.x (JWT)
- testify 1.9+ (testing)

**Mobile (React Native):**
- React Native 0.74+
- TypeScript 5.3+
- React Navigation 6.x
- Redux Toolkit 2.x
- WatermelonDB 0.27+
- react-native-video 6.x
- Victory Native 36.x
- Axios 1.6+

**Infrastructure:**
- AWS ECS Fargate (container hosting)
- AWS RDS PostgreSQL 15 (database)
- AWS S3 (video storage)
- AWS Lambda (video processing)
- AWS CloudFront (CDN)
- GitHub Actions (CI/CD)

### Appendix B: Environment Variables

**Backend (.env):**
```env
# Server
PORT=8080
ENV=production

# Database
DATABASE_URL=postgresql://user:password@host:5432/ascend_db
DATABASE_MAX_CONNECTIONS=20

# Redis
REDIS_URL=redis://host:6379/0

# AWS
AWS_REGION=eu-west-1
AWS_S3_BUCKET=ascend-videos-eu-west-1
AWS_CLOUDFRONT_DOMAIN=d123456.cloudfront.net

# JWT
JWT_ACCESS_SECRET=<strong-secret-key>
JWT_REFRESH_SECRET=<strong-secret-key>
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

# Sentry
SENTRY_DSN=<sentry-dsn>

# PostHog
POSTHOG_API_KEY=<posthog-key>
POSTHOG_HOST=https://eu.posthog.com
```

**Mobile (.env):**
```env
# API
API_BASE_URL=https://api.ascend.app
API_TIMEOUT=30000

# Environment
ENV=production
```

### Appendix C: Deployment Checklist

**Pre-Deployment:**
- [ ] All tests passing (backend + mobile)
- [ ] Database migrations reviewed and tested
- [ ] Environment variables configured
- [ ] SSL certificates configured
- [ ] Monitoring and alerts set up

**Deployment:**
- [ ] Build Docker image and push to ECR
- [ ] Deploy to ECS Fargate
- [ ] Run database migrations
- [ ] Verify API health check
- [ ] Test critical endpoints

**Post-Deployment:**
- [ ] Monitor error rates for 1 hour
- [ ] Check CloudWatch logs
- [ ] Verify video upload and processing
- [ ] Test mobile app against production API
- [ ] Smoke test key user flows

### Appendix D: Contact Information

**Project Lead:** [Name]
**Backend Lead:** [Name]
**Mobile Lead:** [Name]
**DevOps Lead:** [Name]

**Communication Channels:**
- Slack: #ascend-dev
- Daily Standup: 10:00 CET
- Sprint Planning: Mondays 14:00 CET
- Retrospective: Fridays 16:00 CET

---

**Document Status:** Approved for Implementation
**Next Review:** End of Week 4 (Checkpoint 1)
**Last Updated:** October 6, 2025
