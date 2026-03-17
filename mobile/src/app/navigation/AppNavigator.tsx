import React from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createStackNavigator } from '@react-navigation/stack';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { useAuth } from '../../shared/contexts';
import { LoginScreen, RegisterScreen } from '../../features/auth';
import { HomeScreen } from '../../features/home/HomeScreen';
import {
  SessionListScreen,
  CreateSessionScreen,
  SessionDetailScreen,
} from '../../features/sessions';
import { AnalyticsScreen, ExerciseHistoryScreen, AnalysisHistoryScreen } from '../../features/analytics';
import {
  SettingsScreen,
  EditProfileScreen,
  PreferencesScreen,
  DataExportScreen,
  AboutScreen,
} from '../../features/settings';
import { ActivityIndicator, View, StyleSheet, Text } from 'react-native';

export type RootStackParamList = {
  Login: undefined;
  Register: undefined;
  MainTabs: undefined;
  CreateSession: undefined;
  SessionDetail: { sessionId: string };
};

export type TabParamList = {
  Home: undefined;
  Sessions: undefined;
  Analytics: undefined;
  Settings: undefined;
};

export type AnalyticsStackParamList = {
  AnalyticsMain: undefined;
  ExerciseHistory: undefined;
  AnalysisHistory: undefined;
};

export type SettingsStackParamList = {
  SettingsMain: undefined;
  EditProfile: undefined;
  Preferences: undefined;
  DataExport: undefined;
  About: undefined;
  PrivacyPolicy: undefined;
  TermsOfService: undefined;
  Help: undefined;
  ChangePassword: undefined;
  DeleteAccount: undefined;
};

const Stack = createStackNavigator<RootStackParamList>();
const Tab = createBottomTabNavigator<TabParamList>();
const AnalyticsStack = createStackNavigator<AnalyticsStackParamList>();
const SettingsStack = createStackNavigator<SettingsStackParamList>();

const AnalyticsNavigator = () => (
  <AnalyticsStack.Navigator
    screenOptions={{
      headerShown: false,
      cardStyle: { backgroundColor: '#0A0A0A' },
    }}
  >
    <AnalyticsStack.Screen name="AnalyticsMain" component={AnalyticsScreen} />
    <AnalyticsStack.Screen name="ExerciseHistory" component={ExerciseHistoryScreen} />
    <AnalyticsStack.Screen name="AnalysisHistory" component={AnalysisHistoryScreen} />
  </AnalyticsStack.Navigator>
);

const SettingsNavigator = () => (
  <SettingsStack.Navigator
    screenOptions={{
      headerShown: false,
      cardStyle: { backgroundColor: '#0A0A0A' },
    }}
  >
    <SettingsStack.Screen name="SettingsMain" component={SettingsScreen} />
    <SettingsStack.Screen name="EditProfile" component={EditProfileScreen} />
    <SettingsStack.Screen name="Preferences" component={PreferencesScreen} />
    <SettingsStack.Screen name="DataExport" component={DataExportScreen} />
    <SettingsStack.Screen name="About" component={AboutScreen} />
  </SettingsStack.Navigator>
);

const AuthStack = () => (
  <Stack.Navigator
    screenOptions={{
      headerShown: false,
      cardStyle: { backgroundColor: '#0A0A0A' },
    }}
  >
    <Stack.Screen name="Login" component={LoginScreen} />
    <Stack.Screen name="Register" component={RegisterScreen} />
  </Stack.Navigator>
);

const MainTabs = () => (
  <Tab.Navigator
    screenOptions={{
      headerShown: false,
      tabBarStyle: {
        backgroundColor: '#1A1A1A',
        borderTopColor: '#2A2A2A',
        borderTopWidth: 1,
      },
      tabBarActiveTintColor: '#007AFF',
      tabBarInactiveTintColor: '#999',
    }}
  >
    <Tab.Screen
      name="Home"
      component={HomeScreen}
      options={{
        tabBarLabel: 'Home',
        tabBarIcon: ({ color }) => <Text style={{ color, fontSize: 24 }}>🏠</Text>,
      }}
    />
    <Tab.Screen
      name="Sessions"
      component={SessionListScreen}
      options={{
        tabBarLabel: 'Workouts',
        tabBarIcon: ({ color }) => <Text style={{ color, fontSize: 24 }}>💪</Text>,
      }}
    />
    <Tab.Screen
      name="Analytics"
      component={AnalyticsNavigator}
      options={{
        tabBarLabel: 'Analytics',
        tabBarIcon: ({ color }) => <Text style={{ color, fontSize: 24 }}>📊</Text>,
      }}
    />
    <Tab.Screen
      name="Settings"
      component={SettingsNavigator}
      options={{
        tabBarLabel: 'Settings',
        tabBarIcon: ({ color }) => <Text style={{ color, fontSize: 24 }}>⚙️</Text>,
      }}
    />
  </Tab.Navigator>
);

const AppStack = () => (
  <Stack.Navigator
    screenOptions={{
      headerShown: false,
      cardStyle: { backgroundColor: '#0A0A0A' },
    }}
  >
    <Stack.Screen name="MainTabs" component={MainTabs} />
    <Stack.Screen name="CreateSession" component={CreateSessionScreen} />
    <Stack.Screen name="SessionDetail" component={SessionDetailScreen} />
  </Stack.Navigator>
);

export const AppNavigator: React.FC = () => {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#007AFF" />
      </View>
    );
  }

  return (
    <NavigationContainer>
      {isAuthenticated ? <AppStack /> : <AuthStack />}
    </NavigationContainer>
  );
};

const styles = StyleSheet.create({
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#0A0A0A',
  },
});

export default AppNavigator;
