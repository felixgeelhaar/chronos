import { Model } from '@nozbe/watermelondb';
import { field, readonly, date } from '@nozbe/watermelondb/decorators';

export class Video extends Model {
  static table = 'videos';

  @field('server_id') serverId?: string;
  @field('user_id') userId!: string;
  @field('set_id') setId?: string;
  @field('filename') filename!: string;
  @field('file_path') filePath!: string;
  @field('file_size') fileSize!: number;
  @field('duration') duration?: number;
  @field('status') status!: string;
  @field('upload_progress') uploadProgress!: number;
  @field('is_synced') isSynced!: boolean;

  @readonly @date('created_at') createdAt!: Date;
  @readonly @date('updated_at') updatedAt!: Date;
  @date('synced_at') syncedAt?: Date;
}
