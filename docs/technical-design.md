# Project Ascend: Technical Design Document
## Mobile-First Weightlifting Performance Platform - MVP

**Version:** 1.0
**Date:** October 6, 2025
**Authors:** Engineering Team
**Status:** Draft - Ready for Review
**Related Documents:** [PRD v2.0](./prd.md)

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [System Architecture](#2-system-architecture)
3. [Technology Stack](#3-technology-stack)
4. [Data Architecture](#4-data-architecture)
5. [API Design](#5-api-design)
6. [Mobile Application Architecture](#6-mobile-application-architecture)
7. [Video Processing Pipeline](#7-video-processing-pipeline)
8. [Offline-First Synchronization](#8-offline-first-synchronization)
9. [Authentication & Security](#9-authentication--security)
10. [Infrastructure & Deployment](#10-infrastructure--deployment)
11. [Testing Strategy](#11-testing-strategy)
12. [Performance & Optimization](#12-performance--optimization)
13. [Monitoring & Observability](#13-monitoring--observability)
14. [GDPR Compliance Implementation](#14-gdpr-compliance-implementation)
15. [Development Phases](#15-development-phases)
16. [Risk Mitigation](#16-risk-mitigation)

---

## 1. Executive Summary

### 1.1 Purpose

This technical design document specifies the architecture, implementation details, and engineering approach for Project Ascend MVP - an athlete-focused weightlifting performance tracking platform launching in EMEA (Q1 2026).

### 1.2 Design Goals

**Primary Objectives:**
1. **Mobile-First:** Seamless iOS/Android experience optimized for gym environments
2. **Offline-First:** Full functionality without internet connectivity
3. **GDPR-Compliant:** EU data residency and privacy by design
4. **Performant:** < 2s cold start, < 1s screen transitions
5. **Scalable:** Support 1,000 MVP users → 10,000 within 12 months

**Technical Principles:**
- Local-first architecture with cloud sync
- Modular, testable, maintainable codebase
- Security and privacy by default
- Horizontal scalability from day one
- Production-grade from MVP launch

### 1.3 Scope

**In Scope (MVP):**
- React Native mobile app (iOS/Android)
- Node.js REST API backend
- PostgreSQL database
- AWS S3 video storage with FFmpeg processing
- JWT-based authentication
- Offline SQLite storage with sync engine
- ACWR analytics and progress charts
- Video playback and comparison tools

**Out of Scope (Post-MVP):**
- Automated bar path tracking (ML/CV)
- VBT device integrations
- Coach platform features
- Social/community features
- Real-time collaboration
- GraphQL API

### 1.4 Key Technical Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Mobile Framework | React Native | Cross-platform, large ecosystem, faster development |
| Backend Language | Node.js + TypeScript | Type safety, JavaScript ecosystem alignment, async performance |
| Database | PostgreSQL 15+ | ACID compliance, JSON support, mature, excellent performance |
| File Storage | AWS S3 (eu-west-1) | GDPR-compliant EU region, reliable, cost-effective |
| Video Processing | FFmpeg | Industry standard, comprehensive format support |
| State Management | Redux Toolkit | Predictable state, DevTools, middleware ecosystem |
| Local Storage | SQLite | Reliable, fast, supports complex queries offline |
| API Architecture | REST | Simple, well-understood, sufficient for MVP |
| Authentication | JWT + Refresh Tokens | Stateless, scalable, standard approach |
| Deployment | AWS ECS Fargate | Container-based, auto-scaling, managed infrastructure |

---

## 2. System Architecture

### 2.1 High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     CLIENT LAYER (Mobile)                    │
├─────────────────────────────────────────────────────────────┤
│  React Native App (iOS/Android)                             │
│  ├── UI Components (React Navigation)                       │
│  ├── State Management (Redux Toolkit)                       │
│  ├── Local Database (SQLite via WatermelonDB)              │
│  ├── Sync Engine (Delta sync with conflict resolution)      │
│  └── Video Player (react-native-video)                      │
└─────────────────────────────────────────────────────────────┘
                            ↕ HTTPS (TLS 1.3)
┌─────────────────────────────────────────────────────────────┐
│                  APPLICATION LAYER (API)                     │
├─────────────────────────────────────────────────────────────┤
│  AWS Application Load Balancer                              │
│         ↓                                                    │
│  Node.js API Servers (ECS Fargate Tasks)                   │
│  ├── Express.js (REST endpoints)                            │
│  ├── Authentication Middleware (JWT validation)             │
│  ├── Business Logic (Services layer)                        │
│  ├── Data Access Layer (TypeORM repositories)               │
│  └── Background Jobs (Bull Queue + Redis)                   │
└─────────────────────────────────────────────────────────────┘
                            ↕
┌─────────────────────────────────────────────────────────────┐
│                    DATA LAYER                                │
├─────────────────────────────────────────────────────────────┤
│  PostgreSQL 15 (RDS Multi-AZ)                               │
│  ├── Athlete data (profiles, sessions, sets)                │
│  ├── Video metadata                                          │
│  └── Analytics aggregations                                  │
│                                                              │
│  AWS S3 (eu-west-1)                                         │
│  ├── Video files (private bucket)                           │
│  ├── Thumbnails                                              │
│  └── Annotated frames                                        │
│                                                              │
│  Redis (ElastiCache)                                        │
│  ├── Session store                                           │
│  ├── Job queue (video processing)                           │
│  └── Cache layer (1RM, recent sessions)                     │
└─────────────────────────────────────────────────────────────┘
                            ↕
┌─────────────────────────────────────────────────────────────┐
│              SUPPORTING SERVICES                             │
├─────────────────────────────────────────────────────────────┤
│  CloudFront CDN (video delivery)                            │
│  Lambda Functions (video processing workers)                 │
│  Sentry (error tracking)                                     │
│  PostHog (analytics - EU instance)                          │
│  CloudWatch (logs, metrics, alarms)                         │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 Component Breakdown

#### 2.2.1 Mobile Application Components

```typescript
src/
├── app/                    # App initialization
│   ├── App.tsx            # Root component
│   ├── store.ts           # Redux store configuration
│   └── navigation.tsx     # Navigation configuration
├── features/              # Feature modules (domain-driven)
│   ├── auth/              # Authentication
│   ├── workout/           # Workout logging
│   ├── analytics/         # Progress tracking
│   ├── video/             # Video management
│   └── profile/           # User profile
├── shared/                # Shared utilities
│   ├── components/        # Reusable UI components
│   ├── hooks/             # Custom React hooks
│   ├── utils/             # Utility functions
│   └── types/             # TypeScript types
├── services/              # API and data services
│   ├── api/               # API client
│   ├── database/          # SQLite/WatermelonDB
│   ├── sync/              # Sync engine
│   └── storage/           # File storage
└── config/                # Configuration
    ├── constants.ts       # App constants
    └── env.ts             # Environment variables
```

#### 2.2.2 Backend API Components

```go
cmd/
└── api/
    └── main.go            # Application entry point

internal/
├── config/                # Configuration
│   ├── database.go        # Database connection setup
│   ├── redis.go           # Redis client setup
│   └── aws.go             # AWS SDK configuration
├── middleware/            # Gin middleware
│   ├── auth.go            # JWT validation
│   ├── error.go           # Error handling
│   ├── ratelimit.go       # Rate limiting
│   └── cors.go            # CORS configuration
├── domain/                # Domain models
│   ├── athlete.go         # Athlete entity
│   ├── session.go         # Session entity
│   ├── exercise_set.go    # Exercise set entity
│   └── video.go           # Video entity
├── repository/            # Data access layer
│   ├── athlete_repo.go    # Athlete repository
│   ├── session_repo.go    # Session repository
│   └── video_repo.go      # Video repository
├── service/               # Business logic
│   ├── auth_service.go    # Authentication service
│   ├── session_service.go # Session service
│   └── video_service.go   # Video processing
├── handler/               # HTTP handlers (controllers)
│   ├── auth_handler.go    # Auth endpoints
│   ├── session_handler.go # Session endpoints
│   └── analytics_handler.go # Analytics endpoints
├── dto/                   # Data transfer objects (request/response)
│   ├── auth_dto.go        # Authentication DTOs
│   ├── session_dto.go     # Session DTOs
│   └── analytics_dto.go   # Analytics DTOs
├── util/                  # Utility functions
│   ├── jwt.go             # JWT utilities
│   ├── hash.go            # Password hashing
│   └── validator.go       # Custom validators
└── worker/                # Background workers
    ├── video_worker.go    # Video processing jobs
    └── analytics_worker.go # Analytics aggregation

pkg/                       # Shared packages (reusable)
├── logger/                # Logging utilities
├── errors/                # Custom error types
└── response/              # Standard API responses

migrations/                # Database migrations (GORM or sql-migrate)
└── *.sql                  # Migration files

tests/                     # Test files
├── unit/                  # Unit tests
├── integration/           # Integration tests
└── e2e/                   # End-to-end tests
```

### 2.3 Communication Patterns

**Client ↔ API:**
- REST over HTTPS (TLS 1.3)
- JSON payload format
- JWT bearer token authentication
- Gzip compression for responses > 1KB

**API ↔ Database:**
- GORM or sqlc for data access
- pgx connection pooling (max 20 connections per instance)
- Prepared statements for security and performance
- Read replicas for analytics queries (post-MVP)

**API ↔ S3:**
- AWS SDK for video upload/download
- Signed URLs with 1-hour expiry
- Multipart upload for files > 50MB

**Background Jobs:**
- Bull queue backed by Redis
- Video processing: Lambda functions
- Job retry with exponential backoff (3 attempts)

---

## 3. Technology Stack

### 3.1 Mobile Application

| Layer | Technology | Version | Purpose |
|-------|-----------|---------|---------|
| Framework | React Native | 0.74+ | Cross-platform mobile development |
| Language | TypeScript | 5.3+ | Type-safe JavaScript |
| Navigation | React Navigation | 6.x | Screen navigation and routing |
| State Management | Redux Toolkit | 2.x | Global state management |
| Local Database | WatermelonDB | 0.27+ | Offline-first reactive database |
| SQL Engine | SQLite | 3.x | Local data persistence |
| Video Player | react-native-video | 6.x | Video playback with controls |
| Charts | Victory Native | 36.x | Data visualization |
| HTTP Client | Axios | 1.6+ | API requests with interceptors |
| Forms | React Hook Form | 7.x | Form state and validation |
| Date/Time | date-fns | 3.x | Date manipulation (lightweight) |
| Testing | Jest + React Native Testing Library | Latest | Unit and integration testing |
| E2E Testing | Detox | 20.x | End-to-end mobile testing |

**Key Dependencies:**
```json
{
  "dependencies": {
    "react": "18.2.0",
    "react-native": "0.74.0",
    "@react-navigation/native": "^6.1.9",
    "@react-navigation/bottom-tabs": "^6.5.11",
    "@reduxjs/toolkit": "^2.0.1",
    "react-redux": "^9.0.4",
    "@nozbe/watermelondb": "^0.27.1",
    "react-native-video": "^6.0.0",
    "victory-native": "^36.9.0",
    "axios": "^1.6.2",
    "react-hook-form": "^7.49.2",
    "date-fns": "^3.0.6"
  },
  "devDependencies": {
    "@types/react": "^18.2.45",
    "@types/react-native": "^0.72.8",
    "jest": "^29.7.0",
    "@testing-library/react-native": "^12.4.2",
    "detox": "^20.14.6"
  }
}
```

### 3.2 Backend API

| Layer | Technology | Version | Purpose |
|-------|-----------|---------|---------|
| Language | Go | 1.22+ | Compiled, high-performance language |
| Framework | Gin | 1.10+ | Fast HTTP web framework |
| ORM | GORM | 1.25+ | Go ORM with migrations |
| Query Builder | sqlc | 1.25+ | Type-safe SQL from queries (alternative to GORM) |
| Database Driver | pgx | 5.x | PostgreSQL driver for Go |
| Database | PostgreSQL | 15+ | Primary datastore |
| Cache/Queue | go-redis | 9.x | Redis client for Go |
| Job Queue | asynq | 0.24+ | Distributed task queue |
| File Storage | AWS SDK Go v2 | Latest | S3 integration |
| Video Processing | FFmpeg | 6.x | Video compression (exec) |
| Authentication | golang-jwt | 5.x | JWT creation/validation |
| Password Hashing | bcrypt (golang.org/x/crypto) | Latest | Secure password hashing |
| Validation | go-playground/validator | 10.x | Struct validation |
| Testing | testing + testify | Latest | Unit and integration testing |
| Logging | zerolog | 1.32+ | Fast structured logging |
| Monitoring | sentry-go | 0.27+ | Error tracking |

**Key Dependencies (go.mod):**
```go
module github.com/ascend/api

go 1.22

require (
    github.com/gin-gonic/gin v1.10.0
    gorm.io/gorm v1.25.7
    gorm.io/driver/postgres v1.5.7
    github.com/jackc/pgx/v5 v5.5.5        // PostgreSQL driver
    github.com/redis/go-redis/v9 v9.5.1   // Redis client
    github.com/hibiken/asynq v0.24.1      // Job queue
    github.com/aws/aws-sdk-go-v2 v1.26.0
    github.com/aws/aws-sdk-go-v2/service/s3 v1.51.0
    github.com/golang-jwt/jwt/v5 v5.2.1
    golang.org/x/crypto v0.21.0           // bcrypt
    github.com/go-playground/validator/v10 v10.19.0
    github.com/rs/zerolog v1.32.0
    github.com/getsentry/sentry-go v0.27.0
    github.com/stretchr/testify v1.9.0    // Testing assertions
)

// Alternative: sqlc for type-safe queries
// github.com/sqlc-dev/sqlc v1.25.0
```

**Database Access Strategy:**
- **Option 1 (GORM):** Full-featured ORM with migrations, associations, hooks
- **Option 2 (sqlc):** Generate type-safe Go code from SQL queries (better performance, less magic)

### 3.3 Infrastructure

| Component | Technology | Purpose |
|-----------|-----------|---------|
| Container Registry | AWS ECR | Docker image storage |
| Container Orchestration | AWS ECS Fargate | Serverless container hosting |
| Load Balancer | AWS ALB | HTTPS termination, traffic routing |
| Database | AWS RDS PostgreSQL | Managed PostgreSQL (Multi-AZ) |
| Cache | AWS ElastiCache Redis | Managed Redis cluster |
| File Storage | AWS S3 | Video and file storage |
| CDN | AWS CloudFront | Video content delivery |
| Serverless Functions | AWS Lambda | Video processing workers |
| Secrets Management | AWS Secrets Manager | API keys, database credentials |
| Monitoring | AWS CloudWatch | Logs, metrics, alarms |
| Error Tracking | Sentry | Application error monitoring |
| Analytics | PostHog (EU) | User behavior analytics |
| CI/CD | GitHub Actions | Automated testing and deployment |

### 3.4 Development Tools

| Category | Tool | Purpose |
|----------|------|---------|
| Version Control | Git + GitHub | Source code management |
| Package Manager | npm | Dependency management |
| Code Quality | ESLint + Prettier | Linting and formatting |
| Type Checking | TypeScript | Static type checking |
| API Testing | Postman/Insomnia | Manual API testing |
| Database Tools | DBeaver/pgAdmin | Database management |
| Container Tools | Docker Desktop | Local containerization |
| Infrastructure | Terraform | Infrastructure as code (future) |

---

## 4. Data Architecture

### 4.1 Database Schema Design

#### 4.1.1 Core Tables

**athletes**
```sql
CREATE TABLE athletes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    body_weight DECIMAL(5,2), -- kg
    gender VARCHAR(10) CHECK (gender IN ('male', 'female', 'other')),
    birth_year INTEGER,
    timezone VARCHAR(50) DEFAULT 'Europe/London',
    locale VARCHAR(10) DEFAULT 'en-GB',
    units VARCHAR(10) DEFAULT 'metric' CHECK (units IN ('metric', 'imperial')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE -- Soft delete for GDPR
);

CREATE INDEX idx_athletes_email ON athletes(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_athletes_created_at ON athletes(created_at);
```

**one_rep_maxes**
```sql
CREATE TABLE one_rep_maxes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    athlete_id UUID NOT NULL REFERENCES athletes(id) ON DELETE CASCADE,
    exercise VARCHAR(100) NOT NULL,
    weight DECIMAL(5,2) NOT NULL, -- kg
    date DATE NOT NULL,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_1rm_athlete ON one_rep_maxes(athlete_id, exercise, date DESC);
```

**sessions**
```sql
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    athlete_id UUID NOT NULL REFERENCES athletes(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    session_type VARCHAR(20) CHECK (session_type IN ('heavy', 'moderate', 'light', 'technique', 'testing')),
    overall_rpe SMALLINT CHECK (overall_rpe BETWEEN 1 AND 10),
    notes TEXT,
    total_volume_load DECIMAL(10,2), -- Calculated field (denormalized for performance)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    synced_at TIMESTAMP WITH TIME ZONE -- Last sync from mobile
);

CREATE INDEX idx_sessions_athlete_date ON sessions(athlete_id, date DESC);
CREATE INDEX idx_sessions_created_at ON sessions(created_at);
```

**exercise_sets**
```sql
CREATE TABLE exercise_sets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    exercise_name VARCHAR(100) NOT NULL,
    exercise_category VARCHAR(20) CHECK (exercise_category IN ('olympic', 'squat', 'pull', 'press', 'accessory')),
    set_number SMALLINT NOT NULL,
    weight DECIMAL(5,2) NOT NULL, -- kg
    reps SMALLINT NOT NULL CHECK (reps BETWEEN 1 AND 50),
    rpe SMALLINT CHECK (rpe BETWEEN 1 AND 10),
    velocity DECIMAL(4,2), -- m/s (optional VBT data)
    notes TEXT,
    video_id UUID REFERENCES videos(id) ON DELETE SET NULL,
    volume_load DECIMAL(7,2) GENERATED ALWAYS AS (weight * reps) STORED,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sets_session ON exercise_sets(session_id, set_number);
CREATE INDEX idx_sets_exercise ON exercise_sets(exercise_name, created_at DESC);
```

**videos**
```sql
CREATE TABLE videos (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    athlete_id UUID NOT NULL REFERENCES athletes(id) ON DELETE CASCADE,
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    exercise_set_id UUID REFERENCES exercise_sets(id) ON DELETE SET NULL,
    exercise_name VARCHAR(100) NOT NULL,
    weight DECIMAL(5,2),

    -- File metadata
    original_filename VARCHAR(255) NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    duration_seconds DECIMAL(6,2),
    resolution VARCHAR(20), -- e.g., "1920x1080"
    fps INTEGER, -- frames per second

    -- S3 storage
    s3_bucket VARCHAR(100) NOT NULL,
    s3_key VARCHAR(255) NOT NULL,
    thumbnail_s3_key VARCHAR(255),

    -- Processing status
    processing_status VARCHAR(20) DEFAULT 'pending' CHECK (processing_status IN ('pending', 'processing', 'completed', 'failed')),
    processed_at TIMESTAMP WITH TIME ZONE,

    -- User metadata
    quality_rating SMALLINT CHECK (quality_rating BETWEEN 1 AND 5),
    notes TEXT,
    privacy_level VARCHAR(20) DEFAULT 'private' CHECK (privacy_level IN ('private', 'shareable', 'public')),
    share_token VARCHAR(64) UNIQUE, -- For shareable links

    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_videos_athlete ON videos(athlete_id, uploaded_at DESC);
CREATE INDEX idx_videos_session ON videos(session_id);
CREATE INDEX idx_videos_exercise ON videos(exercise_name, uploaded_at DESC);
CREATE INDEX idx_videos_share_token ON videos(share_token) WHERE share_token IS NOT NULL;
```

#### 4.1.2 Authentication & Security Tables

**refresh_tokens**
```sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    athlete_id UUID NOT NULL REFERENCES athletes(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_refresh_tokens_athlete ON refresh_tokens(athlete_id);
CREATE INDEX idx_refresh_tokens_expiry ON refresh_tokens(expires_at) WHERE revoked_at IS NULL;
```

**audit_logs**
```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    athlete_id UUID REFERENCES athletes(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL, -- e.g., 'login', 'data_export', 'data_deletion'
    resource_type VARCHAR(50), -- e.g., 'session', 'video'
    resource_id UUID,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB, -- Additional context
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_athlete ON audit_logs(athlete_id, created_at DESC);
CREATE INDEX idx_audit_action ON audit_logs(action, created_at DESC);
```

#### 4.1.3 Analytics Materialized Views

**athlete_volume_weekly**
```sql
CREATE MATERIALIZED VIEW athlete_volume_weekly AS
SELECT
    athlete_id,
    date_trunc('week', date) AS week_start,
    SUM(total_volume_load) AS weekly_volume,
    COUNT(*) AS session_count,
    AVG(overall_rpe) AS avg_rpe
FROM sessions
WHERE deleted_at IS NULL
GROUP BY athlete_id, date_trunc('week', date);

CREATE UNIQUE INDEX idx_athlete_volume_weekly ON athlete_volume_weekly(athlete_id, week_start);

-- Refresh schedule: Daily at midnight UTC
```

**exercise_1rm_history**
```sql
CREATE MATERIALIZED VIEW exercise_1rm_history AS
SELECT
    athlete_id,
    exercise,
    date,
    weight,
    ROW_NUMBER() OVER (PARTITION BY athlete_id, exercise ORDER BY date DESC) AS recency_rank
FROM one_rep_maxes
ORDER BY athlete_id, exercise, date DESC;

CREATE INDEX idx_1rm_history_athlete_exercise ON exercise_1rm_history(athlete_id, exercise, date DESC);
```

### 4.2 Data Access Patterns

#### 4.2.1 Read-Heavy Operations

**Session List (Primary Screen):**
```sql
-- Optimized query for session feed
SELECT
    s.id, s.date, s.session_type, s.overall_rpe, s.total_volume_load,
    COUNT(es.id) AS total_sets
FROM sessions s
LEFT JOIN exercise_sets es ON s.id = es.session_id
WHERE s.athlete_id = $1
  AND s.date >= $2
GROUP BY s.id
ORDER BY s.date DESC
LIMIT 30;
```

**Analytics Dashboard:**
```sql
-- ACWR Calculation (Acute: 7 days, Chronic: 28 days)
WITH acute AS (
    SELECT SUM(weekly_volume) AS acute_load
    FROM athlete_volume_weekly
    WHERE athlete_id = $1
      AND week_start >= (CURRENT_DATE - INTERVAL '7 days')
),
chronic AS (
    SELECT AVG(weekly_volume) AS chronic_load
    FROM athlete_volume_weekly
    WHERE athlete_id = $1
      AND week_start >= (CURRENT_DATE - INTERVAL '28 days')
)
SELECT
    acute.acute_load,
    chronic.chronic_load,
    CASE
        WHEN chronic.chronic_load > 0
        THEN acute.acute_load / chronic.chronic_load
        ELSE NULL
    END AS acwr
FROM acute, chronic;
```

#### 4.2.2 Write Operations

**Session Creation with Sets (Transaction):**
```typescript
async createSession(sessionData: CreateSessionDTO): Promise<Session> {
  return this.dataSource.transaction(async (manager) => {
    // Create session
    const session = manager.create(Session, {
      athleteId: sessionData.athleteId,
      date: sessionData.date,
      sessionType: sessionData.sessionType,
      overallRpe: sessionData.overallRpe,
    });
    await manager.save(session);

    // Create sets
    const sets = sessionData.sets.map(setData =>
      manager.create(ExerciseSet, {
        sessionId: session.id,
        ...setData,
      })
    );
    await manager.save(sets);

    // Calculate and update total volume
    const totalVolume = sets.reduce((sum, set) => sum + set.volumeLoad, 0);
    session.totalVolumeLoad = totalVolume;
    await manager.save(session);

    return session;
  });
}
```

### 4.3 Data Retention Policy

| Data Type | Retention | Rationale |
|-----------|-----------|-----------|
| Active athlete data | Indefinite | Core user data |
| Inactive athlete data | 12 months | GDPR Article 17 - delete after 12 months inactivity with notice |
| Audit logs | 2 years | Security and compliance |
| Video files | User-controlled | User can delete anytime, count toward 5GB limit |
| Deleted account data | 30 days | GDPR "right to be forgotten" - 30 day grace period |
| Analytics aggregations | 2 years | Historical trend analysis |

**Automated Cleanup Jobs:**
- Daily: Mark inactive accounts (> 365 days no login)
- Weekly: Send deletion warning emails to inactive accounts
- Monthly: Permanently delete accounts marked for deletion (> 30 days)

---

## 5. API Design

### 5.1 RESTful API Principles

**Base URL:** `https://api.ascend.app/v1`

**General Conventions:**
- Resource-based URLs (nouns, not verbs)
- HTTP methods for operations (GET, POST, PUT, DELETE)
- Plural resource names (`/sessions`, not `/session`)
- Nested resources for relationships (`/sessions/:id/sets`)
- Hypermedia links for navigation (HATEOAS-lite)
- Consistent error response format
- Pagination for list endpoints (cursor-based)
- Versioning in URL path (`/v1`)

### 5.2 Authentication Endpoints

#### POST `/auth/register`
**Purpose:** Create new athlete account

**Request:**
```json
{
  "email": "athlete@example.com",
  "password": "SecureP@ssw0rd",
  "name": "John Doe",
  "bodyWeight": 85.5,
  "gender": "male",
  "birthYear": 1995,
  "timezone": "Europe/London",
  "locale": "en-GB"
}
```

**Response:** `201 Created`
```json
{
  "athlete": {
    "id": "uuid",
    "email": "athlete@example.com",
    "name": "John Doe",
    "createdAt": "2025-10-06T12:00:00Z"
  },
  "tokens": {
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "refresh_token_here",
    "expiresIn": 900
  }
}
```

**Errors:**
- `400 Bad Request` - Invalid input data
- `409 Conflict` - Email already exists

#### POST `/auth/login`
**Purpose:** Authenticate athlete

**Request:**
```json
{
  "email": "athlete@example.com",
  "password": "SecureP@ssw0rd"
}
```

**Response:** `200 OK`
```json
{
  "athlete": { /* athlete object */ },
  "tokens": {
    "accessToken": "jwt_token",
    "refreshToken": "refresh_token",
    "expiresIn": 900
  }
}
```

**Errors:**
- `401 Unauthorized` - Invalid credentials
- `429 Too Many Requests` - Rate limit exceeded (5 attempts per 15 minutes)

#### POST `/auth/refresh`
**Purpose:** Refresh access token

**Request:**
```json
{
  "refreshToken": "refresh_token_here"
}
```

**Response:** `200 OK`
```json
{
  "accessToken": "new_jwt_token",
  "refreshToken": "new_refresh_token",
  "expiresIn": 900
}
```

#### POST `/auth/logout`
**Purpose:** Revoke refresh token

**Headers:** `Authorization: Bearer {accessToken}`

**Response:** `204 No Content`

### 5.3 Athlete Profile Endpoints

#### GET `/athlete/profile`
**Purpose:** Get authenticated athlete's profile

**Headers:** `Authorization: Bearer {accessToken}`

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "email": "athlete@example.com",
  "name": "John Doe",
  "bodyWeight": 85.5,
  "gender": "male",
  "birthYear": 1995,
  "timezone": "Europe/London",
  "locale": "en-GB",
  "units": "metric",
  "createdAt": "2025-01-01T00:00:00Z",
  "lastLoginAt": "2025-10-06T12:00:00Z"
}
```

#### PUT `/athlete/profile`
**Purpose:** Update athlete profile

**Request:**
```json
{
  "name": "John Doe Jr.",
  "bodyWeight": 86.0,
  "timezone": "Europe/Berlin"
}
```

**Response:** `200 OK` (updated profile object)

### 5.4 One Rep Max Endpoints

#### GET `/athlete/1rm`
**Purpose:** Get all 1RMs for athlete

**Query Parameters:**
- `exercise` (optional): Filter by exercise name

**Response:** `200 OK`
```json
{
  "oneRepMaxes": [
    {
      "id": "uuid",
      "exercise": "Snatch",
      "weight": 95.0,
      "date": "2025-10-01",
      "notes": "PR!",
      "createdAt": "2025-10-01T18:30:00Z"
    }
  ]
}
```

#### POST `/athlete/1rm`
**Purpose:** Record new 1RM

**Request:**
```json
{
  "exercise": "Clean & Jerk",
  "weight": 125.0,
  "date": "2025-10-06",
  "notes": "New PR!"
}
```

**Response:** `201 Created`

#### GET `/athlete/1rm/history/:exercise`
**Purpose:** Get 1RM history for specific exercise

**Response:** `200 OK`
```json
{
  "exercise": "Snatch",
  "history": [
    {
      "weight": 95.0,
      "date": "2025-10-01"
    },
    {
      "weight": 92.5,
      "date": "2025-09-15"
    }
  ],
  "currentMax": 95.0,
  "prDate": "2025-10-01"
}
```

### 5.5 Session Endpoints

#### GET `/sessions`
**Purpose:** List athlete's sessions

**Query Parameters:**
- `startDate` (ISO 8601): Filter by start date
- `endDate` (ISO 8601): Filter by end date
- `cursor` (UUID): Pagination cursor
- `limit` (integer, max 100, default 30): Results per page

**Response:** `200 OK`
```json
{
  "sessions": [
    {
      "id": "uuid",
      "date": "2025-10-06",
      "startTime": "2025-10-06T09:00:00Z",
      "endTime": "2025-10-06T10:30:00Z",
      "sessionType": "heavy",
      "overallRpe": 8,
      "totalVolumeLoad": 4500.0,
      "notes": "Felt strong today",
      "totalSets": 15
    }
  ],
  "pagination": {
    "nextCursor": "uuid_of_next_session",
    "hasMore": true
  }
}
```

#### POST `/sessions`
**Purpose:** Create new session with sets

**Request:**
```json
{
  "date": "2025-10-06",
  "startTime": "2025-10-06T09:00:00Z",
  "sessionType": "heavy",
  "overallRpe": 8,
  "notes": "Heavy snatch day",
  "sets": [
    {
      "exerciseName": "Snatch",
      "exerciseCategory": "olympic",
      "setNumber": 1,
      "weight": 80.0,
      "reps": 3,
      "rpe": 7
    },
    {
      "exerciseName": "Snatch",
      "exerciseCategory": "olympic",
      "setNumber": 2,
      "weight": 85.0,
      "reps": 2,
      "rpe": 8
    }
  ]
}
```

**Response:** `201 Created`
```json
{
  "id": "uuid",
  "date": "2025-10-06",
  "totalVolumeLoad": 410.0,
  "sets": [/* array of created sets */]
}
```

#### GET `/sessions/:id`
**Purpose:** Get session details including all sets

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "date": "2025-10-06",
  "sessionType": "heavy",
  "overallRpe": 8,
  "totalVolumeLoad": 4500.0,
  "sets": [
    {
      "id": "uuid",
      "exerciseName": "Snatch",
      "setNumber": 1,
      "weight": 80.0,
      "reps": 3,
      "volumeLoad": 240.0,
      "videoId": "uuid_if_video_attached"
    }
  ]
}
```

#### PUT `/sessions/:id`
**Purpose:** Update session metadata (not sets)

**Request:**
```json
{
  "overallRpe": 9,
  "notes": "Actually felt harder than expected"
}
```

**Response:** `200 OK`

#### DELETE `/sessions/:id`
**Purpose:** Delete session (cascades to sets)

**Response:** `204 No Content`

### 5.6 Exercise Set Endpoints

#### POST `/sessions/:sessionId/sets`
**Purpose:** Add sets to existing session

**Request:**
```json
{
  "sets": [
    {
      "exerciseName": "Back Squat",
      "exerciseCategory": "squat",
      "setNumber": 1,
      "weight": 120.0,
      "reps": 5,
      "rpe": 8
    }
  ]
}
```

**Response:** `201 Created`

#### PUT `/sets/:id`
**Purpose:** Update individual set

**Request:**
```json
{
  "weight": 122.5,
  "rpe": 9
}
```

**Response:** `200 OK`

#### DELETE `/sets/:id`
**Purpose:** Delete set

**Response:** `204 No Content`

### 5.7 Analytics Endpoints

#### GET `/analytics/progress`
**Purpose:** Get 1RM progress over time

**Query Parameters:**
- `exercise` (required): Exercise name
- `timeRange` (optional): `30d`, `90d`, `6m`, `1y`, `all` (default: `90d`)

**Response:** `200 OK`
```json
{
  "exercise": "Snatch",
  "timeRange": "90d",
  "dataPoints": [
    {
      "date": "2025-08-01",
      "weight": 90.0
    },
    {
      "date": "2025-09-15",
      "weight": 92.5
    },
    {
      "date": "2025-10-01",
      "weight": 95.0
    }
  ],
  "currentMax": 95.0,
  "changePercent": 5.56,
  "sinclair": 115.2
}
```

#### GET `/analytics/volume`
**Purpose:** Get volume load trends

**Query Parameters:**
- `timeRange` (optional): `30d`, `90d`, `6m`, `1y` (default: `90d`)

**Response:** `200 OK`
```json
{
  "timeRange": "90d",
  "weeklyVolume": [
    {
      "weekStart": "2025-07-08",
      "totalVolume": 12500.0,
      "sessionCount": 4,
      "avgRpe": 7.5
    },
    {
      "weekStart": "2025-07-15",
      "totalVolume": 13200.0,
      "sessionCount": 5,
      "avgRpe": 8.0
    }
  ],
  "volumeByCategory": {
    "olympic": 5000.0,
    "squat": 4500.0,
    "pull": 2500.0,
    "press": 1000.0,
    "accessory": 500.0
  }
}
```

#### GET `/analytics/acwr`
**Purpose:** Get Acute:Chronic Workload Ratio

**Response:** `200 OK`
```json
{
  "acwr": 1.15,
  "acuteLoad": 13500.0,
  "chronicLoad": 11739.0,
  "status": "optimal",
  "recommendation": "Continue current training load",
  "lastCalculated": "2025-10-06T12:00:00Z"
}
```

Status levels:
- `optimal`: ACWR between 0.8 and 1.3 (green)
- `caution`: ACWR between 1.3 and 1.5 (yellow)
- `high_risk`: ACWR > 1.5 (red)

#### GET `/analytics/frequency`
**Purpose:** Get training frequency heatmap

**Query Parameters:**
- `timeRange` (optional): `30d`, `90d`, `6m`, `1y` (default: `90d`)

**Response:** `200 OK`
```json
{
  "timeRange": "90d",
  "dailyActivity": {
    "2025-10-01": 1,
    "2025-10-02": 0,
    "2025-10-03": 1,
    "2025-10-04": 1
  },
  "averageSessionsPerWeek": 4.2,
  "currentStreak": 3,
  "longestStreak": 14
}
```

### 5.8 Video Endpoints

#### GET `/videos`
**Purpose:** List athlete's videos

**Query Parameters:**
- `exercise` (optional): Filter by exercise
- `startDate`, `endDate` (optional): Date range
- `cursor`, `limit`: Pagination

**Response:** `200 OK`
```json
{
  "videos": [
    {
      "id": "uuid",
      "exerciseName": "Snatch",
      "weight": 90.0,
      "thumbnailUrl": "https://cdn.ascend.app/thumbnails/...",
      "duration": 8.5,
      "uploadedAt": "2025-10-06T10:00:00Z",
      "qualityRating": 4
    }
  ],
  "pagination": {
    "nextCursor": "uuid",
    "hasMore": true
  }
}
```

#### POST `/videos/upload`
**Purpose:** Request pre-signed URL for video upload

**Request:**
```json
{
  "fileName": "snatch_90kg.mp4",
  "fileSize": 45678901,
  "exerciseName": "Snatch",
  "weight": 90.0,
  "sessionId": "uuid"
}
```

**Response:** `200 OK`
```json
{
  "videoId": "uuid",
  "uploadUrl": "https://s3.eu-west-1.amazonaws.com/...",
  "fields": {
    "key": "videos/athlete_id/video_id.mp4",
    "bucket": "ascend-videos-eu",
    "X-Amz-Algorithm": "AWS4-HMAC-SHA256",
    "X-Amz-Credential": "...",
    "X-Amz-Date": "...",
    "Policy": "...",
    "X-Amz-Signature": "..."
  },
  "expiresIn": 3600
}
```

**Client Upload Process:**
1. Client calls POST `/videos/upload` to get pre-signed URL
2. Client uploads directly to S3 using multipart upload
3. Client calls POST `/videos/:id/complete` to notify backend
4. Backend queues video processing job

#### POST `/videos/:id/complete`
**Purpose:** Notify backend that upload is complete

**Response:** `200 OK`
```json
{
  "status": "processing",
  "estimatedProcessingTime": 30
}
```

#### GET `/videos/:id`
**Purpose:** Get video details and playback URL

**Response:** `200 OK`
```json
{
  "id": "uuid",
  "exerciseName": "Snatch",
  "weight": 90.0,
  "duration": 8.5,
  "resolution": "1920x1080",
  "fps": 60,
  "fileSize": 42000000,
  "thumbnailUrl": "https://cdn.ascend.app/...",
  "videoUrl": "https://cdn.ascend.app/videos/...",
  "qualityRating": 4,
  "notes": "Good lift, minor balance issue at catch",
  "uploadedAt": "2025-10-06T10:00:00Z"
}
```

**videoUrl** is a signed CloudFront URL valid for 1 hour.

#### PUT `/videos/:id`
**Purpose:** Update video metadata

**Request:**
```json
{
  "qualityRating": 5,
  "notes": "Perfect lift",
  "privacyLevel": "shareable"
}
```

**Response:** `200 OK`

#### DELETE `/videos/:id`
**Purpose:** Delete video

**Response:** `204 No Content`

Backend deletes S3 objects and database records.

#### POST `/videos/:id/share`
**Purpose:** Generate shareable link

**Request:**
```json
{
  "expiresIn": 604800  // 7 days in seconds
}
```

**Response:** `200 OK`
```json
{
  "shareUrl": "https://ascend.app/share/AbCdEf123456",
  "expiresAt": "2025-10-13T12:00:00Z"
}
```

### 5.9 Error Response Format

**Consistent error structure:**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid input data",
    "details": [
      {
        "field": "email",
        "message": "Invalid email format"
      }
    ],
    "requestId": "uuid",
    "timestamp": "2025-10-06T12:00:00Z"
  }
}
```

**Error Codes:**
- `VALIDATION_ERROR` (400)
- `UNAUTHORIZED` (401)
- `FORBIDDEN` (403)
- `NOT_FOUND` (404)
- `CONFLICT` (409)
- `RATE_LIMIT_EXCEEDED` (429)
- `INTERNAL_SERVER_ERROR` (500)
- `SERVICE_UNAVAILABLE` (503)

### 5.10 Rate Limiting

**Per-endpoint limits:**
- Authentication endpoints: 5 requests / 15 minutes
- Read endpoints: 100 requests / minute
- Write endpoints: 30 requests / minute
- Video upload: 10 uploads / hour

**Headers:**
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1696593600
```

**Rate limit exceeded response:**
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Too many requests",
    "retryAfter": 60
  }
}
```

---

## 6. Mobile Application Architecture

### 6.1 Application Structure

**Core Architectural Patterns:**
- **Feature-based organization:** Domain-driven folder structure
- **Separation of concerns:** Presentation, business logic, data layers
- **Unidirectional data flow:** Redux for state management
- **Dependency injection:** React Context for service injection
- **Offline-first:** Local database as source of truth

### 6.2 Navigation Structure

```typescript
// Navigation hierarchy
RootNavigator (Stack)
├── AuthNavigator (Stack)
│   ├── Login
│   ├── Register
│   └── Onboarding
└── MainNavigator (Bottom Tab)
    ├── LogTab (Stack)
    │   ├── SessionList
    │   ├── SessionDetail
    │   ├── CreateSession
    │   └── AddSet
    ├── AnalyticsTab (Stack)
    │   ├── Dashboard
    │   ├── ProgressChart
    │   └── VolumeChart
    ├── VideosTab (Stack)
    │   ├── VideoLibrary
    │   ├── VideoPlayer
    │   └── VideoComparison
    └── ProfileTab (Stack)
        ├── ProfileOverview
        ├── OneRepMaxes
        ├── Settings
        └── ExportData
```

**Navigation Configuration:**
```typescript
// src/app/navigation.tsx
import { NavigationContainer } from '@react-navigation/native';
import { createStackNavigator } from '@react-navigation/stack';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';

const Stack = createStackNavigator();
const Tab = createBottomTabNavigator();

export function RootNavigator() {
  const { isAuthenticated } = useAuth();

  return (
    <NavigationContainer>
      <Stack.Navigator screenOptions={{ headerShown: false }}>
        {!isAuthenticated ? (
          <Stack.Screen name="Auth" component={AuthNavigator} />
        ) : (
          <Stack.Screen name="Main" component={MainNavigator} />
        )}
      </Stack.Navigator>
    </NavigationContainer>
  );
}

function MainNavigator() {
  return (
    <Tab.Navigator>
      <Tab.Screen name="LogTab" component={LogStackNavigator} />
      <Tab.Screen name="AnalyticsTab" component={AnalyticsStackNavigator} />
      <Tab.Screen name="VideosTab" component={VideosStackNavigator} />
      <Tab.Screen name="ProfileTab" component={ProfileStackNavigator} />
    </Tab.Navigator>
  );
}
```

### 6.3 State Management Architecture

**Redux Store Structure:**
```typescript
// src/app/store.ts
import { configureStore } from '@reduxjs/toolkit';

export const store = configureStore({
  reducer: {
    auth: authReducer,
    sessions: sessionsReducer,
    analytics: analyticsReducer,
    videos: videosReducer,
    sync: syncReducer,
    ui: uiReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: false, // WatermelonDB uses non-serializable data
    }).concat(syncMiddleware),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
```

**Feature Slice Example:**
```typescript
// src/features/workout/sessionsSlice.ts
import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import { sessionService } from '@/services/api/sessionService';

interface SessionsState {
  sessions: Session[];
  currentSession: Session | null;
  loading: boolean;
  error: string | null;
  syncStatus: 'idle' | 'syncing' | 'synced' | 'error';
}

export const fetchSessions = createAsyncThunk(
  'sessions/fetch',
  async (params: FetchSessionsParams) => {
    return await sessionService.fetchSessions(params);
  }
);

export const createSession = createAsyncThunk(
  'sessions/create',
  async (sessionData: CreateSessionDTO) => {
    return await sessionService.createSession(sessionData);
  }
);

const sessionsSlice = createSlice({
  name: 'sessions',
  initialState: {
    sessions: [],
    currentSession: null,
    loading: false,
    error: null,
    syncStatus: 'idle',
  } as SessionsState,
  reducers: {
    setCurrentSession: (state, action) => {
      state.currentSession = action.payload;
    },
    clearError: (state) => {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      .addCase(fetchSessions.pending, (state) => {
        state.loading = true;
      })
      .addCase(fetchSessions.fulfilled, (state, action) => {
        state.loading = false;
        state.sessions = action.payload;
      })
      .addCase(fetchSessions.rejected, (state, action) => {
        state.loading = false;
        state.error = action.error.message || 'Failed to fetch sessions';
      });
  },
});

export const { setCurrentSession, clearError } = sessionsSlice.actions;
export default sessionsSlice.reducer;
```

### 6.4 Local Database Schema (WatermelonDB)

**WatermelonDB Models:**
```typescript
// src/services/database/models/Session.ts
import { Model } from '@nozbe/watermelondb';
import { field, date, readonly, children } from '@nozbe/watermelondb/decorators';

export class Session extends Model {
  static table = 'sessions';
  static associations = {
    exercise_sets: { type: 'has_many', foreignKey: 'session_id' },
  };

  @field('athlete_id') athleteId!: string;
  @field('date') date!: string;
  @date('start_time') startTime!: Date;
  @date('end_time') endTime?: Date;
  @field('session_type') sessionType!: string;
  @field('overall_rpe') overallRpe?: number;
  @field('notes') notes?: string;
  @field('total_volume_load') totalVolumeLoad!: number;
  @field('synced') synced!: boolean;
  @readonly @date('created_at') createdAt!: Date;
  @readonly @date('updated_at') updatedAt!: Date;

  @children('exercise_sets') sets: any;
}
```

**Database Schema:**
```typescript
// src/services/database/schema.ts
import { appSchema, tableSchema } from '@nozbe/watermelondb';

export const schema = appSchema({
  version: 1,
  tables: [
    tableSchema({
      name: 'sessions',
      columns: [
        { name: 'athlete_id', type: 'string', isIndexed: true },
        { name: 'date', type: 'string', isIndexed: true },
        { name: 'start_time', type: 'number' },
        { name: 'end_time', type: 'number', isOptional: true },
        { name: 'session_type', type: 'string' },
        { name: 'overall_rpe', type: 'number', isOptional: true },
        { name: 'notes', type: 'string', isOptional: true },
        { name: 'total_volume_load', type: 'number' },
        { name: 'synced', type: 'boolean' },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
      ],
    }),
    tableSchema({
      name: 'exercise_sets',
      columns: [
        { name: 'session_id', type: 'string', isIndexed: true },
        { name: 'exercise_name', type: 'string', isIndexed: true },
        { name: 'exercise_category', type: 'string' },
        { name: 'set_number', type: 'number' },
        { name: 'weight', type: 'number' },
        { name: 'reps', type: 'number' },
        { name: 'rpe', type: 'number', isOptional: true },
        { name: 'velocity', type: 'number', isOptional: true },
        { name: 'notes', type: 'string', isOptional: true },
        { name: 'video_id', type: 'string', isOptional: true },
        { name: 'synced', type: 'boolean' },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
      ],
    }),
    tableSchema({
      name: 'one_rep_maxes',
      columns: [
        { name: 'athlete_id', type: 'string', isIndexed: true },
        { name: 'exercise', type: 'string', isIndexed: true },
        { name: 'weight', type: 'number' },
        { name: 'date', type: 'string' },
        { name: 'notes', type: 'string', isOptional: true },
        { name: 'synced', type: 'boolean' },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
      ],
    }),
    tableSchema({
      name: 'videos',
      columns: [
        { name: 'athlete_id', type: 'string', isIndexed: true },
        { name: 'session_id', type: 'string', isOptional: true },
        { name: 'exercise_set_id', type: 'string', isOptional: true },
        { name: 'exercise_name', type: 'string' },
        { name: 'weight', type: 'number', isOptional: true },
        { name: 'local_uri', type: 'string' },
        { name: 's3_key', type: 'string', isOptional: true },
        { name: 'thumbnail_uri', type: 'string', isOptional: true },
        { name: 'duration', type: 'number' },
        { name: 'file_size', type: 'number' },
        { name: 'upload_status', type: 'string' }, // 'pending', 'uploading', 'completed', 'failed'
        { name: 'quality_rating', type: 'number', isOptional: true },
        { name: 'notes', type: 'string', isOptional: true },
        { name: 'synced', type: 'boolean' },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
      ],
    }),
  ],
});
```

### 6.5 Service Layer Architecture

**API Service Pattern:**
```typescript
// src/services/api/sessionService.ts
import { apiClient } from './client';
import { database } from '@/services/database';
import { Session } from '@/services/database/models';

class SessionService {
  async fetchSessions(params: FetchSessionsParams): Promise<Session[]> {
    try {
      // Try API first
      const response = await apiClient.get('/sessions', { params });

      // Update local database
      await database.write(async () => {
        const sessionsCollection = database.get<Session>('sessions');
        for (const sessionData of response.data.sessions) {
          await sessionsCollection.create(session => {
            Object.assign(session, sessionData);
            session.synced = true;
          });
        }
      });

      return response.data.sessions;
    } catch (error) {
      // Fallback to local database if offline
      if (error.message === 'Network Error') {
        const localSessions = await database
          .get<Session>('sessions')
          .query()
          .fetch();
        return localSessions;
      }
      throw error;
    }
  }

  async createSession(sessionData: CreateSessionDTO): Promise<Session> {
    // Create locally first (optimistic UI)
    const localSession = await database.write(async () => {
      const sessionsCollection = database.get<Session>('sessions');
      return await sessionsCollection.create(session => {
        Object.assign(session, sessionData);
        session.synced = false;
      });
    });

    // Queue for sync
    await syncQueue.add({
      type: 'CREATE_SESSION',
      data: localSession,
    });

    return localSession;
  }
}

export const sessionService = new SessionService();
```

### 6.6 Component Architecture

**Screen Component Example:**
```typescript
// src/features/workout/screens/SessionListScreen.tsx
import React, { useEffect } from 'react';
import { FlatList, RefreshControl } from 'react-native';
import { useAppDispatch, useAppSelector } from '@/app/hooks';
import { fetchSessions } from '../sessionsSlice';
import { SessionCard } from '../components/SessionCard';
import { EmptyState } from '@/shared/components/EmptyState';

export function SessionListScreen({ navigation }) {
  const dispatch = useAppDispatch();
  const { sessions, loading } = useAppSelector(state => state.sessions);
  const { syncStatus } = useAppSelector(state => state.sync);

  useEffect(() => {
    dispatch(fetchSessions({ limit: 30 }));
  }, [dispatch]);

  const handleRefresh = () => {
    dispatch(fetchSessions({ limit: 30 }));
  };

  const handleSessionPress = (session: Session) => {
    navigation.navigate('SessionDetail', { sessionId: session.id });
  };

  return (
    <FlatList
      data={sessions}
      keyExtractor={item => item.id}
      renderItem={({ item }) => (
        <SessionCard
          session={item}
          onPress={() => handleSessionPress(item)}
        />
      )}
      refreshControl={
        <RefreshControl
          refreshing={loading || syncStatus === 'syncing'}
          onRefresh={handleRefresh}
        />
      }
      ListEmptyComponent={
        <EmptyState
          title="No sessions yet"
          message="Log your first workout to get started"
          actionLabel="Create Session"
          onAction={() => navigation.navigate('CreateSession')}
        />
      }
    />
  );
}
```

**Reusable Component Example:**
```typescript
// src/shared/components/SetInput.tsx
import React from 'react';
import { View, TextInput, StyleSheet } from 'react-native';
import { Controller, useFormContext } from 'react-hook-form';
import { theme } from '@/config/theme';

interface SetInputProps {
  setIndex: number;
  onRemove: () => void;
}

export function SetInput({ setIndex, onRemove }: SetInputProps) {
  const { control } = useFormContext();

  return (
    <View style={styles.container}>
      <Controller
        control={control}
        name={`sets.${setIndex}.weight`}
        rules={{ required: true, min: 20, max: 300 }}
        render={({ field: { onChange, value }, fieldState: { error } }) => (
          <TextInput
            style={[styles.input, error && styles.inputError]}
            placeholder="Weight (kg)"
            keyboardType="decimal-pad"
            value={value?.toString()}
            onChangeText={onChange}
          />
        )}
      />
      <Controller
        control={control}
        name={`sets.${setIndex}.reps`}
        rules={{ required: true, min: 1, max: 20 }}
        render={({ field: { onChange, value }, fieldState: { error } }) => (
          <TextInput
            style={[styles.input, error && styles.inputError]}
            placeholder="Reps"
            keyboardType="number-pad"
            value={value?.toString()}
            onChangeText={onChange}
          />
        )}
      />
      {/* RPE input, remove button, etc. */}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    gap: theme.spacing.md,
    marginBottom: theme.spacing.sm,
  },
  input: {
    flex: 1,
    height: 48,
    borderWidth: 1,
    borderColor: theme.colors.border,
    borderRadius: theme.borderRadius.md,
    paddingHorizontal: theme.spacing.md,
    fontSize: 16,
  },
  inputError: {
    borderColor: theme.colors.error,
  },
});
```

### 6.7 Performance Optimization Strategies

**Image Optimization:**
```typescript
// Thumbnail caching
import FastImage from 'react-native-fast-image';

<FastImage
  source={{
    uri: video.thumbnailUrl,
    priority: FastImage.priority.normal,
    cache: FastImage.cacheControl.immutable,
  }}
  style={styles.thumbnail}
  resizeMode={FastImage.resizeMode.cover}
/>
```

**List Virtualization:**
```typescript
// Use FlatList with proper optimization
<FlatList
  data={sessions}
  renderItem={renderItem}
  keyExtractor={item => item.id}
  initialNumToRender={10}
  maxToRenderPerBatch={10}
  updateCellsBatchingPeriod={50}
  windowSize={10}
  removeClippedSubviews={true}
  getItemLayout={(data, index) => ({
    length: 100, // Fixed height for performance
    offset: 100 * index,
    index,
  })}
/>
```

**Memoization:**
```typescript
// Expensive calculations
const acwr = useMemo(() => {
  return calculateACWR(weeklyVolume);
}, [weeklyVolume]);

// Component memoization
export const SessionCard = React.memo(({ session, onPress }) => {
  return (
    <Pressable onPress={() => onPress(session)}>
      {/* ... */}
    </Pressable>
  );
}, (prev, next) => prev.session.id === next.session.id);
```

**Code Splitting:**
```typescript
// Lazy load heavy screens
const VideoPlayerScreen = React.lazy(() =>
  import('./screens/VideoPlayerScreen')
);

// Use Suspense
<Suspense fallback={<LoadingSpinner />}>
  <VideoPlayerScreen />
</Suspense>
```

---

## 7. Video Processing Pipeline

### 7.1 Video Upload Flow

```
Mobile App                    API Server                 AWS Services
    |                              |                           |
    | 1. Request upload URL        |                           |
    |----------------------------->|                           |
    |                              | 2. Generate presigned URL |
    |                              |-------------------------->|
    |                              |                           | S3
    | 3. Presigned URL + metadata  |                           |
    |<-----------------------------|                           |
    |                              |                           |
    | 4. Upload video to S3        |                           |
    |------------------------------------------------->|       |
    |                              |                   |       |
    | 5. Notify upload complete    |                   |       |
    |----------------------------->|                   |       |
    |                              | 6. Queue processing job   |
    |                              |-------------------------->|
    |                              |                   | Lambda |
    |                              |                           |
    | 7. Poll for status           |                           |
    |----------------------------->|                           |
    |                              |                           | 8. Process video
    |                              |                           |    (compress, thumbnail)
    |                              |                           |
    |                              | 9. Update video status    |
    |                              |<--------------------------|
    | 10. Video ready response     |                           |
    |<-----------------------------|                           |
```

### 7.2 Video Compression Strategy

**Target Specifications:**
- **Resolution:** Max 1080p (4K downsampled)
- **Bitrate:** 2-4 Mbps (variable bitrate)
- **Frame Rate:** Preserve original (up to 240fps)
- **Codec:** H.264 (broad compatibility)
- **Container:** MP4
- **Audio:** AAC 128kbps (if present)
- **File Size Reduction:** 60-70%

**FFmpeg Command:**
```bash
ffmpeg -i input.mp4 \
  -vf "scale='min(1920,iw)':'min(1080,ih)':force_original_aspect_ratio=decrease" \
  -c:v libx264 \
  -preset medium \
  -crf 23 \
  -maxrate 4M \
  -bufsize 8M \
  -movflags +faststart \
  -c:a aac \
  -b:a 128k \
  -r $(ffprobe -v error -select_streams v:0 -show_entries stream=r_frame_rate -of csv=p=0 input.mp4) \
  output.mp4
```

**Parameter Explanation:**
- `scale`: Maintain aspect ratio, max 1080p
- `-preset medium`: Balance between speed and compression
- `-crf 23`: Constant Rate Factor (18-28 range, 23 = good quality/size)
- `-maxrate 4M`: Maximum bitrate cap
- `-movflags +faststart`: Enable progressive playback
- `-r`: Preserve original frame rate

### 7.3 Thumbnail Generation

**Strategy:**
- Extract frame at 2-second mark (or 25% of video duration if < 8s)
- Generate 3 sizes: small (160x90), medium (320x180), large (640x360)
- Save as JPEG with 80% quality

**FFmpeg Command:**
```bash
# Extract thumbnail at 2 seconds
ffmpeg -ss 2 -i input.mp4 -vframes 1 -vf "scale=640:360" -q:v 2 thumbnail.jpg
```

### 7.4 Lambda Video Processing Function

```typescript
// lambda/videoProcessor/index.ts
import { S3Event } from 'aws-lambda';
import { S3Client, GetObjectCommand, PutObjectCommand } from '@aws-sdk/client-s3';
import { exec } from 'child_process';
import { promisify } from 'util';
import fs from 'fs';
import path from 'path';

const execAsync = promisify(exec);
const s3Client = new S3Client({ region: 'eu-west-1' });

export async function handler(event: S3Event) {
  for (const record of event.Records) {
    const bucket = record.s3.bucket.name;
    const key = decodeURIComponent(record.s3.object.key.replace(/\+/g, ' '));

    try {
      // Download video from S3
      const inputPath = `/tmp/${path.basename(key)}`;
      await downloadFromS3(bucket, key, inputPath);

      // Process video
      const outputPath = `/tmp/processed_${path.basename(key)}`;
      await compressVideo(inputPath, outputPath);

      // Generate thumbnail
      const thumbnailPath = `/tmp/thumb_${path.basename(key, '.mp4')}.jpg`;
      await generateThumbnail(inputPath, thumbnailPath);

      // Upload processed files
      const processedKey = key.replace('/raw/', '/processed/');
      const thumbnailKey = key.replace('/raw/', '/thumbnails/').replace('.mp4', '.jpg');

      await uploadToS3(bucket, processedKey, outputPath);
      await uploadToS3(bucket, thumbnailKey, thumbnailPath);

      // Update database
      await updateVideoStatus(getVideoIdFromKey(key), {
        processingStatus: 'completed',
        s3Key: processedKey,
        thumbnailS3Key: thumbnailKey,
        fileSize: fs.statSync(outputPath).size,
      });

      // Cleanup
      fs.unlinkSync(inputPath);
      fs.unlinkSync(outputPath);
      fs.unlinkSync(thumbnailPath);

    } catch (error) {
      console.error(`Error processing video ${key}:`, error);
      await updateVideoStatus(getVideoIdFromKey(key), {
        processingStatus: 'failed',
        error: error.message,
      });
    }
  }
}

async function compressVideo(inputPath: string, outputPath: string): Promise<void> {
  const command = `
    ffmpeg -i ${inputPath} \
      -vf "scale='min(1920,iw)':'min(1080,ih)':force_original_aspect_ratio=decrease" \
      -c:v libx264 \
      -preset medium \
      -crf 23 \
      -maxrate 4M \
      -bufsize 8M \
      -movflags +faststart \
      -c:a aac \
      -b:a 128k \
      ${outputPath}
  `;

  await execAsync(command);
}

async function generateThumbnail(inputPath: string, outputPath: string): Promise<void> {
  const command = `
    ffmpeg -ss 2 -i ${inputPath} \
      -vframes 1 \
      -vf "scale=640:360" \
      -q:v 2 \
      ${outputPath}
  `;

  await execAsync(command);
}

async function downloadFromS3(bucket: string, key: string, outputPath: string): Promise<void> {
  const command = new GetObjectCommand({ Bucket: bucket, Key: key });
  const response = await s3Client.send(command);
  const stream = response.Body as NodeJS.ReadableStream;
  const writeStream = fs.createWriteStream(outputPath);

  return new Promise((resolve, reject) => {
    stream.pipe(writeStream);
    writeStream.on('finish', resolve);
    writeStream.on('error', reject);
  });
}

async function uploadToS3(bucket: string, key: string, filePath: string): Promise<void> {
  const fileContent = fs.readFileSync(filePath);
  const command = new PutObjectCommand({
    Bucket: bucket,
    Key: key,
    Body: fileContent,
    ContentType: key.endsWith('.jpg') ? 'image/jpeg' : 'video/mp4',
  });

  await s3Client.send(command);
}
```

### 7.5 Video Storage Structure

**S3 Bucket Organization:**
```
ascend-videos-eu/
├── raw/                          # Uploaded videos (temporary)
│   └── athlete_{id}/
│       └── video_{id}.mp4
├── processed/                    # Compressed videos
│   └── athlete_{id}/
│       └── video_{id}.mp4
├── thumbnails/                   # Generated thumbnails
│   └── athlete_{id}/
│       └── video_{id}.jpg
└── annotations/                  # Saved annotated frames
    └── athlete_{id}/
        └── video_{id}_frame_{n}.jpg
```

**S3 Lifecycle Policy:**
```json
{
  "Rules": [
    {
      "Id": "DeleteRawVideosAfter7Days",
      "Status": "Enabled",
      "Prefix": "raw/",
      "Expiration": {
        "Days": 7
      }
    },
    {
      "Id": "TransitionProcessedToIA",
      "Status": "Enabled",
      "Prefix": "processed/",
      "Transitions": [
        {
          "Days": 90,
          "StorageClass": "STANDARD_IA"
        }
      ]
    }
  ]
}
```

### 7.6 CloudFront CDN Configuration

**Distribution Settings:**
- **Origin:** S3 bucket (ascend-videos-eu)
- **Price Class:** Use Only Europe and North America (optimize for EMEA)
- **Caching:** Cache for 24 hours (videos don't change)
- **Compression:** Enable automatic compression for text/JSON responses
- **SSL:** CloudFront default SSL certificate (*.cloudfront.net)

**Signed URLs:**
```typescript
// Generate signed CloudFront URL
import { getSignedUrl } from '@aws-sdk/cloudfront-signer';

function getVideoUrl(videoKey: string): string {
  const url = `https://${process.env.CLOUDFRONT_DOMAIN}/${videoKey}`;

  return getSignedUrl({
    url,
    keyPairId: process.env.CLOUDFRONT_KEY_PAIR_ID,
    privateKey: process.env.CLOUDFRONT_PRIVATE_KEY,
    dateLessThan: new Date(Date.now() + 3600000).toISOString(), // 1 hour expiry
  });
}
```

### 7.7 Mobile Video Player Implementation

```typescript
// src/features/video/components/VideoPlayer.tsx
import React, { useRef, useState } from 'react';
import { View, StyleSheet, Pressable } from 'react-native';
import Video from 'react-native-video';
import Slider from '@react-native-community/slider';

interface VideoPlayerProps {
  videoUrl: string;
  onEnd?: () => void;
}

export function VideoPlayer({ videoUrl, onEnd }: VideoPlayerProps) {
  const videoRef = useRef<Video>(null);
  const [paused, setPaused] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const [playbackRate, setPlaybackRate] = useState(1.0);

  const handleProgress = (data: { currentTime: number }) => {
    setCurrentTime(data.currentTime);
  };

  const handleLoad = (data: { duration: number }) => {
    setDuration(data.duration);
  };

  const handleSeek = (value: number) => {
    videoRef.current?.seek(value);
  };

  const togglePlayback = () => {
    setPaused(!paused);
  };

  const changePlaybackRate = () => {
    const rates = [0.25, 0.5, 0.75, 1.0];
    const currentIndex = rates.indexOf(playbackRate);
    const nextIndex = (currentIndex + 1) % rates.length;
    setPlaybackRate(rates[nextIndex]);
  };

  return (
    <View style={styles.container}>
      <Pressable onPress={togglePlayback}>
        <Video
          ref={videoRef}
          source={{ uri: videoUrl }}
          style={styles.video}
          resizeMode="contain"
          paused={paused}
          rate={playbackRate}
          onProgress={handleProgress}
          onLoad={handleLoad}
          onEnd={onEnd}
          controls={false}
        />
      </Pressable>

      <View style={styles.controls}>
        <Slider
          style={styles.slider}
          value={currentTime}
          minimumValue={0}
          maximumValue={duration}
          onSlidingComplete={handleSeek}
        />

        <View style={styles.buttons}>
          <Pressable onPress={togglePlayback} style={styles.button}>
            {/* Play/Pause Icon */}
          </Pressable>

          <Pressable onPress={changePlaybackRate} style={styles.button}>
            {/* Speed: {playbackRate}x */}
          </Pressable>
        </View>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    width: '100%',
    aspectRatio: 16 / 9,
    backgroundColor: '#000',
  },
  video: {
    width: '100%',
    height: '100%',
  },
  controls: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
    padding: 10,
  },
  slider: {
    width: '100%',
    height: 40,
  },
  buttons: {
    flexDirection: 'row',
    justifyContent: 'space-around',
  },
  button: {
    padding: 10,
  },
});
```

---

## 8. Offline-First Synchronization

### 8.1 Sync Architecture Overview

**Core Principles:**
1. **Local-first:** All writes go to local database first
2. **Optimistic UI:** Show changes immediately, sync in background
3. **Delta sync:** Only send changed data, not entire dataset
4. **Conflict resolution:** Last-write-wins for MVP (CRDTs for post-MVP)
5. **Background sync:** Queue operations, retry with exponential backoff

### 8.2 Sync State Machine

```
┌──────────┐
│   IDLE   │◄──────────────────┐
└────┬─────┘                   │
     │ Trigger                 │
     │ (app open, pull refresh)│
     ▼                         │
┌──────────┐                   │
│ FETCHING │                   │
│  DELTA   │                   │
└────┬─────┘                   │
     │                         │
     ▼                         │
┌──────────┐                   │
│ APPLYING │                   │
│ CHANGES  │                   │
└────┬─────┘                   │
     │                         │
     ▼                         │
┌──────────┐                   │
│ PUSHING  │                   │
│  LOCAL   │                   │
│ CHANGES  │                   │
└────┬─────┘                   │
     │                         │
     ▼                         │
┌──────────┐   Success         │
│  SYNCED  │───────────────────┘
└────┬─────┘
     │ Error
     ▼
┌──────────┐   Retry Queue
│  ERROR   │──────────────┐
└────┬─────┘              │
     │                    │
     │ Manual Retry or    │
     │ Auto-retry after   │
     │ exponential backoff│
     └────────────────────┘
```

### 8.3 Sync Implementation

**Sync Service:**
```typescript
// src/services/sync/syncService.ts
import { database } from '@/services/database';
import { apiClient } from '@/services/api/client';
import { SyncQueue } from './syncQueue';

interface SyncOperation {
  id: string;
  type: 'CREATE' | 'UPDATE' | 'DELETE';
  entity: 'session' | 'set' | 'video' | '1rm';
  localId: string;
  data: any;
  timestamp: number;
  retryCount: number;
}

class SyncService {
  private queue: SyncQueue;
  private syncInterval: NodeJS.Timeout | null = null;

  constructor() {
    this.queue = new SyncQueue();
  }

  async startAutoSync(intervalMs: number = 30000) {
    this.syncInterval = setInterval(() => {
      this.sync();
    }, intervalMs);
  }

  stopAutoSync() {
    if (this.syncInterval) {
      clearInterval(this.syncInterval);
      this.syncInterval = null;
    }
  }

  async sync(): Promise<SyncResult> {
    try {
      // Step 1: Fetch server changes since last sync
      const lastSyncTimestamp = await this.getLastSyncTimestamp();
      const serverChanges = await this.fetchServerChanges(lastSyncTimestamp);

      // Step 2: Apply server changes to local database
      await this.applyServerChanges(serverChanges);

      // Step 3: Push local changes to server
      const localChanges = await this.getLocalChanges();
      await this.pushLocalChanges(localChanges);

      // Step 4: Update last sync timestamp
      await this.setLastSyncTimestamp(Date.now());

      return { success: true, conflicts: [] };
    } catch (error) {
      console.error('Sync failed:', error);
      return { success: false, error: error.message };
    }
  }

  private async fetchServerChanges(since: number): Promise<ServerChanges> {
    const response = await apiClient.get('/sync/changes', {
      params: { since },
    });
    return response.data;
  }

  private async applyServerChanges(changes: ServerChanges): Promise<void> {
    await database.write(async () => {
      // Apply sessions
      for (const session of changes.sessions) {
        const existingSession = await database
          .get('sessions')
          .find(session.id)
          .catch(() => null);

        if (existingSession) {
          // Update
          await existingSession.update(s => {
            Object.assign(s, session);
            s.synced = true;
          });
        } else {
          // Create
          await database.get('sessions').create(s => {
            Object.assign(s, session);
            s.synced = true;
          });
        }
      }

      // Apply exercise sets
      for (const set of changes.exerciseSets) {
        // Similar logic...
      }

      // Apply 1RMs
      for (const oneRM of changes.oneRepMaxes) {
        // Similar logic...
      }
    });
  }

  private async getLocalChanges(): Promise<SyncOperation[]> {
    const sessions = await database
      .get('sessions')
      .query(Q.where('synced', false))
      .fetch();

    const sets = await database
      .get('exercise_sets')
      .query(Q.where('synced', false))
      .fetch();

    const oneRMs = await database
      .get('one_rep_maxes')
      .query(Q.where('synced', false))
      .fetch();

    return [
      ...sessions.map(s => ({ type: 'CREATE', entity: 'session', data: s })),
      ...sets.map(s => ({ type: 'CREATE', entity: 'set', data: s })),
      ...oneRMs.map(o => ({ type: 'CREATE', entity: '1rm', data: o })),
    ];
  }

  private async pushLocalChanges(changes: SyncOperation[]): Promise<void> {
    for (const change of changes) {
      try {
        switch (change.entity) {
          case 'session':
            await this.pushSession(change);
            break;
          case 'set':
            await this.pushSet(change);
            break;
          case '1rm':
            await this.push1RM(change);
            break;
        }
      } catch (error) {
        // Queue for retry
        await this.queue.add(change);
      }
    }
  }

  private async pushSession(change: SyncOperation): Promise<void> {
    const response = await apiClient.post('/sessions', change.data);
    const serverSession = response.data;

    // Update local record with server ID
    await database.write(async () => {
      const localSession = await database.get('sessions').find(change.localId);
      await localSession.update(s => {
        s.id = serverSession.id; // Update with server-generated UUID
        s.synced = true;
      });
    });
  }

  private async getLastSyncTimestamp(): Promise<number> {
    // Retrieve from AsyncStorage
    const timestamp = await AsyncStorage.getItem('lastSyncTimestamp');
    return timestamp ? parseInt(timestamp, 10) : 0;
  }

  private async setLastSyncTimestamp(timestamp: number): Promise<void> {
    await AsyncStorage.setItem('lastSyncTimestamp', timestamp.toString());
  }
}

export const syncService = new SyncService();
```

**Sync Queue with Retry Logic:**
```typescript
// src/services/sync/syncQueue.ts
import AsyncStorage from '@react-native-async-storage/async-storage';

const QUEUE_KEY = 'sync_queue';
const MAX_RETRIES = 3;

export class SyncQueue {
  private queue: SyncOperation[] = [];

  async load(): Promise<void> {
    const queueData = await AsyncStorage.getItem(QUEUE_KEY);
    this.queue = queueData ? JSON.parse(queueData) : [];
  }

  async save(): Promise<void> {
    await AsyncStorage.setItem(QUEUE_KEY, JSON.stringify(this.queue));
  }

  async add(operation: SyncOperation): Promise<void> {
    this.queue.push({
      ...operation,
      retryCount: 0,
    });
    await this.save();
  }

  async processQueue(): Promise<void> {
    const operations = [...this.queue];
    this.queue = [];

    for (const operation of operations) {
      try {
        await this.processOperation(operation);
      } catch (error) {
        if (operation.retryCount < MAX_RETRIES) {
          // Retry with exponential backoff
          const backoffMs = Math.pow(2, operation.retryCount) * 1000;
          setTimeout(() => {
            this.add({
              ...operation,
              retryCount: operation.retryCount + 1,
            });
          }, backoffMs);
        } else {
          // Max retries exceeded, log error
          console.error('Sync operation failed after max retries:', operation);
        }
      }
    }

    await this.save();
  }

  private async processOperation(operation: SyncOperation): Promise<void> {
    // Execute the operation (API call)
    switch (operation.type) {
      case 'CREATE':
        await apiClient.post(`/${operation.entity}s`, operation.data);
        break;
      case 'UPDATE':
        await apiClient.put(`/${operation.entity}s/${operation.localId}`, operation.data);
        break;
      case 'DELETE':
        await apiClient.delete(`/${operation.entity}s/${operation.localId}`);
        break;
    }
  }
}
```

### 8.4 Conflict Resolution

**Strategy: Last-Write-Wins (MVP)**
```typescript
function resolveConflict(localRecord: any, serverRecord: any): any {
  // Compare timestamps
  if (localRecord.updatedAt > serverRecord.updatedAt) {
    return localRecord; // Local is newer, push to server
  } else {
    return serverRecord; // Server is newer, apply locally
  }
}
```

**Future: Operational Transformation (Post-MVP)**
- For collaborative features, implement CRDTs or OT algorithms
- Handle concurrent edits to same session by multiple users (coach + athlete)

### 8.5 Network State Handling

```typescript
// src/services/network/networkService.ts
import NetInfo from '@react-native-community/netinfo';
import { EventEmitter } from 'events';

class NetworkService extends EventEmitter {
  private isConnected: boolean = true;

  constructor() {
    super();
    this.initialize();
  }

  initialize() {
    NetInfo.addEventListener(state => {
      const wasConnected = this.isConnected;
      this.isConnected = state.isConnected ?? false;

      if (this.isConnected && !wasConnected) {
        // Network restored, trigger sync
        this.emit('connected');
        syncService.sync();
      } else if (!this.isConnected && wasConnected) {
        // Network lost
        this.emit('disconnected');
      }
    });
  }

  getConnectionStatus(): boolean {
    return this.isConnected;
  }
}

export const networkService = new NetworkService();
```

### 8.6 User Feedback for Sync Status

```typescript
// src/shared/components/SyncIndicator.tsx
import React from 'react';
import { View, Text, ActivityIndicator } from 'react-native';
import { useAppSelector } from '@/app/hooks';

export function SyncIndicator() {
  const { syncStatus, lastSyncedAt } = useAppSelector(state => state.sync);
  const { isConnected } = useAppSelector(state => state.network);

  if (!isConnected) {
    return (
      <View style={styles.container}>
        <OfflineIcon />
        <Text>Offline - Changes saved locally</Text>
      </View>
    );
  }

  if (syncStatus === 'syncing') {
    return (
      <View style={styles.container}>
        <ActivityIndicator size="small" />
        <Text>Syncing...</Text>
      </View>
    );
  }

  if (syncStatus === 'synced' && lastSyncedAt) {
    return (
      <View style={styles.container}>
        <CheckIcon />
        <Text>Synced {formatRelativeTime(lastSyncedAt)}</Text>
      </View>
    );
  }

  if (syncStatus === 'error') {
    return (
      <View style={styles.container}>
        <ErrorIcon />
        <Text>Sync failed - Tap to retry</Text>
      </View>
    );
  }

  return null;
}
```

---

## 9. Authentication & Security

### 9.1 Authentication Flow

```
Mobile App                    API Server                  Database
    |                              |                           |
    | 1. POST /auth/login          |                           |
    |   { email, password }        |                           |
    |----------------------------->|                           |
    |                              | 2. Query user by email    |
    |                              |-------------------------->|
    |                              | 3. User record            |
    |                              |<--------------------------|
    |                              | 4. Verify password hash   |
    |                              |    (bcrypt.compare)       |
    |                              |                           |
    |                              | 5. Generate access token  |
    |                              |    (JWT, 15min expiry)    |
    |                              |                           |
    |                              | 6. Generate refresh token |
    |                              |    (UUID, 7day expiry)    |
    |                              |                           |
    |                              | 7. Store refresh token    |
    |                              |-------------------------->|
    |                              |                           |
    | 8. Return tokens             |                           |
    |<-----------------------------|                           |
    | { accessToken, refreshToken }|                           |
    |                              |                           |
    | 9. Store tokens securely     |                           |
    |    (Keychain/Keystore)       |                           |
    |                              |                           |
```

### 9.2 JWT Access Token Structure

```go
// JWTClaims represents the JWT payload
type JWTClaims struct {
    AthleteID string `json:"sub"`   // Athlete ID (UUID)
    Email     string `json:"email"` // Athlete email
    Type      string `json:"type"`  // Token type ("access")
    jwt.RegisteredClaims
}

// Example JWT payload
{
  "sub": "550e8400-e29b-41d4-a716-446655440000",
  "email": "athlete@example.com",
  "iat": 1696593600,
  "exp": 1696594500,
  "type": "access"
}
```

**Token Generation:**
```go
// internal/service/token_service.go
package service

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
)

type TokenService struct {
    accessTokenSecret  string
    refreshTokenSecret string
    refreshTokenRepo   repository.RefreshTokenRepository
}

func NewTokenService(
    accessSecret string,
    refreshSecret string,
    refreshRepo repository.RefreshTokenRepository,
) *TokenService {
    return &TokenService{
        accessTokenSecret:  accessSecret,
        refreshTokenSecret: refreshSecret,
        refreshTokenRepo:   refreshRepo,
    }
}

// GenerateAccessToken creates a JWT access token (15 minutes)
func (s *TokenService) GenerateAccessToken(athleteID, email string) (string, error) {
    claims := JWTClaims{
        AthleteID: athleteID,
        Email:     email,
        Type:      "access",
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.accessTokenSecret))
}

// GenerateRefreshToken creates a random refresh token (7 days)
func (s *TokenService) GenerateRefreshToken(athleteID string) (string, error) {
    // Generate random token
    tokenBytes := make([]byte, 32)
    if _, err := rand.Read(tokenBytes); err != nil {
        return "", fmt.Errorf("failed to generate refresh token: %w", err)
    }
    token := hex.EncodeToString(tokenBytes)

    // Hash token for database storage
    hash := sha256.Sum256([]byte(token))
    tokenHash := hex.EncodeToString(hash[:])

    // Store in database
    expiresAt := time.Now().Add(7 * 24 * time.Hour)
    if err := s.refreshTokenRepo.Create(athleteID, tokenHash, expiresAt); err != nil {
        return "", fmt.Errorf("failed to store refresh token: %w", err)
    }

    return token, nil
}

// VerifyAccessToken validates and parses JWT access token
func (s *TokenService) VerifyAccessToken(tokenString string) (*JWTClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(s.accessTokenSecret), nil
    })

    if err != nil {
        return nil, fmt.Errorf("invalid token: %w", err)
    }

    if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
        return claims, nil
    }

    return nil, fmt.Errorf("invalid token claims")
}

// VerifyRefreshToken validates refresh token and returns athlete ID
func (s *TokenService) VerifyRefreshToken(token string) (string, error) {
    // Hash token
    hash := sha256.Sum256([]byte(token))
    tokenHash := hex.EncodeToString(hash[:])

    // Verify in database
    refreshToken, err := s.refreshTokenRepo.FindByHash(tokenHash)
    if err != nil {
        return "", fmt.Errorf("invalid or expired refresh token: %w", err)
    }

    if refreshToken.RevokedAt != nil {
        return "", fmt.Errorf("refresh token has been revoked")
    }

    if time.Now().After(refreshToken.ExpiresAt) {
        return "", fmt.Errorf("refresh token has expired")
    }

    return refreshToken.AthleteID, nil
}
```

### 9.3 Password Security

**Hashing Strategy:**
- Algorithm: bcrypt
- Cost factor: 12 (2^12 = 4096 iterations)
- Salt: Generated automatically by bcrypt

```go
// internal/service/password_service.go
package service

import (
    "fmt"
    "regexp"

    "golang.org/x/crypto/bcrypt"
)

type PasswordService struct {
    cost int // bcrypt cost factor
}

func NewPasswordService() *PasswordService {
    return &PasswordService{
        cost: 12, // 2^12 = 4096 iterations
    }
}

// HashPassword generates bcrypt hash from plain password
func (s *PasswordService) HashPassword(plainPassword string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(plainPassword), s.cost)
    if err != nil {
        return "", fmt.Errorf("failed to hash password: %w", err)
    }
    return string(hash), nil
}

// ComparePassword verifies plain password against hash
func (s *PasswordService) ComparePassword(plainPassword, hash string) error {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainPassword))
}

// ValidatePasswordStrength checks password meets security requirements
func (s *PasswordService) ValidatePasswordStrength(password string) error {
    if len(password) < 8 {
        return fmt.Errorf("password must be at least 8 characters")
    }

    // At least 1 uppercase
    if matched, _ := regexp.MatchString(`[A-Z]`, password); !matched {
        return fmt.Errorf("password must contain at least one uppercase letter")
    }

    // At least 1 lowercase
    if matched, _ := regexp.MatchString(`[a-z]`, password); !matched {
        return fmt.Errorf("password must contain at least one lowercase letter")
    }

    // At least 1 digit
    if matched, _ := regexp.MatchString(`\d`, password); !matched {
        return fmt.Errorf("password must contain at least one number")
    }

    // At least 1 special character
    if matched, _ := regexp.MatchString(`[@$!%*?&]`, password); !matched {
        return fmt.Errorf("password must contain at least one special character (@$!%%*?&)")
    }

    return nil
}
```

### 9.4 API Authentication Middleware

```typescript
// src/middleware/auth.ts
import { Request, Response, NextFunction } from 'express';
import { tokenService } from '@/modules/auth/services/tokenService';

export interface AuthenticatedRequest extends Request {
  athleteId: string;
  email: string;
}

export async function authenticateToken(
  req: Request,
  res: Response,
  next: NextFunction
) {
  try {
    const authHeader = req.headers.authorization;

    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return res.status(401).json({
        error: {
          code: 'UNAUTHORIZED',
          message: 'Missing or invalid authorization header',
        },
      });
    }

    const token = authHeader.substring(7); // Remove 'Bearer ' prefix
    const payload = tokenService.verifyAccessToken(token);

    // Attach athlete info to request
    (req as AuthenticatedRequest).athleteId = payload.sub;
    (req as AuthenticatedRequest).email = payload.email;

    next();
  } catch (error) {
    return res.status(401).json({
      error: {
        code: 'UNAUTHORIZED',
        message: 'Invalid or expired token',
      },
    });
  }
}
```

### 9.5 Secure Token Storage (Mobile)

**iOS - Keychain:**
```typescript
// src/services/auth/secureStorage.ios.ts
import * as Keychain from 'react-native-keychain';

