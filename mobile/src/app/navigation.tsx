import React from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createStackNavigator } from '@react-navigation/stack';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';

// Placeholder screens - will be replaced with actual feature screens
const PlaceholderScreen = ({ route }: any) => {
  return (
    <div style={{ flex: 1, justifyContent: 'center', alignItems: 'center' }}>
      <h1>{route.name}</h1>
    </div>
  );
};

// Auth Stack Navigation
const AuthStack = createStackNavigator();
const AuthNavigator = () => {
  return (
    <AuthStack.Navigator
      screenOptions={{
        headerShown: false,
      }}
    >
      <AuthStack.Screen name="Login" component={PlaceholderScreen} />
      <AuthStack.Screen name="Register" component={PlaceholderScreen} />
      <AuthStack.Screen name="Onboarding" component={PlaceholderScreen} />
    </AuthStack.Navigator>
  );
};

// Main Tab Navigation
const MainTabs = createBottomTabNavigator();
const MainNavigator = () => {
  return (
    <MainTabs.Navigator
      screenOptions={{
        headerShown: true,
      }}
    >
      <MainTabs.Screen
        name="Log"
        component={PlaceholderScreen}
        options={{
          title: 'Log Session',
        }}
      />
      <MainTabs.Screen
        name="Analytics"
        component={PlaceholderScreen}
        options={{
          title: 'Analytics',
        }}
      />
      <MainTabs.Screen
        name="Videos"
        component={PlaceholderScreen}
        options={{
          title: 'Videos',
        }}
      />
      <MainTabs.Screen
        name="Profile"
        component={PlaceholderScreen}
        options={{
          title: 'Profile',
        }}
      />
    </MainTabs.Navigator>
  );
};

// Root Navigator
const RootStack = createStackNavigator();
export const RootNavigator = () => {
  // TODO: Replace with actual auth state check
  const isAuthenticated = false;

  return (
    <NavigationContainer>
      <RootStack.Navigator screenOptions={{ headerShown: false }}>
        {isAuthenticated ? (
          <RootStack.Screen name="Main" component={MainNavigator} />
        ) : (
          <RootStack.Screen name="Auth" component={AuthNavigator} />
        )}
      </RootStack.Navigator>
    </NavigationContainer>
  );
};

// Type definitions for navigation
export type RootStackParamList = {
  Auth: undefined;
  Main: undefined;
};

export type AuthStackParamList = {
  Login: undefined;
  Register: undefined;
  Onboarding: undefined;
};

export type MainTabParamList = {
  Log: undefined;
  Analytics: undefined;
  Videos: undefined;
  Profile: undefined;
};
