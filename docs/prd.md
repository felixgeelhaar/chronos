# Project Ascend: Product Requirements Document
## The Athlete-First Weightlifting Performance Platform (MVP)

**Version:** 2.0
**Date:** October 6, 2025
**Author:** Product Strategy Team
**Status:** In Development - Athlete-Focused MVP

---

## 1. Executive Summary & Vision

### Long-Term Vision: The Definitive Coach Platform for Elite Weightlifting

Project Ascend will become the industry-standard performance management platform for elite weightlifting coaches, enabling them to systematically develop athletes from youth prospects to Olympic medalists. The platform will provide coaches with comprehensive tools for programming, load management, technique analysis, and roster management—replacing the fragmented workflows that currently define high-performance coaching.

### MVP Strategy: Bottom-Up Market Entry via Athletes

Rather than building for coaches immediately, we're launching athlete-first in EMEA. This bottom-up go-to-market strategy addresses a critical insight: **athletes who already track their training are the perfect customers for coach-athlete platform features.** By building a loyal athlete user base first, we create natural demand for coach features and validate our core value proposition before investing in complex multi-user workflows.

**MVP Focus:** Self-coached and semi-coached competitive weightlifting athletes in Europe, Middle East, and Africa who want professional-grade performance tracking without paying for a full coaching service.

**Geographic Focus:** EMEA launch (Q1 2026) before expanding to North America and Asia-Pacific.

**Vision Statement:** Build the world's most comprehensive weightlifting performance ecosystem—starting with athletes, scaling to elite coaching.

**Core MVP Hypothesis:** Athletes who consistently track their training data, monitor their workload, and analyze their technique will see measurable improvements in performance and reduced injury rates compared to those using traditional tracking methods.

---

## 2. The Problem

Competitive weightlifting athletes face a critical gap between their training ambition and available tools:

### Current Pain Points

**Fragmented Data Tracking**
- Athletes use notebooks, Notes apps, or generic workout trackers that weren't designed for Olympic weightlifting
- Historical training data is scattered across multiple sources, making pattern recognition impossible
- No easy way to answer questions like "What was my volume 4 weeks before my last PR?"

**No Load Management Visibility**
- Athletes can't see their weekly training volume trends
- Overtraining often goes unnoticed until performance drops or injury occurs
- Deload timing is guesswork rather than data-driven

**Subjective Technique Assessment**
- Athletes record videos on their phone but have no tools to analyze them systematically
- Comparing technique across sessions requires manual scrolling through camera rolls
- No objective metrics to track bar path consistency or position improvements

**Programming Inefficiency**
- Following online programs requires manual tracking of percentages based on changing 1RMs
- No automatic calculation of training loads when maxes improve
- Difficult to track compliance and adherence to a program

### The Core Problem Statement

**Self-coached and semi-coached weightlifting athletes lack a purpose-built tool to systematically track, analyze, and optimize their training, forcing them to choose between time-consuming manual tracking or no tracking at all.**

---

## 3. Target Audience: The Weightlifting Athlete

### Primary User Persona: "The Dedicated Lifter"

**Demographics:**
- Age: 18-35
- Training experience: 1-5 years of consistent weightlifting
- Competition level: Local/regional meets, aiming for national qualification
- Training environment: CrossFit gym, weightlifting club, or garage gym
- Coaching situation: Self-coached (60%), remote programming (30%), local coach (10%)

**Psychographics:**
- Highly motivated and consistent (trains 4-6 days/week)
- Data-curious but not data scientists
- Follows weightlifting content on Instagram/YouTube
- Willing to pay for tools that demonstrably improve performance
- Comfortable with technology and mobile apps

**Goals:**
- Increase Snatch and Clean & Jerk totals by 5-10% per year
- Qualify for national-level competitions
- Train consistently without injury
- Understand what's working (and what isn't) in their training
- Track progress objectively to stay motivated

**Pain Points:**
- *"I don't know if I'm doing too much or too little volume"*
- *"I film every session but never actually review the footage systematically"*
- *"My Excel sheet is a mess and I've stopped updating it"*
- *"I can't tell if my technique is improving or just feels different"*
- *"When I hit a plateau, I don't know what to adjust"*

**Current Tools:**
- Notes app or notebook for logging workouts (40%)
- Excel/Google Sheets (30%)
- Generic fitness apps like Strong or JEFIT (20%)
- Nothing systematic (10%)

### Secondary Persona: "The Coached Athlete"

**Profile:** Athlete with a coach but wants personal tracking and analysis tools
**Use Case:** Supplement coach's programming with personal performance insights
**Future Consideration:** Coach integration features (post-MVP)

---

## 4. Geographic Strategy: EMEA First

### 4.1 Why EMEA for MVP Launch?

**Strong Weightlifting Culture**
- Europe has the highest concentration of competitive weightlifters globally
- Established national federations and regular competition schedule
- Strong tradition of data-driven training (Eastern European influence)
- High percentage of English speakers + addressable multi-language market

**Market Characteristics**
- **UK & Ireland:** 15,000+ competitive weightlifters, high app adoption, English-speaking
- **Germany:** 20,000+ lifters, strong engineering/data culture, high willingness to pay
- **France:** 18,000+ lifters, growing CrossFit/weightlifting crossover
- **Nordics:** Tech-savvy early adopters, strong English proficiency
- **Eastern Europe:** Largest weightlifting populations (Poland, Romania, Bulgaria), price-sensitive but high engagement

**Regulatory Advantage**
- GDPR compliance required anyway for any European presence
- Building GDPR-first makes global expansion easier
- Data residency requirements favor early European infrastructure investment

**Competitive Landscape**
- No dominant weightlifting-specific app in EMEA market
- Generic fitness apps (Strong, JEFIT) don't serve weightlifting needs
- WinWoTa legacy users looking for modern alternative
- Less competition than saturated US fitness app market

### 4.2 EMEA-Specific Product Considerations

**Language Support (MVP)**
- English (primary, 100% coverage)
- German (secondary, launch within 3 months)
- French (secondary, launch within 6 months)
- Language picker in onboarding
- Unit preference: kg/lbs toggle (default: kg for EMEA)

**Localization Requirements**
- Date formats: DD/MM/YYYY (European standard)
- 24-hour time format default
- Metric system primary (kg, cm, m)
- Currency: EUR, GBP for future premium features
- Timezone handling: CET, GMT, EET as primary zones

**EMEA Competition Calendar**
- Integration with European Weightlifting Federation calendar
- National championship dates pre-loaded by country
- Competition prep mode with countdown timers

**Privacy & Compliance**
- GDPR compliance from day one (not retrofit)
- Cookie consent (minimal tracking, opt-in for analytics)
- Right to data portability (JSON export)
- Right to erasure (30-day deletion window)
- Data residency: EU-based servers (AWS Frankfurt or Ireland region)

### 4.3 Geographic Expansion Timeline

**Phase 1: EMEA Launch (Months 0-6)**
- Focus countries: UK, Ireland, Germany, France, Nordics
- English + German languages
- EU data residency
- Target: 500 athletes across 5 countries

**Phase 2: EMEA Expansion (Months 6-12)**
- Add French language
- Eastern European marketing push (Poland, Romania, Bulgaria)
- Partnership with European weightlifting clubs
- Target: 2,000 athletes across 10+ countries

**Phase 3: North America (Months 12-18)**
- US/Canada launch
- Add lbs/imperial units prominence
- Localize competition calendars (USAW, USA Powerlifting)
- Target: 5,000 athletes globally

**Phase 4: Asia-Pacific (Months 18-24)**
- Australia, New Zealand, Singapore launch
- Evaluate need for Mandarin, Japanese, Korean languages
- Target: 10,000 athletes globally

