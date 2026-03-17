import { configureStore } from '@reduxjs/toolkit';

// Placeholder reducers - will be replaced with feature slices
const rootReducer = {
  // auth: authSlice,
  // sessions: sessionsSlice,
  // analytics: analyticsSlice,
  // videos: videosSlice,
  // sync: syncSlice,
};

export const store = configureStore({
  reducer: rootReducer,
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        // Ignore these action types
        ignoredActions: ['persist/PERSIST', 'persist/REHYDRATE'],
      },
    }),
});

// Infer the `RootState` and `AppDispatch` types from the store itself
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
