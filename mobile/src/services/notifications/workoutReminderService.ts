import AsyncStorage from '@react-native-async-storage/async-storage';
import { notificationService, LocalNotificationSchedule } from './notificationService';
import database from '../../database';
import { Session } from '../../database/models';
import { Q } from '@nozbe/watermelondb';
import { addDays, addHours, startOfDay, format, differenceInHours } from 'date-fns';

const REMINDER_SETTINGS_KEY = '@ascend:workout_reminders';
const REMINDER_IDS = {
  DAILY: 'workout-reminder-daily',
  REST_DAY: 'workout-reminder-rest-day',
  STREAK: 'workout-reminder-streak',
};

export interface WorkoutReminderSettings {
  enabled: boolean;
  time: string; // HH:mm format
  days: number[]; // 0 = Sunday, 1 = Monday, etc.
  restDayReminder: boolean;
  streakReminder: boolean;
}

const DEFAULT_SETTINGS: WorkoutReminderSettings = {
  enabled: true,
  time: '18:00', // 6 PM
  days: [1, 2, 3, 4, 5], // Monday-Friday
  restDayReminder: true,
  streakReminder: true,
};

class WorkoutReminderService {
  async getSettings(): Promise<WorkoutReminderSettings> {
    try {
      const settingsJson = await AsyncStorage.getItem(REMINDER_SETTINGS_KEY);
      if (settingsJson) {
        return JSON.parse(settingsJson);
      }
    } catch (error) {
      console.error('Failed to get reminder settings:', error);
    }
    return DEFAULT_SETTINGS;
  }

  async updateSettings(settings: Partial<WorkoutReminderSettings>): Promise<void> {
    try {
      const currentSettings = await this.getSettings();
      const newSettings = { ...currentSettings, ...settings };
      await AsyncStorage.setItem(REMINDER_SETTINGS_KEY, JSON.stringify(newSettings));

      // Reschedule reminders with new settings
      await this.scheduleReminders();
    } catch (error) {
      console.error('Failed to update reminder settings:', error);
      throw error;
    }
  }

  async scheduleReminders(): Promise<void> {
    const settings = await this.getSettings();

    // Cancel all existing reminders
    await this.cancelAllReminders();

    if (!settings.enabled) {
      console.log('Workout reminders disabled');
      return;
    }

    // Schedule daily reminders for selected days
    await this.scheduleDailyReminders(settings);

    // Schedule rest day reminder if enabled
    if (settings.restDayReminder) {
      await this.scheduleRestDayReminder(settings);
    }

    // Schedule streak reminder if enabled
    if (settings.streakReminder) {
      await this.scheduleStreakReminder();
    }

    console.log('Workout reminders scheduled');
  }

  private async scheduleDailyReminders(settings: WorkoutReminderSettings): Promise<void> {
    const [hours, minutes] = settings.time.split(':').map(Number);
    const now = new Date();

    // Schedule for each enabled day of the week
    for (const day of settings.days) {
      const nextOccurrence = this.getNextOccurrence(day, hours, minutes);
      const dayName = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'][day];

      await notificationService.scheduleLocalNotification({
        id: `${REMINDER_IDS.DAILY}-${day}`,
        title: '💪 Time to Train!',
        body: `Don't forget your workout for ${dayName}`,
        trigger: {
          type: 'timestamp',
          timestamp: nextOccurrence.getTime(),
        },
        data: {
          type: 'session_reminder',
          day,
        },
      });
    }
  }

  private async scheduleRestDayReminder(settings: WorkoutReminderSettings): Promise<void> {
    try {
      // Check if user has worked out in the last 48 hours
      const twoDaysAgo = addDays(new Date(), -2);
      const sessions = await database
        .get<Session>('sessions')
        .query(Q.where('date', Q.gte(twoDaysAgo.toISOString())))
        .fetch();

      if (sessions.length === 0) {
        // User hasn't worked out recently, send rest day reminder in 24 hours
        const tomorrow = addHours(new Date(), 24);

        await notificationService.scheduleLocalNotification({
          id: REMINDER_IDS.REST_DAY,
          title: '🧘 Rest Day Check-in',
          body: "It's been a while since your last workout. Ready to get back to it?",
          trigger: {
            type: 'timestamp',
            timestamp: tomorrow.getTime(),
          },
          data: {
            type: 'session_reminder',
            restDay: true,
          },
        });
      }
    } catch (error) {
      console.error('Failed to schedule rest day reminder:', error);
    }
  }

  private async scheduleStreakReminder(): Promise<void> {
    try {
      // Calculate current workout streak
      const streak = await this.calculateStreak();

      if (streak >= 3) {
        // User has a good streak, remind them to maintain it
        const tomorrow = addHours(new Date(), 24);

        await notificationService.scheduleLocalNotification({
          id: REMINDER_IDS.STREAK,
          title: `🔥 ${streak}-Day Streak!`,
          body: "You're on fire! Keep the momentum going.",
          trigger: {
            type: 'timestamp',
            timestamp: tomorrow.getTime(),
          },
          data: {
            type: 'progress_update',
            streak,
          },
        });
      }
    } catch (error) {
      console.error('Failed to schedule streak reminder:', error);
    }
  }

  private getNextOccurrence(dayOfWeek: number, hours: number, minutes: number): Date {
    const now = new Date();
    const currentDay = now.getDay();

    // Calculate days until next occurrence
    let daysUntil = dayOfWeek - currentDay;
    if (daysUntil < 0) {
      daysUntil += 7;
    } else if (daysUntil === 0) {
      // If today, check if time has passed
      const targetTime = new Date(now);
      targetTime.setHours(hours, minutes, 0, 0);

      if (targetTime <= now) {
        // Time has passed, schedule for next week
        daysUntil = 7;
      }
    }

    const nextDate = addDays(startOfDay(now), daysUntil);
    nextDate.setHours(hours, minutes, 0, 0);

    return nextDate;
  }

  private async calculateStreak(): Promise<number> {
    try {
      const sessions = await database
        .get<Session>('sessions')
        .query(Q.sortBy('date', Q.desc))
        .fetch();

      let streak = 0;
      let currentDate = startOfDay(new Date());

      for (const session of sessions) {
        const sessionDate = startOfDay(new Date(session.date));
        const daysDiff = differenceInHours(currentDate, sessionDate) / 24;

        if (daysDiff <= 1) {
          streak++;
          currentDate = sessionDate;
        } else if (daysDiff > 2) {
          // More than 2 days gap, streak is broken
          break;
        }
      }

      return streak;
    } catch (error) {
      console.error('Failed to calculate streak:', error);
      return 0;
    }
  }

  async cancelAllReminders(): Promise<void> {
    try {
      await notificationService.cancelScheduledNotification(REMINDER_IDS.DAILY);
      await notificationService.cancelScheduledNotification(REMINDER_IDS.REST_DAY);
      await notificationService.cancelScheduledNotification(REMINDER_IDS.STREAK);

      // Cancel day-specific reminders
      for (let day = 0; day < 7; day++) {
        await notificationService.cancelScheduledNotification(`${REMINDER_IDS.DAILY}-${day}`);
      }

      console.log('All workout reminders cancelled');
    } catch (error) {
      console.error('Failed to cancel reminders:', error);
    }
  }

  async sendTestReminder(): Promise<void> {
    await notificationService.displayNotification({
      title: '💪 Test Reminder',
      body: 'This is a test workout reminder',
      data: { type: 'session_reminder', test: true },
    });
  }
}

export const workoutReminderService = new WorkoutReminderService();
