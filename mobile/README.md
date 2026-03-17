# Ascend Mobile

> 🏔️ Comprehensive weightlifting performance tracking mobile application

Ascend is a React Native mobile application designed to help athletes monitor progress, analyze form through video analysis, and achieve their strength training goals.

## ✨ Features

- 📊 **Comprehensive Workout Logging** - Track exercises, sets, reps, weight, and RPE
- 📈 **Progress Analytics** - ACWR monitoring, volume trends, and performance insights
- 🎥 **Video Recording & Analysis** - Record form, analyze technique, and track improvements
- 📴 **Offline Mode** - Full functionality offline with automatic synchronization
- 💪 **1RM Calculations** - Estimated one-rep max tracking and progression
- 🔔 **Smart Notifications** - Workout reminders, progress updates, and sync notifications
- 📱 **Cross-Platform** - Native iOS and Android experience
- 🎯 **Performance Optimized** - Fast image caching, efficient database queries, smooth animations

## 🛠 Tech Stack

### Core
- **React Native 0.74.1** - Cross-platform mobile framework
- **TypeScript 5.3.3** - Type-safe development
- **React Navigation 6.x** - Navigation and routing

### State & Data
- **Redux Toolkit 2.2.3** - Global state management
- **WatermelonDB 0.27.1** - Offline-first reactive database
- **AsyncStorage 1.23.1** - Key-value storage
- **Axios 1.6.8** - HTTP client

### UI & Performance
- **Victory Native 36.9.2** - Data visualization
- **React Native Fast Image 8.6.3** - Optimized image loading
- **React Native Haptic Feedback 2.2.0** - Tactile feedback
- **React Native SVG 15.2.0** - SVG rendering

### Media & Notifications
- **React Native Video 6.0.0** - Video playback
- **React Native Image Picker 7.1.0** - Camera access
- **Firebase Cloud Messaging 19.0.1** - Push notifications
- **Notifee 7.8.2** - Local notifications
- **React Native Background Fetch 4.2.3** - Background sync

## 📁 Project Structure

```
mobile/
├── src/
│   ├── app/                    # Application root
│   │   ├── App.tsx            # Main app component with providers
│   │   ├── navigation/        # Navigation configuration
│   │   └── store/             # Redux store setup
│   ├── features/              # Feature modules
│   │   ├── auth/             # Authentication (Login, Register)
│   │   ├── sessions/         # Workout sessions (List, Create, Detail)
│   │   ├── analytics/        # Progress analytics and charts
│   │   └── settings/         # App settings and profile
│   ├── shared/               # Shared resources
│   │   ├── components/       # Reusable components (Skeleton, Toast, ErrorBoundary)
│   │   ├── contexts/         # React contexts (Auth)
│   │   ├── hooks/            # Custom hooks (useDebounce, useThrottle)
│   │   └── utils/            # Utility functions (haptics, animations, performance)
│   ├── services/             # Business logic services
│   │   ├── api/              # API client and endpoints
│   │   ├── sync/             # Sync engine for offline support
│   │   ├── notifications/    # Notification and reminder services
│   │   └── background/       # Background task services
│   └── database/             # WatermelonDB configuration
│       ├── schema.ts         # Database schema (6 tables)
│       ├── models/           # Database models
│       ├── queryHelpers.ts   # Optimized query functions
│       └── index.ts          # Database initialization
├── __tests__/                # Test files
│   ├── components/          # Component tests
│   ├── hooks/               # Hook tests
│   └── services/            # Service tests
├── ios/                      # iOS native code
├── android/                  # Android native code
├── jest.config.js           # Jest configuration
├── jest.setup.js            # Test environment setup
└── package.json             # Dependencies
```

## Getting Started

### Prerequisites

- Node.js 18+ and npm/yarn
- Xcode 15+ (for iOS development on macOS)
- Android Studio (for Android development)
- CocoaPods (for iOS dependencies)
- React Native CLI

### Installation

1. **Install dependencies:**
   ```bash
   cd mobile
   npm install
   ```

2. **Install iOS dependencies:**
   ```bash
   cd ios && pod install && cd ..
   ```

3. **Set up environment variables:**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

### Running the App

#### iOS
```bash
npm run ios

# Or for specific simulator
npm run ios -- --simulator="iPhone 15 Pro"
```

#### Android
```bash
npm run android

# Or for specific device
npm run android -- --deviceId=<device_id>
```

#### Metro Bundler
```bash
npm start
```

### Development

#### Type Checking
```bash
npm run type-check
```

#### Linting
```bash
npm run lint

# Auto-fix
npm run lint -- --fix
```

#### Testing
```bash
# Run all tests
npm test

# Watch mode
npm test -- --watch

# Coverage
npm test -- --coverage
```

## 🏗 Architecture

### Offline-First Design

Ascend uses WatermelonDB for offline-first data persistence:

1. **Local Storage** - All data stored locally using SQLite
2. **Optimistic Updates** - Instant UI updates, sync in background
3. **Automatic Sync** - Background synchronization when online (every 15 minutes)
4. **Conflict Resolution** - Server-wins strategy for data conflicts
5. **Retry Queue** - Failed sync operations queued for retry

### Database Schema

