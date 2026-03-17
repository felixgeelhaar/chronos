import { Model } from '@nozbe/watermelondb';
import { field, date, readonly } from '@nozbe/watermelondb/decorators';

export class SyncQueue extends Model {
  static table = 'sync_queue';

  @field('operation') operation!: string; // 'create', 'update', 'delete'
  @field('table_name') tableName!: string;
  @field('record_id') recordId!: string;
  @field('payload') payload!: string; // JSON payload
  @field('status') status!: string; // 'pending', 'syncing', 'failed'
  @field('retry_count') retryCount!: number;
  @field('error_message') errorMessage?: string;
  @readonly @date('created_at') createdAt!: Date;
  @readonly @date('updated_at') updatedAt!: Date;
}
