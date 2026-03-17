import { Database } from '@nozbe/watermelondb';
import SQLiteAdapter from '@nozbe/watermelondb/adapters/sqlite';
import { schema } from './schema';
import { User, Session, Set, OneRepMax, Video, SyncQueue } from './models';

// Configure SQLite adapter
const adapter = new SQLiteAdapter({
  schema,
  // Migrations will go here
  migrations: [],
  // JSI enables better performance on newer React Native versions
  jsi: true,
  // Optional: enable query logging in development
  onSetUpError: (error) => {
    console.error('Database setup error:', error);
  },
});

// Create database instance
export const database = new Database({
  adapter,
  modelClasses: [User, Session, Set, OneRepMax, Video, SyncQueue],
});

// Export models for convenience
export { User, Session, Set, OneRepMax, Video, SyncQueue };
