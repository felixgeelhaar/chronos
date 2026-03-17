import { Model, Q } from '@nozbe/watermelondb';
import { field, readonly, date, children } from '@nozbe/watermelondb/decorators';
import { Set } from './Set';

export class Session extends Model {
  static table = 'sessions';

  static associations = {
    sets: { type: 'has_many' as const, foreignKey: 'session_id' },
  };

  @field('server_id') serverId?: string;
  @field('user_id') userId!: string;
  @field('date') date!: string;
  @field('notes') notes?: string;
  @field('total_volume') totalVolume!: number;
  @field('total_sets') totalSets!: number;
  @field('is_synced') isSynced!: boolean;

  @readonly @date('created_at') createdAt!: Date;
  @readonly @date('updated_at') updatedAt!: Date;
  @date('synced_at') syncedAt?: Date;

  @children('sets') sets!: Q.Query<Set>;
}
