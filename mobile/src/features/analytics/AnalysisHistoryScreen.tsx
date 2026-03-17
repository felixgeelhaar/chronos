import React, { useState, useEffect, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  FlatList,
  TouchableOpacity,
  ActivityIndicator,
  Alert,
  RefreshControl,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { videoService } from '../../services/api';
import { VideoResponse } from '../../services/api/types';
import { format } from 'date-fns';

const EXERCISES = ['All', 'Squat', 'Bench Press', 'Deadlift', 'Overhead Press', 'Barbell Row'];

export const AnalysisHistoryScreen: React.FC = () => {
  const navigation = useNavigation();
  const [videos, setVideos] = useState<VideoResponse[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [selectedExercise, setSelectedExercise] = useState('All');
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);

  useEffect(() => {
    loadVideos(1);
  }, [selectedExercise]);

  const loadVideos = useCallback(async (pageNum: number) => {
    try {
      if (pageNum === 1) {
        setIsLoading(true);
      }

      const response = await videoService.listVideos(pageNum, 20);

      if (pageNum === 1) {
        setVideos(response.videos);
      } else {
        setVideos(prev => [...prev, ...response.videos]);
      }

      setPage(pageNum);
      setHasMore(response.videos.length === 20);
    } catch (error: any) {
      console.error('Failed to load videos:', error);
      Alert.alert('Error', 'Failed to load analysis history');
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  }, []);

  const handleRefresh = useCallback(() => {
    setIsRefreshing(true);
    loadVideos(1);
  }, [loadVideos]);

  const handleLoadMore = useCallback(() => {
    if (!isLoading && hasMore) {
      loadVideos(page + 1);
    }
  }, [isLoading, hasMore, page, loadVideos]);

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return '#34C759';
      case 'processing':
        return '#FF9500';
      case 'failed':
        return '#FF3B30';
      default:
        return '#999';
    }
  };

  const getStatusLabel = (status: string) => {
    switch (status) {
      case 'completed':
        return 'Analyzed';
      case 'processing':
        return 'Processing';
      case 'failed':
        return 'Failed';
      case 'pending':
        return 'Pending';
      default:
        return status;
    }
  };

  const renderVideoCard = ({ item }: { item: VideoResponse }) => (
    <TouchableOpacity
      style={styles.videoCard}
      onPress={() => {
        // Navigate to video detail or play video
        // For now, we'll just show an alert
        Alert.alert('Video', `Video ID: ${item.id}`);
      }}
    >
      <View style={styles.videoCardHeader}>
        <View style={styles.videoInfo}>
          <Text style={styles.videoTitle} numberOfLines={1}>
            {item.filename}
          </Text>
          <Text style={styles.videoDate}>
            {format(new Date(item.created_at), 'MMM d, yyyy • h:mm a')}
          </Text>
        </View>
        <View
          style={[
            styles.statusBadge,
            { backgroundColor: getStatusColor(item.status) },
          ]}
        >
          <Text style={styles.statusText}>{getStatusLabel(item.status)}</Text>
        </View>
      </View>

      {item.status === 'completed' && item.analysis_result && (
        <View style={styles.analysisPreview}>
          {item.analysis_result.reps_detected !== undefined && (
            <View style={styles.analysisStat}>
              <Text style={styles.analysisStatValue}>
                {item.analysis_result.reps_detected}
              </Text>
              <Text style={styles.analysisStatLabel}>Reps</Text>
            </View>
          )}
          {item.analysis_result.form_feedback && item.analysis_result.form_feedback.length > 0 && (
            <View style={styles.analysisStat}>
              <Text style={styles.analysisStatValue}>
                {item.analysis_result.form_feedback.length}
              </Text>
              <Text style={styles.analysisStatLabel}>Issues</Text>
            </View>
          )}
          {item.analysis_result.confidence_score !== undefined && (
            <View style={styles.analysisStat}>
              <Text style={styles.analysisStatValue}>
                {Math.round(item.analysis_result.confidence_score * 100)}%
              </Text>
              <Text style={styles.analysisStatLabel}>Confidence</Text>
            </View>
          )}
        </View>
      )}

      {item.status === 'processing' && (
        <View style={styles.processingContainer}>
          <ActivityIndicator size="small" color="#FF9500" />
          <Text style={styles.processingText}>Analyzing video...</Text>
        </View>
      )}

      <View style={styles.videoCardFooter}>
        <Text style={styles.videoSize}>
          {(item.file_size / (1024 * 1024)).toFixed(1)} MB
        </Text>
        {item.duration && (
          <Text style={styles.videoDuration}>
            {Math.floor(item.duration)}s
          </Text>
        )}
      </View>
    </TouchableOpacity>
  );

  const renderEmpty = () => (
    <View style={styles.emptyContainer}>
      <Text style={styles.emptyIcon}>📊</Text>
      <Text style={styles.emptyTitle}>No Analysis History</Text>
      <Text style={styles.emptyText}>
        Upload videos of your sets to get form analysis and track your technique improvements.
      </Text>
    </View>
  );

  const renderFooter = () => {
    if (!isLoading) return null;
    return (
      <View style={styles.footerLoader}>
        <ActivityIndicator size="small" color="#007AFF" />
      </View>
    );
  };

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>Analysis History</Text>
      </View>

      {/* Exercise Filter */}
      <View style={styles.filterContainer}>
        <FlatList
          horizontal
          showsHorizontalScrollIndicator={false}
          data={EXERCISES}
          keyExtractor={item => item}
          renderItem={({ item }) => (
            <TouchableOpacity
              style={[
                styles.filterButton,
                selectedExercise === item && styles.filterButtonActive,
              ]}
              onPress={() => setSelectedExercise(item)}
            >
              <Text
                style={[
                  styles.filterButtonText,
                  selectedExercise === item && styles.filterButtonTextActive,
                ]}
              >
                {item}
              </Text>
            </TouchableOpacity>
          )}
        />
      </View>

      {isLoading && videos.length === 0 ? (
        <View style={styles.loadingContainer}>
          <ActivityIndicator size="large" color="#007AFF" />
        </View>
      ) : (
        <FlatList
          data={videos}
          keyExtractor={item => item.id}
          renderItem={renderVideoCard}
          ListEmptyComponent={renderEmpty}
          ListFooterComponent={renderFooter}
          onEndReached={handleLoadMore}
          onEndReachedThreshold={0.5}
          refreshControl={
            <RefreshControl
              refreshing={isRefreshing}
              onRefresh={handleRefresh}
              tintColor="#007AFF"
            />
          }
          contentContainerStyle={[
            styles.listContent,
            videos.length === 0 && styles.listContentEmpty,
          ]}
        />
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
    padding: 24,
    paddingBottom: 16,
  },
  title: {
    fontSize: 28,
    fontWeight: '700',
    color: '#FFF',
  },
  filterContainer: {
    paddingHorizontal: 24,
    paddingBottom: 16,
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
    fontSize: 14,
    fontWeight: '600',
    color: '#999',
  },
  filterButtonTextActive: {
    color: '#FFF',
  },
  listContent: {
    padding: 24,
    paddingTop: 0,
  },
  listContentEmpty: {
    flex: 1,
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  videoCard: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  videoCardHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: 12,
  },
  videoInfo: {
    flex: 1,
    marginRight: 12,
  },
  videoTitle: {
    fontSize: 15,
    fontWeight: '600',
    color: '#FFF',
    marginBottom: 4,
  },
  videoDate: {
    fontSize: 12,
    color: '#999',
  },
  statusBadge: {
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: 8,
  },
  statusText: {
    fontSize: 11,
    fontWeight: '700',
    color: '#FFF',
  },
  analysisPreview: {
    flexDirection: 'row',
    gap: 16,
    marginBottom: 12,
  },
  analysisStat: {
    alignItems: 'center',
  },
  analysisStatValue: {
    fontSize: 20,
    fontWeight: '700',
    color: '#007AFF',
    marginBottom: 4,
  },
  analysisStatLabel: {
    fontSize: 11,
    color: '#999',
  },
  processingContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: 8,
  },
  processingText: {
    fontSize: 13,
    color: '#FF9500',
    marginLeft: 8,
  },
  videoCardFooter: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingTop: 12,
    borderTopWidth: 1,
    borderTopColor: '#2A2A2A',
  },
  videoSize: {
    fontSize: 12,
    color: '#666',
  },
  videoDuration: {
    fontSize: 12,
    color: '#666',
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 48,
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
    lineHeight: 20,
  },
  footerLoader: {
    paddingVertical: 20,
    alignItems: 'center',
  },
});

export default AnalysisHistoryScreen;
