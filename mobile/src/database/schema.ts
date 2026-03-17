import { appSchema, tableSchema } from '@nozbe/watermelondb';

export const schema = appSchema({
  version: 1,
  tables: [
    tableSchema({
      name: 'users',
      columns: [
        { name: 'server_id', type: 'string', isIndexed: true },
        { name: 'email', type: 'string', isIndexed: true },
        { name: 'name', type: 'string' },
        { name: 'body_weight', type: 'number', isOptional: true },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
        { name: 'synced_at', type: 'number', isOptional: true },
      ],
    }),
    tableSchema({
      name: 'sessions',
      columns: [
        { name: 'server_id', type: 'string', isIndexed: true, isOptional: true },
        { name: 'user_id', type: 'string', isIndexed: true },
        { name: 'date', type: 'string', isIndexed: true },
        { name: 'notes', type: 'string', isOptional: true },
        { name: 'total_volume', type: 'number' },
        { name: 'total_sets', type: 'number' },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
        { name: 'synced_at', type: 'number', isOptional: true },
        { name: 'is_synced', type: 'boolean' },
      ],
    }),
    tableSchema({
      name: 'sets',
      columns: [
        { name: 'server_id', type: 'string', isIndexed: true, isOptional: true },
        { name: 'session_id', type: 'string', isIndexed: true },
        { name: 'exercise_name', type: 'string', isIndexed: true },
        { name: 'weight', type: 'number' },
        { name: 'reps', type: 'number' },
        { name: 'rpe', type: 'number', isOptional: true },
        { name: 'set_order', type: 'number' },
        { name: 'volume', type: 'number' },
        { name: 'estimated_one_rep_max', type: 'number' },
        { name: 'video_id', type: 'string', isOptional: true },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
        { name: 'synced_at', type: 'number', isOptional: true },
        { name: 'is_synced', type: 'boolean' },
      ],
    }),
    tableSchema({
      name: 'videos',
      columns: [
        { name: 'server_id', type: 'string', isIndexed: true, isOptional: true },
        { name: 'user_id', type: 'string', isIndexed: true },
        { name: 'set_id', type: 'string', isIndexed: true, isOptional: true },
        { name: 'filename', type: 'string' },
        { name: 'file_path', type: 'string' },
        { name: 'file_size', type: 'number' },
        { name: 'duration', type: 'number', isOptional: true },
        { name: 'status', type: 'string' },
        { name: 'upload_progress', type: 'number' },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
        { name: 'synced_at', type: 'number', isOptional: true },
        { name: 'is_synced', type: 'boolean' },
      ],
    }),
    tableSchema({
      name: 'sync_queue',
      columns: [
        { name: 'entity_type', type: 'string', isIndexed: true },
        { name: 'entity_id', type: 'string', isIndexed: true },
        { name: 'operation', type: 'string' }, // 'create', 'update', 'delete'
        { name: 'payload', type: 'string' }, // JSON string
        { name: 'retry_count', type: 'number' },
        { name: 'last_error', type: 'string', isOptional: true },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
      ],
    }),
    tableSchema({
      name: 'one_rep_maxes',
      columns: [
        { name: 'server_id', type: 'string', isIndexed: true, isOptional: true },
        { name: 'user_id', type: 'string', isIndexed: true },
        { name: 'exercise_name', type: 'string', isIndexed: true },
        { name: 'one_rep_max', type: 'number' },
        { name: 'date', type: 'string', isIndexed: true },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
        { name: 'synced_at', type: 'number', isOptional: true },
        { name: 'is_synced', type: 'boolean' },
      ],
    }),
  ],
});
