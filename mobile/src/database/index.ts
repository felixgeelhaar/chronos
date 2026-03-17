import { Database } from '@nozbe/watermelondb';
import SQLiteAdapter from '@nozbe/watermelondb/adapters/sqlite';
import { schema } from './schema';
import { User, Session, Set, Video, SyncQueue, OneRepMax } from './models';

// Create SQLite adapter
const adapter = new SQLiteAdapter({
  schema,
  // Optional: Enable JSI for better performance (requires native setup)
  jsi: false,
  // Optional: migrations for future schema changes
  // migrations,
});

// Create database instance
export const database = new Database({
  adapter,
  modelClasses: [User, Session, Set, Video, SyncQueue, OneRepMax],
});

export default database;
