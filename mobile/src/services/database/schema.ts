import { appSchema, tableSchema } from '@nozbe/watermelondb';

export const schema = appSchema({
  version: 1,
  tables: [
    // User profile
    tableSchema({
      name: 'users',
      columns: [
        { name: 'email', type: 'string' },
        { name: 'name', type: 'string' },
        { name: 'body_weight', type: 'number', isOptional: true },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
        { name: 'synced_at', type: 'number', isOptional: true },
      ],
    }),

    // Training sessions
    tableSchema({
      name: 'sessions',
      columns: [
        { name: 'date', type: 'number' },
        { name: 'notes', type: 'string', isOptional: true },
        { name: 'user_id', type: 'string', isIndexed: true },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
        { name: 'synced_at', type: 'number', isOptional: true },
      ],
    }),

    // Exercise sets
    tableSchema({
      name: 'sets',
      columns: [
        { name: 'session_id', type: 'string', isIndexed: true },
        { name: 'exercise_name', type: 'string' },
        { name: 'weight', type: 'number' },
        { name: 'reps', type: 'number' },
        { name: 'rpe', type: 'number', isOptional: true },
        { name: 'notes', type: 'string', isOptional: true },
        { name: 'set_order', type: 'number' },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
        { name: 'synced_at', type: 'number', isOptional: true },
      ],
    }),

    // One-rep max records
    tableSchema({
      name: 'one_rep_maxes',
      columns: [
        { name: 'user_id', type: 'string', isIndexed: true },
        { name: 'exercise_name', type: 'string', isIndexed: true },
        { name: 'weight', type: 'number' },
        { name: 'date', type: 'number' },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
        { name: 'synced_at', type: 'number', isOptional: true },
      ],
    }),

    // Training videos
    tableSchema({
      name: 'videos',
      columns: [
        { name: 'user_id', type: 'string', isIndexed: true },
        { name: 'session_id', type: 'string', isIndexed: true, isOptional: true },
        { name: 'set_id', type: 'string', isIndexed: true, isOptional: true },
        { name: 'url', type: 'string' },
        { name: 'thumbnail_url', type: 'string', isOptional: true },
        { name: 'duration', type: 'number', isOptional: true },
        { name: 'file_size', type: 'number', isOptional: true },
        { name: 'exercise_name', type: 'string', isOptional: true },
        { name: 'date', type: 'number' },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
        { name: 'synced_at', type: 'number', isOptional: true },
      ],
    }),

    // Sync queue for offline operations
    tableSchema({
      name: 'sync_queue',
      columns: [
        { name: 'operation', type: 'string' }, // 'create', 'update', 'delete'
        { name: 'table_name', type: 'string' },
        { name: 'record_id', type: 'string' },
        { name: 'payload', type: 'string' }, // JSON payload
        { name: 'status', type: 'string' }, // 'pending', 'syncing', 'failed'
        { name: 'retry_count', type: 'number' },
        { name: 'error_message', type: 'string', isOptional: true },
        { name: 'created_at', type: 'number' },
        { name: 'updated_at', type: 'number' },
      ],
    }),
  ],
});