class SecureStorage {
  private readonly SERVICE_NAME = 'app.ascend';

  async storeTokens(accessToken: string, refreshToken: string): Promise<void> {
    await Keychain.setGenericPassword('tokens', JSON.stringify({
      accessToken,
      refreshToken,
    }), {
      service: this.SERVICE_NAME,
      accessible: Keychain.ACCESSIBLE.WHEN_UNLOCKED_THIS_DEVICE_ONLY,
    });
  }

  async getTokens(): Promise<{ accessToken: string; refreshToken: string } | null> {
    const credentials = await Keychain.getGenericPassword({
      service: this.SERVICE_NAME,
    });

    if (credentials) {
      return JSON.parse(credentials.password);
    }

    return null;
  }

  async clearTokens(): Promise<void> {
    await Keychain.resetGenericPassword({
      service: this.SERVICE_NAME,
    });
  }
}

export const secureStorage = new SecureStorage();
```

**Android - Keystore:**
```typescript
// src/services/auth/secureStorage.android.ts
// Similar implementation using Android Keystore via react-native-keychain
```

### 9.6 API Rate Limiting

```typescript
// src/middleware/rateLimiter.ts
import rateLimit from 'express-rate-limit';
import RedisStore from 'rate-limit-redis';
import { redisClient } from '@/config/redis';

