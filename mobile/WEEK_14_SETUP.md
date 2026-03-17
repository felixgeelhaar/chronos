# Week 14: Offline Mode & Sync - Setup Instructions

## Overview

Week 14 implements offline-first functionality using WatermelonDB for local data persistence and a custom sync engine for bidirectional synchronization with the backend API.

## New Dependencies

```json
{
  "@nozbe/watermelondb": "^0.27.1",
  "@nozbe/with-observables": "^1.6.0",
  "@react-native-community/netinfo": "^11.3.1"
}
```

## Installation Steps

```bash
# Install dependencies
npm install

# iOS: Install pods
cd ios && pod install && cd ..

# iOS: Enable JSI (Optional, for better performance)
# Add to ios/Podfile:
# pod 'simdjson', path: '../node_modules/@nozbe/simdjson'

# Android: No additional setup required
```

## Architecture

### Database Schema (`src/database/schema.ts`)

The WatermelonDB schema defines 6 tables:

1. **users** - User profiles with body weight tracking
2. **sessions** - Workout sessions with date, notes, and totals
3. **sets** - Individual exercise sets with weight, reps, RPE
4. **videos** - Video files with upload status tracking
5. **sync_queue** - Queue for failed sync operations with retry logic
6. **one_rep_maxes** - Historical 1RM records

All tables include:
- `server_id` - Remote server ID (null for unsynced records)
- `is_synced` - Boolean flag for sync status
- `synced_at` - Timestamp of last successful sync
- `created_at`, `updated_at` - Standard timestamps

### Database Models (`src/database/models/`)

WatermelonDB models with TypeScript decorators:

- **User.ts** - User profile model
- **Session.ts** - Workout session with has_many relationship to sets
- **Set.ts** - Exercise set with belongs_to relationship to session
- **Video.ts** - Video metadata and upload tracking
- **SyncQueue.ts** - Retry queue for failed operations
- **OneRepMax.ts** - 1RM history records

### Sync Engine (`src/services/sync/SyncEngine.ts`)

Bidirectional sync with conflict resolution:

**Features:**
- Automatic sync on network reconnection
- Periodic background sync (default: 5 minutes)
- Push local changes to server
- Pull remote changes from server
- Retry queue for failed operations
- Network status awareness
- Conflict resolution (server wins by default)

**Sync Flow:**
1. Check network connectivity
2. Check authentication status
3. Push unsynced local changes
4. Pull new remote changes
5. Process retry queue
6. Update sync status

**Status Updates:**
- `isSyncing: boolean` - Currently syncing
- `lastSyncedAt: Date` - Last successful sync
- `pendingChanges: number` - Unsynced local changes
- `errors: string[]` - Recent sync errors

### Sync Status Indicator (`src/shared/components/SyncStatusIndicator.tsx`)

Visual feedback component with two modes:

**Compact Mode:**
- Small circular indicator
- Color-coded status (green/orange/red/blue)
- Icon showing sync state

**Detailed Mode:**
- Full status card
- Sync progress
- Last synced timestamp
- Error messages
- Tap to trigger manual sync

**Status States:**
- 📴 Offline (orange) - No network connection
- ⏳ Pending (orange) - Local changes awaiting sync
- ⚠️ Error (red) - Sync failed
- ✓ Synced (green) - All data synced
- 🔄 Syncing (blue) - Currently syncing

## Usage

### Initialize Sync Engine

Add to `App.tsx` or root component:

```typescript
import { syncEngine } from './services/sync';
import { useEffect } from 'react';

function App() {
  useEffect(() => {
    // Initialize with 5-minute auto-sync
    syncEngine.initialize(300000);

    return () => {
      syncEngine.stop();
    };
  }, []);

  return <AppNavigator />;
}
```

### Manual Sync

```typescript
import { syncEngine } from './services/sync';

// Trigger manual sync
await syncEngine.sync();

// Get current sync status
const status = await syncEngine.getStatus();
console.log('Pending changes:', status.pendingChanges);
```

### Subscribe to Sync Status

```typescript
import { syncEngine } from './services/sync';

const unsubscribe = syncEngine.subscribe((status) => {
  console.log('Sync status:', status);
  if (status.errors.length > 0) {
    Alert.alert('Sync Error', status.errors[0]);
  }
});

// Later: cleanup
unsubscribe();
```

### Offline-First Data Operations

```typescript
import database from './database';
import { Session, Set } from './database/models';

// Create session offline
await database.write(async () => {
  const session = await database.get<Session>('sessions').create(s => {
    s.userId = currentUserId;
    s.date = format(new Date(), 'yyyy-MM-dd');
    s.totalVolume = 0;
    s.totalSets = 0;
    s.isSynced = false; // Mark as unsynced
  });

  // Create sets
  await database.get<Set>('sets').create(set => {
    set.sessionId = session.id;
    set.exerciseName = 'Squat';
    set.weight = 100;
    set.reps = 5;
    set.setOrder = 1;
    set.volume = 500;
    set.estimatedOneRepMax = 112.5;
    set.isSynced = false;
  });
});

// Sync will automatically push to server when online
```

