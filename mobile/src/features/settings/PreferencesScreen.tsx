import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  Switch,
  Alert,
  Modal,
  TextInput,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { workoutReminderService, WorkoutReminderSettings } from '../../services/notifications/workoutReminderService';

const PREFERENCES_KEY = '@ascend:preferences';

interface Preferences {
  weightUnit: 'kg' | 'lbs';
  distanceUnit: 'km' | 'miles';
  notifications: {
    workoutReminders: boolean;
    progressUpdates: boolean;
    syncComplete: boolean;
  };
  privacy: {
    analytics: boolean;
    crashReports: boolean;
  };
  display: {
    showEstimated1RM: boolean;
    showVolume: boolean;
    compactView: boolean;
  };
}

const DEFAULT_PREFERENCES: Preferences = {
  weightUnit: 'kg',
  distanceUnit: 'km',
  notifications: {
    workoutReminders: true,
    progressUpdates: true,
    syncComplete: false,
  },
  privacy: {
    analytics: false,
    crashReports: true,
  },
  display: {
    showEstimated1RM: true,
    showVolume: true,
    compactView: false,
  },
};

export const PreferencesScreen: React.FC = () => {
  const navigation = useNavigation();
  const [preferences, setPreferences] = useState<Preferences>(DEFAULT_PREFERENCES);
  const [isLoading, setIsLoading] = useState(true);
  const [reminderSettings, setReminderSettings] = useState<WorkoutReminderSettings | null>(null);
  const [showReminderModal, setShowReminderModal] = useState(false);

  useEffect(() => {
    loadPreferences();
    loadReminderSettings();
  }, []);

  const loadPreferences = async () => {
    try {
      const stored = await AsyncStorage.getItem(PREFERENCES_KEY);
      if (stored) {
        setPreferences(JSON.parse(stored));
      }
    } catch (error) {
      console.error('Failed to load preferences:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const savePreferences = async (newPreferences: Preferences) => {
    try {
      await AsyncStorage.setItem(PREFERENCES_KEY, JSON.stringify(newPreferences));
      setPreferences(newPreferences);
    } catch (error) {
      console.error('Failed to save preferences:', error);
      Alert.alert('Error', 'Failed to save preferences');
    }
  };

  const updatePreference = (path: string[], value: any) => {
    const newPreferences = { ...preferences };
    let current: any = newPreferences;

    for (let i = 0; i < path.length - 1; i++) {
      current = current[path[i]];
    }
    current[path[path.length - 1]] = value;

    savePreferences(newPreferences);

    // If workout reminders toggle changed, update reminder service
    if (path[0] === 'notifications' && path[1] === 'workoutReminders') {
      updateReminderService(value);
    }
  };

  const loadReminderSettings = async () => {
    try {
      const settings = await workoutReminderService.getSettings();
      setReminderSettings(settings);
    } catch (error) {
      console.error('Failed to load reminder settings:', error);
    }
  };

  const updateReminderService = async (enabled: boolean) => {
    try {
      await workoutReminderService.updateSettings({ enabled });
      await loadReminderSettings();
    } catch (error) {
      console.error('Failed to update reminder service:', error);
      Alert.alert('Error', 'Failed to update workout reminders');
    }
  };

  const saveReminderSettings = async (newSettings: Partial<WorkoutReminderSettings>) => {
    try {
      await workoutReminderService.updateSettings(newSettings);
      await loadReminderSettings();
      setShowReminderModal(false);
      Alert.alert('Success', 'Reminder settings updated');
    } catch (error) {
      console.error('Failed to save reminder settings:', error);
      Alert.alert('Error', 'Failed to save reminder settings');
    }
  };

  const toggleReminderDay = (day: number) => {
    if (!reminderSettings) return;

    const newDays = reminderSettings.days.includes(day)
      ? reminderSettings.days.filter(d => d !== day)
      : [...reminderSettings.days, day].sort();

    const newSettings = { ...reminderSettings, days: newDays };
    setReminderSettings(newSettings);
  };

  const SettingSection: React.FC<{ title: string; children: React.ReactNode }> = ({
    title,
    children,
  }) => (
    <View style={styles.section}>
      <Text style={styles.sectionTitle}>{title}</Text>
      <View style={styles.sectionContent}>{children}</View>
    </View>
  );

  const SettingRow: React.FC<{
    title: string;
    subtitle?: string;
    value: boolean;
    onValueChange: (value: boolean) => void;
  }> = ({ title, subtitle, value, onValueChange }) => (
    <View style={styles.settingRow}>
      <View style={styles.settingText}>
        <Text style={styles.settingTitle}>{title}</Text>
        {subtitle && <Text style={styles.settingSubtitle}>{subtitle}</Text>}
      </View>
      <Switch
        value={value}
        onValueChange={onValueChange}
        trackColor={{ false: '#2A2A2A', true: '#007AFF' }}
        thumbColor={value ? '#FFF' : '#666'}
      />
    </View>
  );

  const UnitSelector: React.FC<{
    title: string;
    options: Array<{ label: string; value: string }>;
    selected: string;
    onSelect: (value: string) => void;
  }> = ({ title, options, selected, onSelect }) => (
    <View style={styles.unitSelector}>
      <Text style={styles.settingTitle}>{title}</Text>
      <View style={styles.unitOptions}>
        {options.map(option => (
          <TouchableOpacity
            key={option.value}
            style={[
              styles.unitOption,
              selected === option.value && styles.unitOptionSelected,
            ]}
            onPress={() => onSelect(option.value)}
          >
            <Text
              style={[
                styles.unitOptionText,
                selected === option.value && styles.unitOptionTextSelected,
              ]}
            >
              {option.label}
            </Text>
          </TouchableOpacity>
        ))}
      </View>
    </View>
  );

  if (isLoading) {
    return (
      <View style={styles.container}>
        <Text style={styles.loadingText}>Loading preferences...</Text>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <TouchableOpacity onPress={() => navigation.goBack()}>
          <Text style={styles.backButton}>← Back</Text>
        </TouchableOpacity>
        <Text style={styles.title}>Preferences</Text>
        <View style={{ width: 60 }} />
      </View>

      <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
        {/* Units */}
        <SettingSection title="Units">
          <UnitSelector
            title="Weight"
            options={[
              { label: 'Kilograms (kg)', value: 'kg' },
              { label: 'Pounds (lbs)', value: 'lbs' },
            ]}
            selected={preferences.weightUnit}
            onSelect={(value) => updatePreference(['weightUnit'], value)}
          />
          <View style={styles.divider} />
          <UnitSelector
            title="Distance"
            options={[
              { label: 'Kilometers (km)', value: 'km' },
              { label: 'Miles', value: 'miles' },
            ]}
            selected={preferences.distanceUnit}
            onSelect={(value) => updatePreference(['distanceUnit'], value)}
          />
        </SettingSection>

        {/* Notifications */}
        <SettingSection title="Notifications">
          <SettingRow
            title="Workout Reminders"
            subtitle="Get reminded to log your workouts"
            value={preferences.notifications.workoutReminders}
            onValueChange={(value) =>
              updatePreference(['notifications', 'workoutReminders'], value)
            }
          />
          <View style={styles.divider} />
          <SettingRow
            title="Progress Updates"
            subtitle="Weekly summary of your progress"
            value={preferences.notifications.progressUpdates}
            onValueChange={(value) =>
              updatePreference(['notifications', 'progressUpdates'], value)
            }
          />
          <View style={styles.divider} />
          <SettingRow
            title="Sync Complete"
            subtitle="Notify when data sync is complete"
            value={preferences.notifications.syncComplete}
            onValueChange={(value) =>
              updatePreference(['notifications', 'syncComplete'], value)
            }
          />
          {preferences.notifications.workoutReminders && reminderSettings && (
            <>
              <View style={styles.divider} />
              <TouchableOpacity
                style={styles.reminderConfigButton}
                onPress={() => setShowReminderModal(true)}
              >
                <View style={styles.reminderConfigText}>
                  <Text style={styles.settingTitle}>⏰ Configure Reminders</Text>
                  <Text style={styles.settingSubtitle}>
                    {reminderSettings.time} • {reminderSettings.days.length} days selected
                  </Text>
                </View>
                <Text style={styles.chevron}>›</Text>
              </TouchableOpacity>
            </>
          )}
        </SettingSection>

        {/* Display */}
        <SettingSection title="Display">
          <SettingRow
            title="Show Estimated 1RM"
            subtitle="Display calculated one-rep max for sets"
            value={preferences.display.showEstimated1RM}
            onValueChange={(value) =>
              updatePreference(['display', 'showEstimated1RM'], value)
            }
          />
          <View style={styles.divider} />
          <SettingRow
            title="Show Volume"
            subtitle="Display total volume for exercises"
            value={preferences.display.showVolume}
            onValueChange={(value) =>
              updatePreference(['display', 'showVolume'], value)
            }
          />
          <View style={styles.divider} />
          <SettingRow
            title="Compact View"
            subtitle="Use compact layout for workout lists"
            value={preferences.display.compactView}
            onValueChange={(value) =>
              updatePreference(['display', 'compactView'], value)
            }
          />
        </SettingSection>

        {/* Privacy */}
        <SettingSection title="Privacy">
          <SettingRow
            title="Analytics"
            subtitle="Help improve the app with usage data"
            value={preferences.privacy.analytics}
            onValueChange={(value) =>
              updatePreference(['privacy', 'analytics'], value)
            }
          />
          <View style={styles.divider} />
          <SettingRow
            title="Crash Reports"
            subtitle="Automatically send crash reports"
            value={preferences.privacy.crashReports}
            onValueChange={(value) =>
              updatePreference(['privacy', 'crashReports'], value)
            }
          />
        </SettingSection>

        <View style={styles.footer}>
          <Text style={styles.footerText}>
            Changes are saved automatically
          </Text>
        </View>
      </ScrollView>

      {/* Reminder Configuration Modal */}
      {reminderSettings && (
        <Modal
          visible={showReminderModal}
          animationType="slide"
          transparent={true}
          onRequestClose={() => setShowReminderModal(false)}
        >
          <View style={styles.modalOverlay}>
            <View style={styles.modalContainer}>
              <View style={styles.modalHeader}>
                <Text style={styles.modalTitle}>Configure Reminders</Text>
                <TouchableOpacity onPress={() => setShowReminderModal(false)}>
                  <Text style={styles.modalClose}>✕</Text>
                </TouchableOpacity>
              </View>

              <ScrollView style={styles.modalContent}>
                {/* Time Selector */}
                <View style={styles.modalSection}>
                  <Text style={styles.modalSectionTitle}>Reminder Time</Text>
                  <TextInput
                    style={styles.timeInput}
                    value={reminderSettings.time}
                    onChangeText={(time) =>
                      setReminderSettings({ ...reminderSettings, time })
                    }
                    placeholder="HH:MM"
                    placeholderTextColor="#666"
                    keyboardType="numbers-and-punctuation"
                  />
                  <Text style={styles.helperText}>24-hour format (e.g., 18:00 for 6 PM)</Text>
                </View>

                {/* Day Selector */}
                <View style={styles.modalSection}>
                  <Text style={styles.modalSectionTitle}>Reminder Days</Text>
                  <View style={styles.daySelector}>
                    {['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'].map((day, index) => (
                      <TouchableOpacity
                        key={index}
                        style={[
                          styles.dayButton,
                          reminderSettings.days.includes(index) && styles.dayButtonSelected,
                        ]}
                        onPress={() => toggleReminderDay(index)}
                      >
                        <Text
                          style={[
                            styles.dayButtonText,
                            reminderSettings.days.includes(index) &&
                              styles.dayButtonTextSelected,
                          ]}
                        >
                          {day}
                        </Text>
                      </TouchableOpacity>
                    ))}
                  </View>
                </View>

                {/* Additional Options */}
                <View style={styles.modalSection}>
                  <Text style={styles.modalSectionTitle}>Additional Reminders</Text>
                  <View style={styles.modalOption}>
                    <View style={styles.modalOptionText}>
                      <Text style={styles.modalOptionTitle}>Rest Day Check-in</Text>
                      <Text style={styles.modalOptionSubtitle}>
                        Remind me after 2 days without a workout
                      </Text>
                    </View>
                    <Switch
                      value={reminderSettings.restDayReminder}
                      onValueChange={(value) =>
                        setReminderSettings({ ...reminderSettings, restDayReminder: value })
                      }
                      trackColor={{ false: '#2A2A2A', true: '#007AFF' }}
                      thumbColor={reminderSettings.restDayReminder ? '#FFF' : '#666'}
                    />
                  </View>
                  <View style={styles.modalOptionDivider} />
                  <View style={styles.modalOption}>
                    <View style={styles.modalOptionText}>
                      <Text style={styles.modalOptionTitle}>Streak Reminder</Text>
                      <Text style={styles.modalOptionSubtitle}>
                        Celebrate and maintain workout streaks
                      </Text>
                    </View>
                    <Switch
                      value={reminderSettings.streakReminder}
                      onValueChange={(value) =>
                        setReminderSettings({ ...reminderSettings, streakReminder: value })
                      }
                      trackColor={{ false: '#2A2A2A', true: '#007AFF' }}
                      thumbColor={reminderSettings.streakReminder ? '#FFF' : '#666'}
                    />
                  </View>
                </View>
              </ScrollView>

              <View style={styles.modalActions}>
                <TouchableOpacity
                  style={styles.modalButtonSecondary}
                  onPress={() => setShowReminderModal(false)}
                >
                  <Text style={styles.modalButtonSecondaryText}>Cancel</Text>
                </TouchableOpacity>
                <TouchableOpacity
                  style={styles.modalButtonPrimary}
                  onPress={() => saveReminderSettings(reminderSettings)}
                >
                  <Text style={styles.modalButtonPrimaryText}>Save</Text>
                </TouchableOpacity>
              </View>
            </View>
          </View>
        </Modal>
      )}
    </View>
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
  backButton: {
    fontSize: 16,
    color: '#007AFF',
  },
  loadingText: {
    fontSize: 16,
    color: '#999',
    textAlign: 'center',
    marginTop: 100,
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
    marginBottom: 12,
  },
  sectionContent: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    padding: 16,
  },
  settingRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 8,
  },
  settingText: {
    flex: 1,
    marginRight: 16,
  },
  settingTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#FFF',
    marginBottom: 4,
  },
  settingSubtitle: {
    fontSize: 13,
    color: '#999',
    lineHeight: 18,
  },
  divider: {
    height: 1,
    backgroundColor: '#2A2A2A',
    marginVertical: 16,
  },
  unitSelector: {
    paddingVertical: 8,
  },
  unitOptions: {
    flexDirection: 'row',
    gap: 8,
    marginTop: 12,
  },
  unitOption: {
    flex: 1,
    backgroundColor: '#2A2A2A',
    borderRadius: 8,
    padding: 12,
    alignItems: 'center',
    borderWidth: 2,
    borderColor: 'transparent',
  },
  unitOptionSelected: {
    backgroundColor: '#1A3A5A',
    borderColor: '#007AFF',
  },
  unitOptionText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#999',
  },
  unitOptionTextSelected: {
    color: '#007AFF',
  },
  footer: {
    alignItems: 'center',
    paddingVertical: 20,
  },
  footerText: {
    fontSize: 12,
    color: '#666',
  },
  reminderConfigButton: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 8,
  },
  reminderConfigText: {
    flex: 1,
  },
  chevron: {
    fontSize: 24,
    color: '#666',
    marginLeft: 8,
  },
  modalOverlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    justifyContent: 'flex-end',
  },
  modalContainer: {
    backgroundColor: '#1A1A1A',
    borderTopLeftRadius: 24,
    borderTopRightRadius: 24,
    maxHeight: '80%',
  },
  modalHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 24,
    borderBottomWidth: 1,
    borderBottomColor: '#2A2A2A',
  },
  modalTitle: {
    fontSize: 20,
    fontWeight: '700',
    color: '#FFF',
  },
  modalClose: {
    fontSize: 24,
    color: '#999',
  },
  modalContent: {
    padding: 24,
  },
  modalSection: {
    marginBottom: 32,
  },
  modalSectionTitle: {
    fontSize: 14,
    fontWeight: '600',
    color: '#999',
    textTransform: 'uppercase',
    letterSpacing: 0.5,
    marginBottom: 12,
  },
  timeInput: {
    backgroundColor: '#2A2A2A',
    borderRadius: 12,
    padding: 16,
    fontSize: 16,
    color: '#FFF',
    borderWidth: 1,
    borderColor: '#3A3A3A',
  },
  helperText: {
    fontSize: 12,
    color: '#666',
    marginTop: 8,
  },
  daySelector: {
    flexDirection: 'row',
    gap: 8,
  },
  dayButton: {
    flex: 1,
    aspectRatio: 1,
    backgroundColor: '#2A2A2A',
    borderRadius: 12,
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 2,
    borderColor: 'transparent',
  },
  dayButtonSelected: {
    backgroundColor: '#1A3A5A',
    borderColor: '#007AFF',
  },
  dayButtonText: {
    fontSize: 13,
    fontWeight: '600',
    color: '#999',
  },
  dayButtonTextSelected: {
    color: '#007AFF',
  },
  modalOption: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    paddingVertical: 12,
  },
  modalOptionText: {
    flex: 1,
    marginRight: 16,
  },
  modalOptionTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#FFF',
    marginBottom: 4,
  },
  modalOptionSubtitle: {
    fontSize: 13,
    color: '#999',
    lineHeight: 18,
  },
  modalOptionDivider: {
    height: 1,
    backgroundColor: '#2A2A2A',
    marginVertical: 12,
  },
  modalActions: {
    flexDirection: 'row',
    gap: 12,
    padding: 24,
    borderTopWidth: 1,
    borderTopColor: '#2A2A2A',
  },
  modalButtonSecondary: {
    flex: 1,
    backgroundColor: '#2A2A2A',
    borderRadius: 12,
    padding: 16,
    alignItems: 'center',
  },
  modalButtonSecondaryText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#999',
  },
  modalButtonPrimary: {
    flex: 1,
    backgroundColor: '#007AFF',
    borderRadius: 12,
    padding: 16,
    alignItems: 'center',
  },
  modalButtonPrimaryText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#FFF',
  },
});

export default PreferencesScreen;
