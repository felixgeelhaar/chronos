import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  Linking,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';

export const AboutScreen: React.FC = () => {
  const navigation = useNavigation();

  const openURL = (url: string) => {
    Linking.openURL(url).catch(err => console.error('Failed to open URL:', err));
  };

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <TouchableOpacity onPress={() => navigation.goBack()}>
          <Text style={styles.backButton}>← Back</Text>
        </TouchableOpacity>
        <Text style={styles.title}>About</Text>
        <View style={{ width: 60 }} />
      </View>

      <ScrollView style={styles.content} contentContainerStyle={styles.contentContainer}>
        <View style={styles.logoContainer}>
          <Text style={styles.logo}>🏔️</Text>
          <Text style={styles.appName}>Ascend</Text>
          <Text style={styles.version}>Version 1.0.0</Text>
        </View>

        <Text style={styles.description}>
          Ascend is a comprehensive weightlifting performance tracking application designed to
          help athletes monitor progress, analyze form, and achieve their strength training goals.
        </Text>

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Features</Text>
          <View style={styles.featuresList}>
            <Text style={styles.feature}>📊 Comprehensive workout logging</Text>
            <Text style={styles.feature}>📈 Progress analytics and ACWR monitoring</Text>
            <Text style={styles.feature}>🎥 Video recording and form analysis</Text>
            <Text style={styles.feature}>📴 Offline mode with automatic sync</Text>
            <Text style={styles.feature}>💪 Estimated 1RM calculations</Text>
            <Text style={styles.feature}>📱 Cross-platform mobile experience</Text>
          </View>
        </View>

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Credits</Text>
          <Text style={styles.creditText}>
            Developed with ❤️ for the strength training community
          </Text>
        </View>

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Open Source</Text>
          <Text style={styles.creditText}>
            Built with React Native, WatermelonDB, and other amazing open source projects
          </Text>
        </View>

        <View style={styles.linksContainer}>
          <TouchableOpacity
            style={styles.linkButton}
            onPress={() => openURL('https://ascend.app/privacy')}
          >
            <Text style={styles.linkIcon}>🔒</Text>
            <Text style={styles.linkText}>Privacy Policy</Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={styles.linkButton}
            onPress={() => openURL('https://ascend.app/terms')}
          >
            <Text style={styles.linkIcon}>📜</Text>
            <Text style={styles.linkText}>Terms of Service</Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={styles.linkButton}
            onPress={() => openURL('https://ascend.app')}
          >
            <Text style={styles.linkIcon}>🌐</Text>
            <Text style={styles.linkText}>Visit Website</Text>
          </TouchableOpacity>
        </View>

        <View style={styles.footer}>
          <Text style={styles.footerText}>© 2025 Ascend</Text>
          <Text style={styles.footerText}>All rights reserved</Text>
        </View>
      </ScrollView>
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
  content: {
    flex: 1,
  },
  contentContainer: {
    padding: 24,
  },
  logoContainer: {
    alignItems: 'center',
    paddingVertical: 40,
  },
  logo: {
    fontSize: 80,
    marginBottom: 16,
  },
  appName: {
    fontSize: 32,
    fontWeight: '700',
    color: '#FFF',
    marginBottom: 8,
  },
  version: {
    fontSize: 16,
    color: '#999',
  },
  description: {
    fontSize: 16,
    color: '#CCC',
    lineHeight: 24,
    textAlign: 'center',
    marginBottom: 40,
  },
  section: {
    marginBottom: 32,
  },
  sectionTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#FFF',
    marginBottom: 16,
  },
  featuresList: {
    gap: 12,
  },
  feature: {
    fontSize: 15,
    color: '#CCC',
    lineHeight: 22,
  },
  creditText: {
    fontSize: 15,
    color: '#999',
    lineHeight: 22,
  },
  linksContainer: {
    marginTop: 20,
    marginBottom: 40,
    gap: 12,
  },
  linkButton: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  linkIcon: {
    fontSize: 24,
    marginRight: 12,
  },
  linkText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#007AFF',
  },
  footer: {
    alignItems: 'center',
    paddingVertical: 20,
  },
  footerText: {
    fontSize: 13,
    color: '#666',
    marginBottom: 4,
  },
});

export default AboutScreen;