---

## 5. MVP Feature Set & Requirements

The MVP focuses on three core athlete workflows: **Track → Analyze → Improve**

### 5.1 Training Log & Workout Tracking

**User Story:** *"As an athlete, I want to quickly log my daily training so I can build a complete performance history without spending more than 2-3 minutes per session."*

#### Acceptance Criteria

**Quick Workout Entry**
- [ ] Mobile-first interface optimized for gym use (large touch targets, simple navigation)
- [ ] Pre-loaded exercise library with 50+ weightlifting movements and variations
  - Olympic lifts: Snatch, Clean, Jerk, and all common variations
  - Squats: Back, Front, Overhead, variations
  - Pulls: Clean Pull, Snatch Pull, Deadlift, RDL
  - Accessories: Press variations, rows, core work
- [ ] Each exercise tagged with movement category (Olympic lift, squat, pull, accessory)
- [ ] Create custom exercises with optional video reference links

**Set-by-Set Logging**
- [ ] Log sets with: weight lifted, reps completed, perceived effort (1-10 RPE)
- [ ] Optional fields: bar velocity (m/s), notes, video attachment
- [ ] Auto-calculate volume load (sets × reps × weight) per exercise and session
- [ ] Smart defaults: pre-fill previous session's weights for efficiency
- [ ] Percentage-based programming: input "80% of back squat" and app calculates from current 1RM

**Session Management**
- [ ] Start/end session timestamps for duration tracking
- [ ] Overall session RPE (1-10 scale)
- [ ] Session type tags: Heavy, Moderate, Light, Technique, Testing
- [ ] Optional session notes (max 500 characters)
- [ ] Mark sessions as "planned" vs "actual" to track adherence

**1RM Tracking & Management**
- [ ] Manual 1RM input for key lifts (Snatch, C&J, Back Squat, Front Squat)
- [ ] 1RM history log with date stamps
- [ ] Visual indicator when 1RM PR is achieved
- [ ] Auto-suggest 1RM updates based on logged heavy singles/triples
- [ ] Calculate and display Sinclair/Robi points for competitive lifts

#### Technical Requirements

- **Performance:** Workout entry screen loads in < 1 second
- **Offline Support:** Full offline functionality; sync when connection restored
- **Data Validation:**
  - Weight limits: 20kg - 300kg (adjustable in settings)
  - Reps: 1-20
  - RPE: 1-10
  - Required fields: exercise, weight, reps
- **Storage:** Local-first architecture with cloud backup
- **Export:** Ability to export training log as CSV

---

### 5.2 Performance Analytics & Progress Tracking

**User Story:** *"As an athlete, I want to visualize my training trends and progress so I can objectively understand if my program is working and identify patterns in my performance."*

#### Acceptance Criteria

**Lift Progress Dashboard**
- [ ] Charts showing 1RM progress over time for key lifts (Snatch, C&J, BS, FS)
- [ ] Time range filters: Last 30 days, 90 days, 6 months, 1 year, All time
- [ ] Personal records list with dates (e.g., "Snatch PR: 95kg on Sept 15, 2025")
- [ ] Comparative view: Current 1RM vs 3/6/12 months ago
- [ ] Sinclair/Robi coefficient tracking for competitive total

**Volume Load Analytics**
- [ ] Total weekly volume load chart (last 12 weeks)
- [ ] Volume breakdown by movement category (Olympic lifts, squats, pulls, accessories)
- [ ] Session-by-session volume comparison
- [ ] Moving average trendline (4-week rolling average)
- [ ] Color-coded volume zones:
  - Green: Within normal training range
  - Yellow: Elevated volume (>20% above 4-week average)
  - Red: High volume spike (>30% above 4-week average)

**Training Load Management**
- [ ] Acute:Chronic Workload Ratio (ACWR) calculation
  - Acute load: Last 7 days total volume
  - Chronic load: Last 28 days average weekly volume
  - Display ACWR with risk zones:
    - 0.8-1.3: Optimal (green)
    - 1.3-1.5: Caution (yellow)
    - >1.5: High risk (red)
- [ ] Visual alerts when ACWR exceeds 1.5
- [ ] Weekly volume recommendations based on ACWR trends

**Training Frequency Insights**
- [ ] Heatmap calendar showing training frequency (sessions per week)
- [ ] Streak tracking: Current consecutive training days
- [ ] Adherence metrics: Sessions completed vs planned (if using program features)

**Exercise-Specific Analytics**
- [ ] Volume trends for individual exercises
- [ ] Average intensity (% of 1RM) over time for key lifts
- [ ] Rep range distribution analysis

#### Technical Requirements

- **Performance:** Dashboard loads in < 2 seconds with 1 year of data
- **Data Refresh:** Analytics update in real-time after workout logging
- **Visualization:** Interactive charts with zoom/pan capabilities
- **Calculations:**
  - Volume Load: Σ(sets × reps × weight) aggregated by day/week/month
  - ACWR: (7-day volume) / (28-day average weekly volume)
  - Moving averages: 4-week rolling window
- **Export:** Ability to export analytics data as CSV or PDF report

---

### 5.3 Video Analysis & Technique Tracking

**User Story:** *"As an athlete, I want to upload and review my lift videos with analysis tools so I can objectively track my technique development and identify movement patterns."*

#### Acceptance Criteria

**Video Library & Organization**
- [ ] Upload videos from phone camera roll or record directly in-app
- [ ] Automatic video compression (target: < 50MB file size for 30-second clip)
- [ ] Tag videos with:
  - Exercise type (from exercise library)
  - Weight lifted
  - Session date (auto-linked to training log)
  - Subjective quality rating (1-5 stars)
  - Optional text notes (max 200 characters)
- [ ] Search/filter videos by exercise, date range, weight range
- [ ] Thumbnail grid view for easy browsing
- [ ] Sort by: Date, exercise type, weight, rating

**Video Playback & Analysis**
- [ ] Slow-motion playback controls (0.25x, 0.5x, 0.75x, 1x)
- [ ] Frame-by-frame navigation (for 60fps+ videos)
- [ ] Pause/play with large touch targets
- [ ] Zoom capability for detailed position review
- [ ] Optional: Grid overlay for vertical bar path reference

**Comparison Tools**
- [ ] Side-by-side video comparison (compare 2 lifts)
- [ ] Synchronized playback controls
- [ ] Common use cases:
  - Today's lift vs previous PR
  - Heavy attempt vs light technique work
  - Before/after technique adjustment
- [ ] Swipe between videos for quick comparison

**Manual Technique Annotation (Simple Drawing Tools)**
- [ ] Draw lines and angles on paused video frames
- [ ] Color options for lines (3-4 colors)
- [ ] Save annotated frame as image
- [ ] Common annotations:
  - Bar path line
  - Hip/knee/ankle angles
  - Shoulder position reference
- [ ] Undo/redo functionality
- [ ] Clear all drawings option

**Video Sharing**
- [ ] Generate shareable link for individual videos
- [ ] Privacy settings: Private, shareable with link, public
- [ ] Export video with annotations as MP4

#### Technical Requirements

- **Video Support:**
  - Formats: MP4, MOV, HEVC
  - Max resolution: 1080p (4K down-sampled)
  - Max file size: 100MB upload, 50MB after compression
  - Frame rates: 30fps, 60fps, 120fps, 240fps
- **Storage:** 5GB free storage per athlete (approx. 100 videos)
- **Performance:**
  - Video upload with compression: < 30 seconds for 30-second clip
  - Playback starts: < 2 seconds after selection
  - Side-by-side load time: < 3 seconds
