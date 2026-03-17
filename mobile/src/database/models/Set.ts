import { Model } from '@nozbe/watermelondb';
import { field, readonly, date, relation } from '@nozbe/watermelondb/decorators';
import { Session } from './Session';

export class Set extends Model {
  static table = 'sets';

  static associations = {
    sessions: { type: 'belongs_to' as const, key: 'session_id' },
  };

  @field('server_id') serverId?: string;
  @field('session_id') sessionId!: string;
  @field('exercise_name') exerciseName!: string;
  @field('weight') weight!: number;
  @field('reps') reps!: number;
  @field('rpe') rpe?: number;
  @field('set_order') setOrder!: number;
  @field('volume') volume!: number;
  @field('estimated_one_rep_max') estimatedOneRepMax!: number;
  @field('video_id') videoId?: string;
  @field('is_synced') isSynced!: boolean;

  @readonly @date('created_at') createdAt!: Date;
  @readonly @date('updated_at') updatedAt!: Date;
  @date('synced_at') syncedAt?: Date;

  @relation('sessions', 'session_id') session!: Session;
}
