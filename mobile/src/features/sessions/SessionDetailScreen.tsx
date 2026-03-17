import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  ScrollView,
  TouchableOpacity,
  StyleSheet,
  ActivityIndicator,
  Alert,
} from 'react-native';
import { format } from 'date-fns';
import { sessionService, videoService, SessionResponse, SetResponse } from '../../services/api';
import { VideoPlayer, VideoAnalysisDisplay } from '../../shared/components';

interface SessionDetailScreenProps {
  navigation: any;
  route: {
    params: {
      sessionId: string;
    };
  };
}

export const SessionDetailScreen: React.FC<SessionDetailScreenProps> = ({
  navigation,
  route,
}) => {
  const { sessionId } = route.params;
  const [session, setSession] = useState<SessionResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    loadSession();
  }, [sessionId]);

  const loadSession = async () => {
    try {
      setIsLoading(true);
      const data = await sessionService.getSession(sessionId);
      setSession(data);
    } catch (error) {
      console.error('Failed to load session:', error);
      Alert.alert('Error', 'Failed to load workout details');
      navigation.goBack();
    } finally {
      setIsLoading(false);
    }
  };

  const handleDelete = () => {
    Alert.alert(
      'Delete Workout',
      'Are you sure you want to delete this workout? This action cannot be undone.',
      [
        { text: 'Cancel', style: 'cancel' },
        {
          text: 'Delete',
          style: 'destructive',
          onPress: async () => {
            try {
              await sessionService.deleteSession(sessionId);
              navigation.goBack();
            } catch (error) {
              Alert.alert('Error', 'Failed to delete workout');
            }
          },
        },
      ]
    );
  };

  if (isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="large" color="#007AFF" />
      </View>
    );
  }

  if (!session) {
    return null;
  }

  // Group sets by exercise
  const groupedSets = session.sets.reduce((acc, set) => {
    if (!acc[set.exercise_name]) {
      acc[set.exercise_name] = [];
    }
    acc[set.exercise_name].push(set);
    return acc;
  }, {} as Record<string, SetResponse[]>);

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <TouchableOpacity onPress={() => navigation.goBack()}>
          <Text style={styles.backButton}>← Back</Text>
        </TouchableOpacity>
        <TouchableOpacity onPress={handleDelete}>
          <Text style={styles.deleteButton}>Delete</Text>
        </TouchableOpacity>
      </View>

      <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
        <View style={styles.titleSection}>
          <Text style={styles.date}>
            {format(new Date(session.date), 'EEEE, MMMM d, yyyy')}
          </Text>
        </View>

        <View style={styles.statsSection}>
          <View style={styles.statCard}>
            <Text style={styles.statLabel}>Total Volume</Text>
            <Text style={styles.statValue}>{session.total_volume.toFixed(0)} kg</Text>
          </View>
          <View style={styles.statCard}>
            <Text style={styles.statLabel}>Total Sets</Text>
            <Text style={styles.statValue}>{session.total_sets}</Text>
          </View>
          <View style={styles.statCard}>
            <Text style={styles.statLabel}>Exercises</Text>
            <Text style={styles.statValue}>{Object.keys(groupedSets).length}</Text>
          </View>
        </View>

        {session.notes && (
          <View style={styles.notesSection}>
            <Text style={styles.sectionTitle}>Notes</Text>
            <Text style={styles.notesText}>{session.notes}</Text>
          </View>
        )}

        <View style={styles.exercisesSection}>
          <Text style={styles.sectionTitle}>Exercises</Text>
          {Object.entries(groupedSets).map(([exerciseName, sets]) => {
            const totalVolume = sets.reduce((sum, set) => sum + set.volume, 0);
            const maxWeight = Math.max(...sets.map(s => s.weight));

            return (
              <View key={exerciseName} style={styles.exerciseCard}>
                <View style={styles.exerciseHeader}>
                  <Text style={styles.exerciseName}>{exerciseName}</Text>
                  <View style={styles.exerciseStats}>
                    <Text style={styles.exerciseStatText}>
                      {sets.length} sets • {totalVolume.toFixed(0)} kg
                    </Text>
                  </View>
                </View>

                <View style={styles.setHeader}>
                  <Text style={[styles.setHeaderText, { flex: 0.5 }]}>Set</Text>
                  <Text style={[styles.setHeaderText, { flex: 1 }]}>Weight</Text>
                  <Text style={[styles.setHeaderText, { flex: 1 }]}>Reps</Text>
                  <Text style={[styles.setHeaderText, { flex: 1 }]}>RPE</Text>
                  <Text style={[styles.setHeaderText, { flex: 1 }]}>Volume</Text>
                </View>

                {sets.map((set, index) => (
                  <View key={set.id}>
                    <View style={styles.setRow}>
                      <Text style={[styles.setCell, { flex: 0.5 }]}>{index + 1}</Text>
                      <Text style={[styles.setCell, { flex: 1 }]}>{set.weight} kg</Text>
                      <Text style={[styles.setCell, { flex: 1 }]}>{set.reps}</Text>
                      <Text style={[styles.setCell, { flex: 1 }]}>
                        {set.rpe ? set.rpe.toFixed(1) : '-'}
                      </Text>
                      <Text style={[styles.setCell, { flex: 1 }]}>
                        {set.volume.toFixed(0)} kg
                      </Text>
                    </View>
                    {set.video_id && (
                      <View style={styles.videoContainer}>
                        <VideoPlayer
                          videoId={set.video_id}
                          videoUrl={videoService.getVideoUrl(set.video_id)}
                          thumbnailUrl={videoService.getThumbnailUrl(set.video_id)}
                          autoPlay={false}
                          controls={true}
                          height={220}
                        />
                        <View style={styles.analysisContainer}>
                          <VideoAnalysisDisplay
                            videoId={set.video_id}
                            compact={true}
                          />
                        </View>
                      </View>
                    )}
                  </View>
                ))}

                <View style={styles.exerciseFooter}>
                  <Text style={styles.exerciseFooterText}>
                    Max Weight: {maxWeight} kg • Est. 1RM: {Math.max(...sets.map(s => s.estimated_one_rep_max)).toFixed(1)} kg
                  </Text>
                </View>
              </View>
            );
          })}
        </View>

        <View style={{ height: 40 }} />
      </ScrollView>
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
    padding: 16,
    paddingTop: 60,
    borderBottomWidth: 1,
    borderBottomColor: '#1A1A1A',
  },
  backButton: {
    fontSize: 16,
    color: '#007AFF',
  },
  deleteButton: {
    fontSize: 16,
    color: '#FF4444',
  },
  content: {
    flex: 1,
  },
  titleSection: {
    padding: 24,
  },
  date: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#FFF',
  },
  statsSection: {
    flexDirection: 'row',
    paddingHorizontal: 24,
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
    marginBottom: 4,
  },
  statValue: {
    fontSize: 20,
    fontWeight: '600',
    color: '#FFF',
  },
  notesSection: {
    padding: 24,
    paddingTop: 0,
  },
  sectionTitle: {
    fontSize: 14,
    fontWeight: '600',
    color: '#999',
    marginBottom: 12,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  notesText: {
    fontSize: 16,
    color: '#FFF',
    lineHeight: 24,
  },
  exercisesSection: {
    padding: 24,
    paddingTop: 0,
  },
  exerciseCard: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    marginBottom: 16,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  exerciseHeader: {
    marginBottom: 16,
  },
  exerciseName: {
    fontSize: 18,
    fontWeight: '600',
    color: '#FFF',
    marginBottom: 4,
  },
  exerciseStats: {
    flexDirection: 'row',
    alignItems: 'center',
  },
  exerciseStatText: {
    fontSize: 14,
    color: '#999',
  },
  setHeader: {
    flexDirection: 'row',
    paddingBottom: 8,
    borderBottomWidth: 1,
    borderBottomColor: '#2A2A2A',
    marginBottom: 8,
  },
  setHeaderText: {
    fontSize: 12,
    color: '#666',
    fontWeight: '600',
  },
  setRow: {
    flexDirection: 'row',
    paddingVertical: 8,
  },
  setCell: {
    fontSize: 14,
    color: '#FFF',
  },
  videoContainer: {
    marginTop: 12,
    marginBottom: 8,
  },
  analysisContainer: {
    marginTop: 12,
  },
  exerciseFooter: {
    marginTop: 12,
    paddingTop: 12,
    borderTopWidth: 1,
    borderTopColor: '#2A2A2A',
  },
  exerciseFooterText: {
    fontSize: 12,
    color: '#999',
  },
});

export default SessionDetailScreen;
