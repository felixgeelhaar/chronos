import { Model } from '@nozbe/watermelondb';
import { field, readonly, date } from '@nozbe/watermelondb/decorators';

export class User extends Model {
  static table = 'users';

  @field('server_id') serverId!: string;
  @field('email') email!: string;
  @field('name') name!: string;
  @field('body_weight') bodyWeight?: number;

  @readonly @date('created_at') createdAt!: Date;
  @readonly @date('updated_at') updatedAt!: Date;
  @date('synced_at') syncedAt?: Date;
}
