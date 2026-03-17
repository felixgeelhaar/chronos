import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  Alert,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import { useAuth } from '../../shared/contexts';
import { SyncStatusIndicator } from '../../shared/components';

type SettingsNavigationProp = StackNavigationProp<any>;

export const SettingsScreen: React.FC = () => {
  const navigation = useNavigation<SettingsNavigationProp>();
  const { user, logout } = useAuth();

  const handleLogout = () => {
    Alert.alert(
      'Logout',
      'Are you sure you want to logout?',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Logout',
          style: 'destructive',
          onPress: async () => {
            try {
              await logout();
            } catch (error) {
              console.error('Logout error:', error);
              Alert.alert('Error', 'Failed to logout');
            }
          },
        },
      ]
    );
  };

  const SettingItem: React.FC<{
    title: string;
    subtitle?: string;
    icon: string;
    onPress: () => void;
    showChevron?: boolean;
  }> = ({ title, subtitle, icon, onPress, showChevron = true }) => (
    <TouchableOpacity style={styles.settingItem} onPress={onPress}>
      <View style={styles.settingItemLeft}>
        <Text style={styles.settingIcon}>{icon}</Text>
        <View style={styles.settingTextContainer}>
          <Text style={styles.settingTitle}>{title}</Text>
          {subtitle && <Text style={styles.settingSubtitle}>{subtitle}</Text>}
        </View>
      </View>
      {showChevron && <Text style={styles.chevron}>›</Text>}
    </TouchableOpacity>
  );

  const SettingSection: React.FC<{ title: string; children: React.ReactNode }> = ({
    title,
    children,
  }) => (
    <View style={styles.section}>
      <Text style={styles.sectionTitle}>{title}</Text>
      <View style={styles.sectionContent}>{children}</View>
    </View>
  );

  return (
    <ScrollView style={styles.container} contentContainerStyle={styles.content}>
      <View style={styles.header}>
        <Text style={styles.title}>Settings</Text>
        <SyncStatusIndicator />
      </View>

      {/* User Profile Section */}
      <SettingSection title="Profile">
        <SettingItem
          icon="👤"
          title="Edit Profile"
          subtitle={user?.name || 'Update your profile'}
          onPress={() => navigation.navigate('EditProfile')}
        />
        <SettingItem
          icon="⚖️"
          title="Body Weight"
          subtitle={user?.body_weight ? `${user.body_weight} kg` : 'Not set'}
          onPress={() => navigation.navigate('EditProfile')}
        />
      </SettingSection>

      {/* App Preferences */}
      <SettingSection title="Preferences">
        <SettingItem
          icon="⚙️"
          title="App Preferences"
          subtitle="Units, notifications, and more"
          onPress={() => navigation.navigate('Preferences')}
        />
        <SettingItem
          icon="🌙"
          title="Theme"
          subtitle="Dark mode (default)"
          onPress={() => Alert.alert('Theme', 'Light mode coming soon!')}
        />
      </SettingSection>

      {/* Data Management */}
      <SettingSection title="Data">
        <SettingItem
          icon="💾"
          title="Export Data"
          subtitle="Download your workout data"
          onPress={() => navigation.navigate('DataExport')}
        />
        <SettingItem
          icon="🔄"
          title="Sync Status"
          subtitle="Manage offline data sync"
          onPress={() => Alert.alert('Sync', 'View detailed sync status')}
        />
      </SettingSection>

      {/* Account Management */}
      <SettingSection title="Account">
        <SettingItem
          icon="🔒"
          title="Change Password"
          subtitle="Update your password"
          onPress={() => navigation.navigate('ChangePassword')}
        />
        <SettingItem
          icon="🗑️"
          title="Delete Account"
          subtitle="Permanently delete your account"
          onPress={() => navigation.navigate('DeleteAccount')}
        />
      </SettingSection>

      {/* About */}
      <SettingSection title="About">
        <SettingItem
          icon="ℹ️"
          title="About Ascend"
          subtitle="Version 1.0.0"
          onPress={() => navigation.navigate('About')}
        />
        <SettingItem
          icon="📄"
          title="Privacy Policy"
          subtitle="How we handle your data"
          onPress={() => navigation.navigate('PrivacyPolicy')}
        />
        <SettingItem
          icon="📝"
          title="Terms of Service"
          subtitle="Terms and conditions"
          onPress={() => navigation.navigate('TermsOfService')}
        />
        <SettingItem
          icon="❓"
          title="Help & Support"
          subtitle="Get help or send feedback"
          onPress={() => navigation.navigate('Help')}
        />
      </SettingSection>

      {/* Logout */}
      <TouchableOpacity style={styles.logoutButton} onPress={handleLogout}>
        <Text style={styles.logoutIcon}>🚪</Text>
        <Text style={styles.logoutText}>Logout</Text>
      </TouchableOpacity>

      <View style={styles.footer}>
        <Text style={styles.footerText}>Ascend Mobile v1.0.0</Text>
        <Text style={styles.footerText}>© 2025 Ascend. All rights reserved.</Text>
      </View>
    </ScrollView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0A0A0A',
  },
  content: {
    padding: 24,
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 32,
  },
  title: {
    fontSize: 32,
    fontWeight: '700',
    color: '#FFF',
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
    overflow: 'hidden',
  },
  settingItem: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#2A2A2A',
  },
  settingItemLeft: {
    flexDirection: 'row',
    alignItems: 'center',
    flex: 1,
  },
  settingIcon: {
    fontSize: 24,
    marginRight: 12,
  },
  settingTextContainer: {
    flex: 1,
  },
  settingTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#FFF',
    marginBottom: 2,
  },
  settingSubtitle: {
    fontSize: 13,
    color: '#999',
  },
  chevron: {
    fontSize: 24,
    color: '#666',
    fontWeight: '300',
  },
  logoutButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#FF3B30',
    marginBottom: 32,
  },
  logoutIcon: {
    fontSize: 20,
    marginRight: 8,
  },
  logoutText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#FF3B30',
  },
  footer: {
    alignItems: 'center',
    paddingVertical: 20,
  },
  footerText: {
    fontSize: 12,
    color: '#666',
    marginBottom: 4,
  },
});

export default SettingsScreen;
