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
import { analyticsService } from '../../services/api';
import { ExerciseHistoryResponse } from '../../services/api/types';
import { LineChart } from '../../shared/components';
import { format, subMonths, subYears } from 'date-fns';

const COMMON_EXERCISES = ['Squat', 'Bench Press', 'Deadlift', 'Overhead Press', 'Barbell Row', 'Pull-up'];
const DATE_RANGES = [
  { label: '1M', value: 1, unit: 'month' as const },
  { label: '3M', value: 3, unit: 'month' as const },
  { label: '6M', value: 6, unit: 'month' as const },
  { label: '1Y', value: 1, unit: 'year' as const },
  { label: 'All', value: null, unit: null },
];

export const ExerciseHistoryScreen: React.FC = () => {
  const [selectedExercise, setSelectedExercise] = useState<string>('Squat');
  const [selectedRange, setSelectedRange] = useState(3); // 3 months default
  const [historyData, setHistoryData] = useState<ExerciseHistoryResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const loadHistory = useCallback(async () => {
    try {
      const range = DATE_RANGES.find((r) => r.value === selectedRange);
      let startDate: string | undefined;

      if (range && range.value !== null) {
        const date =
          range.unit === 'month'
            ? subMonths(new Date(), range.value)
            : subYears(new Date(), range.value);
        startDate = format(date, 'yyyy-MM-dd');
      }

      const data = await analyticsService.getExerciseHistory(
        selectedExercise,
        startDate,
        undefined
      );

      setHistoryData(data);
    } catch (error: any) {
      console.error('Failed to load exercise history:', error);
      Alert.alert('Error', error.response?.data?.error || 'Failed to load exercise history');
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  }, [selectedExercise, selectedRange]);

  useEffect(() => {
    setIsLoading(true);
    loadHistory();
  }, [loadHistory]);

  const handleRefresh = useCallback(() => {
    setIsRefreshing(true);
    loadHistory();
  }, [loadHistory]);

  if (isLoading && !historyData) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#007AFF" />
      </View>
    );
  }

  const chartData = historyData?.history?.map((point) => ({
    x: new Date(point.date),
    y: point.estimated_one_rep_max,
  })) || [];

  const recentPRs = historyData?.history
    ?.filter((point) => point.is_pr)
    .slice(0, 5) || [];

  const maxOneRM = historyData?.max_one_rep_max || 0;
  const currentOneRM = historyData?.history?.[historyData.history.length - 1]?.estimated_one_rep_max || 0;
  const improvement = maxOneRM > 0 && currentOneRM > 0
    ? ((currentOneRM - (historyData?.history?.[0]?.estimated_one_rep_max || 0)) / (historyData?.history?.[0]?.estimated_one_rep_max || 1)) * 100
    : 0;

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
      <Text style={styles.header}>Exercise History</Text>

      {/* Exercise Selection */}
      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Select Exercise</Text>
        <ScrollView horizontal showsHorizontalScrollIndicator={false}>
          {COMMON_EXERCISES.map((exercise) => (
            <TouchableOpacity
              key={exercise}
              style={[
                styles.exerciseButton,
                selectedExercise === exercise && styles.exerciseButtonActive,
              ]}
              onPress={() => setSelectedExercise(exercise)}
            >
              <Text
                style={[
                  styles.exerciseButtonText,
                  selectedExercise === exercise && styles.exerciseButtonTextActive,
                ]}
              >
                {exercise}
              </Text>
            </TouchableOpacity>
          ))}
        </ScrollView>
      </View>

      {/* Date Range Selection */}
      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Time Period</Text>
        <View style={styles.dateRangeContainer}>
          {DATE_RANGES.map((range) => (
            <TouchableOpacity
              key={range.label}
              style={[
                styles.dateButton,
                selectedRange === range.value && styles.dateButtonActive,
              ]}
              onPress={() => setSelectedRange(range.value)}
            >
              <Text
                style={[
                  styles.dateButtonText,
                  selectedRange === range.value && styles.dateButtonTextActive,
                ]}
              >
                {range.label}
              </Text>
            </TouchableOpacity>
          ))}
        </View>
      </View>

      {/* Statistics Cards */}
      {historyData && historyData.history.length > 0 && (
        <View style={styles.statsContainer}>
          <View style={styles.statCard}>
            <Text style={styles.statLabel}>Current 1RM</Text>
            <Text style={styles.statValue}>{currentOneRM.toFixed(1)} kg</Text>
          </View>
          <View style={styles.statCard}>
            <Text style={styles.statLabel}>Max 1RM</Text>
            <Text style={styles.statValue}>{maxOneRM.toFixed(1)} kg</Text>
          </View>
          <View style={styles.statCard}>
            <Text style={styles.statLabel}>Improvement</Text>
            <Text style={[styles.statValue, improvement >= 0 ? styles.statPositive : styles.statNegative]}>
              {improvement >= 0 ? '+' : ''}{improvement.toFixed(1)}%
            </Text>
          </View>
        </View>
      )}

      {/* 1RM Progression Chart */}
      {chartData.length > 0 && (
        <View style={styles.section}>
          <View style={styles.chartContainer}>
            <LineChart
              data={chartData}
              title="1RM Progression"
              yLabel="Estimated 1RM (kg)"
              color="#007AFF"
              height={240}
            />
          </View>
        </View>
      )}

      {/* Recent PRs */}
      {recentPRs.length > 0 && (
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Recent Personal Records</Text>
          <View style={styles.prList}>
            {recentPRs.map((pr, index) => (
              <View key={index} style={styles.prCard}>
                <View style={styles.prHeader}>
                  <Text style={styles.prDate}>{format(new Date(pr.date), 'MMM d, yyyy')}</Text>
                  <View style={styles.prBadge}>
                    <Text style={styles.prBadgeText}>PR</Text>
                  </View>
                </View>
                <Text style={styles.prValue}>{pr.estimated_one_rep_max.toFixed(1)} kg</Text>
                <Text style={styles.prDetails}>
                  {pr.weight}kg × {pr.reps} reps
                  {pr.rpe && ` @ RPE ${pr.rpe}`}
                </Text>
              </View>
            ))}
          </View>
        </View>
      )}

      {/* Empty State */}
      {(!historyData || historyData.history.length === 0) && !isLoading && (
        <View style={styles.emptyState}>
          <Text style={styles.emptyIcon}>📈</Text>
          <Text style={styles.emptyTitle}>No History for {selectedExercise}</Text>
          <Text style={styles.emptyText}>
            Start logging sets for this exercise to track your 1RM progression over time.
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
    marginBottom: 24,
  },
  section: {
    marginBottom: 24,
  },
  sectionTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#FFF',
    marginBottom: 12,
  },
  exerciseButton: {
    paddingHorizontal: 16,
    paddingVertical: 10,
    borderRadius: 8,
    backgroundColor: '#1A1A1A',
    borderWidth: 1,
    borderColor: '#2A2A2A',
    marginRight: 8,
  },
  exerciseButtonActive: {
    backgroundColor: '#007AFF',
    borderColor: '#007AFF',
  },
  exerciseButtonText: {
    color: '#999',
    fontSize: 14,
    fontWeight: '600',
  },
  exerciseButtonTextActive: {
    color: '#FFF',
  },
  dateRangeContainer: {
    flexDirection: 'row',
    gap: 8,
  },
  dateButton: {
    flex: 1,
    paddingVertical: 8,
    borderRadius: 8,
    backgroundColor: '#1A1A1A',
    borderWidth: 1,
    borderColor: '#2A2A2A',
    alignItems: 'center',
  },
  dateButtonActive: {
    backgroundColor: '#007AFF',
    borderColor: '#007AFF',
  },
  dateButtonText: {
    color: '#999',
    fontSize: 14,
    fontWeight: '600',
  },
  dateButtonTextActive: {
    color: '#FFF',
  },
  statsContainer: {
    flexDirection: 'row',
    gap: 12,
    marginBottom: 24,
  },
  statCard: {
    flex: 1,
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  statLabel: {
    fontSize: 12,
    color: '#999',
    marginBottom: 8,
  },
  statValue: {
    fontSize: 20,
    fontWeight: '700',
    color: '#FFF',
  },
  statPositive: {
    color: '#34C759',
  },
  statNegative: {
    color: '#FF3B30',
  },
  chartContainer: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  prList: {
    gap: 12,
  },
  prCard: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  prHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  prDate: {
    fontSize: 14,
    color: '#999',
  },
  prBadge: {
    backgroundColor: '#34C759',
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 8,
  },
  prBadgeText: {
    fontSize: 12,
    fontWeight: '700',
    color: '#FFF',
  },
  prValue: {
    fontSize: 24,
    fontWeight: '700',
    color: '#007AFF',
    marginBottom: 4,
  },
  prDetails: {
    fontSize: 14,
    color: '#CCC',
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
});

export default ExerciseHistoryScreen;
