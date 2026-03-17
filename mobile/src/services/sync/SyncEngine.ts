import { Q } from '@nozbe/watermelondb';
import NetInfo from '@react-native-community/netinfo';
import database from '../../database';
import { Session, Set, SyncQueue } from '../../database/models';
import { sessionService, apiClient } from '../api';

export interface SyncStatus {
  isSyncing: boolean;
  lastSyncedAt?: Date;
  pendingChanges: number;
  errors: string[];
}

class SyncEngine {
  private isSyncing = false;
  private syncInterval?: NodeJS.Timeout;
  private listeners: Array<(status: SyncStatus) => void> = [];

  /**
   * Initialize sync engine with auto-sync
   */
  async initialize(autoSyncIntervalMs: number = 300000) {
    // Listen for network changes
    NetInfo.addEventListener(state => {
      if (state.isConnected && !this.isSyncing) {
        this.sync();
      }
    });

    // Set up periodic sync
    if (autoSyncIntervalMs > 0) {
      this.syncInterval = setInterval(() => {
        this.sync();
      }, autoSyncIntervalMs);
    }

    // Initial sync
    const netState = await NetInfo.fetch();
    if (netState.isConnected) {
      await this.sync();
    }
  }

  /**
   * Stop sync engine
   */
  stop() {
    if (this.syncInterval) {
      clearInterval(this.syncInterval);
      this.syncInterval = undefined;
    }
  }

  /**
   * Subscribe to sync status changes
   */
  subscribe(listener: (status: SyncStatus) => void) {
    this.listeners.push(listener);
    return () => {
      this.listeners = this.listeners.filter(l => l !== listener);
    };
  }

  /**
   * Get current sync status
   */
  async getStatus(): Promise<SyncStatus> {
    const pendingQueue = await database.get<SyncQueue>('sync_queue').query().fetchCount();
    const unsyncedSessions = await database
      .get<Session>('sessions')
      .query(Q.where('is_synced', false))
      .fetchCount();

    return {
      isSyncing: this.isSyncing,
      pendingChanges: pendingQueue + unsyncedSessions,
      errors: [],
    };
  }

  /**
   * Notify listeners of status change
   */
  private notifyListeners(status: SyncStatus) {
    this.listeners.forEach(listener => listener(status));
  }

  /**
   * Main sync method
   */
  async sync(): Promise<void> {
    // Check if already syncing
    if (this.isSyncing) {
      console.log('[Sync] Already syncing, skipping');
      return;
    }

    // Check network connectivity
    const netState = await NetInfo.fetch();
    if (!netState.isConnected) {
      console.log('[Sync] No network connection, skipping sync');
      return;
    }

    // Check authentication
    const isAuthenticated = await apiClient.isAuthenticated();
    if (!isAuthenticated) {
      console.log('[Sync] Not authenticated, skipping sync');
      return;
    }

    try {
      this.isSyncing = true;
      this.notifyListeners(await this.getStatus());

      console.log('[Sync] Starting sync...');

      // Step 1: Push local changes to server
      await this.pushLocalChanges();

      // Step 2: Pull remote changes from server
      await this.pullRemoteChanges();

      // Step 3: Process sync queue
      await this.processSyncQueue();

      console.log('[Sync] Sync completed successfully');
    } catch (error) {
      console.error('[Sync] Sync failed:', error);
      const status = await this.getStatus();
      status.errors.push(error instanceof Error ? error.message : 'Unknown error');
      this.notifyListeners(status);
    } finally {
      this.isSyncing = false;
      this.notifyListeners(await this.getStatus());
    }
  }

  /**
   * Push unsynced local changes to server
   */
  private async pushLocalChanges(): Promise<void> {
    const unsyncedSessions = await database
      .get<Session>('sessions')
      .query(Q.where('is_synced', false))
      .fetch();

    console.log(`[Sync] Pushing ${unsyncedSessions.length} sessions...`);

    for (const session of unsyncedSessions) {
      try {
        const sets = await session.sets.fetch();

        // If session has no server ID, it's a new session
        if (!session.serverId) {
          // Create new session on server
          const response = await sessionService.createSession({
            date: session.date,
            notes: session.notes,
            sets: sets.map(set => ({
              exercise_name: set.exerciseName,
              weight: set.weight,
              reps: set.reps,
              rpe: set.rpe,
              set_order: set.setOrder,
            })),
          });

          // Update local session with server ID
          await database.write(async () => {
            await session.update(s => {
              s.serverId = response.id;
              s.isSynced = true;
              s.syncedAt = new Date();
            });

            // Update sets with server IDs
            for (let i = 0; i < sets.length; i++) {
              const set = sets[i];
              const serverSet = response.sets[i];
              await set.update(s => {
                s.serverId = serverSet.id;
                s.isSynced = true;
                s.syncedAt = new Date();
              });
            }
          });

          console.log(`[Sync] Created session ${response.id}`);
        } else {
          // Update existing session on server
          await sessionService.updateSession(session.serverId, {
            date: session.date,
            notes: session.notes,
          });

          await database.write(async () => {
            await session.update(s => {
              s.isSynced = true;
              s.syncedAt = new Date();
            });
          });

          console.log(`[Sync] Updated session ${session.serverId}`);
        }
      } catch (error) {
        console.error(`[Sync] Failed to sync session ${session.id}:`, error);
        // Add to sync queue for retry
        await this.addToSyncQueue('session', session.id, 'update', {
          sessionId: session.id,
          serverId: session.serverId,
        });
      }
    }
  }

