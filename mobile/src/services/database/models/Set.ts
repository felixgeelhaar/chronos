import { Model } from '@nozbe/watermelondb';
import { field, date, readonly, relation } from '@nozbe/watermelondb/decorators';
import type { Session } from './Session';

export class Set extends Model {
  static table = 'sets';

  static associations = {
    sessions: { type: 'belongs_to' as const, key: 'session_id' },
  };

  @field('session_id') sessionId!: string;
  @field('exercise_name') exerciseName!: string;
  @field('weight') weight!: number;
  @field('reps') reps!: number;
  @field('rpe') rpe?: number;
  @field('notes') notes?: string;
  @field('set_order') setOrder!: number;
  @readonly @date('created_at') createdAt!: Date;
  @readonly @date('updated_at') updatedAt!: Date;
  @date('synced_at') syncedAt?: Date;

  @relation('sessions', 'session_id') session!: Session;
}
