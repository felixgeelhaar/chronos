import { Q } from '@nozbe/watermelondb';
import database from './index';
import { Session } from './models/Session';
import { Set } from './models/Set';
import { OneRepMax } from './models/OneRepMax';

/**
 * Optimized database query helpers
 * Provides efficient, memoizable queries with proper indexing
 */

export const queryHelpers = {
  /**
   * Get recent sessions with limit
   * Optimized with indexed date column and fetch limit
   */
  async getRecentSessions(limit: number = 20): Promise<Session[]> {
    return await database
      .get<Session>('sessions')
      .query(Q.sortBy('date', Q.desc), Q.take(limit))
      .fetch();
  },

  /**
   * Get sessions for a date range
   * Uses indexed date column for efficient filtering
   */
  async getSessionsInRange(startDate: string, endDate: string): Promise<Session[]> {
    return await database
      .get<Session>('sessions')
      .query(
        Q.where('date', Q.gte(startDate)),
        Q.where('date', Q.lte(endDate)),
        Q.sortBy('date', Q.desc)
      )
      .fetch();
  },

  /**
   * Get sets for an exercise across all sessions
   * Uses indexed exercise_name for efficient lookup
   */
  async getSetsForExercise(exerciseName: string, limit?: number): Promise<Set[]> {
    const query = database
      .get<Set>('sets')
      .query(
        Q.where('exercise_name', exerciseName),
        Q.sortBy('created_at', Q.desc)
      );

    if (limit) {
      query.extend(Q.take(limit));
    }

    return await query.fetch();
  },

  /**
   * Get unique exercise names with counts
   * Efficient aggregation query
   */
  async getExerciseList(): Promise<Array<{ name: string; count: number }>> {
    const sets = await database
      .get<Set>('sets')
      .query(Q.sortBy('exercise_name'))
      .fetch();

    // Group by exercise name and count
    const exerciseMap = new Map<string, number>();
    sets.forEach((set) => {
      const count = exerciseMap.get(set.exerciseName) || 0;
      exerciseMap.set(set.exerciseName, count + 1);
    });

    return Array.from(exerciseMap.entries())
      .map(([name, count]) => ({ name, count }))
      .sort((a, b) => b.count - a.count);
  },

  /**
   * Get 1RM history for an exercise
   * Uses compound index on user_id and exercise_name
   */
  async get1RMHistory(
    userId: string,
    exerciseName: string,
    limit?: number
  ): Promise<OneRepMax[]> {
    const query = database
      .get<OneRepMax>('one_rep_maxes')
      .query(
        Q.where('user_id', userId),
        Q.where('exercise_name', exerciseName),
        Q.sortBy('date', Q.desc)
      );

    if (limit) {
      query.extend(Q.take(limit));
    }

    return await query.fetch();
  },

  /**
   * Get unsynced records for sync queue processing
   * Uses indexed is_synced column
   */
  async getUnsyncedSessions(): Promise<Session[]> {
    return await database
      .get<Session>('sessions')
      .query(Q.where('is_synced', false))
      .fetch();
  },

  /**
   * Get total volume for a date range
   * Optimized aggregation with indexed date
   */
  async getTotalVolumeInRange(startDate: string, endDate: string): Promise<number> {
    const sessions = await database
      .get<Session>('sessions')
      .query(
        Q.where('date', Q.gte(startDate)),
        Q.where('date', Q.lte(endDate))
      )
      .fetch();

    return sessions.reduce((sum, session) => sum + session.totalVolume, 0);
  },

  /**
   * Get workout frequency for a date range
   * Returns number of sessions per day
   */
  async getWorkoutFrequency(startDate: string, endDate: string): Promise<Map<string, number>> {
    const sessions = await database
      .get<Session>('sessions')
      .query(
        Q.where('date', Q.gte(startDate)),
        Q.where('date', Q.lte(endDate)),
        Q.sortBy('date', Q.asc)
      )
      .fetch();

    const frequencyMap = new Map<string, number>();
    sessions.forEach((session) => {
      const date = session.date.split('T')[0]; // Get date only
      const count = frequencyMap.get(date) || 0;
      frequencyMap.set(date, count + 1);
    });

    return frequencyMap;
  },

  /**
   * Get exercise performance trends
   * Returns max weight, volume trends for an exercise
   */
  async getExerciseTrends(
    exerciseName: string,
    startDate: string,
    endDate: string
  ): Promise<
    Array<{
      date: string;
      maxWeight: number;
      totalVolume: number;
      totalReps: number;
    }>
  > {
    const sets = await database
      .get<Set>('sets')
      .query(
        Q.where('exercise_name', exerciseName),
        Q.where('created_at', Q.gte(new Date(startDate).getTime())),
        Q.where('created_at', Q.lte(new Date(endDate).getTime())),
        Q.sortBy('created_at', Q.asc)
      )
      .fetch();

    // Group by date
    const dateMap = new Map<
      string,
      { maxWeight: number; totalVolume: number; totalReps: number }
    >();

    sets.forEach((set) => {
      const date = new Date(set.createdAt).toISOString().split('T')[0];
      const existing = dateMap.get(date) || {
        maxWeight: 0,
        totalVolume: 0,
        totalReps: 0,
      };

      dateMap.set(date, {
        maxWeight: Math.max(existing.maxWeight, set.weight),
        totalVolume: existing.totalVolume + set.volume,
        totalReps: existing.totalReps + set.reps,
      });
    });

    return Array.from(dateMap.entries()).map(([date, stats]) => ({
      date,
      ...stats,
    }));
  },

  /**
   * Clean up old synced records
   * Useful for maintaining database size
   */
  async cleanupOldSyncedRecords(daysToKeep: number = 90): Promise<void> {
    const cutoffDate = new Date();
    cutoffDate.setDate(cutoffDate.getDate() - daysToKeep);
    const cutoffTime = cutoffDate.getTime();

    await database.write(async () => {
      const oldSyncQueue = await database
        .get('sync_queue')
        .query(
          Q.where('updated_at', Q.lt(cutoffTime)),
          Q.where('retry_count', Q.gt(5))
        )
        .fetch();

      await Promise.all(oldSyncQueue.map((record) => record.markAsDeleted()));
    });
  },

  /**
   * Batch fetch sessions with their sets
   * More efficient than individual queries
   */
  async getSessionsWithSets(sessionIds: string[]): Promise<Session[]> {
    const sessions = await database
      .get<Session>('sessions')
      .query(Q.where('id', Q.oneOf(sessionIds)))
      .fetch();

    // Prefetch sets for all sessions
    await Promise.all(sessions.map((session) => session.sets.fetch()));

    return sessions;
  },
};

export default queryHelpers;