  /**
   * Pull remote changes from server
   */
  private async pullRemoteChanges(): Promise<void> {
    // For now, we'll just sync recent sessions
    // In a production app, you'd want to implement incremental sync with timestamps
    console.log('[Sync] Pulling remote changes...');

    try {
      const response = await sessionService.listSessions(1, 50);

      for (const remoteSession of response.sessions) {
        // Check if session exists locally
        const localSessions = await database
          .get<Session>('sessions')
          .query(Q.where('server_id', remoteSession.id))
          .fetch();

        if (localSessions.length === 0) {
          // Session doesn't exist locally, create it
          await database.write(async () => {
            const session = await database.get<Session>('sessions').create(s => {
              s.serverId = remoteSession.id;
              s.userId = remoteSession.user_id;
              s.date = remoteSession.date;
              s.notes = remoteSession.notes;
              s.totalVolume = remoteSession.total_volume;
              s.totalSets = remoteSession.total_sets;
              s.isSynced = true;
              s.syncedAt = new Date();
            });

            // Create sets
            for (const remoteSet of remoteSession.sets) {
              await database.get<Set>('sets').create(s => {
                s.serverId = remoteSet.id;
                s.sessionId = session.id;
                s.exerciseName = remoteSet.exercise_name;
                s.weight = remoteSet.weight;
                s.reps = remoteSet.reps;
                s.rpe = remoteSet.rpe;
                s.setOrder = remoteSet.set_order;
                s.volume = remoteSet.volume;
                s.estimatedOneRepMax = remoteSet.estimated_one_rep_max;
                s.videoId = remoteSet.video_id;
                s.isSynced = true;
                s.syncedAt = new Date();
              });
            }
          });

          console.log(`[Sync] Pulled session ${remoteSession.id}`);
        } else {
          // Session exists, check if it needs updating
          const localSession = localSessions[0];
          const localUpdated = localSession.updatedAt.getTime();
          const remoteUpdated = new Date(remoteSession.updated_at).getTime();

          if (remoteUpdated > localUpdated) {
            // Remote is newer, update local
            await database.write(async () => {
              await localSession.update(s => {
                s.date = remoteSession.date;
                s.notes = remoteSession.notes;
                s.totalVolume = remoteSession.total_volume;
                s.totalSets = remoteSession.total_sets;
                s.isSynced = true;
                s.syncedAt = new Date();
              });
            });

            console.log(`[Sync] Updated session ${remoteSession.id}`);
          }
        }
      }
    } catch (error) {
      console.error('[Sync] Failed to pull remote changes:', error);
      throw error;
    }
  }

  /**
   * Process items in sync queue
   */
  private async processSyncQueue(): Promise<void> {
    const queueItems = await database
      .get<SyncQueue>('sync_queue')
      .query(Q.sortBy('created_at', Q.asc))
      .fetch();

    console.log(`[Sync] Processing ${queueItems.length} queue items...`);

    for (const item of queueItems) {
      try {
        const payload = JSON.parse(item.payload);

        // Process based on entity type and operation
        // This is a simplified example - in production you'd have more complex logic
        console.log(`[Sync] Processing queue item ${item.id}`);

        // Remove from queue on success
        await database.write(async () => {
          await item.destroyPermanently();
        });
      } catch (error) {
        console.error(`[Sync] Failed to process queue item ${item.id}:`, error);

        // Update retry count
        await database.write(async () => {
          await item.update(i => {
            i.retryCount += 1;
            i.lastError = error instanceof Error ? error.message : 'Unknown error';
          });
        });

        // If retry count exceeds threshold, log error but continue
        if (item.retryCount >= 5) {
          console.error(`[Sync] Queue item ${item.id} exceeded retry limit`);
        }
      }
    }
  }

  /**
   * Add item to sync queue for retry
   */
  private async addToSyncQueue(
    entityType: string,
    entityId: string,
    operation: 'create' | 'update' | 'delete',
    payload: any
  ): Promise<void> {
    await database.write(async () => {
      await database.get<SyncQueue>('sync_queue').create(item => {
        item.entityType = entityType;
        item.entityId = entityId;
        item.operation = operation;
        item.payload = JSON.stringify(payload);
        item.retryCount = 0;
      });
    });
  }
}

export const syncEngine = new SyncEngine();
export default syncEngine;