// Authentication endpoints (stricter limits)
export const authLimiter = rateLimit({
  store: new RedisStore({
    client: redisClient,
    prefix: 'rl:auth:',
  }),
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 5, // 5 requests per window
  message: {
    error: {
      code: 'RATE_LIMIT_EXCEEDED',
      message: 'Too many authentication attempts. Please try again later.',
    },
  },
  standardHeaders: true,
  legacyHeaders: false,
});

// General API endpoints
export const apiLimiter = rateLimit({
  store: new RedisStore({
    client: redisClient,
    prefix: 'rl:api:',
  }),
  windowMs: 60 * 1000, // 1 minute
  max: 100, // 100 requests per minute
  message: {
    error: {
      code: 'RATE_LIMIT_EXCEEDED',
      message: 'Too many requests. Please slow down.',
    },
  },
  standardHeaders: true,
  legacyHeaders: false,
  keyGenerator: (req) => {
    // Use athlete ID if authenticated, otherwise IP
    return (req as AuthenticatedRequest).athleteId || req.ip;
  },
});

// Video upload endpoints (special limits)
export const uploadLimiter = rateLimit({
  store: new RedisStore({
    client: redisClient,
    prefix: 'rl:upload:',
  }),
  windowMs: 60 * 60 * 1000, // 1 hour
  max: 10, // 10 uploads per hour
  keyGenerator: (req) => (req as AuthenticatedRequest).athleteId,
});
```

### 9.7 HTTPS & TLS Configuration

**ALB HTTPS Listener:**
```typescript
// AWS CDK infrastructure code (example)
const listener = loadBalancer.addListener('HttpsListener', {
  port: 443,
  protocol: ApplicationProtocol.HTTPS,
  certificates: [certificate],
  sslPolicy: SslPolicy.TLS13_RES, // TLS 1.3 only
});

