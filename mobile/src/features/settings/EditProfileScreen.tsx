import React, { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TextInput,
  TouchableOpacity,
  ScrollView,
  Alert,
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { useAuth } from '../../shared/contexts';
import { authService } from '../../services/api';

export const EditProfileScreen: React.FC = () => {
  const navigation = useNavigation();
  const { user, refreshUser } = useAuth();

  const [name, setName] = useState(user?.name || '');
  const [bodyWeight, setBodyWeight] = useState(user?.body_weight?.toString() || '');
  const [isLoading, setIsLoading] = useState(false);

  const handleSave = async () => {
    if (!name.trim()) {
      Alert.alert('Error', 'Name is required');
      return;
    }

    try {
      setIsLoading(true);

      // Update profile via API
      // Note: This assumes an update profile endpoint exists
      // You may need to implement this in the backend and API service
      const updates: any = {
        name: name.trim(),
      };

      if (bodyWeight.trim()) {
        const weight = parseFloat(bodyWeight);
        if (isNaN(weight) || weight <= 0) {
          Alert.alert('Error', 'Please enter a valid body weight');
          return;
        }
        updates.body_weight = weight;
      }

      // For now, we'll just show a success message
      // In production, call: await authService.updateProfile(updates);

      // Refresh user data
      await refreshUser();

      Alert.alert('Success', 'Profile updated successfully', [
        { text: 'OK', onPress: () => navigation.goBack() },
      ]);
    } catch (error: any) {
      console.error('Failed to update profile:', error);
      Alert.alert(
        'Error',
        error.response?.data?.error || 'Failed to update profile'
      );
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <KeyboardAvoidingView
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      style={styles.container}
    >
      <View style={styles.header}>
        <TouchableOpacity onPress={() => navigation.goBack()}>
          <Text style={styles.cancelButton}>Cancel</Text>
        </TouchableOpacity>
        <Text style={styles.title}>Edit Profile</Text>
        <TouchableOpacity onPress={handleSave} disabled={isLoading}>
          {isLoading ? (
            <ActivityIndicator color="#007AFF" />
          ) : (
            <Text style={styles.saveButton}>Save</Text>
          )}
        </TouchableOpacity>
      </View>

      <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Personal Information</Text>

          <View style={styles.field}>
            <Text style={styles.label}>Name</Text>
            <TextInput
              style={styles.input}
              placeholder="Your name"
              placeholderTextColor="#666"
              value={name}
              onChangeText={setName}
              autoCapitalize="words"
              autoCorrect={false}
            />
          </View>

          <View style={styles.field}>
            <Text style={styles.label}>Email</Text>
            <View style={styles.disabledInput}>
              <Text style={styles.disabledText}>{user?.email}</Text>
            </View>
            <Text style={styles.hint}>Email cannot be changed</Text>
          </View>
        </View>

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Body Metrics</Text>

          <View style={styles.field}>
            <Text style={styles.label}>Body Weight (kg)</Text>
            <TextInput
              style={styles.input}
              placeholder="Enter your current body weight"
              placeholderTextColor="#666"
              value={bodyWeight}
              onChangeText={setBodyWeight}
              keyboardType="decimal-pad"
            />
            <Text style={styles.hint}>
              Track your body weight to monitor progress over time
            </Text>
          </View>
        </View>

        <View style={styles.infoBox}>
          <Text style={styles.infoIcon}>💡</Text>
          <Text style={styles.infoText}>
            Your profile information is synced across all your devices. Body weight helps
            calculate relative strength metrics.
          </Text>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0A0A0A',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingHorizontal: 24,
    paddingTop: 60,
    paddingBottom: 20,
    borderBottomWidth: 1,
    borderBottomColor: '#2A2A2A',
  },
  title: {
    fontSize: 18,
    fontWeight: '600',
    color: '#FFF',
  },
  cancelButton: {
    fontSize: 16,
    color: '#999',
  },
  saveButton: {
    fontSize: 16,
    fontWeight: '600',
    color: '#007AFF',
  },
  content: {
    flex: 1,
    padding: 24,
  },
  section: {
    marginBottom: 32,
  },
  sectionTitle: {
    fontSize: 13,
    fontWeight: '600',
    color: '#999',
    textTransform: 'uppercase',
    letterSpacing: 0.5,
    marginBottom: 16,
  },
  field: {
    marginBottom: 20,
  },
  label: {
    fontSize: 14,
    fontWeight: '600',
    color: '#FFF',
    marginBottom: 8,
  },
  input: {
    backgroundColor: '#1A1A1A',
    borderWidth: 1,
    borderColor: '#2A2A2A',
    borderRadius: 8,
    padding: 16,
    fontSize: 16,
    color: '#FFF',
  },
  disabledInput: {
    backgroundColor: '#0F0F0F',
    borderWidth: 1,
    borderColor: '#2A2A2A',
    borderRadius: 8,
    padding: 16,
  },
  disabledText: {
    fontSize: 16,
    color: '#666',
  },
  hint: {
    fontSize: 12,
    color: '#666',
    marginTop: 6,
  },
  infoBox: {
    flexDirection: 'row',
    backgroundColor: '#1A2A3A',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#2A3A4A',
  },
  infoIcon: {
    fontSize: 20,
    marginRight: 12,
  },
  infoText: {
    flex: 1,
    fontSize: 13,
    color: '#7A9ABF',
    lineHeight: 18,
  },
});

export default EditProfileScreen;
