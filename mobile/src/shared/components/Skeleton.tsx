import React, { useEffect, useRef } from 'react';
import { View, StyleSheet, Animated, ViewStyle } from 'react-native';

interface SkeletonProps {
  width?: number | string;
  height?: number;
  borderRadius?: number;
  style?: ViewStyle;
}

export const Skeleton: React.FC<SkeletonProps> = ({
  width = '100%',
  height = 20,
  borderRadius = 4,
  style,
}) => {
  const opacity = useRef(new Animated.Value(0.3)).current;

  useEffect(() => {
    Animated.loop(
      Animated.sequence([
        Animated.timing(opacity, {
          toValue: 1,
          duration: 800,
          useNativeDriver: true,
        }),
        Animated.timing(opacity, {
          toValue: 0.3,
          duration: 800,
          useNativeDriver: true,
        }),
      ])
    ).start();
  }, [opacity]);

  return (
    <Animated.View
      style={[
        styles.skeleton,
        {
          width,
          height,
          borderRadius,
          opacity,
        },
        style,
      ]}
    />
  );
};

export const SessionListSkeleton: React.FC = () => (
  <View style={styles.sessionCard}>
    <View style={styles.sessionHeader}>
      <Skeleton width={120} height={20} />
      <Skeleton width={80} height={16} />
    </View>
    <View style={styles.sessionStats}>
      <Skeleton width={60} height={14} />
      <Skeleton width={60} height={14} />
      <Skeleton width={60} height={14} />
    </View>
    <View style={styles.sessionExercises}>
      <Skeleton width="100%" height={16} style={styles.exerciseLine} />
      <Skeleton width="80%" height={16} style={styles.exerciseLine} />
    </View>
  </View>
);

export const ExerciseCardSkeleton: React.FC = () => (
  <View style={styles.exerciseCard}>
    <View style={styles.exerciseHeader}>
      <Skeleton width={150} height={20} />
      <Skeleton width={40} height={16} />
    </View>
    <View style={styles.exerciseSets}>
      {[1, 2, 3].map((key) => (
        <View key={key} style={styles.setRow}>
          <Skeleton width={40} height={14} />
          <Skeleton width={60} height={14} />
          <Skeleton width={60} height={14} />
        </View>
      ))}
    </View>
  </View>
);

export const AnalyticsChartSkeleton: React.FC = () => (
  <View style={styles.chartCard}>
    <Skeleton width={180} height={24} style={styles.chartTitle} />
    <View style={styles.chartContainer}>
      <View style={styles.chartBars}>
        {[1, 2, 3, 4, 5, 6, 7].map((key) => (
          <Skeleton
            key={key}
            width={30}
            height={Math.random() * 100 + 50}
            style={styles.chartBar}
          />
        ))}
      </View>
    </View>
  </View>
);

const styles = StyleSheet.create({
  skeleton: {
    backgroundColor: '#2A2A2A',
  },
  sessionCard: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  sessionHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  sessionStats: {
    flexDirection: 'row',
    gap: 16,
    marginBottom: 12,
  },
  sessionExercises: {
    gap: 8,
  },
  exerciseLine: {
    marginBottom: 4,
  },
  exerciseCard: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
  },
  exerciseHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 16,
  },
  exerciseSets: {
    gap: 12,
  },
  setRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingHorizontal: 8,
  },
  chartCard: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 20,
    marginBottom: 16,
  },
  chartTitle: {
    marginBottom: 20,
  },
  chartContainer: {
    height: 200,
  },
  chartBars: {
    flexDirection: 'row',
    justifyContent: 'space-around',
    alignItems: 'flex-end',
    height: '100%',
  },
  chartBar: {
    alignSelf: 'flex-end',
  },
});

export default Skeleton;