Six core tables with proper indexing:
- **users** - User profiles and body weight tracking
- **sessions** - Workout sessions with date, volume, and set count
- **sets** - Individual exercise sets with weight, reps, RPE, and 1RM
- **videos** - Video metadata and upload status
- **one_rep_maxes** - Historical 1RM tracking per exercise
- **sync_queue** - Pending sync operations with retry logic

### State Management

- **Redux Toolkit** - Global application state
- **Local Component State** - React hooks for UI state
- **Database Observables** - Reactive queries with WatermelonDB
- **Context API** - Authentication and user context

### Performance Optimizations

#### Image Performance
- **FastImage** - Memory and disk caching
- **Preloading** - Background loading of upcoming images
- **Progressive Loading** - Show placeholders while loading

#### Database Optimization
- **Indexed Queries** - All frequent lookups use indexed columns
- **Batch Operations** - Multiple operations in single transaction
- **Lazy Loading** - Fetch data only when needed
- **Query Helpers** - Pre-optimized common queries

#### Rendering Performance
- **React.memo** - Prevent unnecessary re-renders
- **useMemo/useCallback** - Memoize expensive computations
- **Debouncing** - Input optimization (search, form validation)
- **Throttling** - Event rate limiting (scroll, gestures)

#### UI/UX Polish
- **Skeleton Screens** - Loading states with animated placeholders
- **Haptic Feedback** - Tactile response for all interactions
- **Smooth Animations** - 60 FPS animations using native driver
- **Toast Notifications** - Non-intrusive feedback

## 🧪 Testing Strategy

### Test Coverage Goals
- Branches: 70%
- Functions: 70%
- Lines: 70%
- Statements: 70%

### Unit Tests
- Component rendering and behavior
- Custom hook functionality
- Utility function logic
- Service methods

### Integration Tests
- API integration with mocked responses
- Database operations and queries
- Sync engine workflow
- Navigation flows

### Test Files
```
src/
  shared/
    hooks/
      __tests__/
        useDebounce.test.ts
        useThrottle.test.ts
    components/
      __tests__/
        Skeleton.test.tsx
        Toast.test.tsx
        ErrorBoundary.test.tsx
  services/
    sync/
      __tests__/
        SyncEngine.test.ts
```

### Running Tests
```bash
# All tests
npm test

# Watch mode during development
npm test -- --watch

# Coverage report
npm test -- --coverage

# Update snapshots
npm test -- -u
```

## 📊 Performance Monitoring

Built-in performance monitoring utilities:

```typescript
import { performanceMonitor } from '@/shared/utils/performanceMonitor';

// Measure async operations
await performanceMonitor.measure('fetchSessions', async () => {
  return await sessionService.getSessions();
});

// Measure sync operations
const result = await performanceMonitor.measureSync('calculateStats', () => {
  return calculateWorkoutStats(data);
});

// Generate report
performanceMonitor.logReport();
```

## 🔔 Notifications

### Push Notifications (Firebase)
- **Session Reminders** - Configurable workout reminders
- **Progress Updates** - Weekly summaries and achievements
- **Sync Complete** - Background sync notifications
- **Analysis Complete** - Video analysis results

### Local Notifications
- **Workout Streaks** - Celebrate consecutive workout days
- **Rest Day Reminders** - Prompt after 2+ days inactive
- **Custom Schedules** - Day-of-week and time configuration

### Background Tasks
- **Auto Sync** - Every 15 minutes when online
- **Headless Mode** - Android background sync even when app closed
- **Network Aware** - Only sync when connected

## 🎨 Features

### Offline-First Architecture

The app works fully offline using WatermelonDB for local storage:
- All data writes happen locally first
- Automatic sync when connection available
- Optimistic UI updates for fast UX
- Conflict resolution on sync

### Navigation Structure

- **Auth Stack:** Login, Register, Onboarding
- **Main Tabs:**
  - **Log:** Session creation and history
  - **Analytics:** Progress charts and ACWR
  - **Videos:** Video library and analysis
  - **Profile:** 1RMs, settings, export data

### State Management

Redux Toolkit with feature-based slices:
- `auth`: Authentication state
- `sessions`: Workout sessions
- `analytics`: Progress data
- `videos`: Video metadata
- `sync`: Sync status

## Building for Production

### iOS

1. **Update version in Xcode**
2. **Archive build:**
   ```bash
   cd ios
   xcodebuild -workspace AscendMobile.xcworkspace \
     -scheme AscendMobile \
     -configuration Release \
     -archivePath build/AscendMobile.xcarchive \
     archive
   ```
3. **Upload to App Store Connect**

### Android

1. **Update version in `android/app/build.gradle`**
2. **Generate release APK:**
   ```bash
   cd android
   ./gradlew assembleRelease
   ```
3. **Output:** `android/app/build/outputs/apk/release/app-release.apk`
4. **Upload to Play Console**

## Environment Variables

```env
API_BASE_URL=https://api.ascend.app
API_TIMEOUT=30000
ENV=development
```

## Troubleshooting

### iOS Build Fails
```bash
cd ios
rm -rf Pods Podfile.lock
pod install
cd ..
```

### Android Build Fails
```bash
cd android
./gradlew clean
cd ..
```

### Metro Cache Issues
```bash
npm start -- --reset-cache
```

## Contributing

1. Create feature branch from `develop`
2. Write tests for new features
3. Ensure lint and type checks pass
4. Submit PR with clear description

## License

Proprietary - All rights reserved
