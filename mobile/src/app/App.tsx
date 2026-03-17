import React, { useEffect } from 'react';
import { Provider } from 'react-redux';
import { SafeAreaProvider } from 'react-native-safe-area-context';
import { GestureHandlerRootView } from 'react-native-gesture-handler';
import { store } from './store';
import { AuthProvider } from '../shared/contexts';
import { AppNavigator } from './navigation/AppNavigator';
import { ErrorBoundary } from '../shared/components/ErrorBoundary';
import { ToastProvider } from '../shared/components/ToastProvider';
import { notificationService } from '../services/notifications/notificationService';
import { backgroundTaskService } from '../services/background/backgroundTaskService';
import { workoutReminderService } from '../services/notifications/workoutReminderService';

const App = () => {
  useEffect(() => {
    const initializeServices = async () => {
      try {
        // Initialize notification service
        await notificationService.initialize();
        console.log('Notification service initialized');

        // Initialize background task service
        await backgroundTaskService.initialize();
        console.log('Background task service initialized');

        // Schedule workout reminders
        await workoutReminderService.scheduleReminders();
        console.log('Workout reminders scheduled');
      } catch (error) {
        console.error('Failed to initialize services:', error);
      }
    };

    initializeServices();

    // Cleanup on unmount
    return () => {
      backgroundTaskService.stop();
    };
  }, []);

  return (
    <ErrorBoundary>
      <GestureHandlerRootView style={{ flex: 1 }}>
        <SafeAreaProvider>
          <Provider store={store}>
            <AuthProvider>
              <ToastProvider>
                <AppNavigator />
              </ToastProvider>
            </AuthProvider>
          </Provider>
        </SafeAreaProvider>
      </GestureHandlerRootView>
    </ErrorBoundary>
  );
};

export default App;
