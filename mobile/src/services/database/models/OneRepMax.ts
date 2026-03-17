import { Model } from '@nozbe/watermelondb';
import { field, date, readonly } from '@nozbe/watermelondb/decorators';

export class OneRepMax extends Model {
  static table = 'one_rep_maxes';

  @field('user_id') userId!: string;
  @field('exercise_name') exerciseName!: string;
  @field('weight') weight!: number;
  @field('date') date!: number;
  @readonly @date('created_at') createdAt!: Date;
  @readonly @date('updated_at') updatedAt!: Date;
  @date('synced_at') syncedAt?: Date;
}