listener.addTargets('EcsTarget', {
  port: 3000,
  protocol: ApplicationProtocol.HTTP,
  targets: [ecsService],
  healthCheck: {
    path: '/health',
    interval: Duration.seconds(30),
  },
});
```

**Security Headers:**
```typescript
// src/middleware/securityHeaders.ts
import helmet from 'helmet';

export const securityHeaders = helmet({
  hsts: {
    maxAge: 31536000, // 1 year
    includeSubDomains: true,
    preload: true,
  },
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      styleSrc: ["'self'", "'unsafe-inline'"],
      scriptSrc: ["'self'"],
      imgSrc: ["'self'", 'data:', 'https://cdn.ascend.app'],
      connectSrc: ["'self'", 'https://api.ascend.app'],
      fontSrc: ["'self'"],
      objectSrc: ["'none'"],
      mediaSrc: ["'self'", 'https://cdn.ascend.app'],
      frameSrc: ["'none'"],
    },
  },
  referrerPolicy: { policy: 'strict-origin-when-cross-origin' },
  noSniff: true,
  xssFilter: true,
  hidePoweredBy: true,
});
```

### 9.8 GDPR Data Protection

**Data Encryption:**
- At rest: AES-256 encryption for RDS database
- In transit: TLS 1.3 for all API communication
- Video files: Server-side encryption (SSE-S3)

**Personal Data Handling:**
```typescript
// src/modules/athletes/services/gdprService.ts
class GDPRService {
  async exportAthleteData(athleteId: string): Promise<ExportData> {
    // Collect all athlete data
    const profile = await this.athleteRepository.findOne(athleteId);
    const sessions = await this.sessionRepository.find({ athleteId });
    const videos = await this.videoRepository.find({ athleteId });
    const oneRMs = await this.oneRMRepository.find({ athleteId });

    // Create export package (JSON format)
    return {
      exportDate: new Date(),
      athlete: profile,
      sessions,
      videos: videos.map(v => ({
        ...v,
        downloadUrl: this.generateVideoDownloadUrl(v.s3Key),
      })),
      oneRepMaxes: oneRMs,
    };
  }

