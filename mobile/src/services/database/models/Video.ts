import { Model } from '@nozbe/watermelondb';
import { field, date, readonly, relation } from '@nozbe/watermelondb/decorators';
import type { Session } from './Session';

export class Video extends Model {
  static table = 'videos';

  static associations = {
    sessions: { type: 'belongs_to' as const, key: 'session_id' },
  };

  @field('user_id') userId!: string;
  @field('session_id') sessionId?: string;
  @field('set_id') setId?: string;
  @field('url') url!: string;
  @field('thumbnail_url') thumbnailUrl?: string;
  @field('duration') duration?: number;
  @field('file_size') fileSize?: number;
  @field('exercise_name') exerciseName?: string;
  @field('date') date!: number;
  @readonly @date('created_at') createdAt!: Date;
  @readonly @date('updated_at') updatedAt!: Date;
  @date('synced_at') syncedAt?: Date;

  @relation('sessions', 'session_id') session?: Session;
}
