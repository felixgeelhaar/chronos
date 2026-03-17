import { Model, Q } from '@nozbe/watermelondb';
import { field, date, readonly, children, lazy } from '@nozbe/watermelondb/decorators';
import type { Set } from './Set';
import type { Video } from './Video';

export class Session extends Model {
  static table = 'sessions';

  static associations = {
    sets: { type: 'has_many' as const, foreignKey: 'session_id' },
    videos: { type: 'has_many' as const, foreignKey: 'session_id' },
  };

  @field('date') date!: number;
  @field('notes') notes?: string;
  @field('user_id') userId!: string;
  @readonly @date('created_at') createdAt!: Date;
  @readonly @date('updated_at') updatedAt!: Date;
  @date('synced_at') syncedAt?: Date;

  @children('sets') sets!: Q.Query<Set>;
  @children('videos') videos!: Q.Query<Video>;

  @lazy totalSets = this.sets.observeCount();
}