  async deleteAthleteData(athleteId: string): Promise<void> {
    // Soft delete athlete record
    await this.athleteRepository.update(athleteId, {
      deletedAt: new Date(),
      email: `deleted_${athleteId}@ascend.app`, // Anonymize
      name: 'Deleted User',
    });

    // Schedule permanent deletion after 30 days
    await this.jobQueue.add({
      type: 'PERMANENT_DELETE',
      athleteId,
      executeAt: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000),
    });
  }

  async permanentlyDeleteAthleteData(athleteId: string): Promise<void> {
    // Delete all associated data
    await this.sessionRepository.delete({ athleteId });
    await this.oneRMRepository.delete({ athleteId });

    // Delete videos from S3
    const videos = await this.videoRepository.find({ athleteId });
    for (const video of videos) {
      await this.s3Client.deleteObject({
        Bucket: process.env.S3_BUCKET,
        Key: video.s3Key,
      });
      if (video.thumbnailS3Key) {
        await this.s3Client.deleteObject({
          Bucket: process.env.S3_BUCKET,
          Key: video.thumbnailS3Key,
        });
      }
    }
    await this.videoRepository.delete({ athleteId });

    // Delete athlete record
    await this.athleteRepository.delete(athleteId);

    // Log deletion for audit
    await this.auditLogRepository.create({
      action: 'PERMANENT_DELETION',
      athleteId,
      metadata: { reason: 'GDPR right to be forgotten' },
    });
  }
}

export const gdprService = new GDPRService();
```

---

## 10. Infrastructure & Deployment

### 10.1 Cloud Infrastructure (AWS eu-west-1)

**Regional Architecture:**
```
┌─────────────────────────────────────────────────────────────────┐
│                     AWS eu-west-1 (Ireland)                      │
│                                                                   │
│  ┌──────────────┐         ┌──────────────────┐                 │
│  │  CloudFront  │────────▶│  ALB (Application│                 │
│  │  (Global CDN)│         │  Load Balancer)  │                 │
│  └──────────────┘         └──────────────────┘                 │
│         │                          │                             │
│         │                          ▼                             │
│         │                  ┌──────────────────┐                 │
│         │                  │   ECS Fargate    │                 │
│         │                  │  (API Services)  │                 │
│         │                  │                  │                 │
│         │                  │  • Auth Service  │                 │
│         │                  │  • API Service   │                 │
│         │                  │  • Sync Service  │                 │
│         │                  └──────────────────┘                 │
│         │                          │                             │
│         │                          │                             │
│         ▼                          ▼                             │
│  ┌──────────────┐         ┌──────────────────┐                 │
│  │  S3 Bucket   │         │  RDS PostgreSQL  │                 │
│  │  (Videos)    │         │  (Multi-AZ)      │                 │
│  │              │         │                  │                 │
│  │  • Original  │         │  • Primary: AZ-A │                 │
│  │  • Compressed│         │  • Standby: AZ-B │                 │
│  │  • Thumbnails│         └──────────────────┘                 │
│  └──────────────┘                  │                             │
│         │                          │                             │
│         │                          ▼                             │
│         │                  ┌──────────────────┐                 │
│         │                  │  ElastiCache     │                 │
│         │                  │  (Redis Cluster) │                 │
│         │                  └──────────────────┘                 │
│         │                                                        │
│         ▼                                                        │
│  ┌──────────────┐         ┌──────────────────┐                 │
│  │  Lambda      │         │  CloudWatch      │                 │
│  │  (Video      │         │  (Logs, Metrics) │                 │
│  │  Processing) │         └──────────────────┘                 │
│  └──────────────┘                                               │
└─────────────────────────────────────────────────────────────────┘
```

### 10.2 Container Configuration

**Dockerfile (Multi-stage build):**
```dockerfile
# Stage 1: Build
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY migrations/ ./migrations/

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -o api ./cmd/api

# Stage 2: Production
FROM alpine:3.19

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /app

# Create non-root user
RUN addgroup -g 1001 -S ascend && \
    adduser -S ascend -u 1001 -G ascend

# Copy binary and migrations from builder
COPY --from=builder --chown=ascend:ascend /build/api ./api
COPY --from=builder --chown=ascend:ascend /build/migrations ./migrations

# Set timezone to UTC
ENV TZ=UTC

# Expose port
EXPOSE 8080

# Switch to non-root user
USER ascend

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=40s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Start application
CMD ["./api"]
```

**ECS Task Definition:**
```json
{
  "family": "ascend-api",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "1024",
  "memory": "2048",
  "executionRoleArn": "arn:aws:iam::ACCOUNT_ID:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::ACCOUNT_ID:role/ascendApiTaskRole",
  "containerDefinitions": [
    {
      "name": "api",
      "image": "ACCOUNT_ID.dkr.ecr.eu-west-1.amazonaws.com/ascend-api:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {
          "name": "ENV",
          "value": "production"
        },
        {
          "name": "PORT",
          "value": "8080"
        },
        {
          "name": "AWS_REGION",
          "value": "eu-west-1"
        },
        {
          "name": "LOG_LEVEL",
          "value": "info"
        }
      ],
      "secrets": [
        {
          "name": "DATABASE_URL",
          "valueFrom": "arn:aws:secretsmanager:eu-west-1:ACCOUNT_ID:secret:ascend/db-url"
        },
        {
          "name": "JWT_ACCESS_SECRET",
          "valueFrom": "arn:aws:secretsmanager:eu-west-1:ACCOUNT_ID:secret:ascend/jwt-access"
        },
        {
          "name": "JWT_REFRESH_SECRET",
          "valueFrom": "arn:aws:secretsmanager:eu-west-1:ACCOUNT_ID:secret:ascend/jwt-refresh"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/ascend-api",
          "awslogs-region": "eu-west-1",
          "awslogs-stream-prefix": "api"
        }
      },
      "healthCheck": {
        "command": [
          "CMD-SHELL",
          "curl -f http://localhost:3000/health || exit 1"
        ],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      }
    }
  ]
}
```

### 10.3 Database Infrastructure

**RDS PostgreSQL Configuration:**
```terraform
resource "aws_db_instance" "ascend_postgres" {
  identifier           = "ascend-postgres-prod"
  engine              = "postgres"
  engine_version      = "15.4"
  instance_class      = "db.t4g.medium"
  allocated_storage   = 100
  storage_type        = "gp3"
  storage_encrypted   = true
  kms_key_id         = aws_kms_key.rds_encryption.arn

  # High Availability
  multi_az           = true

  # Database Configuration
  db_name            = "ascend_production"
  username           = "ascend_admin"
  password           = random_password.db_password.result
  port               = 5432

  # Backup Configuration
  backup_retention_period = 30
  backup_window          = "03:00-04:00"
  maintenance_window     = "Mon:04:00-Mon:05:00"

  # Monitoring
  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]
  monitoring_interval             = 60
  monitoring_role_arn            = aws_iam_role.rds_monitoring.arn

  # Performance Insights
  performance_insights_enabled    = true
  performance_insights_retention_period = 7

  # Security
  vpc_security_group_ids = [aws_security_group.rds.id]
  db_subnet_group_name   = aws_db_subnet_group.ascend.name

  # Protection
  deletion_protection = true
  skip_final_snapshot = false
  final_snapshot_identifier = "ascend-postgres-final-snapshot-${formatdate("YYYY-MM-DD-hhmm", timestamp())}"

  tags = {
    Name        = "Ascend Production Database"
    Environment = "production"
    Backup      = "critical"
  }
}

# Read Replica (for analytics queries)
resource "aws_db_instance" "ascend_postgres_replica" {
  identifier          = "ascend-postgres-replica"
  replicate_source_db = aws_db_instance.ascend_postgres.identifier
  instance_class      = "db.t4g.small"

  # Replica configuration
  auto_minor_version_upgrade = true

  # Monitoring
  monitoring_interval = 60
  monitoring_role_arn = aws_iam_role.rds_monitoring.arn

  tags = {
    Name        = "Ascend Read Replica"
    Environment = "production"
    Purpose     = "analytics"
  }
}
```

**ElastiCache Redis Configuration:**
```terraform
resource "aws_elasticache_replication_group" "ascend_redis" {
  replication_group_id       = "ascend-redis-cluster"
  replication_group_description = "Redis cluster for Ascend caching and session storage"

  engine               = "redis"
  engine_version       = "7.0"
  node_type           = "cache.t4g.micro"
  num_cache_clusters  = 2
  port                = 6379

  # High Availability
  automatic_failover_enabled = true
  multi_az_enabled          = true

  # Security
  at_rest_encryption_enabled = true
  transit_encryption_enabled = true
  auth_token_enabled        = true
  auth_token               = random_password.redis_token.result

  # Subnet and Security Groups
  subnet_group_name  = aws_elasticache_subnet_group.ascend.name
  security_group_ids = [aws_security_group.redis.id]

  # Backup
  snapshot_retention_limit = 7
  snapshot_window         = "03:00-05:00"

  # Maintenance
  maintenance_window = "sun:05:00-sun:06:00"

  tags = {
    Name        = "Ascend Redis Cluster"
    Environment = "production"
  }
}
```

### 10.4 CI/CD Pipeline

**GitHub Actions Workflow (.github/workflows/deploy.yml):**
```yaml
name: Deploy to Production

on:
  push:
    branches:
      - main
  workflow_dispatch:

env:
  AWS_REGION: eu-west-1
  ECR_REPOSITORY: ascend-api
  ECS_CLUSTER: ascend-production
  ECS_SERVICE: api
  CONTAINER_NAME: api

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_DB: test_db
          POSTGRES_USER: test_user
          POSTGRES_PASSWORD: test_password
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true

      - name: Verify dependencies
        run: go mod verify

      - name: Run linter
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          args: --timeout=5m

      - name: Run unit tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Run integration tests
        env:
          DATABASE_URL: postgresql://test_user:test_password@localhost:5432/test_db
        run: go test -v -tags=integration ./tests/integration/...

      - name: Check test coverage
        run: |
          go tool cover -func=coverage.out
          COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print substr($3, 1, length($3)-1)}')
          echo "Total coverage: $COVERAGE%"
          if (( $(echo "$COVERAGE < 80" | bc -l) )); then
            echo "Coverage $COVERAGE% is below 80% threshold"
            exit 1
          fi

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out
          fail_ci_if_error: true

  security-scan:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          format: 'sarif'
          output: 'trivy-results.sarif'

      - name: Upload Trivy results to GitHub Security
        uses: github/codeql-action/upload-sarif@v2
        with:
          sarif_file: 'trivy-results.sarif'

      - name: Run Gosec security scanner
        uses: securego/gosec@master
        with:
          args: '-no-fail -fmt sarif -out gosec-results.sarif ./...'

      - name: Run Nancy (dependency vulnerability scanner)
        run: |
          go install github.com/sonatype-nexus-community/nancy@latest
          go list -json -m all | nancy sleuth

  build-and-deploy:
    needs: [test, security-scan]
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: ${{ env.AWS_REGION }}

      - name: Login to Amazon ECR
        id: login-ecr
        uses: aws-actions/amazon-ecr-login@v2

      - name: Build, tag, and push image to Amazon ECR
        id: build-image
        env:
          ECR_REGISTRY: ${{ steps.login-ecr.outputs.registry }}
          IMAGE_TAG: ${{ github.sha }}
        run: |
          docker build -t $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG .
          docker tag $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG $ECR_REGISTRY/$ECR_REPOSITORY:latest
          docker push $ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG
          docker push $ECR_REGISTRY/$ECR_REPOSITORY:latest
          echo "image=$ECR_REGISTRY/$ECR_REPOSITORY:$IMAGE_TAG" >> $GITHUB_OUTPUT

      - name: Download task definition
        run: |
          aws ecs describe-task-definition \
            --task-definition ascend-api \
            --query taskDefinition > task-definition.json

      - name: Fill in the new image ID in the Amazon ECS task definition
        id: task-def
        uses: aws-actions/amazon-ecs-render-task-definition@v1
        with:
          task-definition: task-definition.json
          container-name: ${{ env.CONTAINER_NAME }}
          image: ${{ steps.build-image.outputs.image }}

      - name: Deploy Amazon ECS task definition
        uses: aws-actions/amazon-ecs-deploy-task-definition@v1
        with:
          task-definition: ${{ steps.task-def.outputs.task-definition }}
          service: ${{ env.ECS_SERVICE }}
          cluster: ${{ env.ECS_CLUSTER }}
          wait-for-service-stability: true

      - name: Verify deployment
        run: |
          # Wait for new tasks to be running
          sleep 30

          # Check health endpoint
          ENDPOINT=$(aws elbv2 describe-load-balancers \
            --names ascend-alb \
            --query 'LoadBalancers[0].DNSName' \
            --output text)

          curl -f https://$ENDPOINT/health || exit 1

          echo "✅ Deployment successful!"

  notify:
    needs: [build-and-deploy]
    runs-on: ubuntu-latest
    if: always()
    steps:
      - name: Notify Slack on success
        if: success()
        uses: slackapi/slack-github-action@v1
        with:
          payload: |
            {
              "text": "✅ Production deployment successful",
              "blocks": [
                {
                  "type": "section",
                  "text": {
                    "type": "mrkdwn",
                    "text": "*Production Deployment Successful* ✅\n*Commit:* ${{ github.sha }}\n*Author:* ${{ github.actor }}"
                  }
                }
              ]
            }
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}

      - name: Notify Slack on failure
        if: failure()
        uses: slackapi/slack-github-action@v1
        with:
          payload: |
            {
              "text": "❌ Production deployment failed",
              "blocks": [
                {
                  "type": "section",
                  "text": {
                    "type": "mrkdwn",
                    "text": "*Production Deployment Failed* ❌\n*Commit:* ${{ github.sha }}\n*Author:* ${{ github.actor }}\n<${{ github.server_url }}/${{ github.repository }}/actions/runs/${{ github.run_id }}|View Logs>"
                  }
                }
              ]
            }
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
```

### 10.5 Database Migrations

**Migration Strategy:**
```typescript
// src/database/migrations/runMigrations.ts
import { DataSource } from 'typeorm';
import { Logger } from '../utils/logger';

