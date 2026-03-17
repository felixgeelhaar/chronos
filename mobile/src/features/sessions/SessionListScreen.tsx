import React, { useState, useEffect, useCallback } from 'react';
import {
  View,
  Text,
  FlatList,
  TouchableOpacity,
  StyleSheet,
  RefreshControl,
  ActivityIndicator,
} from 'react-native';
import { format } from 'date-fns';
import { sessionService, SessionResponse } from '../../services/api';

interface SessionListScreenProps {
  navigation: any;
}

export const SessionListScreen: React.FC<SessionListScreenProps> = ({ navigation }) => {
  const [sessions, setSessions] = useState<SessionResponse[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [page, setPage] = useState(1);
  const [hasMore, setHasMore] = useState(true);

  useEffect(() => {
    loadSessions();
  }, []);

  const loadSessions = async (pageNum: number = 1) => {
    try {
      if (pageNum === 1) {
        setIsLoading(true);
      }

      const response = await sessionService.listSessions(pageNum, 20);

      if (pageNum === 1) {
        setSessions(response.sessions);
      } else {
        setSessions(prev => [...prev, ...response.sessions]);
      }

      setHasMore(pageNum < response.total_pages);
      setPage(pageNum);
    } catch (error) {
      console.error('Failed to load sessions:', error);
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  };

  const handleRefresh = useCallback(() => {
    setIsRefreshing(true);
    loadSessions(1);
  }, []);

  const handleLoadMore = useCallback(() => {
    if (!isLoading && hasMore) {
      loadSessions(page + 1);
    }
  }, [isLoading, hasMore, page]);

  const renderSession = ({ item }: { item: SessionResponse }) => (
    <TouchableOpacity
      style={styles.sessionCard}
      onPress={() => navigation.navigate('SessionDetail', { sessionId: item.id })}
    >
      <View style={styles.sessionHeader}>
        <Text style={styles.sessionDate}>
          {format(new Date(item.date), 'EEEE, MMMM d, yyyy')}
        </Text>
        <View style={styles.sessionBadge}>
          <Text style={styles.sessionBadgeText}>{item.total_sets} sets</Text>
        </View>
      </View>

      <View style={styles.sessionStats}>
        <View style={styles.stat}>
          <Text style={styles.statLabel}>Volume</Text>
          <Text style={styles.statValue}>{item.total_volume.toFixed(0)} kg</Text>
        </View>
        <View style={styles.stat}>
          <Text style={styles.statLabel}>Exercises</Text>
          <Text style={styles.statValue}>
            {new Set(item.sets.map(s => s.exercise_name)).size}
          </Text>
        </View>
      </View>

      {item.notes && (
        <Text style={styles.sessionNotes} numberOfLines={2}>
          {item.notes}
        </Text>
      )}
    </TouchableOpacity>
  );

  const renderEmpty = () => (
    <View style={styles.emptyContainer}>
      <Text style={styles.emptyTitle}>No Sessions Yet</Text>
      <Text style={styles.emptyText}>
        Start logging your workouts to track your progress
      </Text>
      <TouchableOpacity
        style={styles.createButton}
        onPress={() => navigation.navigate('CreateSession')}
      >
        <Text style={styles.createButtonText}>Log Your First Workout</Text>
      </TouchableOpacity>
    </View>
  );

  const renderFooter = () => {
    if (!isLoading || page === 1) return null;

    return (
      <View style={styles.loadingFooter}>
        <ActivityIndicator color="#007AFF" />
      </View>
    );
  };

  if (isLoading && page === 1) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#007AFF" />
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>Workouts</Text>
        <TouchableOpacity
          style={styles.addButton}
          onPress={() => navigation.navigate('CreateSession')}
        >
          <Text style={styles.addButtonText}>+ New</Text>
        </TouchableOpacity>
      </View>

      <FlatList
        data={sessions}
        renderItem={renderSession}
        keyExtractor={(item) => item.id}
        contentContainerStyle={sessions.length === 0 ? styles.emptyList : styles.list}
        ListEmptyComponent={renderEmpty}
        ListFooterComponent={renderFooter}
        refreshControl={
          <RefreshControl
            refreshing={isRefreshing}
            onRefresh={handleRefresh}
            tintColor="#007AFF"
          />
        }
        onEndReached={handleLoadMore}
        onEndReachedThreshold={0.5}
      />
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0A0A0A',
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#0A0A0A',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 24,
    paddingTop: 60,
  },
  title: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#FFF',
  },
  addButton: {
    backgroundColor: '#007AFF',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 8,
  },
  addButtonText: {
    color: '#FFF',
    fontSize: 16,
    fontWeight: '600',
  },
  list: {
    padding: 24,
    paddingTop: 0,
  },
  emptyList: {
    flexGrow: 1,
  },
  sessionCard: {
    backgroundColor: '#1A1A1A',
    borderRadius: 16,
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
  sessionDate: {
    fontSize: 16,
    fontWeight: '600',
    color: '#FFF',
    flex: 1,
  },
  sessionBadge: {
    backgroundColor: '#2A2A2A',
    paddingHorizontal: 12,
    paddingVertical: 4,
    borderRadius: 12,
  },
  sessionBadgeText: {
    color: '#999',
    fontSize: 12,
    fontWeight: '600',
  },
  sessionStats: {
    flexDirection: 'row',
    gap: 24,
  },
  stat: {
    flex: 1,
  },
  statLabel: {
    fontSize: 12,
    color: '#999',
    marginBottom: 4,
  },
  statValue: {
    fontSize: 18,
    fontWeight: '600',
    color: '#FFF',
  },
  sessionNotes: {
    fontSize: 14,
    color: '#999',
    marginTop: 12,
    lineHeight: 20,
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 40,
  },
  emptyTitle: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#FFF',
    marginBottom: 12,
  },
  emptyText: {
    fontSize: 16,
    color: '#999',
    textAlign: 'center',
    marginBottom: 32,
  },
  createButton: {
    backgroundColor: '#007AFF',
    paddingHorizontal: 24,
    paddingVertical: 12,
    borderRadius: 12,
  },
  createButtonText: {
    color: '#FFF',
    fontSize: 16,
    fontWeight: '600',
  },
  loadingFooter: {
    paddingVertical: 20,
  },
});

export default SessionListScreen;
