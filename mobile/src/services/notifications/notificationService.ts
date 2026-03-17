import messaging, { FirebaseMessagingTypes } from '@react-native-firebase/messaging';
import notifee, { AndroidImportance, EventType } from '@notifee/react-native';
import { Platform, PermissionsAndroid, Alert } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { apiClient } from '../api/client';

const FCM_TOKEN_KEY = '@ascend:fcm_token';
const NOTIFICATION_CHANNEL_ID = 'ascend-notifications';

export interface NotificationPayload {
  title: string;
  body: string;
  data?: Record<string, any>;
}

export interface LocalNotificationSchedule {
  id: string;
  title: string;
  body: string;
  trigger: {
    type: 'timestamp' | 'interval';
    timestamp?: number;
    timeUnit?: 'seconds' | 'minutes' | 'hours' | 'days';
    value?: number;
    repeats?: boolean;
  };
  data?: Record<string, any>;
}

class NotificationService {
  private initialized = false;

  async initialize(): Promise<void> {
    if (this.initialized) return;

    try {
      // Create notification channel for Android
      await this.createNotificationChannel();

      // Request permissions
      const hasPermission = await this.requestPermission();
      if (!hasPermission) {
        console.warn('Notification permission denied');
        return;
      }

      // Get FCM token and register with backend
      await this.registerFCMToken();

      // Set up foreground message handler
      this.setupForegroundHandler();

      // Set up background message handler
      messaging().setBackgroundMessageHandler(this.handleBackgroundMessage);

      // Set up notification tap handler
      this.setupNotificationTapHandler();

      this.initialized = true;
      console.log('Notification service initialized');
    } catch (error) {
      console.error('Failed to initialize notification service:', error);
    }
  }

  private async createNotificationChannel(): Promise<void> {
    if (Platform.OS === 'android') {
      await notifee.createChannel({
        id: NOTIFICATION_CHANNEL_ID,
        name: 'Ascend Notifications',
        importance: AndroidImportance.HIGH,
        sound: 'default',
        vibration: true,
      });
    }
  }

  private async requestPermission(): Promise<boolean> {
    try {
      if (Platform.OS === 'ios') {
        const authStatus = await messaging().requestPermission();
        return (
          authStatus === messaging.AuthorizationStatus.AUTHORIZED ||
          authStatus === messaging.AuthorizationStatus.PROVISIONAL
        );
      } else {
        // Android 13+ requires notification permission
        if (Platform.Version >= 33) {
          const granted = await PermissionsAndroid.request(
            PermissionsAndroid.PERMISSIONS.POST_NOTIFICATIONS
          );
          return granted === PermissionsAndroid.RESULTS.GRANTED;
        }
        return true; // Earlier Android versions don't require runtime permission
      }
    } catch (error) {
      console.error('Permission request failed:', error);
      return false;
    }
  }

  private async registerFCMToken(): Promise<void> {
    try {
      const token = await messaging().getToken();
      const storedToken = await AsyncStorage.getItem(FCM_TOKEN_KEY);

      if (token !== storedToken) {
        // Register token with backend
        await apiClient.post('/v1/notifications/register', { token });
        await AsyncStorage.setItem(FCM_TOKEN_KEY, token);
        console.log('FCM token registered:', token);
      }

      // Listen for token refresh
      messaging().onTokenRefresh(async (newToken) => {
        await apiClient.post('/v1/notifications/register', { token: newToken });
        await AsyncStorage.setItem(FCM_TOKEN_KEY, newToken);
        console.log('FCM token refreshed:', newToken);
      });
    } catch (error) {
      console.error('Failed to register FCM token:', error);
    }
  }

  private setupForegroundHandler(): void {
    messaging().onMessage(async (remoteMessage) => {
      console.log('Foreground notification received:', remoteMessage);
      await this.displayNotification({
        title: remoteMessage.notification?.title || 'Ascend',
        body: remoteMessage.notification?.body || '',
        data: remoteMessage.data,
      });
    });
  }

  private async handleBackgroundMessage(
    remoteMessage: FirebaseMessagingTypes.RemoteMessage
  ): Promise<void> {
    console.log('Background notification received:', remoteMessage);
    // Background handler - notifications are automatically displayed by the system
  }

