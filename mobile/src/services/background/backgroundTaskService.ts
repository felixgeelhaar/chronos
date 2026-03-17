import BackgroundFetch from 'react-native-background-fetch';
import { syncEngine } from '../sync/SyncEngine';
import { notificationService } from '../notifications/notificationService';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { Platform } from 'react-native';

const BACKGROUND_SYNC_TASK = 'com.ascend.backgroundSync';
const LAST_SYNC_KEY = '@ascend:last_background_sync';

class BackgroundTaskService {
  private initialized = false;

  async initialize(): Promise<void> {
    if (this.initialized) return;

    try {
      // Configure background fetch
      const status = await BackgroundFetch.configure(
        {
          minimumFetchInterval: 15, // minutes (minimum allowed is 15 on iOS)
          stopOnTerminate: false, // Continue background tasks even after app is closed
          enableHeadless: true, // Enable headless tasks (Android)
          startOnBoot: true, // Start on device boot (Android)
          requiredNetworkType: BackgroundFetch.NETWORK_TYPE_ANY, // Run on any network
          requiresCharging: false, // Don't require charging
          requiresDeviceIdle: false, // Don't require device idle
          requiresBatteryNotLow: false, // Don't require battery not low
          requiresStorageNotLow: false, // Don't require storage not low
        },
        this.onBackgroundFetch,
        this.onBackgroundFetchTimeout
      );

      console.log('[BackgroundFetch] Status:', status);

      // Schedule periodic sync task
      await BackgroundFetch.scheduleTask({
        taskId: BACKGROUND_SYNC_TASK,
        delay: 900000, // 15 minutes in milliseconds
        periodic: true,
        forceAlarmManager: true, // Use AlarmManager for more reliable scheduling (Android)
        enableHeadless: true,
        stopOnTerminate: false,
        startOnBoot: true,
      });

      this.initialized = true;
      console.log('[BackgroundTaskService] Initialized');
    } catch (error) {
      console.error('[BackgroundTaskService] Initialization failed:', error);
    }
  }

  private async onBackgroundFetch(taskId: string): Promise<void> {
    console.log('[BackgroundFetch] Task started:', taskId);

    try {
      // Perform background sync
      const syncResult = await syncEngine.sync();

      // Store last sync time
      await AsyncStorage.setItem(LAST_SYNC_KEY, new Date().toISOString());

      // Show notification if there were changes synced
      const preferences = await this.getNotificationPreferences();
      if (preferences.syncComplete && syncResult.changesSynced > 0) {
        await notificationService.displayNotification({
          title: 'Sync Complete',
          body: `${syncResult.changesSynced} workout${syncResult.changesSynced > 1 ? 's' : ''} synced`,
          data: { type: 'sync_complete' },
        });
      }

      console.log('[BackgroundFetch] Sync completed:', syncResult);
      BackgroundFetch.finish(taskId);
    } catch (error) {
      console.error('[BackgroundFetch] Task failed:', error);
      BackgroundFetch.finish(taskId);
    }
  }

  private async onBackgroundFetchTimeout(taskId: string): Promise<void> {
    console.warn('[BackgroundFetch] Task timeout:', taskId);
    BackgroundFetch.finish(taskId);
  }

  private async getNotificationPreferences(): Promise<{
    workoutReminders: boolean;
    progressUpdates: boolean;
    syncComplete: boolean;
  }> {
    try {
      const prefsJson = await AsyncStorage.getItem('@ascend:preferences');
      if (prefsJson) {
        const prefs = JSON.parse(prefsJson);
        return prefs.notifications || {
          workoutReminders: true,
          progressUpdates: true,
          syncComplete: true,
        };
      }
    } catch (error) {
      console.error('Failed to get notification preferences:', error);
    }

    return {
      workoutReminders: true,
      progressUpdates: true,
      syncComplete: true,
    };
  }

  async getLastSyncTime(): Promise<Date | null> {
    try {
      const lastSync = await AsyncStorage.getItem(LAST_SYNC_KEY);
      return lastSync ? new Date(lastSync) : null;
    } catch (error) {
      console.error('Failed to get last sync time:', error);
      return null;
    }
  }

  async forceSync(): Promise<void> {
    console.log('[BackgroundTaskService] Force sync triggered');
    await this.onBackgroundFetch('manual-sync');
  }

  async stop(): Promise<void> {
    try {
      await BackgroundFetch.stop();
      console.log('[BackgroundTaskService] Stopped');
    } catch (error) {
      console.error('[BackgroundTaskService] Stop failed:', error);
    }
  }

  getStatus(): Promise<number> {
    return BackgroundFetch.status();
  }
}

export const backgroundTaskService = new BackgroundTaskService();

// Headless task handler for Android (runs even when app is not running)
if (Platform.OS === 'android') {
  BackgroundFetch.registerHeadlessTask(async (event) => {
    const { taskId, timeout } = event;
    console.log('[BackgroundFetch] Headless task:', taskId);

    if (timeout) {
      console.warn('[BackgroundFetch] Headless task timeout');
      BackgroundFetch.finish(taskId);
      return;
    }

    try {
      await syncEngine.sync();
      console.log('[BackgroundFetch] Headless sync completed');
    } catch (error) {
      console.error('[BackgroundFetch] Headless task failed:', error);
    } finally {
      BackgroundFetch.finish(taskId);
    }
  });
}