const logger = new Logger('Migrations');

export async function runMigrations(dataSource: DataSource): Promise<void> {
  try {
    logger.info('Starting database migrations...');

    // Check pending migrations
    const pendingMigrations = await dataSource.showMigrations();

    if (!pendingMigrations) {
      logger.info('No pending migrations');
      return;
    }

    // Run migrations
    const migrations = await dataSource.runMigrations({
      transaction: 'all', // Run all migrations in a single transaction
    });

    logger.info(`Successfully ran ${migrations.length} migrations`);

    // Log migration details
    for (const migration of migrations) {
      logger.info(`✓ ${migration.name}`);
    }
  } catch (error) {
    logger.error('Migration failed', error);
    throw error;
  }
}

// Pre-deployment migration check script
export async function validateMigrations(dataSource: DataSource): Promise<boolean> {
  try {
    // Check if there are any pending migrations
    const pendingMigrations = await dataSource.showMigrations();

    if (pendingMigrations) {
      logger.warn('Warning: There are pending migrations that need to be applied');
      return false;
    }

    logger.info('✓ All migrations are up to date');
    return true;
  } catch (error) {
    logger.error('Failed to validate migrations', error);
    return false;
  }
}
```

**Migration Deployment Process:**
1. **Pre-deployment:** Run migration validation against production database replica
2. **Deployment window:** Schedule maintenance window for breaking schema changes
3. **Backup:** Automated RDS snapshot before migration
4. **Migration:** Run migrations using TypeORM CLI
5. **Validation:** Run smoke tests to verify schema changes
6. **Rollback plan:** Keep previous RDS snapshot for 30 days

### 10.6 Environment Configuration

**Environment Variables (AWS Secrets Manager):**
```json
{
  "ascend/production/api": {
    "NODE_ENV": "production",
    "PORT": "3000",
    "LOG_LEVEL": "info",

    "DATABASE_URL": "postgresql://user:pass@ascend-postgres-prod.xxx.eu-west-1.rds.amazonaws.com:5432/ascend_production",
    "DATABASE_POOL_MIN": "2",
    "DATABASE_POOL_MAX": "20",
    "DATABASE_SSL": "true",

    "REDIS_URL": "rediss://:token@ascend-redis-cluster.xxx.cache.amazonaws.com:6379",

    "JWT_ACCESS_SECRET": "...",
    "JWT_REFRESH_SECRET": "...",
    "JWT_ACCESS_EXPIRY": "15m",
    "JWT_REFRESH_EXPIRY": "7d",

    "AWS_REGION": "eu-west-1",
    "S3_BUCKET": "ascend-videos-prod",
    "S3_CLOUDFRONT_DOMAIN": "d123456789.cloudfront.net",

    "SENTRY_DSN": "https://xxx@sentry.io/xxx",
    "SENTRY_ENVIRONMENT": "production",

    "RATE_LIMIT_WINDOW_MS": "900000",
    "RATE_LIMIT_MAX_REQUESTS": "100"
  }
}
```

---

## 11. Testing Strategy

### 11.1 Testing Pyramid

**Test Distribution:**
- **Unit Tests:** 70% coverage target
- **Integration Tests:** 20% coverage target
- **End-to-End Tests:** 10% coverage target

### 11.2 Unit Testing

**Framework:** Jest with TypeScript

**Test Structure:**
```typescript
// src/modules/sessions/services/__tests__/sessionService.test.ts
import { SessionService } from '../sessionService';
import { mockRepository } from '../../../../test/utils/mockRepository';
import { createMockSession } from '../../../../test/fixtures/sessions';

describe('SessionService', () => {
  let sessionService: SessionService;
  let sessionRepository: ReturnType<typeof mockRepository>;
  let athleteRepository: ReturnType<typeof mockRepository>;

  beforeEach(() => {
    sessionRepository = mockRepository();
    athleteRepository = mockRepository();
    sessionService = new SessionService(sessionRepository, athleteRepository);
  });

  describe('createSession', () => {
    it('should create a session with correct volume load calculation', async () => {
      // Arrange
      const athleteId = 'athlete-123';
      const sessionData = {
        date: '2025-10-06',
        sessionType: 'heavy',
        sets: [
          { exercise: 'Snatch', weight: 80, reps: 3 },
          { exercise: 'Snatch', weight: 85, reps: 2 },
          { exercise: 'Snatch', weight: 90, reps: 1 },
        ],
      };

      athleteRepository.findOne.mockResolvedValue({
        id: athleteId,
        email: 'athlete@test.com',
      });

      sessionRepository.save.mockImplementation((session) =>
        Promise.resolve({ ...session, id: 'session-123' })
      );

      // Act
      const result = await sessionService.createSession(athleteId, sessionData);

      // Assert
      expect(result.totalVolumeLoad).toBe(80*3 + 85*2 + 90*1); // 500
      expect(result.totalSets).toBe(3);
      expect(sessionRepository.save).toHaveBeenCalledWith(
        expect.objectContaining({
          athleteId,
          date: new Date('2025-10-06'),
          sessionType: 'heavy',
          totalVolumeLoad: 500,
        })
      );
    });

    it('should throw error when athlete does not exist', async () => {
      // Arrange
      athleteRepository.findOne.mockResolvedValue(null);

      // Act & Assert
      await expect(
        sessionService.createSession('invalid-id', {
          date: '2025-10-06',
          sets: [],
        })
      ).rejects.toThrow('Athlete not found');
    });
  });

  describe('calculateACWR', () => {
    it('should calculate ACWR correctly', async () => {
      // Arrange
      const athleteId = 'athlete-123';
      const mockSessions = [
        createMockSession({ date: '2025-10-06', totalVolumeLoad: 5000 }),
        createMockSession({ date: '2025-10-05', totalVolumeLoad: 4500 }),
        createMockSession({ date: '2025-10-04', totalVolumeLoad: 4800 }),
        // ... more sessions for chronic load calculation
      ];

      sessionRepository.find.mockResolvedValue(mockSessions);

      // Act
      const acwr = await sessionService.calculateACWR(athleteId);

      // Assert
      expect(acwr.acuteLoad).toBeDefined();
      expect(acwr.chronicLoad).toBeDefined();
      expect(acwr.ratio).toBeGreaterThan(0);
      expect(acwr.status).toMatch(/optimal|elevated|high/);
    });
  });
});
```

**Coverage Requirements:**
```json
// jest.config.js
{
  "coverageThreshold": {
    "global": {
      "branches": 80,
      "functions": 80,
      "lines": 80,
      "statements": 80
    },
    "./src/modules/sessions/": {
      "branches": 90,
      "functions": 90,
      "lines": 90,
      "statements": 90
    },
    "./src/modules/auth/": {
      "branches": 95,
      "functions": 95,
      "lines": 95,
      "statements": 95
    }
  }
}
```

### 11.3 Integration Testing

**Database Integration Tests:**
```typescript
// src/modules/sessions/integration/__tests__/sessionAPI.integration.test.ts
import request from 'supertest';
import { app } from '../../../../app';
import { setupTestDatabase, teardownTestDatabase } from '../../../../test/utils/testDatabase';
import { createAuthenticatedUser } from '../../../../test/utils/auth';

describe('Session API Integration Tests', () => {
  let authToken: string;
  let athleteId: string;

  beforeAll(async () => {
    await setupTestDatabase();
    const { token, athlete } = await createAuthenticatedUser();
    authToken = token;
    athleteId = athlete.id;
  });

  afterAll(async () => {
    await teardownTestDatabase();
  });

  describe('POST /sessions', () => {
    it('should create a session and persist to database', async () => {
      // Arrange
      const sessionData = {
        date: '2025-10-06',
        startTime: '2025-10-06T09:00:00Z',
        sessionType: 'heavy',
        sets: [
          {
            exercise: 'Snatch',
            weight: 80,
            reps: 3,
            rpe: 7,
            videoUrl: null,
          },
        ],
      };

      // Act
      const response = await request(app)
        .post('/sessions')
        .set('Authorization', `Bearer ${authToken}`)
        .send(sessionData)
        .expect(201);

      // Assert
      expect(response.body.id).toBeDefined();
      expect(response.body.totalVolumeLoad).toBe(240);

      // Verify persistence
      const fetchResponse = await request(app)
        .get(`/sessions/${response.body.id}`)
        .set('Authorization', `Bearer ${authToken}`)
        .expect(200);

      expect(fetchResponse.body.id).toBe(response.body.id);
      expect(fetchResponse.body.sets).toHaveLength(1);
    });

    it('should return 401 without authentication', async () => {
      await request(app)
        .post('/sessions')
        .send({})
        .expect(401);
    });

    it('should validate session data', async () => {
      const response = await request(app)
        .post('/sessions')
        .set('Authorization', `Bearer ${authToken}`)
        .send({
          date: 'invalid-date',
          sets: [],
        })
        .expect(400);

      expect(response.body.errors).toBeDefined();
    });
  });

  describe('GET /analytics/acwr', () => {
    beforeEach(async () => {
      // Seed test data
      const sessions = Array.from({ length: 28 }, (_, i) => ({
        date: new Date(Date.now() - i * 24 * 60 * 60 * 1000).toISOString(),
        sets: [{ exercise: 'Snatch', weight: 80, reps: 5 }],
      }));

      for (const session of sessions) {
        await request(app)
          .post('/sessions')
          .set('Authorization', `Bearer ${authToken}`)
          .send(session);
      }
    });

    it('should calculate ACWR from database records', async () => {
      const response = await request(app)
        .get('/analytics/acwr')
        .set('Authorization', `Bearer ${authToken}`)
        .expect(200);

      expect(response.body.acuteLoad).toBeGreaterThan(0);
      expect(response.body.chronicLoad).toBeGreaterThan(0);
      expect(response.body.ratio).toBeGreaterThan(0);
      expect(response.body.status).toMatch(/optimal|elevated|high/);
    });
  });
});
```

### 11.4 End-to-End Testing (Mobile)

**Framework:** Detox for React Native

**E2E Test Structure:**
```typescript
// e2e/workoutFlow.e2e.ts
describe('Workout Recording Flow', () => {
  beforeAll(async () => {
    await device.launchApp({
      newInstance: true,
      permissions: { camera: 'YES', photos: 'YES' },
    });
  });

  beforeEach(async () => {
    await device.reloadReactNative();
  });

  it('should complete full workout recording flow', async () => {
    // Login
    await element(by.id('email-input')).typeText('test@athlete.com');
    await element(by.id('password-input')).typeText('password123');
    await element(by.id('login-button')).tap();

    // Wait for home screen
    await waitFor(element(by.id('home-screen')))
      .toBeVisible()
      .withTimeout(5000);

    // Start new session
    await element(by.id('new-session-button')).tap();

    // Select session type
    await element(by.text('Heavy Day')).tap();

    // Add exercise
    await element(by.id('add-exercise-button')).tap();
    await element(by.id('exercise-search')).typeText('Snatch');
    await element(by.text('Snatch')).atIndex(0).tap();

    // Add set
    await element(by.id('weight-input')).typeText('80');
    await element(by.id('reps-input')).typeText('3');
    await element(by.id('rpe-slider')).swipe('right', 'fast', 0.7);
    await element(by.id('add-set-button')).tap();

    // Verify set was added
    await expect(element(by.text('80 kg × 3'))).toBeVisible();

    // Save session
    await element(by.id('save-session-button')).tap();

    // Verify success message
    await expect(element(by.text('Session saved successfully'))).toBeVisible();

    // Verify session appears in history
    await element(by.id('tab-history')).tap();
    await expect(element(by.text('Snatch'))).toBeVisible();
    await expect(element(by.text('80 kg × 3'))).toBeVisible();
  });

  it('should record video for a set', async () => {
    // Navigate to session screen
    await element(by.id('new-session-button')).tap();
    await element(by.text('Heavy Day')).tap();
    await element(by.id('add-exercise-button')).tap();
    await element(by.text('Snatch')).tap();

    // Open video recorder
    await element(by.id('record-video-button')).tap();

    // Record video (mock)
    await element(by.id('camera-record-button')).tap();
    await new Promise(resolve => setTimeout(resolve, 3000));
    await element(by.id('camera-stop-button')).tap();

    // Confirm video
    await element(by.id('video-confirm-button')).tap();

    // Verify video thumbnail appears
    await expect(element(by.id('video-thumbnail'))).toBeVisible();
  });

  it('should work offline and sync when back online', async () => {
    // Disable network
    await device.disableSynchronization();
    await device.setURLBlacklist(['.*']);

    // Record session offline
    await element(by.id('new-session-button')).tap();
    await element(by.text('Heavy Day')).tap();
    // ... add exercise and sets
    await element(by.id('save-session-button')).tap();

    // Verify offline indicator
    await expect(element(by.id('offline-badge'))).toBeVisible();

    // Re-enable network
    await device.setURLBlacklist([]);
    await device.enableSynchronization();

    // Trigger sync
    await element(by.id('sync-button')).tap();

    // Wait for sync completion
    await waitFor(element(by.text('Sync complete')))
      .toBeVisible()
      .withTimeout(10000);

    // Verify data synced
    await expect(element(by.id('offline-badge'))).not.toBeVisible();
  });
});
```

### 11.5 Performance Testing

**Load Testing with k6:**
```javascript
// tests/load/api-load-test.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

const failureRate = new Rate('failed_requests');

export const options = {
  stages: [
    { duration: '2m', target: 100 },  // Ramp up to 100 users
    { duration: '5m', target: 100 },  // Stay at 100 users
    { duration: '2m', target: 200 },  // Ramp up to 200 users
    { duration: '5m', target: 200 },  // Stay at 200 users
    { duration: '2m', target: 0 },    // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'], // 95% < 500ms, 99% < 1s
    failed_requests: ['rate<0.01'],                  // < 1% error rate
  },
};

const BASE_URL = __ENV.API_URL || 'https://api.ascend.app';

// Authenticate and get token
function authenticate() {
  const loginRes = http.post(`${BASE_URL}/auth/login`, JSON.stringify({
    email: `loadtest-${__VU}@test.com`,
    password: 'LoadTest123!',
  }), {
    headers: { 'Content-Type': 'application/json' },
  });

  check(loginRes, {
    'login successful': (r) => r.status === 200,
    'access token present': (r) => r.json('accessToken') !== undefined,
  });

  return loginRes.json('accessToken');
}

export default function () {
  const token = authenticate();
  const headers = {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  };

  // Test 1: List sessions
  let res = http.get(`${BASE_URL}/sessions?limit=20`, { headers });
  check(res, {
    'list sessions status 200': (r) => r.status === 200,
    'list sessions response time < 200ms': (r) => r.timings.duration < 200,
  });
  failureRate.add(res.status !== 200);

  sleep(1);

  // Test 2: Create session
  const sessionPayload = JSON.stringify({
    date: new Date().toISOString().split('T')[0],
    startTime: new Date().toISOString(),
    sessionType: 'medium',
    sets: [
      { exercise: 'Snatch', weight: 80, reps: 3, rpe: 7 },
      { exercise: 'Snatch', weight: 85, reps: 2, rpe: 8 },
    ],
  });

  res = http.post(`${BASE_URL}/sessions`, sessionPayload, { headers });
  check(res, {
    'create session status 201': (r) => r.status === 201,
    'create session response time < 500ms': (r) => r.timings.duration < 500,
  });
  failureRate.add(res.status !== 201);

  sleep(1);

  // Test 3: Get analytics
  res = http.get(`${BASE_URL}/analytics/acwr`, { headers });
  check(res, {
    'analytics status 200': (r) => r.status === 200,
    'analytics response time < 800ms': (r) => r.timings.duration < 800,
  });
  failureRate.add(res.status !== 200);

  sleep(2);
}

export function handleSummary(data) {
  return {
    'load-test-results.json': JSON.stringify(data),
    stdout: textSummary(data, { indent: ' ', enableColors: true }),
  };
}
```

### 11.6 Test Data Management

**Test Database Seeding:**
```typescript
// test/utils/seed.ts
import { DataSource } from 'typeorm';
import { hash } from 'bcrypt';

export async function seedTestData(dataSource: DataSource): Promise<void> {
  const athleteRepo = dataSource.getRepository('Athlete');
  const sessionRepo = dataSource.getRepository('Session');

  // Create test athletes
  const athletes = await athleteRepo.save([
    {
      email: 'athlete1@test.com',
      passwordHash: await hash('password123', 10),
      name: 'Test Athlete 1',
      bodyWeight: 85.0,
      gender: 'male',
    },
    {
      email: 'athlete2@test.com',
      passwordHash: await hash('password123', 10),
      name: 'Test Athlete 2',
      bodyWeight: 65.0,
      gender: 'female',
    },
  ]);

  // Create test sessions (28 days of data for ACWR)
  const sessions = [];
  for (let i = 0; i < 28; i++) {
    const date = new Date();
    date.setDate(date.getDate() - i);

    sessions.push({
      athleteId: athletes[0].id,
      date,
      sessionType: i % 3 === 0 ? 'heavy' : 'medium',
      totalVolumeLoad: 4000 + Math.random() * 2000,
      totalSets: 12 + Math.floor(Math.random() * 8),
    });
  }

  await sessionRepo.save(sessions);
}
```

---

## 12. Performance & Optimization

### 12.1 Database Performance

**Indexing Strategy:**
```sql
-- Sessions table indexes
CREATE INDEX idx_sessions_athlete_date ON sessions(athlete_id, date DESC);
CREATE INDEX idx_sessions_date ON sessions(date DESC);
CREATE INDEX idx_sessions_athlete_created ON sessions(athlete_id, created_at DESC);

-- Exercise sets indexes
CREATE INDEX idx_exercise_sets_session ON exercise_sets(session_id);
CREATE INDEX idx_exercise_sets_exercise ON exercise_sets(exercise);
CREATE INDEX idx_exercise_sets_athlete_exercise ON exercise_sets(athlete_id, exercise, created_at DESC);

-- Videos indexes
CREATE INDEX idx_videos_athlete ON videos(athlete_id, created_at DESC);
CREATE INDEX idx_videos_set ON videos(set_id);
CREATE INDEX idx_videos_status ON videos(processing_status, created_at DESC);

-- One Rep Max indexes
CREATE INDEX idx_1rm_athlete_exercise ON one_rep_maxes(athlete_id, exercise, date DESC);

-- Refresh tokens indexes
CREATE INDEX idx_refresh_tokens_athlete ON refresh_tokens(athlete_id, expires_at DESC);
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token) WHERE revoked = false;
```

**Query Optimization Examples:**
```typescript
// ❌ BAD: N+1 query problem
async getSessionsWithSets(athleteId: string) {
  const sessions = await this.sessionRepo.find({ where: { athleteId } });

  for (const session of sessions) {
    session.sets = await this.setRepo.find({ where: { sessionId: session.id } });
  }

  return sessions;
}

// ✅ GOOD: Use eager loading with join
async getSessionsWithSets(athleteId: string) {
  return await this.sessionRepo.find({
    where: { athleteId },
    relations: ['sets'],
    order: { date: 'DESC' },
    take: 30,
  });
}

