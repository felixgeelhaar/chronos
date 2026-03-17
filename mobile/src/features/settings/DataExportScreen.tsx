import React, { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  Alert,
  ActivityIndicator,
  Platform,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import Share from 'react-native-share';
import RNFS from 'react-native-fs';
import database from '../../database';
import { Session } from '../../database/models';
import { format } from 'date-fns';

export const DataExportScreen: React.FC = () => {
  const navigation = useNavigation();
  const [isExporting, setIsExporting] = useState(false);
  const [exportType, setExportType] = useState<'json' | 'csv'>('json');

  const exportData = async () => {
    try {
      setIsExporting(true);

      // Fetch all sessions with sets
      const sessions = await database.get<Session>('sessions').query().fetch();

      const data = await Promise.all(
        sessions.map(async session => {
          const sets = await session.sets.fetch();
          return {
            id: session.serverId || session.id,
            date: session.date,
            notes: session.notes,
            totalVolume: session.totalVolume,
            totalSets: session.totalSets,
            sets: sets.map(set => ({
              exerciseName: set.exerciseName,
              weight: set.weight,
              reps: set.reps,
              rpe: set.rpe,
              setOrder: set.setOrder,
              volume: set.volume,
              estimatedOneRepMax: set.estimatedOneRepMax,
              videoId: set.videoId,
            })),
          };
        })
      );

      let fileContent: string;
      let filename: string;
      let mimeType: string;

      if (exportType === 'json') {
        fileContent = JSON.stringify(
          {
            exportedAt: new Date().toISOString(),
            version: '1.0',
            sessions: data,
          },
          null,
          2
        );
        filename = `ascend-export-${format(new Date(), 'yyyy-MM-dd')}.json`;
        mimeType = 'application/json';
      } else {
        // CSV export
        const csvRows: string[] = [
          'Date,Exercise,Weight (kg),Reps,RPE,Set Order,Volume (kg),Estimated 1RM (kg),Notes',
        ];

        for (const session of data) {
          for (const set of session.sets) {
            csvRows.push(
              [
                session.date,
                set.exerciseName,
                set.weight,
                set.reps,
                set.rpe || '',
                set.setOrder,
                set.volume,
                set.estimatedOneRepMax,
                session.notes ? `"${session.notes.replace(/"/g, '""')}"` : '',
              ].join(',')
            );
          }
        }

        fileContent = csvRows.join('\n');
        filename = `ascend-export-${format(new Date(), 'yyyy-MM-dd')}.csv`;
        mimeType = 'text/csv';
      }

      // Write to file
      const filePath = `${RNFS.DocumentDirectoryPath}/${filename}`;
      await RNFS.writeFile(filePath, fileContent, 'utf8');

      // Share file
      await Share.open({
        title: 'Export Workout Data',
        message: 'Your Ascend workout data export',
        url: Platform.OS === 'android' ? `file://${filePath}` : filePath,
        type: mimeType,
        subject: 'Ascend Workout Data Export',
      });

      Alert.alert(
        'Export Complete',
        `Successfully exported ${data.length} workout sessions`
      );
    } catch (error: any) {
      if (error.message !== 'User did not share') {
        console.error('Export failed:', error);
        Alert.alert('Export Failed', 'Failed to export data. Please try again.');
      }
    } finally {
      setIsExporting(false);
    }
  };

  const ExportOption: React.FC<{
    title: string;
    description: string;
    type: 'json' | 'csv';
    icon: string;
  }> = ({ title, description, type, icon }) => (
    <TouchableOpacity
      style={[
        styles.exportOption,
        exportType === type && styles.exportOptionSelected,
      ]}
      onPress={() => setExportType(type)}
    >
      <View style={styles.exportOptionContent}>
        <Text style={styles.exportIcon}>{icon}</Text>
        <View style={styles.exportOptionText}>
          <Text style={styles.exportOptionTitle}>{title}</Text>
          <Text style={styles.exportOptionDescription}>{description}</Text>
        </View>
        <View
          style={[
            styles.radio,
            exportType === type && styles.radioSelected,
          ]}
        >
          {exportType === type && <View style={styles.radioDot} />}
        </View>
      </View>
    </TouchableOpacity>
  );

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <TouchableOpacity onPress={() => navigation.goBack()}>
          <Text style={styles.backButton}>← Back</Text>
        </TouchableOpacity>
        <Text style={styles.title}>Export Data</Text>
        <View style={{ width: 60 }} />
      </View>

      <ScrollView style={styles.content} contentContainerStyle={styles.contentContainer}>
        <Text style={styles.sectionTitle}>Export Format</Text>
        <Text style={styles.sectionDescription}>
          Choose how you'd like to export your workout data
        </Text>

        <ExportOption
          type="json"
          icon="📄"
          title="JSON Format"
          description="Structured data format, ideal for developers and data analysis tools"
        />

        <ExportOption
          type="csv"
          icon="📊"
          title="CSV Format"
          description="Spreadsheet format, open in Excel, Google Sheets, or Numbers"
        />

        <View style={styles.infoBox}>
          <Text style={styles.infoIcon}>💡</Text>
          <View style={styles.infoContent}>
            <Text style={styles.infoTitle}>What's Included</Text>
            <Text style={styles.infoText}>
              • All workout sessions with dates and notes{'\n'}
              • Exercise details: weight, reps, RPE{'\n'}
              • Calculated metrics: volume, estimated 1RM{'\n'}
              • Video references (IDs only){'\n'}
              • Export timestamp for reference
            </Text>
          </View>
        </View>

        <TouchableOpacity
          style={[styles.exportButton, isExporting && styles.exportButtonDisabled]}
          onPress={exportData}
          disabled={isExporting}
        >
          {isExporting ? (
            <>
              <ActivityIndicator color="#FFF" size="small" />
              <Text style={styles.exportButtonText}>Exporting...</Text>
            </>
          ) : (
            <>
              <Text style={styles.exportButtonIcon}>📤</Text>
              <Text style={styles.exportButtonText}>Export Data</Text>
            </>
          )}
        </TouchableOpacity>

        <View style={styles.privacyNote}>
          <Text style={styles.privacyNoteIcon}>🔒</Text>
          <Text style={styles.privacyNoteText}>
            Your data export is generated locally on your device and never sent to our servers
            during the export process.
          </Text>
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
  sectionTitle: {
    fontSize: 20,
    fontWeight: '600',
    color: '#FFF',
    marginBottom: 8,
  },
  sectionDescription: {
    fontSize: 14,
    color: '#999',
    marginBottom: 24,
    lineHeight: 20,
  },
  exportOption: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    borderWidth: 2,
    borderColor: '#2A2A2A',
    padding: 16,
    marginBottom: 12,
  },
  exportOptionSelected: {
    borderColor: '#007AFF',
    backgroundColor: '#0A1A2A',
  },
  exportOptionContent: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  exportIcon: {
    fontSize: 32,
    marginRight: 16,
  },
  exportOptionText: {
    flex: 1,
  },
  exportOptionTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#FFF',
    marginBottom: 4,
  },
  exportOptionDescription: {
    fontSize: 13,
    color: '#999',
    lineHeight: 18,
  },
  radio: {
    width: 24,
    height: 24,
    borderRadius: 12,
    borderWidth: 2,
    borderColor: '#666',
    justifyContent: 'center',
    alignItems: 'center',
  },
  radioSelected: {
    borderColor: '#007AFF',
  },
  radioDot: {
    width: 12,
    height: 12,
    borderRadius: 6,
    backgroundColor: '#007AFF',
  },
  infoBox: {
    flexDirection: 'row',
    backgroundColor: '#1A2A1A',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#2A3A2A',
    marginTop: 24,
    marginBottom: 24,
  },
  infoIcon: {
    fontSize: 24,
    marginRight: 12,
  },
  infoContent: {
    flex: 1,
  },
  infoTitle: {
    fontSize: 14,
    fontWeight: '600',
    color: '#7ACA7A',
    marginBottom: 8,
  },
  infoText: {
    fontSize: 13,
    color: '#8AAA8A',
    lineHeight: 20,
  },
  exportButton: {
    flexDirection: 'row',
    backgroundColor: '#007AFF',
    borderRadius: 12,
    padding: 18,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 16,
  },
  exportButtonDisabled: {
    opacity: 0.6,
  },
  exportButtonIcon: {
    fontSize: 20,
    marginRight: 8,
  },
  exportButtonText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#FFF',
  },
  privacyNote: {
    flexDirection: 'row',
    backgroundColor: '#1A1A2A',
    borderRadius: 8,
    padding: 12,
    borderWidth: 1,
    borderColor: '#2A2A3A',
  },
  privacyNoteIcon: {
    fontSize: 16,
    marginRight: 8,
  },
  privacyNoteText: {
    flex: 1,
    fontSize: 12,
    color: '#8A8AAA',
    lineHeight: 16,
  },
});

export default DataExportScreen;