- **Offline:** Download videos for offline playback
- **Processing:** Server-side video compression on upload

#### MVP Scope Limitations (Deferred to Post-MVP)
- ❌ Automated bar path tracking (requires ML/CV)
- ❌ Automated position detection
- ❌ AI-powered technique feedback
- ✅ Manual drawing tools sufficient for MVP

---

### 5.4 Simple Program Following (Optional for MVP)

**User Story:** *"As an athlete following a pre-written program, I want to input my weekly workouts in advance so the app can calculate percentages and track my adherence."*

#### Acceptance Criteria

**Program Template Input**
- [ ] Create weekly program templates
- [ ] Add daily workouts to calendar view
- [ ] Exercises auto-calculate based on percentage of current 1RM
  - Example: "Back Squat 5×3 @ 80%" auto-fills as 120kg if 1RM = 150kg
- [ ] Update all percentages when 1RM changes
- [ ] Save program templates for reuse (e.g., "Russian Squat Cycle")

**Program Execution**
- [ ] Today's workout view shows planned exercises
- [ ] Log actual weights/reps against planned
- [ ] Adherence tracking: planned vs actual completion
- [ ] Substitute exercise option if modification needed

**Program Library (Future Feature)**
- [ ] Pre-built programs from coaches (Phase 2)
- [ ] Community-shared programs (Phase 2)

#### Technical Requirements

- **Calculations:** Percentage calculations update immediately when 1RM changed
- **Flexibility:** Allow athlete to override calculated weights
- **Validation:** Warn if logged weight significantly differs from planned (>15% variance)

#### MVP Decision Point
**Recommendation:** Include basic program template feature in MVP if development timeline allows. If tight on timeline, defer to Phase 2 and focus on core tracking + analytics.

---

## 6. Technical Architecture

### 6.1 Technology Stack

**Mobile Application**
- **Framework:** React Native (cross-platform iOS/Android)
- **State Management:** Redux Toolkit or Zustand
- **Local Storage:** SQLite (via react-native-sqlite-storage)
- **Offline Sync:** WatermelonDB or custom sync engine
- **Video Player:** react-native-video with custom controls
- **Charts:** Victory Native or Recharts

**Backend Services**
- **API:** Node.js with Express or Fastify
- **Database:** PostgreSQL 15+ (primary datastore)
- **File Storage:** AWS S3 (eu-west-1 Ireland or eu-central-1 Frankfurt for GDPR)
- **Authentication:** JWT with refresh tokens, OAuth2 (Google, Apple)
- **Video Processing:** FFmpeg for server-side compression
- **Hosting:** AWS ECS/Fargate in EU regions (Ireland or Frankfurt)

**Infrastructure**
- **Region:** AWS eu-west-1 (Ireland) primary for GDPR compliance
- **CDN:** CloudFront with EU edge locations
- **Monitoring:** Sentry (errors), PostHog (analytics with EU data residency)
- **CI/CD:** GitHub Actions
- **Deployment:** Blue-green deployment strategy
- **Compliance:** GDPR-compliant logging, data retention policies

### 6.2 Data Model (Core Entities)

```typescript
// Athlete
interface Athlete {
  id: string;
  email: string;
  name: string;
  bodyWeight: number; // kg
  gender: 'male' | 'female';
  birthYear: number;
  createdAt: Date;
  updatedAt: Date;
}

// One Rep Max tracking
interface OneRepMax {
  id: string;
  athleteId: string;
  exercise: string; // "Snatch", "Clean & Jerk", "Back Squat", etc.
  weight: number; // kg
  date: Date;
  notes?: string;
  createdAt: Date;
}

// Training Session
interface Session {
  id: string;
  athleteId: string;
  date: Date;
  startTime?: Date;
  endTime?: Date;
  sessionType: 'heavy' | 'moderate' | 'light' | 'technique' | 'testing';
  overallRPE?: number; // 1-10
  notes?: string;
  totalVolumeLoad: number; // calculated
  createdAt: Date;
  updatedAt: Date;
}

// Exercise Set within a session
interface ExerciseSet {
  id: string;
  sessionId: string;
  exerciseName: string;
  exerciseCategory: 'olympic' | 'squat' | 'pull' | 'press' | 'accessory';
  setNumber: number;
  weight: number; // kg
  reps: number;
  rpe?: number; // 1-10
  velocity?: number; // m/s (optional)
  notes?: string;
  videoId?: string; // linked video
  volumeLoad: number; // weight * reps (calculated)
  createdAt: Date;
}

// Video
interface Video {
  id: string;
  athleteId: string;
  sessionId?: string; // optional link to session
  exerciseSetId?: string; // optional link to specific set
  exerciseName: string;
  weight?: number;
  fileName: string;
  fileSize: number; // bytes
  duration: number; // seconds
  thumbnailUrl: string;
  videoUrl: string; // S3/R2 URL
  qualityRating?: number; // 1-5 stars
  notes?: string;
  privacyLevel: 'private' | 'shareable' | 'public';
  uploadedAt: Date;
}

// Program Template (if included in MVP)
interface ProgramTemplate {
  id: string;
  athleteId: string;
  name: string;
  description?: string;
  durationWeeks: number;
  createdAt: Date;
}

interface ProgramWorkout {
  id: string;
  programTemplateId: string;
  weekNumber: number;
  dayNumber: number;
  exercises: ProgramExercise[];
}

interface ProgramExercise {
  exerciseName: string;
  sets: number;
  reps: number;
  intensityType: 'percentage' | 'absolute' | 'rpe';
  intensityValue: number; // e.g., 80 for 80%, or 100 for 100kg
  rest: number; // seconds
}
```

### 6.3 API Endpoints (Key Routes)

```
Authentication
POST   /api/auth/register
POST   /api/auth/login
POST   /api/auth/refresh
POST   /api/auth/logout

Athlete Profile
GET    /api/athlete/profile
PUT    /api/athlete/profile
GET    /api/athlete/settings
PUT    /api/athlete/settings

One Rep Maxes
GET    /api/athlete/1rm
POST   /api/athlete/1rm
PUT    /api/athlete/1rm/:id
DELETE /api/athlete/1rm/:id
GET    /api/athlete/1rm/history/:exercise

Training Sessions
GET    /api/sessions?startDate=&endDate=
POST   /api/sessions
GET    /api/sessions/:id
PUT    /api/sessions/:id
DELETE /api/sessions/:id

Exercise Sets
POST   /api/sessions/:sessionId/sets
PUT    /api/sets/:id
DELETE /api/sets/:id

Analytics
GET    /api/analytics/progress?exercise=&timeRange=
GET    /api/analytics/volume?timeRange=
GET    /api/analytics/acwr
GET    /api/analytics/frequency?timeRange=

Videos
GET    /api/videos?exercise=&startDate=&endDate=
POST   /api/videos/upload
GET    /api/videos/:id
PUT    /api/videos/:id
DELETE /api/videos/:id
GET    /api/videos/:id/download
POST   /api/videos/:id/share

Program Templates (if in MVP)
GET    /api/programs
POST   /api/programs
GET    /api/programs/:id
PUT    /api/programs/:id
DELETE /api/programs/:id
```

### 6.4 Security & Privacy (GDPR-Compliant)

**Authentication**
- JWT access tokens (15-minute expiry)
- Refresh tokens (7-day expiry, stored in httpOnly cookie)
- Rate limiting: 100 requests/minute per user
- Account lockout: 5 failed login attempts = 15-minute lockout

**Data Protection**
- Encryption at rest: AES-256 for database
- Encryption in transit: TLS 1.3 for all API calls
- Video storage: Private S3 buckets with signed URLs (1-hour expiry)
- Password hashing: bcrypt with cost factor 12