// ✅ BETTER: Use query builder for complex queries
async getSessionsWithSetsOptimized(athleteId: string, startDate: Date, endDate: Date) {
  return await this.sessionRepo
    .createQueryBuilder('session')
    .leftJoinAndSelect('session.sets', 'sets')
    .where('session.athleteId = :athleteId', { athleteId })
    .andWhere('session.date BETWEEN :startDate AND :endDate', { startDate, endDate })
    .orderBy('session.date', 'DESC')
    .addOrderBy('sets.createdAt', 'ASC')
    .limit(30)
    .getMany();
}
```

**Connection Pooling:**
```typescript
// src/database/config.ts
export const databaseConfig: DataSourceOptions = {
  type: 'postgres',
  url: process.env.DATABASE_URL,
  ssl: process.env.DATABASE_SSL === 'true' ? { rejectUnauthorized: false } : false,

  // Connection pool configuration
  extra: {
    max: parseInt(process.env.DATABASE_POOL_MAX || '20'), // Max connections
    min: parseInt(process.env.DATABASE_POOL_MIN || '2'),  // Min connections
    idleTimeoutMillis: 30000,  // Close idle connections after 30s
    connectionTimeoutMillis: 10000, // Fail after 10s if can't connect
    maxUses: 7500, // Retire connections after 7500 uses
  },

  // Query logging in development
  logging: process.env.NODE_ENV === 'development' ? ['query', 'error'] : ['error'],

  // Entity caching
  cache: {
    type: 'ioredis',
    options: {
      host: process.env.REDIS_HOST,
      port: parseInt(process.env.REDIS_PORT || '6379'),
    },
    duration: 60000, // Cache for 60 seconds
  },
};
```

### 12.2 API Response Optimization

**Response Compression:**
```typescript
// src/middleware/compression.ts
import compression from 'compression';

export const compressionMiddleware = compression({
  threshold: 1024, // Only compress responses > 1KB
  level: 6, // Compression level (0-9)
  filter: (req, res) => {
    // Don't compress streaming responses
    if (req.headers['x-no-compression']) {
      return false;
    }
    return compression.filter(req, res);
  },
});
```

**Pagination & Cursor-based Navigation:**
```typescript
// src/modules/sessions/controllers/sessionController.ts
export class SessionController {
  async listSessions(req: Request, res: Response) {
    const { cursor, limit = 30, startDate, endDate } = req.query;
    const athleteId = req.user.id;

    const query = this.sessionRepo
      .createQueryBuilder('session')
      .where('session.athleteId = :athleteId', { athleteId })
      .orderBy('session.date', 'DESC')
      .take(parseInt(limit as string) + 1); // Fetch one extra to determine if there's more

    // Cursor-based pagination
    if (cursor) {
      query.andWhere('session.id < :cursor', { cursor });
    }

    // Date filters
    if (startDate) {
      query.andWhere('session.date >= :startDate', { startDate });
    }
    if (endDate) {
      query.andWhere('session.date <= :endDate', { endDate });
    }

    const sessions = await query.getMany();

    // Check if there are more results
    const hasMore = sessions.length > parseInt(limit as string);
    if (hasMore) {
      sessions.pop(); // Remove the extra item
    }

    res.json({
      sessions,
      pagination: {
        hasMore,
        nextCursor: hasMore ? sessions[sessions.length - 1].id : null,
      },
    });
  }
}
```

**ETags for Caching:**
```typescript
// src/middleware/etag.ts
import etag from 'etag';

export function etagMiddleware(req: Request, res: Response, next: NextFunction) {
  const originalSend = res.send;

  res.send = function (data: any): Response {
    if (req.method === 'GET' && res.statusCode === 200) {
      const etagValue = etag(data);
      res.setHeader('ETag', etagValue);

      // Check if client has cached version
      if (req.headers['if-none-match'] === etagValue) {
        res.status(304).end();
        return res;
      }
    }

    return originalSend.call(this, data);
  };

  next();
}
```

### 12.3 Mobile App Performance

**React Native Performance Optimizations:**
```typescript
// src/components/SessionList.tsx
import React, { memo, useCallback } from 'react';
import { FlatList, ListRenderItem } from 'react-native';

interface Session {
  id: string;
  date: string;
  totalVolumeLoad: number;
  sets: number;
}

// Memoize list items
const SessionItem = memo(({ session }: { session: Session }) => (
  <View style={styles.sessionCard}>
    <Text>{session.date}</Text>
    <Text>{session.totalVolumeLoad} kg</Text>
  </View>
));

export const SessionList: React.FC<{ sessions: Session[] }> = ({ sessions }) => {
  // Use keyExtractor for stable keys
  const keyExtractor = useCallback((item: Session) => item.id, []);

  // Memoize render function
  const renderItem: ListRenderItem<Session> = useCallback(
    ({ item }) => <SessionItem session={item} />,
    []
  );

  return (
    <FlatList
      data={sessions}
      keyExtractor={keyExtractor}
      renderItem={renderItem}
      // Performance optimizations
      removeClippedSubviews={true} // Unmount components outside viewport
      maxToRenderPerBatch={10} // Render 10 items per batch
      updateCellsBatchingPeriod={50} // Update every 50ms
      initialNumToRender={10} // Render 10 items initially
      windowSize={5} // Keep 5 screens of content in memory
      // Enable list optimization
      getItemLayout={(data, index) => ({
        length: 80, // Fixed height for each item
        offset: 80 * index,
        index,
      })}
    />
  );
};
```

**Image Optimization:**
```typescript
// src/components/VideoThumbnail.tsx
import FastImage from 'react-native-fast-image';

export const VideoThumbnail: React.FC<{ uri: string }> = ({ uri }) => (
  <FastImage
    source={{
      uri,
      priority: FastImage.priority.normal,
      cache: FastImage.cacheControl.immutable,
    }}
    style={{ width: 120, height: 90 }}
    resizeMode={FastImage.resizeMode.cover}
  />
);
```

### 12.4 Video Processing Optimization

**Parallel Processing:**
```typescript
// src/modules/videos/services/videoProcessor.ts
import { SQS } from 'aws-sdk';
import { Lambda } from 'aws-sdk';

export class VideoProcessorService {
  private sqs: SQS;
  private lambda: Lambda;

  async processVideo(videoId: string, s3Key: string): Promise<void> {
    // Send to SQS queue for async processing
    await this.sqs.sendMessage({
      QueueUrl: process.env.VIDEO_PROCESSING_QUEUE_URL!,
      MessageBody: JSON.stringify({
        videoId,
        s3Key,
        operations: [
          { type: 'compress', targetSize: '1080p' },
          { type: 'thumbnail', timestamp: 1.0 },
          { type: 'metadata', extract: true },
        ],
      }),
    }).promise();

    // Update video status
    await this.videoRepo.update(videoId, {
      processingStatus: 'queued',
    });
  }

  // Lambda function handler
  async lambdaHandler(event: any): Promise<void> {
    const { videoId, s3Key, operations } = JSON.parse(event.Records[0].body);

    try {
      // Process operations in parallel
      await Promise.all([
        this.compressVideo(s3Key),
        this.generateThumbnail(s3Key),
        this.extractMetadata(s3Key),
      ]);

      // Update video status
      await this.videoRepo.update(videoId, {
        processingStatus: 'completed',
        processedAt: new Date(),
      });
    } catch (error) {
      await this.videoRepo.update(videoId, {
        processingStatus: 'failed',
        errorMessage: error.message,
      });
    }
  }
}
```

### 12.5 Caching Strategy

**Redis Caching Implementation:**
```typescript
// src/services/cacheService.ts
import Redis from 'ioredis';

export class CacheService {
  private redis: Redis;

  constructor() {
    this.redis = new Redis(process.env.REDIS_URL!);
  }

  async get<T>(key: string): Promise<T | null> {
    const cached = await this.redis.get(key);
    return cached ? JSON.parse(cached) : null;
  }

  async set(key: string, value: any, ttlSeconds: number = 3600): Promise<void> {
    await this.redis.setex(key, ttlSeconds, JSON.stringify(value));
  }

  async delete(key: string): Promise<void> {
    await this.redis.del(key);
  }

  async deletePattern(pattern: string): Promise<void> {
    const keys = await this.redis.keys(pattern);
    if (keys.length > 0) {
      await this.redis.del(...keys);
    }
  }
}

// Usage example
export class SessionService {
  constructor(
    private sessionRepo: Repository<Session>,
    private cache: CacheService
  ) {}

  async getSession(sessionId: string): Promise<Session | null> {
    // Try cache first
    const cacheKey = `session:${sessionId}`;
    const cached = await this.cache.get<Session>(cacheKey);

    if (cached) {
      return cached;
    }

    // Fetch from database
    const session = await this.sessionRepo.findOne({
      where: { id: sessionId },
      relations: ['sets'],
    });

    if (session) {
      // Cache for 1 hour
      await this.cache.set(cacheKey, session, 3600);
    }

    return session;
  }

  async updateSession(sessionId: string, data: Partial<Session>): Promise<Session> {
    const session = await this.sessionRepo.update(sessionId, data);

    // Invalidate cache
    await this.cache.delete(`session:${sessionId}`);
    await this.cache.deletePattern(`athlete:${session.athleteId}:sessions:*`);

    return session;
  }
}
```

**Cache Invalidation Strategy:**
- **Time-based:** Sessions cached for 1 hour, athlete profiles for 24 hours
- **Event-based:** Invalidate on create/update/delete operations
- **Pattern-based:** Use Redis key patterns for bulk invalidation (e.g., `athlete:{id}:*`)

### 12.6 Performance Monitoring

**Application Performance Monitoring:**
```typescript
// src/monitoring/apm.ts
import * as Sentry from '@sentry/node';
import { ProfilingIntegration } from '@sentry/profiling-node';

export function initializeAPM() {
  Sentry.init({
    dsn: process.env.SENTRY_DSN,
    environment: process.env.SENTRY_ENVIRONMENT,
    tracesSampleRate: process.env.NODE_ENV === 'production' ? 0.1 : 1.0,
    profilesSampleRate: 0.1, // Profile 10% of transactions
    integrations: [
      new ProfilingIntegration(),
      new Sentry.Integrations.Http({ tracing: true }),
      new Sentry.Integrations.Express({ app: true }),
    ],
  });
}

// Custom transaction tracking
export function trackTransaction(name: string, operation: string) {
  return Sentry.startTransaction({
    name,
    op: operation,
  });
}

// Usage in controller
export class SessionController {
  async createSession(req: Request, res: Response) {
    const transaction = trackTransaction('create-session', 'http.server');

    try {
      const span = transaction.startChild({ op: 'db.query', description: 'Save session' });
      const session = await this.sessionService.createSession(req.body);
      span.finish();

      res.status(201).json(session);
    } catch (error) {
      Sentry.captureException(error);
      throw error;
    } finally {
      transaction.finish();
    }
  }
}
```

**Performance Budget:**
- API response time (p95): < 500ms
- API response time (p99): < 1000ms
- Mobile app startup time: < 2s
- Screen transition time: < 300ms
- Video upload time (10MB): < 30s
- Offline sync time (100 records): < 5s

---

## 13. Monitoring & Observability

### 13.1 Logging Strategy

**Structured Logging Implementation:**
```typescript
// src/utils/logger.ts
import winston from 'winston';

const logFormat = winston.format.combine(
  winston.format.timestamp({ format: 'YYYY-MM-DD HH:mm:ss' }),
  winston.format.errors({ stack: true }),
  winston.format.json()
);

export const logger = winston.createLogger({
  level: process.env.LOG_LEVEL || 'info',
  format: logFormat,
  defaultMeta: {
    service: 'ascend-api',
    environment: process.env.NODE_ENV,
  },
  transports: [
    // Console output for development
    new winston.transports.Console({
      format: winston.format.combine(
        winston.format.colorize(),
        winston.format.simple()
      ),
    }),
    // CloudWatch Logs for production
    ...(process.env.NODE_ENV === 'production'
      ? [
          new winston.transports.Stream({
            stream: process.stdout,
            format: logFormat,
          }),
        ]
      : []),
  ],
});

// Context logger for request tracing
export class ContextLogger {
  constructor(private context: Record<string, any>) {}

  info(message: string, meta?: Record<string, any>) {
    logger.info(message, { ...this.context, ...meta });
  }

  warn(message: string, meta?: Record<string, any>) {
    logger.warn(message, { ...this.context, ...meta });
  }

  error(message: string, error?: Error, meta?: Record<string, any>) {
    logger.error(message, {
      ...this.context,
      ...meta,
      error: error ? {
        message: error.message,
        stack: error.stack,
        name: error.name,
      } : undefined,
    });
  }
}

// Request logging middleware
export function requestLogger(req: Request, res: Response, next: NextFunction) {
  const requestId = req.headers['x-request-id'] || uuidv4();
  const startTime = Date.now();

  req.logger = new ContextLogger({
    requestId,
    method: req.method,
    path: req.path,
    athleteId: req.user?.id,
  });

  res.on('finish', () => {
    const duration = Date.now() - startTime;

    req.logger.info('Request completed', {
      statusCode: res.statusCode,
      duration,
      userAgent: req.headers['user-agent'],
    });
  });

  next();
}
```

### 13.2 Metrics Collection

**CloudWatch Metrics:**
```typescript
// src/monitoring/metrics.ts
import { CloudWatch } from 'aws-sdk';

export class MetricsService {
  private cloudwatch: CloudWatch;
  private namespace = 'Ascend/API';

  constructor() {
    this.cloudwatch = new CloudWatch({ region: process.env.AWS_REGION });
  }

  async recordMetric(
    metricName: string,
    value: number,
    unit: string = 'Count',
    dimensions?: Record<string, string>
  ): Promise<void> {
    await this.cloudwatch.putMetricData({
      Namespace: this.namespace,
      MetricData: [
        {
          MetricName: metricName,
          Value: value,
          Unit: unit,
          Timestamp: new Date(),
          Dimensions: dimensions
            ? Object.entries(dimensions).map(([Name, Value]) => ({ Name, Value }))
            : undefined,
        },
      ],
    }).promise();
  }

  async recordSessionCreated(athleteId: string, sessionType: string): Promise<void> {
    await this.recordMetric('SessionCreated', 1, 'Count', {
      SessionType: sessionType,
    });
  }

  async recordAPILatency(endpoint: string, latency: number): Promise<void> {
    await this.recordMetric('APILatency', latency, 'Milliseconds', {
      Endpoint: endpoint,
    });
  }

  async recordVideoProcessed(duration: number, success: boolean): Promise<void> {
    await this.recordMetric('VideoProcessed', 1, 'Count', {
      Status: success ? 'Success' : 'Failed',
    });

    if (success) {
      await this.recordMetric('VideoProcessingDuration', duration, 'Seconds');
    }
  }
}

// Usage in service
export class SessionService {
  constructor(
    private sessionRepo: Repository<Session>,
    private metrics: MetricsService
  ) {}

  async createSession(data: CreateSessionDto): Promise<Session> {
    const session = await this.sessionRepo.save(data);

    // Record metric
    await this.metrics.recordSessionCreated(data.athleteId, data.sessionType);

    return session;
  }
}
```

### 13.3 Distributed Tracing

**AWS X-Ray Integration:**
```typescript
// src/middleware/xray.ts
import AWSXRay from 'aws-xray-sdk';
import { Express } from 'express';

export function setupXRay(app: Express) {
  // Enable X-Ray tracing
  AWSXRay.captureAWS(require('aws-sdk'));
  AWSXRay.captureHTTPsGlobal(require('http'));
  AWSXRay.captureHTTPsGlobal(require('https'));

  // Add X-Ray middleware
  app.use(AWSXRay.express.openSegment('Ascend-API'));

  // Database query tracing
  AWSXRay.capturePostgres(require('pg'));

  // Close segment at the end of request
  app.use(AWSXRay.express.closeSegment());
}

// Custom subsegment tracking
export async function traceOperation<T>(
  name: string,
  operation: () => Promise<T>
): Promise<T> {
  const segment = AWSXRay.getSegment();

  if (!segment) {
    return operation();
  }

  const subsegment = segment.addNewSubsegment(name);

  try {
    const result = await operation();
    subsegment.close();
    return result;
  } catch (error) {
    subsegment.addError(error);
    subsegment.close();
    throw error;
  }
}

// Usage
async function processVideo(videoId: string): Promise<void> {
  await traceOperation('video-compression', async () => {
    // Video compression logic
  });

  await traceOperation('thumbnail-generation', async () => {
    // Thumbnail generation logic
  });
}
```

### 13.4 Alerting Configuration

**CloudWatch Alarms:**
```terraform
# terraform/monitoring/alarms.tf

# High API error rate
resource "aws_cloudwatch_metric_alarm" "api_error_rate" {
  alarm_name          = "ascend-api-high-error-rate"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "5XXError"
  namespace           = "AWS/ApplicationELB"
  period              = 300
  statistic           = "Sum"
  threshold           = 10
  alarm_description   = "API error rate is above 10 errors in 5 minutes"
  alarm_actions       = [aws_sns_topic.alerts.arn]

  dimensions = {
    LoadBalancer = aws_lb.ascend_alb.arn_suffix
  }
}

# High API latency
resource "aws_cloudwatch_metric_alarm" "api_latency" {
  alarm_name          = "ascend-api-high-latency"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "TargetResponseTime"
  namespace           = "AWS/ApplicationELB"
  period              = 300
  statistic           = "Average"
  threshold           = 1.0
  alarm_description   = "API p99 latency is above 1 second"
  alarm_actions       = [aws_sns_topic.alerts.arn]

  dimensions = {
    LoadBalancer = aws_lb.ascend_alb.arn_suffix
  }
}

# Database connection issues
resource "aws_cloudwatch_metric_alarm" "database_connections" {
  alarm_name          = "ascend-db-high-connections"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "DatabaseConnections"
  namespace           = "AWS/RDS"
  period              = 300
  statistic           = "Average"
  threshold           = 80
  alarm_description   = "Database connections above 80% of max"
  alarm_actions       = [aws_sns_topic.alerts.arn]

  dimensions = {
    DBInstanceIdentifier = aws_db_instance.ascend_postgres.id
  }
}

# Failed video processing
resource "aws_cloudwatch_metric_alarm" "video_processing_failures" {
  alarm_name          = "ascend-video-processing-failures"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "VideoProcessed"
  namespace           = "Ascend/API"
  period              = 600
  statistic           = "Sum"
  threshold           = 5
  alarm_description   = "More than 5 video processing failures in 10 minutes"
  alarm_actions       = [aws_sns_topic.alerts.arn]

  dimensions = {
    Status = "Failed"
  }
}
```

### 13.5 Health Checks

**Comprehensive Health Check Endpoint:**
```typescript
// src/controllers/healthController.ts
import { Request, Response } from 'express';
import { DataSource } from 'typeorm';
import Redis from 'ioredis';
import { S3 } from 'aws-sdk';

export class HealthController {
  constructor(
    private dataSource: DataSource,
    private redis: Redis,
    private s3: S3
  ) {}

  async healthCheck(req: Request, res: Response): Promise<void> {
    const checks = await Promise.allSettled([
      this.checkDatabase(),
      this.checkRedis(),
      this.checkS3(),
    ]);

    const [database, redis, s3] = checks;

    const isHealthy = checks.every((check) => check.status === 'fulfilled');

    res.status(isHealthy ? 200 : 503).json({
      status: isHealthy ? 'healthy' : 'unhealthy',
      timestamp: new Date().toISOString(),
      version: process.env.APP_VERSION || '1.0.0',
      checks: {
        database: this.formatCheckResult(database),
        redis: this.formatCheckResult(redis),
        s3: this.formatCheckResult(s3),
      },
    });
  }

  private async checkDatabase(): Promise<{ status: string; latency: number }> {
    const start = Date.now();
    await this.dataSource.query('SELECT 1');
    const latency = Date.now() - start;

    return { status: 'up', latency };
  }

  private async checkRedis(): Promise<{ status: string; latency: number }> {
    const start = Date.now();
    await this.redis.ping();
    const latency = Date.now() - start;

    return { status: 'up', latency };
  }

  private async checkS3(): Promise<{ status: string; latency: number }> {
    const start = Date.now();
    await this.s3.headBucket({ Bucket: process.env.S3_BUCKET! }).promise();
    const latency = Date.now() - start;

    return { status: 'up', latency };
  }