### Query Local Data

```typescript
import database from './database';
import { Session } from './database/models';
import { Q } from '@nozbe/watermelondb';

// Query sessions
const recentSessions = await database
  .get<Session>('sessions')
  .query(
    Q.sortBy('date', Q.desc),
    Q.take(10)
  )
  .fetch();

// Query with relationships
const session = await database.get<Session>('sessions').find(sessionId);
const sets = await session.sets.fetch();
```

## Conflict Resolution

The sync engine uses **server-wins** conflict resolution by default:

1. Local changes are pushed first
2. If a conflict occurs, server version is kept
3. Local changes are overwritten with server data
4. Failed operations are added to sync queue for retry

### Custom Conflict Resolution

To implement custom logic, modify `SyncEngine.pullRemoteChanges()`:

```typescript
// Check timestamps to determine winner
const localUpdated = localSession.updatedAt.getTime();
const remoteUpdated = new Date(remoteSession.updated_at).getTime();

if (remoteUpdated > localUpdated) {
  // Server is newer, update local
  await localSession.update(/* ... */);
} else if (localUpdated > remoteUpdated) {
  // Local is newer, push to server
  await sessionService.updateSession(/* ... */);
}
```

## Sync Queue & Retry Logic

Failed operations are queued for retry:

- **Max retries:** 5 attempts
- **Retry strategy:** Exponential backoff (not yet implemented)
- **Error tracking:** Last error message stored
- **Manual retry:** Clear queue to force retry

```typescript
// View sync queue
const queue = await database.get<SyncQueue>('sync_queue').query().fetch();

// Clear failed items after fixing issue
await database.write(async () => {
  for (const item of queue) {
    await item.destroyPermanently();
  }
});
```

## Performance Optimization

### Enable JSI (JavaScript Interface)

For improved performance, enable JSI in WatermelonDB:

**iOS (Podfile):**
```ruby
pod 'simdjson', path: '../node_modules/@nozbe/simdjson'
```

**Update database config:**
```typescript
const adapter = new SQLiteAdapter({
  schema,
  jsi: true, // Enable JSI
});
```

### Batch Operations

Use `batch()` for multiple operations:

```typescript
await database.write(async () => {
  await database.batch(
    session.prepareUpdate(s => { s.notes = 'Updated'; }),
    set1.prepareUpdate(s => { s.weight = 105; }),
    set2.prepareUpdate(s => { s.weight = 110; }),
  );
});
```

## Testing Offline Mode

### Simulate Network Conditions

**iOS Simulator:**
Settings → Developer → Network Link Conditioner → Enable

**Android Emulator:**
Extended Controls (three dots) → Cellular → Data Status → Denied

**Physical Device:**
Enable Airplane Mode

### Test Scenarios

1. **Create session offline → Go online → Verify sync**
2. **Edit session offline → Server also edited → Check conflict resolution**
3. **Delete locally → Sync → Verify server deletion**
4. **Sync fails → Check retry queue → Fix and retry**

## Troubleshooting

### Database Migration Errors

If schema changes cause errors:

```bash
# iOS
xcrun simctl --set simulator erase all

# Android
adb uninstall com.ascendmobile
```

### Sync Loop/Infinite Syncing

Check for:
- Incorrect `is_synced` flag updates
- Server returning different data than sent
- Timestamp comparison logic errors

### Memory Issues

Large datasets can cause memory issues:
- Implement pagination in queries
- Use `observe()` instead of `fetch()` for reactive updates
- Limit query results with `Q.take()`

### JSI Not Working

Verify:
1. Pod installation completed
2. `jsi: true` in adapter config
3. App fully rebuilt after enabling JSI

## Migration Strategy

To migrate existing data from API-only to offline-first:

1. **Initial sync on login:**
```typescript
// After successful login
await syncEngine.sync();
```

2. **Fetch and store recent data:**
```typescript
// Pull last 100 sessions
const sessions = await sessionService.listSessions(1, 100);
for (const session of sessions.sessions) {
  // Store in local database
}
```

3. **Switch to local-first:**
```typescript
// Query local database first
const localSessions = await database.get<Session>('sessions').query().fetch();

// Sync in background
syncEngine.sync();
```

## Next Steps

Week 15 will add:
- Push notifications for sync completion
- Background sync with iOS/Android background tasks
- Selective sync (choose what to sync)
- Data export/backup functionality

## Additional Resources

- [WatermelonDB Documentation](https://nozbe.github.io/WatermelonDB/)
- [React Native NetInfo](https://github.com/react-native-netinfo/react-native-netinfo)
- [Offline-First Apps](https://offlinefirst.org/)