**Privacy Controls (GDPR-Compliant)**
- Athlete data is private by default
- Video sharing requires explicit per-video privacy setting
- Data export: Full data export in JSON/CSV format (GDPR Article 20)
- Data deletion: Permanent deletion within 30 days of request (GDPR Article 17)
- Cookie consent: Minimal tracking, explicit opt-in for analytics
- Data processing agreements: Clear terms for video/data processing
- Data retention: 12-month inactive account retention, then deletion notice

**Access Control**
- Athletes can only access their own data
- No cross-athlete data visibility (for MVP; coach features post-MVP)
- API endpoints validate athlete ID from JWT token
- Audit logging for all data access (GDPR compliance)

**GDPR-Specific Features**
- Data portability: Export all user data in machine-readable JSON format
- Right to be forgotten: One-click account + data deletion
- Privacy policy: Clear, simple language (not legal jargon)
- Cookie banner: Minimal tracking, clear opt-in/opt-out
- Data processing transparency: Show what data is collected and why

### 6.5 Offline Strategy

**Core Principle:** App remains fully functional without internet connection

**Local-First Architecture**
- All data stored in local SQLite database
- Writes happen locally first (optimistic UI updates)
- Background sync when connection available
- Conflict resolution: Last-write-wins for MVP (CRDTs for future)

**Sync Strategy**
- Delta sync: Only changed records transmitted
- Sync triggers: App open, session end, manual pull-to-refresh
- Queue failed uploads for retry with exponential backoff
- Visual indicators: "Synced" vs "Pending sync" status

**Video Offline Handling**
- Videos uploaded when WiFi available (user preference)
- Cellular upload: Optional, requires user consent per upload
- Downloaded videos available offline
- Thumbnail caching for offline browsing

---

## 7. Non-Functional Requirements

### 7.1 Performance

**Mobile App Performance**
- Cold start time: < 2 seconds on mid-range device (2-year-old iPhone/Android)
- Screen transitions: < 200ms
- Workout entry form: < 1 second load time
- Analytics dashboard: < 2 seconds load with 1 year of data
- Video playback start: < 2 seconds
- Offline operation: No performance degradation without connectivity

**Backend Performance**
- API response time: < 300ms (p95), < 500ms (p99)
- Video upload processing: < 30 seconds for 30-second clip
- Database queries: < 100ms for analytics aggregations
- Concurrent users: Support 1,000 active users (MVP scale)

**Video Performance**
- Compression target: Reduce file size by 60-70% without quality loss
- Streaming: Adaptive bitrate for smooth playback on poor connections
- Thumbnail generation: < 5 seconds after upload

### 7.2 Reliability

**Uptime & Availability**
- Target SLA: 99.5% uptime (3.6 hours downtime/month acceptable for MVP)
- Planned maintenance: Off-peak hours with 48-hour notice

**Data Durability**
- Database backups: Daily automated backups, 30-day retention
- Video backups: Replicated to secondary region (AWS S3 replication)
- Recovery Time Objective (RTO): < 4 hours
- Recovery Point Objective (RPO): < 24 hours

**Error Handling**
- Graceful degradation: Analytics unavailable → show cached data
- Sync failures: Queue for retry, notify user after 3 failed attempts
- Video upload failures: Local storage + retry queue

### 7.3 Usability

**Mobile-First Design Principles**
- Large touch targets (minimum 44×44 points)
- One-handed operation for primary flows
- Optimized for gym environment (bright lighting, sweaty hands)
- Minimal text entry required
- Smart defaults to reduce taps

**Onboarding**
- First-time user tutorial: < 2 minutes
- Sample workout pre-loaded for demonstration
- Tooltips for key features
- Progressive disclosure: Advanced features hidden initially

**Accessibility**
- WCAG 2.1 Level AA compliance (target)
- Screen reader support for core features
- Sufficient color contrast (4.5:1 for normal text)
- Text scaling support (up to 200%)

### 7.4 Scalability (Future-Proofing)

**MVP Scale Targets**
- Users: 1,000 athletes
- Database: 100,000 sessions, 1,000,000 sets
- Video storage: 5TB total (5GB per user × 1,000 users)
- API throughput: 1,000 req/min peak

**Phase 2 Scale Targets (12 months post-launch)**
- Users: 10,000 athletes
- Database: 1M sessions, 10M sets
- Video storage: 50TB
- API throughput: 10,000 req/min peak

**Scaling Strategy**
- Horizontal scaling: Stateless API servers behind load balancer
- Database: Read replicas for analytics queries
- Caching: Redis for frequently accessed data (1RM, recent sessions)
- Video CDN: CloudFront for global delivery

### 7.5 Monitoring & Observability

**Application Monitoring**
- Error tracking: Sentry for crash reporting and error alerts
- Performance monitoring: New Relic or DataDog APM
- Uptime monitoring: Pingdom or UptimeRobot (5-minute intervals)

**Key Metrics to Track**
- Application errors: < 0.1% error rate
- API latency: p50, p95, p99 response times
- Database query performance: Slow query log (>100ms)
- Video processing queue depth
- Sync success rate: > 99%

**Alerts**
- Critical: API error rate > 1%, database down, auth service down
- Warning: API p95 latency > 500ms, sync success < 95%
- Email/SMS alerts to on-call engineer

**User Analytics**
- Daily/Monthly Active Users (DAU/MAU)
- Feature usage: Session logging rate, video uploads, analytics views
- Retention: Day 1, Day 7, Day 30 retention rates
- User flows: Funnel analysis for onboarding, session creation

---

## 8. User Experience & Design

### 8.1 Key User Flows

**Flow 1: First-Time User Onboarding**
1. Download app from App Store/Play Store
2. Create account (email + password or OAuth)
3. Enter profile: Name, body weight, gender, birth year
4. Input current 1RMs for key lifts (Snatch, C&J, Back Squat)
5. Watch 60-second tutorial video
6. View pre-loaded sample workout
7. Log first real session

**Flow 2: Daily Training Session Logging (Primary Flow)**
1. Open app → Tap "Start Workout"
2. Session auto-populated with today's date/time
3. Select exercise from library or recent exercises
4. Log set 1: Weight, Reps, RPE → "Add Set"
5. Repeat for subsequent sets
6. (Optional) Attach video to set
7. Move to next exercise
8. When done → "Finish Workout"
9. Enter overall session RPE
10. View session summary (total volume, duration)
11. Auto-sync to cloud

**Flow 3: Video Analysis**
1. From session summary → Tap video thumbnail
2. Video opens in full-screen player
3. Scrub through video, pause at key positions
4. Tap "Annotate" → Draw lines for bar path
5. Save annotated frame
6. Tap "Compare" → Select previous PR lift
7. Side-by-side view with synchronized playback
8. Rate lift quality (1-5 stars)

**Flow 4: Progress Check**
1. Open app → Navigate to "Analytics" tab
2. View lift progress chart (default: Snatch, last 90 days)
3. Tap chart → See all logged sessions for that lift
4. Change time range to 6 months
5. View volume load trend → Notice upward trend
6. Check ACWR → Currently 1.1 (green, optimal)
7. Navigate to "Frequency" → See 4-day/week average

### 8.2 Information Architecture

**Primary Navigation (Bottom Tab Bar)**
1. **Log** (Home) - Daily workout logging
2. **Analytics** - Progress charts and insights
3. **Videos** - Video library and analysis
4. **Profile** - Settings, 1RMs, account

**Secondary Navigation (In-screen)**
- Contextual actions (swipe gestures, long-press menus)
- Modal overlays for forms (avoid stacked screens)
- Pull-to-refresh for data sync

