import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ActivityIndicator,
  TouchableOpacity,
  ScrollView,
} from 'react-native';
import { videoService } from '../../services/api';
import { VideoAnalysisResponse } from '../../services/api/types';

interface VideoAnalysisDisplayProps {
  videoId: string;
  compact?: boolean;
  onAnalysisLoaded?: (analysis: VideoAnalysisResponse) => void;
}

export const VideoAnalysisDisplay: React.FC<VideoAnalysisDisplayProps> = ({
  videoId,
  compact = false,
  onAnalysisLoaded,
}) => {
  const [analysis, setAnalysis] = useState<VideoAnalysisResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isExpanded, setIsExpanded] = useState(!compact);

  useEffect(() => {
    loadAnalysis();
  }, [videoId]);

  const loadAnalysis = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const data = await videoService.getVideoAnalysis(videoId);
      setAnalysis(data);
      onAnalysisLoaded?.(data);
    } catch (err: any) {
      console.error('Failed to load video analysis:', err);
      setError(err.response?.data?.error || 'Failed to load analysis');
    } finally {
      setIsLoading(false);
    }
  };

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
        return 'Completed';
      case 'processing':
        return 'Processing...';
      case 'failed':
        return 'Failed';
      case 'pending':
        return 'Pending';
      default:
        return status;
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return '#FF3B30';
      case 'warning':
        return '#FF9500';
      case 'info':
        return '#007AFF';
      default:
        return '#999';
    }
  };

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'critical':
        return '🔴';
      case 'warning':
        return '⚠️';
      case 'info':
        return 'ℹ️';
      default:
        return '•';
    }
  };

  if (isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <ActivityIndicator size="small" color="#007AFF" />
        <Text style={styles.loadingText}>Analyzing video...</Text>
      </View>
    );
  }

  if (error) {
    return (
      <View style={styles.errorContainer}>
        <Text style={styles.errorIcon}>⚠️</Text>
        <Text style={styles.errorText}>{error}</Text>
        <TouchableOpacity style={styles.retryButton} onPress={loadAnalysis}>
          <Text style={styles.retryButtonText}>Retry</Text>
        </TouchableOpacity>
      </View>
    );
  }

  if (!analysis) {
    return null;
  }

  // If status is not completed, show processing state
  if (analysis.status !== 'completed') {
    return (
      <View style={styles.processingContainer}>
        <View style={styles.statusBadge}>
          <View
            style={[
              styles.statusIndicator,
              { backgroundColor: getStatusColor(analysis.status) },
            ]}
          />
          <Text style={styles.statusText}>{getStatusLabel(analysis.status)}</Text>
        </View>
        {analysis.status === 'processing' && (
          <Text style={styles.processingText}>
            Your video is being analyzed. This usually takes 1-2 minutes.
          </Text>
        )}
        {analysis.status === 'failed' && (
          <>
            <Text style={styles.processingText}>
              Analysis failed. Please try again or upload a new video.
            </Text>
            <TouchableOpacity
              style={styles.retryButton}
              onPress={async () => {
                try {
                  await videoService.reprocessVideo(videoId);
                  loadAnalysis();
                } catch (err) {
                  console.error('Failed to reprocess video:', err);
                }
              }}
            >
              <Text style={styles.retryButtonText}>Reprocess Video</Text>
            </TouchableOpacity>
          </>
        )}
      </View>
    );
  }

  // Compact view
  if (compact && !isExpanded) {
    return (
      <TouchableOpacity
        style={styles.compactContainer}
        onPress={() => setIsExpanded(true)}
      >
        <View style={styles.compactHeader}>
          <Text style={styles.compactTitle}>📊 Analysis Results</Text>
          <Text style={styles.expandIcon}>▼</Text>
        </View>
        <View style={styles.compactStats}>
          <View style={styles.compactStat}>
            <Text style={styles.compactStatValue}>{analysis.reps_detected || 0}</Text>
            <Text style={styles.compactStatLabel}>Reps</Text>
          </View>
          {analysis.form_feedback && analysis.form_feedback.length > 0 && (
            <View style={styles.compactStat}>
              <Text style={styles.compactStatValue}>{analysis.form_feedback.length}</Text>
              <Text style={styles.compactStatLabel}>Issues</Text>
            </View>
          )}
          {analysis.confidence_score !== undefined && (
            <View style={styles.compactStat}>
              <Text style={styles.compactStatValue}>
                {Math.round(analysis.confidence_score * 100)}%
              </Text>
              <Text style={styles.compactStatLabel}>Confidence</Text>
            </View>
          )}
        </View>
      </TouchableOpacity>
    );
  }

  return (
    <View style={styles.container}>
      {compact && (
        <TouchableOpacity
          style={styles.collapseButton}
          onPress={() => setIsExpanded(false)}
        >
          <Text style={styles.expandIcon}>▲</Text>
        </TouchableOpacity>
      )}

      <Text style={styles.title}>📊 Video Analysis</Text>

      {/* Reps Detected */}
      <View style={styles.section}>
        <Text style={styles.sectionTitle}>Reps Detected</Text>
        <View style={styles.repsCard}>
          <Text style={styles.repsValue}>{analysis.reps_detected || 0}</Text>
          {analysis.confidence_score !== undefined && (
            <Text style={styles.confidenceText}>
              {Math.round(analysis.confidence_score * 100)}% confidence
            </Text>
          )}
        </View>
      </View>

      {/* Form Feedback */}
      {analysis.form_feedback && analysis.form_feedback.length > 0 && (
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Form Feedback</Text>
          {analysis.form_feedback.map((feedback, index) => (
            <View
              key={index}
              style={[
                styles.feedbackCard,
                { borderLeftColor: getSeverityColor(feedback.severity) },
              ]}
            >
              <View style={styles.feedbackHeader}>
                <Text style={styles.feedbackIcon}>
                  {getSeverityIcon(feedback.severity)}
                </Text>
                <Text style={styles.feedbackType}>{feedback.type}</Text>
                {feedback.timestamp !== undefined && (
                  <Text style={styles.feedbackTimestamp}>
                    @ {feedback.timestamp.toFixed(1)}s
                  </Text>
                )}
              </View>
              <Text style={styles.feedbackMessage}>{feedback.message}</Text>
              {feedback.suggestion && (
                <Text style={styles.feedbackSuggestion}>💡 {feedback.suggestion}</Text>
              )}
            </View>
          ))}
        </View>
      )}

      {/* Metadata */}
      {analysis.metadata && Object.keys(analysis.metadata).length > 0 && (
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Details</Text>
          <View style={styles.metadataContainer}>
            {Object.entries(analysis.metadata).map(([key, value]) => (
              <View key={key} style={styles.metadataRow}>
                <Text style={styles.metadataKey}>
                  {key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}:
                </Text>
                <Text style={styles.metadataValue}>{String(value)}</Text>
              </View>
            ))}
          </View>
        </View>
      )}

      {/* Processed At */}
      {analysis.processed_at && (
        <Text style={styles.processedAt}>
          Analyzed {new Date(analysis.processed_at).toLocaleString()}
        </Text>
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  loadingContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  loadingText: {
    fontSize: 14,
    color: '#999',
    marginLeft: 12,
  },
  errorContainer: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#FF3B30',
    alignItems: 'center',
  },
  errorIcon: {
    fontSize: 32,
    marginBottom: 8,
  },
  errorText: {
    fontSize: 14,
    color: '#FF3B30',
    textAlign: 'center',
    marginBottom: 12,
  },
  retryButton: {
    backgroundColor: '#2A2A2A',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 8,
  },
  retryButtonText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#007AFF',
  },
  processingContainer: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  statusBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 12,
  },
  statusIndicator: {
    width: 8,
    height: 8,
    borderRadius: 4,
    marginRight: 8,
  },
  statusText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#FFF',
  },
  processingText: {
    fontSize: 13,
    color: '#999',
    lineHeight: 20,
  },
  compactContainer: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 12,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  compactHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 12,
  },
  compactTitle: {
    fontSize: 14,
    fontWeight: '600',
    color: '#FFF',
  },
  expandIcon: {
    fontSize: 12,
    color: '#999',
  },
  compactStats: {
    flexDirection: 'row',
    gap: 16,
  },
  compactStat: {
    alignItems: 'center',
  },
  compactStatValue: {
    fontSize: 18,
    fontWeight: '700',
    color: '#007AFF',
    marginBottom: 4,
  },
  compactStatLabel: {
    fontSize: 11,
    color: '#999',
  },
  collapseButton: {
    alignSelf: 'flex-end',
    padding: 4,
  },
  title: {
    fontSize: 16,
    fontWeight: '600',
    color: '#FFF',
    marginBottom: 16,
  },
  section: {
    marginBottom: 16,
  },
  sectionTitle: {
    fontSize: 12,
    fontWeight: '600',
    color: '#999',
    textTransform: 'uppercase',
    marginBottom: 8,
    letterSpacing: 0.5,
  },
  repsCard: {
    backgroundColor: '#2A2A2A',
    borderRadius: 8,
    padding: 16,
    alignItems: 'center',
  },
  repsValue: {
    fontSize: 36,
    fontWeight: '700',
    color: '#007AFF',
    marginBottom: 4,
  },
  confidenceText: {
    fontSize: 12,
    color: '#999',
  },
  feedbackCard: {
    backgroundColor: '#2A2A2A',
    borderRadius: 8,
    padding: 12,
    marginBottom: 8,
    borderLeftWidth: 3,
  },
  feedbackHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: 8,
  },
  feedbackIcon: {
    fontSize: 16,
    marginRight: 8,
  },
  feedbackType: {
    fontSize: 13,
    fontWeight: '600',
    color: '#FFF',
    flex: 1,
    textTransform: 'capitalize',
  },
  feedbackTimestamp: {
    fontSize: 11,
    color: '#666',
  },
  feedbackMessage: {
    fontSize: 13,
    color: '#CCC',
    lineHeight: 18,
    marginBottom: 6,
  },
  feedbackSuggestion: {
    fontSize: 12,
    color: '#999',
    fontStyle: 'italic',
    lineHeight: 16,
  },
  metadataContainer: {
    backgroundColor: '#2A2A2A',
    borderRadius: 8,
    padding: 12,
  },
  metadataRow: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    paddingVertical: 4,
  },
  metadataKey: {
    fontSize: 12,
    color: '#999',
  },
  metadataValue: {
    fontSize: 12,
    color: '#FFF',
    fontWeight: '500',
  },
  processedAt: {
    fontSize: 11,
    color: '#666',
    textAlign: 'center',
    marginTop: 8,
  },
});

export default VideoAnalysisDisplay;
