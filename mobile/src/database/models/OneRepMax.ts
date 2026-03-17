import { Model } from '@nozbe/watermelondb';
import { field, readonly, date } from '@nozbe/watermelondb/decorators';

export class OneRepMax extends Model {
  static table = 'one_rep_maxes';

  @field('server_id') serverId?: string;
  @field('user_id') userId!: string;
  @field('exercise_name') exerciseName!: string;
  @field('one_rep_max') oneRepMax!: number;
  @field('date') date!: string;
  @field('is_synced') isSynced!: boolean;

  @readonly @date('created_at') createdAt!: Date;
  @readonly @date('updated_at') updatedAt!: Date;
  @date('synced_at') syncedAt?: Date;
}