### 8.3 Design System

**Visual Language**
- Modern, clean, minimal design
- Focus on data readability and large touch targets
- Dark mode support for gym environments
- Accent color for primary actions (e.g., "Add Set" button)

**Typography**
- System fonts (San Francisco iOS, Roboto Android)
- Font sizes: 17pt body, 28pt headlines, 13pt captions
- Semibold weight for numbers/data emphasis

**Color Palette**
- Primary: Blue (#007AFF iOS style) or custom brand color
- Success: Green (PRs, good ACWR)
- Warning: Yellow/Orange (elevated volume)
- Danger: Red (high ACWR risk, errors)
- Neutral: Grays for secondary text and borders

**Components**
- Buttons: High-contrast, minimum 44pt height
- Forms: Large input fields, clear labels
- Charts: Interactive with tooltips, color-coded zones
- Cards: Rounded corners, subtle shadows

---

## 9. Success Metrics & KPIs

### 9.1 Product Metrics (Primary)

**Adoption**
- Target: 500 registered athletes in first 3 months post-launch
- Target: 50% conversion from registration to first logged session

**Engagement**
- Target: 60% Weekly Active Users (WAU) / Monthly Active Users (MAU)
- Target: Average 4 sessions logged per active user per week
- Target: 30% of users upload at least 1 video per week

**Retention**
- Target: Day 7 retention > 50%
- Target: Day 30 retention > 30%
- Target: 6-month retention > 20%

**Feature Usage**
- Analytics dashboard: 70% of weekly active users view analytics at least once/week
- Video analysis: 40% of weekly active users upload/view videos
- Program templates: 30% of users create at least 1 program (if in MVP)

### 9.2 Business Metrics (Future)

**Revenue (Post-MVP when monetization introduced)**
- Freemium conversion rate: > 10% to paid tier
- Average Revenue Per User (ARPU): $5-10/month
- Customer Lifetime Value (LTV): > $120

**Growth**
- Month-over-month user growth: > 20%
- Viral coefficient: > 0.3 (users inviting other users)
- App Store rating: > 4.5 stars

### 9.3 Technical Metrics

**Performance**
- API p95 latency: < 300ms
- App crash rate: < 0.5%
- Video upload success rate: > 95%
- Sync success rate: > 99%

**Infrastructure**
- Server CPU utilization: < 60% average
- Database query time: < 100ms average
- Video storage costs: < $0.10/GB/month

### 9.4 Qualitative Metrics

**User Satisfaction**
- Net Promoter Score (NPS): > 50
- Customer Satisfaction (CSAT): > 4.0/5.0
- Support ticket volume: < 5 tickets/100 active users/month

**Feedback Themes** (collect during beta)
- "This is exactly what I needed for my training"
- "I can finally see my volume trends clearly"
- "Video comparison feature is a game-changer"

### 9.5 Validation Metrics (MVP Hypothesis Testing)

**Hypothesis 1:** Athletes will consistently log sessions if it takes < 3 minutes
- **Metric:** Average session logging time
- **Target:** < 3 minutes for 80% of sessions
- **Measure:** Track time from "Start Workout" to "Finish Workout"

**Hypothesis 2:** Visual analytics will increase training motivation
- **Metric:** Session logging frequency before/after first analytics view
- **Target:** 20% increase in session frequency after engaging with analytics
- **Measure:** Compare sessions/week in first 7 days vs days 8-30

**Hypothesis 3:** Video tools will drive engagement and retention
- **Metric:** Retention rate for users who upload videos vs those who don't
- **Target:** 50% higher Day 30 retention for video users
- **Measure:** Cohort analysis by video upload behavior

---

## 10. MVP Scope Definition

### 10.1 MVP Inclusions (Must-Have)

**Core Features**
✅ Mobile app (iOS and Android)
✅ User registration and authentication
✅ Exercise library (50+ exercises)
✅ Session logging with set-by-set tracking
✅ 1RM tracking and management
✅ Volume load calculation and analytics
✅ Lift progress charts (Snatch, C&J, Squats)
✅ ACWR calculation and visualization
✅ Training frequency heatmap
✅ Video upload and storage
✅ Video playback with slow-motion
✅ Side-by-side video comparison
✅ Manual video annotation (drawing tools)
✅ Offline functionality with sync
✅ Data export (CSV)

**Technical Must-Haves**
✅ Local-first architecture (SQLite + cloud sync)
✅ Video compression on upload
✅ Secure authentication (JWT)
✅ Basic error tracking (Sentry)
✅ Privacy controls for video sharing

### 10.2 MVP Exclusions (Deferred to Post-MVP)

**Phase 2 Features (3-6 months post-launch)**
❌ Automated bar path tracking (requires ML/CV)
❌ AI-powered technique feedback
❌ VBT device integrations (GymAware, PUSH, etc.)
❌ Force plate integrations
❌ Wellness tracking (sleep, soreness, stress)
❌ Nutrition logging
❌ Body composition tracking (beyond weight)
❌ Social features (following other athletes, leaderboards)
❌ Coach-athlete relationship features
❌ Team/gym management
❌ Community program marketplace
❌ Advanced statistical modeling (e.g., fatigue curves)

**Technical Enhancements (Post-MVP)**
❌ Real-time collaborative features
❌ Advanced caching strategies
❌ GraphQL API (stick with REST for MVP)
❌ Microservices architecture (monolith for MVP)
❌ Multi-region deployment
❌ Advanced analytics (ML-powered insights)

### 10.3 MVP Development Phases

**Phase 1: Foundation (Weeks 1-4)**
- [ ] Set up development environment and infrastructure
- [ ] Database schema design and migrations
- [ ] API authentication and core endpoints
- [ ] Mobile app navigation and basic screens
- [ ] Exercise library data model

**Phase 2: Core Logging (Weeks 5-8)**
- [ ] Session creation and logging UI
- [ ] Set-by-set input forms
- [ ] 1RM management interface
- [ ] Volume load calculations
- [ ] Local storage and offline mode

**Phase 3: Analytics (Weeks 9-11)**
- [ ] Progress charts implementation
- [ ] ACWR calculation and visualization
- [ ] Frequency heatmap
- [ ] Dashboard UI and data refresh

**Phase 4: Video (Weeks 12-15)**
- [ ] Video upload and compression pipeline
- [ ] Video library and playback
- [ ] Slow-motion and frame controls
- [ ] Side-by-side comparison
- [ ] Annotation tools

**Phase 5: Polish & Testing (Weeks 16-18)**
- [ ] Bug fixes and performance optimization
- [ ] User onboarding flow
- [ ] Beta testing with 20-30 athletes
- [ ] Analytics instrumentation
- [ ] App Store preparation

**Total Timeline: 18 weeks (4.5 months) to MVP**

---

## 11. Go-to-Market Strategy (EMEA Launch)

### 11.1 EMEA Launch Plan

**Pre-Launch (Weeks -8 to 0)**
- **Beta Testing (30 athletes across 5 EMEA countries)**
  - UK: 10 athletes (English speakers, app adoption validation)
  - Germany: 8 athletes (test German language, pricing sensitivity)
  - France: 5 athletes (Francophone market validation)
  - Nordics: 4 athletes (tech-savvy early adopters)
  - Ireland: 3 athletes (EU timezone, English-speaking)
- **Localization Sprint**
  - English UI/UX finalization
  - German translation (professional translation service)
  - Date/time format testing (DD/MM/YYYY, 24h clock)
  - kg default, currency display (EUR/GBP)
- **Landing Page & Waitlist**
  - Multilingual landing page (English, German)
  - Waitlist with country selection
  - Clear GDPR messaging and privacy assurances
- **Influencer Outreach (EMEA-specific)**
  - UK: Strength Shop, British Weightlifting community
  - Germany: German Weightlifting Federation (BVDG) connections
  - European Instagram influencers (10k-50k followers)

**Launch Week (Week 0 - Q1 2026)**
- **App Store Submission**
  - iOS App Store (EU region)
  - Google Play Store (EU region)
  - App Store localization (English, German)
- **Announcement Strategy**
  - Launch announcement on European weightlifting forums
  - Reddit: r/weightlifting, r/weightroom (EMEA time zones)
  - Facebook: European Weightlifting groups
  - Instagram: Targeted European hashtags (#weightliftingUK, #kraftdreikampf)
- **Partnership Announcements**
  - Partner with 2-3 European weightlifting clubs for pilot
  - Co-marketing with European equipment brands (Eleiko, ATX, Rogue Europe)
- **Press Outreach**
  - European fitness/tech publications
  - National weightlifting federation newsletters
  - CrossFit/functional fitness media (strong overlap in Europe)

**Post-Launch (Weeks 1-12)**
- **Content Strategy (Localized)**
  - Weekly blog posts in English (German translation within 3 months)
  - Technique analysis content featuring European athletes
  - Competition prep guides for European championships
  - European coach interviews and testimonials
- **Community Building**
  - Daily engagement in European weightlifting communities
  - Weekly progress reports shared publicly
  - User success stories (with permission)
  - Virtual meetups with early adopters (EU time zones)
- **Iteration & Feedback**
  - In-app surveys (English + German)
  - 1-on-1 user interviews (10-15 athletes)
  - Weekly feature prioritization based on feedback
  - Bug fixes prioritized by region impact

### 11.2 EMEA-Specific Acquisition Channels

**Primary Channels (Months 0-6)**

1. **European Weightlifting Communities**
   - UK: British Weight Lifting forums, Strength Shop community
   - Germany: BVDG forums, German CrossFit community
   - France: FFH (French Federation) community
   - Multi-country: r/weightlifting (peak EU hours posting)
   - Facebook: All Things Gym EU, European Weightlifting groups

2. **Local Club Partnerships**
   - Partner with 5-10 European weightlifting clubs for pilot programs
   - Offer club bulk onboarding (free for first 6 months)
   - Club coach referral program (future coach feature early access)
   - Local competition sponsorships (UK, Germany, France nationals)

3. **European Influencer Marketing**
   - Target: 10 micro-influencers (5k-20k followers) across EMEA
   - Focus: Technique analysis, progress tracking content
   - Budget: €200-500 per post (product placement + affiliate link)
   - Key influencers: UK garage gym lifters, German strength coaches

4. **Content Marketing (European Focus)**
   - Blog: European training methodologies (Bulgarian, Russian, German systems)
   - YouTube: Technique breakdowns of European champions
   - Podcast: European weightlifting podcast appearances
   - Competition calendars: European Championships, National qualifiers

**Secondary Channels (Months 3-12)**

5. **Paid Advertising (Limited Budget)**
   - Instagram ads (€1,000/month test budget)
   - Targeting: EMEA, ages 20-35, interests: weightlifting, CrossFit, strength
   - Facebook ads: European weightlifting groups (retargeting)
   - Google App Campaigns (ASO boost)

6. **SEO & App Store Optimization**
   - Keywords: "weightlifting tracker UK", "Gewichtheben App", "haltérophilie app"
   - App Store localization (English, German, French)
   - Backlinks from European weightlifting blogs/forums

7. **Federation & Organization Partnerships**
   - British Weight Lifting (BWL) partnership discussions
   - BVDG (German) newsletter feature
   - EWF (European Weightlifting Federation) awareness

### 11.3 Pricing Strategy: Free MVP → Coach Monetization

**Phase 1: Free for All Athletes (Months 0-12)**
- **Rationale:** Build network effects, validate product-market fit in EMEA
- No feature limitations, no paywall
- Focus: User acquisition, engagement, retention
- Target: 500 athletes (Month 6), 2,000 athletes (Month 12)

**Phase 2: Introduce Coach Platform Beta (Months 12-18)**
- **Athletes remain free forever** (no athlete paywall)
- **Coach subscription introduced:**
  - Coach tier: €39/month or €390/year (12 months for 10)
  - Features: Athlete roster management, program assignment, analytics dashboard
  - Target: 50 coaches (Month 18), each managing 5-20 athletes
- **Network effect:** Coaches bring athletes to platform, athletes create demand for coach features

**Phase 3: Scale Coach Platform (Months 18-24)**
- Refine coach pricing based on EMEA willingness-to-pay data
- Introduce coach tiers: Solo (€39/mo), Team (€99/mo for 3 coaches), Club (€249/mo)
- Maintain athlete platform free to maximize network effects
- Revenue target: €10k MRR by Month 24

**Why Free Athletes + Paid Coaches Model Works:**
1. **Network Effects:** More athletes = more valuable for coaches
2. **Lower CAC:** Athletes recruit each other organically
3. **Upsell Path:** Athletes become coaches, coaches bring athletes
4. **Competitive Moat:** Switching costs high once coach-athlete relationships formed

### 11.4 Launch Budget (EMEA MVP)

**Total Launch Budget: €15,000**

| Category | Budget | Allocation |
|----------|--------|------------|
| Product Development | €0 | Internal team |
| Beta Testing | €1,000 | Beta tester incentives (€30 gift cards × 30) |
| Localization | €2,000 | Professional German/French translation |
| Landing Page & Web | €1,500 | Design, development, hosting (1 year) |
| Influencer Marketing | €3,000 | 10 micro-influencers (€300 each) |
| Paid Advertising | €3,000 | Instagram/Facebook ads (€1k/mo × 3 months) |
| Club Partnerships | €2,000 | Competition sponsorships, club materials |
| Content Creation | €1,500 | Video production, blog content |
| Legal & Compliance | €1,000 | GDPR legal review, privacy policy, terms |

**Month 6 Review:** Adjust budget allocation based on channel performance

---

## 12. Risks & Mitigation Strategies

### 12.1 Product Risks

**Risk 1: Low User Adoption**
- **Likelihood:** Medium
- **Impact:** High
- **Mitigation:**
  - Extensive beta testing with target users
  - Onboarding optimization (maximize Day 1 retention)
  - Referral program for early adopters
  - Active presence in weightlifting communities

**Risk 2: Low Engagement/Session Logging Compliance**
- **Likelihood:** Medium
- **Impact:** High
- **Mitigation:**
  - Make logging as frictionless as possible (< 3 min goal)
  - Push notifications for workout reminders (opt-in)
  - Weekly progress emails to re-engage users
  - Gamification elements (streaks, milestones)

**Risk 3: Video Storage Costs Spiral**
- **Likelihood:** Low
- **Impact:** Medium
- **Mitigation:**
  - Aggressive video compression (60-70% reduction)
  - Storage caps (5GB free tier)
  - Option for users to manage/delete old videos
  - Monitor storage costs per user closely

### 12.2 Technical Risks

**Risk 4: Offline Sync Conflicts**
- **Likelihood:** Low
- **Impact:** Medium
- **Mitigation:**
  - Last-write-wins strategy for MVP (simple)
  - Timestamp-based conflict detection
  - Alert users to potential data conflicts
  - Clear sync status indicators

**Risk 5: Video Playback Performance Issues**
- **Likelihood:** Medium
- **Impact:** Medium
- **Mitigation:**
  - Extensive testing on low-end devices
  - Adaptive bitrate streaming
  - Local caching of recently viewed videos
  - Graceful degradation for slow connections

**Risk 6: Scalability Bottlenecks**
- **Likelihood:** Low (for MVP scale)
- **Impact:** Medium
- **Mitigation:**
  - Load testing before launch
  - Monitoring and alerting on key metrics
  - Horizontal scaling plan documented
  - Database indexing for frequent queries

### 12.3 Market Risks

**Risk 7: Competitor Launches Similar Product**
- **Likelihood:** Medium
- **Impact:** Medium
- **Mitigation:**
  - Fast execution to market (18-week timeline)
  - Focus on weightlifting-specific features (not general fitness)
  - Build strong early community
  - Iterate quickly based on user feedback

**Risk 8: Low Willingness to Pay (Future)**
- **Likelihood:** Medium
- **Impact:** High (post-MVP)
- **Mitigation:**
  - Validate value prop during free MVP phase
  - Gather WTP data through surveys
  - Offer annual discount (reduce churn)
  - Ensure free tier provides real value (not crippled)

---

## 13. Product Evolution: Athlete MVP → Elite Coach Platform

### Strategic Roadmap Overview

Our product evolution follows a deliberate bottom-up strategy: **build for athletes first, scale to elite coaches second, dominate the high-performance coaching market third.** Each phase builds on the previous, creating compounding network effects and a defensible moat.

**Phase 1 (Months 0-6):** Athlete MVP - Prove core value proposition with self-coached athletes
**Phase 2 (Months 6-12):** Enhanced Athlete Features - Increase engagement and retention
**Phase 3 (Months 12-18):** Coach Platform Beta - Introduce coach-athlete collaboration features
**Phase 4 (Months 18-24):** Elite Coach Platform - Full-featured coach management system
**Phase 5 (Months 24+):** AI & Automation - Advanced analytics and intelligent programming

---

### 13.1 Phase 2: Enhanced Athlete Features (Months 6-12)

**Goal:** Increase athlete engagement, retention, and create demand for coach features

**Social & Community Features**
- **Rationale:** Athletes want to share progress and learn from peers
- Follow other athletes (privacy controls, opt-in)
- Public PR feed (celebrate personal records)
- Comment on lifts (technique feedback from community)
- Training crew features (train together virtually)
- **Success Metric:** 40% of athletes engage with social features weekly

**Advanced Load Management**
- Training Monotony and Strain calculations (Foster model)
- Fatigue:Fitness modeling for readiness assessment
- Deload week recommendations based on ACWR trends
- Personalized volume tolerance algorithms
- **Success Metric:** 60% of athletes view load management insights weekly

**Body Composition & Competition Prep**
- Photo-based progress tracking (before/after comparisons)
- Weight class management for competition prep
- Sinclair/Robi coefficient auto-updates with weight changes
- Competition countdown timers and peak preparation
- **Success Metric:** 30% of athletes use competition prep features

**Program Marketplace (Community-Driven)**
- Athletes share their training programs (free)
- Program reviews and ratings
- Clone and customize popular programs
- **Seeds demand for coach-created programs** → Future monetization
- **Success Metric:** 500+ community programs shared, 25% adoption rate

---

### 13.2 Phase 3: Coach Platform Beta (Months 12-18)

**Goal:** Launch coach-athlete collaboration features and validate coach monetization model

**Strategic Context:** By Month 12, we have 2,000+ athletes using the platform daily. Many are asking for coach features to share data with their remote coaches. This creates organic demand for a two-sided marketplace.

#### 13.2.1 Coach Onboarding & Dashboard

**Coach Registration**
- Separate coach account type (€39/month subscription)
- Coach profile with credentials, certifications, specializations
- Portfolio of athletes managed
- Public coach profile for athlete discovery

**Coach Dashboard (Mission Control)**
- Roster view: All athletes with key metrics (recent sessions, ACWR, compliance %)
- Athlete comparison: Side-by-side analytics across roster
- Flagging system: Auto-flag athletes with high ACWR, low compliance, or performance drops
- Bulk actions: Assign programs to multiple athletes at once
- Calendar view: See entire roster's training schedule

**Success Metric:** 50 coaches onboard (Month 18), each managing 5-20 athletes (average: 10 athletes/coach)

#### 13.2.2 Coach-Athlete Relationship Management

**Athlete Invitations**
- Coaches invite athletes via email/link
- Athletes approve coach access to their data (privacy controls)
- Athletes can have multiple coaches (e.g., weightlifting coach + nutrition coach)
- Revoke access at any time

**Data Visibility Controls**
- Athlete chooses what data coach can see (sessions, videos, analytics, notes)
- Coach can view historical data after athlete approval
- Audit log of coach data access

**Communication Tools**
- In-app messaging (coach ↔ athlete)
- Session feedback: Coach comments on specific sessions
- Video feedback: Coach annotates videos with technique cues
- Weekly check-ins: Structured feedback forms

**Success Metric:** 70% of coached athletes engage with feedback features weekly

#### 13.2.3 Coach Programming Tools

**Program Builder (Coach Version)**
- Multi-week program templates
- Block periodization support (accumulation, intensification, realization, taper)
- Exercise prescription with % of 1RM, RPE, or absolute weights
- Auto-regulation rules (e.g., "reduce volume by 20% if ACWR > 1.5")
- Program library: Save and reuse templates

**Program Assignment**
- Assign programs to individual athletes or athlete groups
- Clone programs across athletes with auto-scaled percentages
- Mid-cycle adjustments (modify program for entire roster or individual athletes)
- Program analytics: Track athlete compliance with prescribed programming

**Auto-Regulation Features**
- Load recommendations based on ACWR, session RPE, and readiness
- Auto-adjust training loads if athlete misses sessions
- Fatigue-based deload triggers

**Success Metric:** 80% of coaches use program builder, assign avg. 2 programs/athlete

#### 13.2.4 Coach Business Tools

**Athlete Management**
- Onboarding workflows for new athletes
- Progress reports: Auto-generated monthly summaries for athlete check-ins
- Goal tracking: Set and track athlete goals (e.g., "Increase snatch by 10kg in 12 weeks")

**Payment Integration (Phase 3.5, Month 15+)**
- Stripe integration for coach subscription billing
- Future: Coach-athlete payment processing (coaches charge athletes via platform)
- Invoicing and receipt generation

**Success Metric:** 90% coach subscription retention, <5% churn

---

### 13.3 Phase 4: Elite Coach Platform (Months 18-24)

**Goal:** Become the definitive platform for elite weightlifting coaches managing high-performance athletes

**Strategic Context:** With coach platform validated (50+ coaches, 500+ coached athletes), we now build features that differentiate us from generic coaching apps. Our focus is on **professional coaches managing national/international level athletes.**

#### 13.3.1 Advanced Analytics for Coaches

**Roster-Wide Analytics**
- Aggregate analytics across entire roster (avg. volume, avg. ACWR, compliance rates)
- Cohort analysis: Compare athlete groups (youth vs seniors, men vs women)
- Outlier detection: Automatically flag athletes deviating from norms

**Performance Forecasting**
- Predictive models for competition performance based on training trends
- Peak performance timing: Optimal taper length for each athlete
- Injury risk scoring based on load management history

**Comparative Analytics**
- Benchmark athletes against national/international standards
- Technique comparison: Side-by-side video analysis across roster
- Exercise efficacy: Which exercises correlate with performance improvements?

**Success Metric:** 70% of coaches use advanced analytics monthly

#### 13.3.2 VBT & Hardware Integrations

**Device Integrations**
- APIs for VBT devices: GymAware, PUSH, Vitruve, RepOne
- Force plate integrations: Hawkin Dynamics, VALD ForceDecks
- Auto-sync athlete data from devices to platform
- Real-time velocity feedback during training

**VBT-Based Programming**
- Velocity-based load prescriptions (e.g., "3 reps @ 0.8 m/s")
- Readiness assessment via velocity drop-offs
- Load-velocity profiling for individualized programming

**Success Metric:** 30% of coaches integrate VBT devices

#### 13.3.3 Team & Club Management

**Multi-Coach Organizations**
- Gym/club accounts with multiple coach seats
- Role-based permissions (head coach, assistant coach, intern)
- Shared athlete rosters across coaching staff
- Internal notes and communication between coaches

**Athlete Development Pathways**
- Long-term athlete development tracking (youth → senior)
- Transfer athletes between coaches (e.g., youth coach → senior coach)
- Historical data preserved across coaching transitions

**Club Analytics**
- Club-wide performance dashboards
- Competition prep tracking for entire team
- Resource allocation: Coach workload balancing

**Success Metric:** 10 clubs/organizations onboard, 3+ coaches per club

#### 13.3.4 Competition Management

**Meet Prep Tools**
- Competition calendar integration (IWF, EWF, national federations)
- Athlete roster for upcoming competitions
- Weigh-in tracking and weight class management
- Attempt selection calculators (Snatch, C&J)

**Competition Day Tools**
- Live competition tracking (warm-up attempts, competition lifts)
- Real-time coach notes and strategy adjustments
- Post-competition analysis and debrief reports

**Success Metric:** 100+ competitions managed via platform (Month 24)

---

### 13.4 Phase 5: AI & Automation (Months 24-36)

**Goal:** Leverage ML/AI to provide coaching superpowers—automated insights that would take human coaches hours to identify

#### 13.4.1 Automated Video Analysis

**Computer Vision for Technique**
- Automated bar path tracking (finally viable with enough video data)
- Position detection (start, transition, catch, finish)
- Automated angle measurements (hip, knee, ankle, torso)
- Fault detection: Identify common technique errors automatically

**Technique Scoring**
- AI-powered technique ratings (1-10 scale)
- Consistency scores: Track technique reliability across sessions
- Technique trends: Identify gradual improvements or regressions

**Success Metric:** 90% bar path tracking accuracy, 50% of athletes use automated analysis

#### 13.4.2 Predictive Analytics & Insights

**Injury Risk Prediction**
- ML models trained on load management + injury history
- Risk scoring: Predict injury likelihood for next 4 weeks
- Proactive alerts to coaches when athletes enter high-risk zones

**Performance Forecasting**
- Predict competition totals based on training trends
- Optimal peak timing for major competitions
- Personalized recovery recommendations

**Fatigue & Readiness Models**
- Integrated fatigue scoring (load + wellness + VBT data)
- Readiness predictions for daily training
- Auto-adjust programming based on readiness

**Success Metric:** 15% reduction in injury rates, 5% improvement in competition performance

#### 13.4.3 Intelligent Programming

**AI Program Generation**
- Input athlete profile, competition date, current maxes → Generate periodized program
- Auto-regulation built-in (adjust based on athlete response)
- Exercise selection based on individual athlete weaknesses

**Adaptive Programming**
- Real-time program adjustments based on athlete performance
- Intelligent load recommendations (ACWR-aware, fatigue-aware)
- Competition peaking algorithms

**Success Metric:** 500+ coaches use AI programming tools, 80% satisfaction rate

---

### 13.5 Long-Term Vision (3-5 Years)

**Market Position:** Industry-standard platform for elite weightlifting coaching
- 50,000+ athletes globally
- 1,000+ professional coaches
- Used by national federations and Olympic training centers
- Revenue: €500k+ MRR, primarily from coach subscriptions

**Platform Ecosystem:**
- Open API for 3rd party integrations (nutrition apps, recovery tools, wearables)
- Developer community building plugins and extensions
- White-label solutions for national federations

**Expansion Opportunities:**
- Adjacent sports: Powerlifting, CrossFit, strongman, track & field (throws)
- Broader strength & conditioning market
- Corporate wellness and athletic development programs

---

## 14. Appendices

### 14.1 Exercise Library (MVP Seed Data)

**Olympic Lifts & Variations (20)**
- Snatch (full, power, hang, blocks, muscle)
- Clean (full, power, hang, blocks, muscle)
- Jerk (split, power, push, squat)
- Clean & Jerk (full, power, variations)

**Squats (12)**
- Back Squat (high bar, low bar, pause, tempo)
- Front Squat (full, pause, tempo)
- Overhead Squat
- Bulgarian Split Squat
- Goblet Squat

**Pulls (10)**
- Clean Pull (from floor, hang, deficit)
- Snatch Pull (from floor, hang, deficit)
- Deadlift (conventional, sumo, Romanian)
- Good Morning

**Presses (8)**
- Push Press
- Strict Press
- Push Jerk
- Bench Press
- Incline Press
- Dips

**Accessories (10)**
- Rows (barbell, dumbbell, cable)
- Pull-ups/Chin-ups
- Plank variations
- Ab wheel
- Back extensions
- Lunges

**Total: 60 exercises in library**

### 14.2 Glossary of Terms

- **ACWR (Acute:Chronic Workload Ratio):** Ratio of recent training load (acute, 7 days) to longer-term average (chronic, 28 days), used to assess injury risk
- **CMJ (Countermovement Jump):** Vertical jump test for lower body power
- **IMTP (Isometric Mid-Thigh Pull):** Strength test measuring maximal pulling force
- **RPE (Rate of Perceived Exertion):** Subjective scale (1-10) for effort level
- **Sinclair Coefficient:** Formula for comparing weightlifting totals across body weights
- **VBT (Velocity-Based Training):** Training methodology using bar speed to prescribe loads
- **Volume Load:** Total work performed, calculated as sets × reps × weight

### 14.3 Competitive Analysis

**Existing Solutions:**
1. **Strong/JEFIT** - Generic workout trackers, not weightlifting-specific
2. **WinWoTa** - Legacy desktop software, outdated UI, Windows-only
3. **Excel/Google Sheets** - DIY solutions, time-consuming, no analytics
4. **Coaches' Apps** (TrainHeroic, etc.) - Designed for coaches, not athletes

**Ascend's Differentiation:**
- Built specifically for Olympic weightlifting (not general fitness)
- Mobile-first, offline-capable
- Integrated video analysis tools
- Automatic load management analytics (ACWR)
- Athlete-centric (not coach-centric) for MVP
- Modern, intuitive UX designed for gym environment

### 14.4 References & Inspiration

**Scientific Foundation:**
- Gabbett, T. J. (2016). The training-injury prevention paradox (ACWR research)
- Banyard et al. (2019). Velocity-based training literature
- Mann et al. (2010). Periodization studies in weightlifting

**Product Inspiration:**
- Strava: Social features, progress tracking, data visualization
- MyFitnessPal: Logging simplicity, large database
- Hudl: Video analysis tools for sports coaching

---

## Document Change Log

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | Oct 2, 2025 | Lead Product Strategist | Initial coach-focused PRD |
| 2.0 | Oct 6, 2025 | Product Strategy Team | Complete rewrite for athlete-focused MVP, added technical architecture, expanded requirements |

---

**Next Steps:**
1. Review and approve PRD with stakeholders
2. Create detailed wireframes for key user flows
3. Set up development environment and project infrastructure
4. Begin Phase 1 development (Foundation)
5. Recruit beta testing cohort (target: 30 athletes)

**Questions or Feedback:** Please submit PRD feedback via GitHub issues or contact product team directly.
