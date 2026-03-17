import React from 'react';
import { View, Text, TouchableOpacity, StyleSheet } from 'react-native';
import { useAuth } from '../../shared/contexts';
import { SyncStatusIndicator } from '../../shared/components';

interface HomeScreenProps {
  navigation: any;
}

export const HomeScreen: React.FC<HomeScreenProps> = ({ navigation }) => {
  const { user, logout } = useAuth();

  const handleLogout = async () => {
    try {
      await logout();
    } catch (error) {
      console.error('Logout error:', error);
    }
  };

  return (
    <View style={styles.container}>
      <View style={styles.headerRow}>
        <View style={styles.header}>
          <Text style={styles.title}>Welcome to Ascend</Text>
          <Text style={styles.subtitle}>Hello, {user?.name}!</Text>
        </View>
        <SyncStatusIndicator />
      </View>

      <View style={styles.content}>
        <Text style={styles.comingSoon}>More features coming soon...</Text>
        <Text style={styles.description}>
          Track your workouts, analyze your progress, and reach new heights.
        </Text>
      </View>

      <TouchableOpacity style={styles.logoutButton} onPress={handleLogout}>
        <Text style={styles.logoutButtonText}>Logout</Text>
      </TouchableOpacity>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0A0A0A',
    padding: 24,
  },
  headerRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginTop: 60,
    marginBottom: 40,
  },
  header: {
    flex: 1,
  },
  title: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#FFF',
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 18,
    color: '#999',
  },
  content: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  comingSoon: {
    fontSize: 24,
    fontWeight: '600',
    color: '#FFF',
    marginBottom: 16,
  },
  description: {
    fontSize: 16,
    color: '#999',
    textAlign: 'center',
    paddingHorizontal: 40,
  },
  logoutButton: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  logoutButtonText: {
    color: '#FF4444',
    fontSize: 16,
    fontWeight: '600',
  },
});

export default HomeScreen;