  private formatCheckResult(
    result: PromiseSettledResult<{ status: string; latency: number }>
  ): any {
    if (result.status === 'fulfilled') {
      return result.value;
    }
    return {
      status: 'down',
      error: result.reason.message,
    };
  }
}
```

### 13.6 Dashboards

**CloudWatch Dashboard Configuration:**
```json
{
  "widgets": [
    {
      "type": "metric",
      "properties": {
        "metrics": [
          ["AWS/ApplicationELB", "RequestCount", { "stat": "Sum" }],
          [".", "HTTPCode_Target_2XX_Count", { "stat": "Sum" }],
          [".", "HTTPCode_Target_4XX_Count", { "stat": "Sum" }],
          [".", "HTTPCode_Target_5XX_Count", { "stat": "Sum" }]
        ],
        "period": 300,
        "stat": "Sum",
        "region": "eu-west-1",
        "title": "API Request Volume"
      }
    },
    {
      "type": "metric",
      "properties": {
        "metrics": [
          ["AWS/ApplicationELB", "TargetResponseTime", { "stat": "Average" }],
          ["...", { "stat": "p95" }],
          ["...", { "stat": "p99" }]
        ],
        "period": 300,
        "region": "eu-west-1",
        "title": "API Latency"
      }
    },
    {
      "type": "metric",
      "properties": {
        "metrics": [
          ["AWS/RDS", "DatabaseConnections", { "stat": "Average" }],
          [".", "CPUUtilization", { "stat": "Average" }],
          [".", "FreeableMemory", { "stat": "Average" }]
        ],
        "period": 300,
        "region": "eu-west-1",
        "title": "Database Health"
      }
    },
    {
      "type": "metric",
      "properties": {
        "metrics": [
          ["Ascend/API", "SessionCreated", { "stat": "Sum" }],
          [".", "VideoProcessed", { "stat": "Sum", "dimensions": { "Status": "Success" } }],
          ["...", { "dimensions": { "Status": "Failed" } }]
        ],
        "period": 300,
        "region": "eu-west-1",
        "title": "Business Metrics"
      }
    }
  ]
}
```

---

## 14. GDPR Compliance Implementation

### 14.1 Data Privacy by Design

**Privacy-First Architecture:**
1. **Data Minimization:** Only collect essential user data
2. **Purpose Limitation:** Use data only for stated purposes
3. **Storage Limitation:** Retain data only as long as necessary
4. **Pseudonymization:** Use UUIDs instead of sequential IDs
5. **Encryption:** All personal data encrypted at rest and in transit

### 14.2 User Consent Management

**Cookie Consent Implementation:**
```typescript
// src/modules/consent/services/consentService.ts
export interface ConsentRecord {
  athleteId: string;
  consentType: 'marketing' | 'analytics' | 'essential';
  granted: boolean;
  timestamp: Date;
  ipAddress: string;
  userAgent: string;
}

export class ConsentService {
  async recordConsent(consent: ConsentRecord): Promise<void> {
    await this.consentRepo.save({
      ...consent,
      expiresAt: new Date(Date.now() + 365 * 24 * 60 * 60 * 1000), // 1 year
    });
  }

  async getConsent(athleteId: string, consentType: string): Promise<boolean> {
    const consent = await this.consentRepo.findOne({
      where: {
        athleteId,
        consentType,
        expiresAt: MoreThan(new Date()),
      },
      order: { timestamp: 'DESC' },
    });

    return consent?.granted ?? false;
  }

  async withdrawConsent(athleteId: string, consentType: string): Promise<void> {
    await this.recordConsent({
      athleteId,
      consentType,
      granted: false,
      timestamp: new Date(),
      ipAddress: '',
      userAgent: '',
    });
  }
}
```

### 14.3 Right to Access (Data Export)

**Complete Data Export:**
```typescript
// src/modules/gdpr/services/dataExportService.ts
export class DataExportService {
  async exportAllAthleteData(athleteId: string): Promise<Buffer> {
    // Collect all athlete data
    const [profile, sessions, videos, oneRMs, consents] = await Promise.all([
      this.athleteRepo.findOne(athleteId),
      this.sessionRepo.find({ where: { athleteId }, relations: ['sets'] }),
      this.videoRepo.find({ where: { athleteId } }),
      this.oneRMRepo.find({ where: { athleteId } }),
      this.consentRepo.find({ where: { athleteId } }),
    ]);

    // Generate video download URLs (valid for 24 hours)
    const videoData = await Promise.all(
      videos.map(async (video) => ({
        ...video,
        downloadUrl: await this.s3Service.getSignedUrl(video.s3Key, 86400),
      }))
    );

    const exportData = {
      exportDate: new Date().toISOString(),
      athlete: {
        id: profile.id,
        email: profile.email,
        name: profile.name,
        bodyWeight: profile.bodyWeight,
        gender: profile.gender,
        timezone: profile.timezone,
        createdAt: profile.createdAt,
      },
      sessions: sessions.map((s) => ({
        id: s.id,
        date: s.date,
        sessionType: s.sessionType,
        totalVolumeLoad: s.totalVolumeLoad,
        sets: s.sets,
        notes: s.notes,
      })),
      videos: videoData,
      oneRepMaxes: oneRMs,
      consents: consents,
    };

    // Create JSON file
    return Buffer.from(JSON.stringify(exportData, null, 2), 'utf-8');
  }

  async sendDataExportEmail(athleteId: string): Promise<void> {
    const athlete = await this.athleteRepo.findOne(athleteId);
    const exportBuffer = await this.exportAllAthleteData(athleteId);

    // Upload to S3 with expiration
    const s3Key = `exports/${athleteId}/${Date.now()}-data-export.json`;
    await this.s3Service.upload(s3Key, exportBuffer, {
      expiresIn: 7 * 24 * 60 * 60, // 7 days
    });

    // Generate download URL
    const downloadUrl = await this.s3Service.getSignedUrl(s3Key, 7 * 24 * 60 * 60);

    // Send email
    await this.emailService.sendDataExport(athlete.email, downloadUrl);

    // Log export request
    await this.auditLogRepo.create({
      action: 'DATA_EXPORT_REQUESTED',
      athleteId,
      metadata: { s3Key },
    });
  }
}
```

### 14.4 Right to Erasure (Data Deletion)

**Complete Data Deletion:**
```typescript
// src/modules/gdpr/services/dataDeletionService.ts
export class DataDeletionService {
  async requestDeletion(athleteId: string): Promise<void> {
    // Soft delete athlete record
    await this.athleteRepo.update(athleteId, {
      deletedAt: new Date(),
      email: `deleted_${athleteId}@ascend.app`, // Anonymize
      name: 'Deleted User',
      passwordHash: null,
    });

    // Schedule permanent deletion after 30 days (grace period)
    await this.queueService.schedule({
      type: 'PERMANENT_DELETE_ATHLETE',
      athleteId,
      executeAt: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000),
    });

    // Notify athlete
    await this.emailService.sendDeletionConfirmation(
      athlete.email,
      new Date(Date.now() + 30 * 24 * 60 * 60 * 1000)
    );
  }

  async permanentlyDeleteAthlete(athleteId: string): Promise<void> {
    // Delete all related data
    await this.dataSource.transaction(async (manager) => {
      // Delete sessions and sets
      await manager.delete('exercise_sets', { athleteId });
      await manager.delete('sessions', { athleteId });

      // Delete 1RMs
      await manager.delete('one_rep_maxes', { athleteId });

      // Delete consents
      await manager.delete('consents', { athleteId });

      // Delete videos from S3
      const videos = await manager.find('videos', { where: { athleteId } });
      for (const video of videos) {
        await this.s3Service.deleteObject(video.s3Key);
        if (video.thumbnailS3Key) {
          await this.s3Service.deleteObject(video.thumbnailS3Key);
        }
      }
      await manager.delete('videos', { athleteId });

      // Delete refresh tokens
      await manager.delete('refresh_tokens', { athleteId });

      // Delete athlete record
      await manager.delete('athletes', { id: athleteId });

      // Create audit log (retained for legal compliance)
      await manager.save('audit_logs', {
        action: 'PERMANENT_DELETION_COMPLETED',
        athleteId,
        timestamp: new Date(),
        metadata: { reason: 'GDPR right to be forgotten' },
      });
    });
  }

  async cancelDeletion(athleteId: string): Promise<void> {
    // Restore athlete record
    await this.athleteRepo.update(athleteId, {
      deletedAt: null,
    });

    // Cancel scheduled permanent deletion
    await this.queueService.cancelScheduled({
      type: 'PERMANENT_DELETE_ATHLETE',
      athleteId,
    });
  }
}
```

### 14.5 Data Breach Notification

**Breach Detection & Notification:**
```typescript
// src/modules/security/services/breachNotificationService.ts
export class BreachNotificationService {
  async detectBreach(event: SecurityEvent): Promise<void> {
    // Analyze security event
    const isBreach = await this.analyzeSecurityEvent(event);

    if (isBreach) {
      await this.handleBreach(event);
    }
  }

  private async handleBreach(event: SecurityEvent): Promise<void> {
    // Log breach
    await this.auditLogRepo.create({
      action: 'SECURITY_BREACH_DETECTED',
      metadata: {
        type: event.type,
        severity: event.severity,
        affectedUsers: event.affectedUserIds.length,
        timestamp: new Date(),
      },
    });

    // Notify DPA within 72 hours (if required)
    if (this.requiresDPANotification(event)) {
      await this.notifyDataProtectionAuthority(event);
    }

    // Notify affected users
    if (this.requiresUserNotification(event)) {
      await this.notifyAffectedUsers(event);
    }

    // Send internal alert
    await this.alertService.sendCriticalAlert({
      title: 'Security Breach Detected',
      description: `Breach type: ${event.type}, Affected users: ${event.affectedUserIds.length}`,
      severity: 'CRITICAL',
    });
  }

  private requiresDPANotification(event: SecurityEvent): boolean {
    // High risk to rights and freedoms of individuals
    return (
      event.severity === 'HIGH' ||
      event.type === 'DATA_LEAK' ||
      event.affectedUserIds.length > 100
    );
  }

  private async notifyAffectedUsers(event: SecurityEvent): Promise<void> {
    for (const userId of event.affectedUserIds) {
      const user = await this.athleteRepo.findOne(userId);

      await this.emailService.sendBreachNotification(user.email, {
        breachType: event.type,
        dataAffected: event.dataTypes,
        actionRequired: event.actionRequired,
        contactInfo: 'privacy@ascend.app',
      });
    }
  }
}
```

### 14.6 Privacy Policy & Terms

**Key GDPR Compliance Points:**

1. **Legal Basis for Processing:**
   - Consent: Marketing communications, analytics
   - Contract: Service delivery, account management
   - Legitimate Interest: Service improvement, fraud prevention

2. **Data Retention:**
   - Active accounts: Indefinite (while account is active)
   - Deleted accounts: 30-day grace period, then permanent deletion
   - Audit logs: 7 years (legal compliance requirement)
   - Backups: 30 days (encrypted, access-controlled)

3. **Third-Party Data Sharing:**
   - AWS (hosting, storage)
   - Sentry (error tracking)
   - Stripe (future payment processing)
   - No data sold to third parties

4. **User Rights:**
   - Right to access (data export)
   - Right to rectification (profile updates)
   - Right to erasure (account deletion)
   - Right to data portability (JSON export)
   - Right to object (marketing opt-out)
   - Right to withdraw consent (any time)

---

## 15. Development Phases

### 15.1 Phase 1: MVP Core (Months 0-3)

**Month 1: Foundation**
- ✅ Project setup and repository structure
- ✅ Database schema and migrations
- ✅ Authentication system (JWT + refresh tokens)
- ✅ Basic API endpoints (auth, profile)
- ✅ Mobile app scaffold (React Native)
- ✅ CI/CD pipeline setup

**Month 2: Core Features**
- ✅ Session recording (manual entry)
- ✅ Exercise library (10 core lifts)
- ✅ Set tracking (weight, reps, RPE)
- ✅ Volume load calculation
- ✅ Basic analytics (weekly/monthly volume)
- ✅ Offline-first sync implementation

**Month 3: Video & Polish**
- ✅ Video recording and upload
- ✅ Video compression pipeline
- ✅ ACWR calculation and display
- ✅ 1RM tracking
- ✅ User onboarding flow
- ✅ Beta testing with 30 athletes

**Deliverables:**
- Functional mobile app (iOS + Android)
- RESTful API
- GDPR-compliant infrastructure
- Beta testing feedback report

### 15.2 Phase 2: Enhanced Features (Months 4-6)

**Month 4: Analytics Enhancement**
- Exercise-specific progress tracking
- Volume trends visualization
- PR notifications
- Comparison with past performance
- Export data (CSV/JSON)

**Month 5: UX Improvements**
- Video playback with slow-motion
- Set templates (pre-fill common workouts)
- Quick-add exercises
- Session notes and tags
- Dark mode theme

**Month 6: Social Features**
- Share progress on social media
- Public profile (optional)
- Leaderboards (opt-in)
- Follow other athletes
- Community features MVP

**Deliverables:**
- Enhanced analytics dashboard
- Improved user experience
- Social sharing capabilities
- 500 active athletes target

### 15.3 Phase 3: Coach Platform Beta (Months 7-12)

**Months 7-8: Coach Features Foundation**
- Coach account type
- Roster management (add/remove athletes)
- View athlete sessions
- Comment on athlete videos
- Assign workouts (basic)

**Months 9-10: Programming Tools**
- Workout templates
- Program builder (weekly cycles)
- Load prescription
- Auto-calculate percentages
- Send programs to athletes

**Months 11-12: Analytics for Coaches**
- Athlete comparison dashboard
- Team analytics
- Injury risk indicators (ACWR alerts)
- Progress reports
- Coach-athlete messaging

**Deliverables:**
- Functional coach platform
- 20 beta coaches onboarded
- Athlete-coach workflow validated
- Pricing model finalized

### 15.4 Phase 4: Scale & Optimize (Months 13-18)

**Months 13-14: Performance & Scale**
- Database query optimization
- CDN optimization
- Background job processing
- Mobile app performance tuning
- Load testing (1000 concurrent users)

**Months 15-16: Advanced Features**
- AI form analysis (beta)
- Exercise recommendations
- Fatigue scoring
- Velocity-based training integration
- Wearable device integration

**Months 17-18: Internationalization**
- Multi-language support (German, French)
- Local exercise databases
- Regional pricing
- European expansion marketing

**Deliverables:**
- Scalable infrastructure (10,000+ users)
- Advanced AI features
- Multi-language support
- European market presence

### 15.5 Development Milestones

**Key Metrics by Phase:**

| Phase | Timeframe | Athletes | Coaches | Sessions/Month | Revenue (MRR) |
|-------|-----------|----------|---------|----------------|---------------|
| Phase 1 (MVP) | Months 0-3 | 30 (beta) | 0 | 300 | €0 |
| Phase 2 (Enhanced) | Months 4-6 | 500 | 0 | 6,000 | €0 |
| Phase 3 (Coach Beta) | Months 7-12 | 2,000 | 20 | 30,000 | €780 |
| Phase 4 (Scale) | Months 13-18 | 10,000 | 200 | 180,000 | €7,800 |

---

## 16. Risk Mitigation

### 16.1 Technical Risks

**Risk 1: Data Loss**
- **Probability:** Low
- **Impact:** Critical
- **Mitigation:**
  - Multi-AZ database with automated backups
  - Point-in-time recovery (30 days)
  - Offline-first architecture (local data persistence)
  - Regular backup testing (monthly)
  - S3 versioning for videos

**Risk 2: Performance Degradation at Scale**
- **Probability:** Medium
- **Impact:** High
- **Mitigation:**
  - Load testing before each major release
  - Database read replicas for analytics
  - Redis caching layer
  - CDN for video delivery
  - Auto-scaling ECS services

**Risk 3: Mobile App Crashes**
- **Probability:** Medium
- **Impact:** High
- **Mitigation:**
  - Sentry error tracking
  - Comprehensive unit and E2E testing
  - Phased rollout (5% → 20% → 50% → 100%)
  - Kill switch for problematic features
  - Crash rate monitoring (< 1% threshold)

**Risk 4: Video Processing Failures**
- **Probability:** Medium
- **Impact:** Medium
- **Mitigation:**
  - Retry mechanism with exponential backoff
  - Dead letter queue for failed jobs
  - Manual processing queue
  - User notification on failures
  - Original video always preserved

### 16.2 Security Risks

**Risk 5: Data Breach**
- **Probability:** Low
- **Impact:** Critical
- **Mitigation:**
  - Encryption at rest (AES-256) and in transit (TLS 1.3)
  - Regular security audits (quarterly)
  - Penetration testing (annually)
  - GDPR compliance framework
  - Incident response plan
  - Security awareness training

**Risk 6: Account Takeover**
- **Probability:** Medium
- **Impact:** High
- **Mitigation:**
  - Bcrypt password hashing (cost factor 10)
  - Rate limiting on login attempts
  - Email verification on signup
  - Password strength requirements
  - Account activity monitoring
  - Optional 2FA (future)

**Risk 7: DDoS Attack**
- **Probability:** Low
- **Impact:** High
- **Mitigation:**
  - CloudFlare DDoS protection
  - Rate limiting per IP
  - AWS WAF rules
  - Auto-scaling infrastructure
  - Incident response plan

### 16.3 Business Risks

**Risk 8: Low User Adoption**
- **Probability:** Medium
- **Impact:** Critical
- **Mitigation:**
  - Beta testing with target users
  - Regular user feedback sessions
  - Iterate based on user needs
  - Strong onboarding flow
  - Community building (Discord/forum)
  - Referral program

**Risk 9: Competitor Launches Similar Product**
- **Probability:** Medium
- **Impact:** High
- **Mitigation:**
  - Focus on weightlifting niche
  - Build strong community
  - Superior UX/UI
  - Offline-first advantage
  - Fast iteration cycles
  - Patent/trademark key features

**Risk 10: GDPR Non-Compliance**
- **Probability:** Low
- **Impact:** Critical
- **Mitigation:**
  - Privacy by design architecture
  - Regular GDPR audits
  - Legal counsel review
  - User consent management
  - Data processing agreements
  - Breach notification procedures

### 16.4 Operational Risks

**Risk 11: Key Team Member Departure**
- **Probability:** Medium
- **Impact:** High
- **Mitigation:**
  - Comprehensive documentation
  - Knowledge sharing sessions
  - Code review process
  - Onboarding documentation
  - Backup contacts for critical systems

**Risk 12: AWS Service Outage**
- **Probability:** Low
- **Impact:** High
- **Mitigation:**
  - Multi-AZ deployment
  - CloudFront caching
  - Status page for users
  - Communication plan
  - Service Level Agreement monitoring

**Risk 13: Budget Overrun**
- **Probability:** Medium
- **Impact:** Medium
- **Mitigation:**
  - AWS cost monitoring and alerts
  - Resource tagging and allocation
  - Monthly cost reviews
  - Scalable infrastructure (pay as you grow)
  - Cost optimization opportunities (reserved instances, spot instances)

### 16.5 Risk Monitoring

**Continuous Risk Assessment:**
- Monthly risk review meetings
- Quarterly security audits
- Annual penetration testing
- User feedback monitoring
- Performance metrics tracking
- Cost variance analysis

**Escalation Procedures:**
1. **Low Impact:** Team lead handles, documented in issue tracker
2. **Medium Impact:** CTO notified, mitigation plan within 24 hours
3. **High Impact:** Executive team notified, immediate action required
4. **Critical Impact:** All stakeholders notified, emergency response activated

---

## Conclusion

This Technical Design Document provides a comprehensive blueprint for building Project Ascend MVP - a mobile-first, offline-capable training management platform for weightlifting athletes in EMEA.

**Key Technical Highlights:**
- **Offline-First Architecture:** WatermelonDB + sync engine ensures seamless offline operation
- **GDPR Compliance:** Privacy by design with EU data residency (AWS eu-west-1)
- **Scalable Infrastructure:** AWS ECS Fargate with auto-scaling, multi-AZ database
- **Production-Grade Quality:** Comprehensive testing strategy (70% unit, 20% integration, 10% E2E)
- **Performance Optimized:** Redis caching, CDN delivery, query optimization
- **Observable System:** CloudWatch metrics, X-Ray tracing, Sentry error tracking
- **Security First:** TLS 1.3, AES-256 encryption, JWT authentication, rate limiting

**Development Timeline:**
- **Phase 1 (Months 0-3):** MVP with session tracking, video recording, ACWR analytics
- **Phase 2 (Months 4-6):** Enhanced analytics, UX improvements, social features
- **Phase 3 (Months 7-12):** Coach platform beta with roster management and programming tools
- **Phase 4 (Months 13-18):** Scale to 10,000+ users, AI features, European expansion

This document serves as the definitive technical reference for the engineering team and should be updated as architecture decisions evolve.