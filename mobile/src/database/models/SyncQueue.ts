import { Model } from '@nozbe/watermelondb';
import { field, readonly, date } from '@nozbe/watermelondb/decorators';

export class SyncQueue extends Model {
  static table = 'sync_queue';

  @field('entity_type') entityType!: string;
  @field('entity_id') entityId!: string;
  @field('operation') operation!: 'create' | 'update' | 'delete';
  @field('payload') payload!: string;
  @field('retry_count') retryCount!: number;
  @field('last_error') lastError?: string;

  @readonly @date('created_at') createdAt!: Date;
  @readonly @date('updated_at') updatedAt!: Date;
}