  private setupNotificationTapHandler(): void {
    // Handle notification tap when app is in background
    notifee.onBackgroundEvent(async ({ type, detail }) => {
      if (type === EventType.PRESS) {
        console.log('Notification tapped (background):', detail.notification);
        await this.handleNotificationTap(detail.notification?.data);
      }
    });

    // Handle notification tap when app is in foreground
    notifee.onForegroundEvent(({ type, detail }) => {
      if (type === EventType.PRESS) {
        console.log('Notification tapped (foreground):', detail.notification);
        this.handleNotificationTap(detail.notification?.data);
      }
    });

    // Handle initial notification (app opened from quit state)
    messaging()
      .getInitialNotification()
      .then((remoteMessage) => {
        if (remoteMessage) {
          console.log('Notification opened app:', remoteMessage);
          this.handleNotificationTap(remoteMessage.data);
        }
      });
  }

  private async handleNotificationTap(data?: Record<string, any>): Promise<void> {
    if (!data) return;

    // Handle different notification types
    const { type, sessionId, videoId } = data;

    switch (type) {
      case 'session_reminder':
        // Navigate to create session screen
        console.log('Navigate to create session');
        break;
      case 'sync_complete':
        // Navigate to home screen
        console.log('Navigate to home');
        break;
      case 'analysis_complete':
        // Navigate to video analysis
        if (videoId) {
          console.log('Navigate to video analysis:', videoId);
        }
        break;
      case 'progress_update':
        // Navigate to analytics
        console.log('Navigate to analytics');
        break;
      default:
        console.log('Unknown notification type:', type);
    }
  }

  async displayNotification(payload: NotificationPayload): Promise<void> {
    try {
      await notifee.displayNotification({
        title: payload.title,
        body: payload.body,
        android: {
          channelId: NOTIFICATION_CHANNEL_ID,
          smallIcon: 'ic_notification',
          pressAction: {
            id: 'default',
          },
        },
        ios: {
          sound: 'default',
        },
        data: payload.data,
      });
    } catch (error) {
      console.error('Failed to display notification:', error);
    }
  }

  async scheduleLocalNotification(schedule: LocalNotificationSchedule): Promise<void> {
    try {
      const trigger: any =
        schedule.trigger.type === 'timestamp'
          ? { type: 'timestamp', timestamp: schedule.trigger.timestamp }
          : {
              type: 'interval',
              interval: {
                timeUnit: schedule.trigger.timeUnit || 'hours',
                value: schedule.trigger.value || 1,
              },
              repeats: schedule.trigger.repeats || false,
            };

      await notifee.createTriggerNotification(
        {
          id: schedule.id,
          title: schedule.title,
          body: schedule.body,
          android: {
            channelId: NOTIFICATION_CHANNEL_ID,
            smallIcon: 'ic_notification',
            pressAction: {
              id: 'default',
            },
          },
          ios: {
            sound: 'default',
          },
          data: schedule.data,
        },
        trigger
      );

      console.log('Local notification scheduled:', schedule.id);
    } catch (error) {
      console.error('Failed to schedule notification:', error);
    }
  }

  async cancelScheduledNotification(id: string): Promise<void> {
    try {
      await notifee.cancelNotification(id);
      console.log('Notification cancelled:', id);
    } catch (error) {
      console.error('Failed to cancel notification:', error);
    }
  }

  async cancelAllScheduledNotifications(): Promise<void> {
    try {
      await notifee.cancelAllNotifications();
      console.log('All notifications cancelled');
    } catch (error) {
      console.error('Failed to cancel all notifications:', error);
    }
  }

  async getBadgeCount(): Promise<number> {
    if (Platform.OS === 'ios') {
      return await notifee.getBadgeCount();
    }
    return 0;
  }

  async setBadgeCount(count: number): Promise<void> {
    if (Platform.OS === 'ios') {
      await notifee.setBadgeCount(count);
    }
  }

  async clearBadge(): Promise<void> {
    await this.setBadgeCount(0);
  }
}

export const notificationService = new NotificationService();
