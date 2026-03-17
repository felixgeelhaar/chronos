import React, { useState, useEffect, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
  RefreshControl,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { StackNavigationProp } from '@react-navigation/stack';
import { AnalyticsStackParamList } from '../../app/navigation/AppNavigator';
import { analyticsService } from '../../services/api';
import { ACWRResponse, VolumeProgressResponse } from '../../services/api/types';
import { LineChart, BarChart } from '../../shared/components';
import { format, subMonths } from 'date-fns';

const COMMON_EXERCISES = ['All', 'Squat', 'Bench Press', 'Deadlift', 'Overhead Press', 'Barbell Row'];

type AnalyticsScreenNavigationProp = StackNavigationProp<AnalyticsStackParamList, 'AnalyticsMain'>;

export const AnalyticsScreen: React.FC = () => {
  const navigation = useNavigation<AnalyticsScreenNavigationProp>();
  const [selectedExercise, setSelectedExercise] = useState<string>('All');
  const [acwrData, setAcwrData] = useState<ACWRResponse | null>(null);
  const [volumeData, setVolumeData] = useState<VolumeProgressResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const loadAnalytics = useCallback(async () => {
    try {
      const exercise = selectedExercise === 'All' ? undefined : selectedExercise;

      const [acwr, volume] = await Promise.all([
        analyticsService.getACWR(exercise),
        analyticsService.getVolumeProgress('month', exercise),
      ]);

      setAcwrData(acwr);
      setVolumeData(volume);
    } catch (error: any) {
      console.error('Failed to load analytics:', error);
      Alert.alert('Error', error.response?.data?.error || 'Failed to load analytics');
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  }, [selectedExercise]);

  useEffect(() => {
    loadAnalytics();
  }, [loadAnalytics]);

  const handleRefresh = useCallback(() => {
    setIsRefreshing(true);
    loadAnalytics();
  }, [loadAnalytics]);

  const getACWRStatus = (ratio: number) => {
    if (ratio >= 0.8 && ratio <= 1.3) {
      return { label: 'Optimal', color: '#34C759', icon: '✓' };
    } else if (ratio > 1.3) {
      return { label: 'High Risk', color: '#FF3B30', icon: '⚠' };
    } else {
      return { label: 'Undertraining', color: '#FF9500', icon: '!' };
    }
  };

  if (isLoading && !acwrData) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#007AFF" />
      </View>
    );
  }

  const acwrStatus = acwrData ? getACWRStatus(acwrData.current_ratio) : null;
  const acwrChartData = acwrData?.history?.map((point) => ({
    x: new Date(point.date),
    y: point.ratio,
  })) || [];

  const volumeChartData = volumeData?.data?.map((point) => ({
    x: point.period,
    y: point.volume,
  })) || [];

  return (
    <ScrollView
      style={styles.container}
      contentContainerStyle={styles.content}
      refreshControl={
        <RefreshControl
          refreshing={isRefreshing}
          onRefresh={handleRefresh}
          tintColor="#007AFF"
        />
      }
    >
      <Text style={styles.header}>Progress Analytics</Text>

      {/* Exercise Filter */}
      <View style={styles.filterContainer}>
        <ScrollView horizontal showsHorizontalScrollIndicator={false}>
          {COMMON_EXERCISES.map((exercise) => (
            <TouchableOpacity
              key={exercise}
              style={[
                styles.filterButton,
                selectedExercise === exercise && styles.filterButtonActive,
              ]}
              onPress={() => setSelectedExercise(exercise)}
            >
              <Text
                style={[
                  styles.filterButtonText,
                  selectedExercise === exercise && styles.filterButtonTextActive,
                ]}
              >
                {exercise}
              </Text>
            </TouchableOpacity>
          ))}
        </ScrollView>
      </View>

      {/* ACWR Section */}
      {acwrData && (
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Injury Risk Monitor (ACWR)</Text>

          {/* ACWR Current Status */}
          <View style={styles.acwrCard}>
            <View style={styles.acwrHeader}>
              <Text style={styles.acwrLabel}>Current Ratio</Text>
              {acwrStatus && (
                <View style={[styles.statusBadge, { backgroundColor: acwrStatus.color }]}>
                  <Text style={styles.statusIcon}>{acwrStatus.icon}</Text>
                  <Text style={styles.statusText}>{acwrStatus.label}</Text>
                </View>
              )}
            </View>
            <Text style={styles.acwrValue}>{acwrData.current_ratio.toFixed(2)}</Text>
            <View style={styles.acwrDetails}>
              <View style={styles.acwrDetailItem}>
                <Text style={styles.acwrDetailLabel}>Acute Load (7d)</Text>
                <Text style={styles.acwrDetailValue}>{acwrData.acute_load.toFixed(0)} kg</Text>
              </View>
              <View style={styles.acwrDetailDivider} />
              <View style={styles.acwrDetailItem}>
                <Text style={styles.acwrDetailLabel}>Chronic Load (28d)</Text>
                <Text style={styles.acwrDetailValue}>{acwrData.chronic_load.toFixed(0)} kg</Text>
              </View>
            </View>
          </View>

          {/* ACWR Trend */}
          {acwrChartData.length > 0 && (
            <View style={styles.chartContainer}>
              <LineChart
                data={acwrChartData}
                title="ACWR Trend (Last 30 Days)"
                yLabel="Ratio"
                color="#007AFF"
                height={200}
              />
              <View style={styles.chartLegend}>
                <View style={styles.legendItem}>
                  <View style={[styles.legendColor, { backgroundColor: '#34C759' }]} />
                  <Text style={styles.legendText}>Optimal (0.8-1.3)</Text>
                </View>
                <View style={styles.legendItem}>
                  <View style={[styles.legendColor, { backgroundColor: '#FF3B30' }]} />
                  <Text style={styles.legendText}>High Risk (&gt;1.3)</Text>
                </View>
                <View style={styles.legendItem}>
                  <View style={[styles.legendColor, { backgroundColor: '#FF9500' }]} />
                  <Text style={styles.legendText}>Undertraining (&lt;0.8)</Text>
                </View>
              </View>
            </View>
          )}
        </View>
      )}

      {/* Volume Progress Section */}
      {volumeData && volumeData.data.length > 0 && (
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Volume Progress</Text>

          {/* Summary Stats */}
          <View style={styles.summaryCards}>
            <View style={styles.summaryCard}>
              <Text style={styles.summaryLabel}>Total Volume</Text>
              <Text style={styles.summaryValue}>
                {volumeData.total_volume.toFixed(0)} kg
              </Text>
            </View>
            <View style={styles.summaryCard}>
              <Text style={styles.summaryLabel}>Avg per Period</Text>
              <Text style={styles.summaryValue}>
                {volumeData.average_volume.toFixed(0)} kg
              </Text>
            </View>
          </View>

          {/* Volume Chart */}
          <View style={styles.chartContainer}>
            <BarChart
              data={volumeChartData}
              title="Monthly Volume"
              yLabel="Total Volume (kg)"
              color="#34C759"
              height={220}
            />
          </View>

          {/* Navigation Buttons */}
          <TouchableOpacity
            style={styles.historyButton}
            onPress={() => navigation.navigate('ExerciseHistory')}
          >
            <Text style={styles.historyButtonText}>View Exercise History & 1RM Progress</Text>
            <Text style={styles.historyButtonIcon}>→</Text>
          </TouchableOpacity>

          <TouchableOpacity
            style={[styles.historyButton, styles.analysisButton]}
            onPress={() => navigation.navigate('AnalysisHistory')}
          >
            <Text style={styles.historyButtonText}>View Video Analysis History</Text>
            <Text style={styles.historyButtonIcon}>→</Text>
          </TouchableOpacity>
        </View>
      )}

      {/* Empty State */}
      {(!acwrData || !volumeData || volumeData.data.length === 0) && !isLoading && (
        <View style={styles.emptyState}>
          <Text style={styles.emptyIcon}>📊</Text>
          <Text style={styles.emptyTitle}>No Analytics Data Yet</Text>
          <Text style={styles.emptyText}>
            Start logging workouts to see your progress analytics and injury risk monitoring.
          </Text>
        </View>
      )}
    </ScrollView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0A0A0A',
  },
  content: {
    padding: 16,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#0A0A0A',
  },
  header: {
    fontSize: 28,
    fontWeight: '700',
    color: '#FFF',
    marginBottom: 20,
  },
  filterContainer: {
    marginBottom: 24,
  },
  filterButton: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 20,
    backgroundColor: '#1A1A1A',
    borderWidth: 1,
    borderColor: '#2A2A2A',
    marginRight: 8,
  },
  filterButtonActive: {
    backgroundColor: '#007AFF',
    borderColor: '#007AFF',
  },
  filterButtonText: {
    color: '#999',
    fontSize: 14,
    fontWeight: '600',
  },
  filterButtonTextActive: {
    color: '#FFF',
  },
  section: {
    marginBottom: 32,
  },
  sectionTitle: {
    fontSize: 20,
    fontWeight: '600',
    color: '#FFF',
    marginBottom: 16,
  },
  acwrCard: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 20,
    borderWidth: 1,
    borderColor: '#2A2A2A',
    marginBottom: 16,
  },
  acwrHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  acwrLabel: {
    fontSize: 14,
    color: '#999',
    fontWeight: '500',
  },
  statusBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 12,
  },
  statusIcon: {
    color: '#FFF',
    fontSize: 12,
    marginRight: 4,
  },
  statusText: {
    color: '#FFF',
    fontSize: 12,
    fontWeight: '600',
  },
  acwrValue: {
    fontSize: 48,
    fontWeight: '700',
    color: '#007AFF',
    marginBottom: 16,
  },
  acwrDetails: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  acwrDetailItem: {
    flex: 1,
  },
  acwrDetailDivider: {
    width: 1,
    height: 40,
    backgroundColor: '#2A2A2A',
    marginHorizontal: 16,
  },
  acwrDetailLabel: {
    fontSize: 12,
    color: '#999',
    marginBottom: 4,
  },
  acwrDetailValue: {
    fontSize: 18,
    fontWeight: '600',
    color: '#FFF',
  },
  chartContainer: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  chartLegend: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    marginTop: 12,
    paddingTop: 12,
    borderTopWidth: 1,
    borderTopColor: '#2A2A2A',
  },
  legendItem: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  legendColor: {
    width: 12,
    height: 12,
    borderRadius: 6,
    marginRight: 6,
  },
  legendText: {
    fontSize: 11,
    color: '#999',
  },
  summaryCards: {
    flexDirection: 'row',
    marginBottom: 16,
    gap: 12,
  },
  summaryCard: {
    flex: 1,
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  summaryLabel: {
    fontSize: 12,
    color: '#999',
    marginBottom: 8,
  },
  summaryValue: {
    fontSize: 24,
    fontWeight: '700',
    color: '#34C759',
  },
  emptyState: {
    alignItems: 'center',
    paddingVertical: 60,
  },
  emptyIcon: {
    fontSize: 64,
    marginBottom: 16,
  },
  emptyTitle: {
    fontSize: 18,
    fontWeight: '600',
    color: '#FFF',
    marginBottom: 8,
  },
  emptyText: {
    fontSize: 14,
    color: '#999',
    textAlign: 'center',
    paddingHorizontal: 32,
  },
  historyButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    backgroundColor: '#007AFF',
    borderRadius: 12,
    padding: 16,
    marginTop: 16,
  },
  historyButtonText: {
    color: '#FFF',
    fontSize: 16,
    fontWeight: '600',
  },
  historyButtonIcon: {
    color: '#FFF',
    fontSize: 20,
    fontWeight: '600',
  },
  analysisButton: {
    backgroundColor: '#34C759',
  },
});

export default AnalyticsScreen;
