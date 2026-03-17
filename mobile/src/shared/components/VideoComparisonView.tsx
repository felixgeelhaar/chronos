import React, { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Dimensions,
} from 'react-native';
import { VideoPlayer } from './VideoPlayer';

interface VideoComparisonViewProps {
  video1: {
    id: string;
    url: string;
    label: string;
    date?: string;
  };
  video2: {
    id: string;
    url: string;
    label: string;
    date?: string;
  };
  defaultMode?: 'side-by-side' | 'stacked';
}

export const VideoComparisonView: React.FC<VideoComparisonViewProps> = ({
  video1,
  video2,
  defaultMode = 'side-by-side',
}) => {
  const [mode, setMode] = useState<'side-by-side' | 'stacked'>(defaultMode);
  const screenWidth = Dimensions.get('window').width;

  const videoHeight = mode === 'side-by-side' ? 220 : 280;
  const videoWidth = mode === 'side-by-side' ? (screenWidth - 56) / 2 : screenWidth - 48;

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>Compare Videos</Text>
        <View style={styles.modeToggle}>
          <TouchableOpacity
            style={[
              styles.modeButton,
              mode === 'side-by-side' && styles.modeButtonActive,
            ]}
            onPress={() => setMode('side-by-side')}
          >
            <Text
              style={[
                styles.modeButtonText,
                mode === 'side-by-side' && styles.modeButtonTextActive,
              ]}
            >
              Side by Side
            </Text>
          </TouchableOpacity>
          <TouchableOpacity
            style={[
              styles.modeButton,
              mode === 'stacked' && styles.modeButtonActive,
            ]}
            onPress={() => setMode('stacked')}
          >
            <Text
              style={[
                styles.modeButtonText,
                mode === 'stacked' && styles.modeButtonTextActive,
              ]}
            >
              Stacked
            </Text>
          </TouchableOpacity>
        </View>
      </View>

      <View
        style={[
          styles.videosContainer,
          mode === 'stacked' && styles.videosContainerStacked,
        ]}
      >
        <View style={[styles.videoWrapper, { width: videoWidth }]}>
          <View style={styles.videoLabel}>
            <Text style={styles.videoLabelText}>{video1.label}</Text>
            {video1.date && (
              <Text style={styles.videoDate}>{video1.date}</Text>
            )}
          </View>
          <VideoPlayer
            videoId={video1.id}
            videoUrl={video1.url}
            autoPlay={false}
            controls={true}
            height={videoHeight}
          />
        </View>

        <View style={[styles.videoWrapper, { width: videoWidth }]}>
          <View style={styles.videoLabel}>
            <Text style={styles.videoLabelText}>{video2.label}</Text>
            {video2.date && (
              <Text style={styles.videoDate}>{video2.date}</Text>
            )}
          </View>
          <VideoPlayer
            videoId={video2.id}
            videoUrl={video2.url}
            autoPlay={false}
            controls={true}
            height={videoHeight}
          />
        </View>
      </View>

      <View style={styles.instructions}>
        <Text style={styles.instructionsText}>
          💡 Compare your form and technique between sets or over time
        </Text>
      </View>
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
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 16,
  },
  title: {
    fontSize: 16,
    fontWeight: '600',
    color: '#FFF',
  },
  modeToggle: {
    flexDirection: 'row',
    backgroundColor: '#2A2A2A',
    borderRadius: 8,
    padding: 2,
  },
  modeButton: {
    paddingHorizontal: 12,
    paddingVertical: 6,
    borderRadius: 6,
  },
  modeButtonActive: {
    backgroundColor: '#007AFF',
  },
  modeButtonText: {
    fontSize: 11,
    fontWeight: '600',
    color: '#999',
  },
  modeButtonTextActive: {
    color: '#FFF',
  },
  videosContainer: {
    flexDirection: 'row',
    gap: 8,
    marginBottom: 12,
  },
  videosContainerStacked: {
    flexDirection: 'column',
    gap: 16,
  },
  videoWrapper: {
    flex: 1,
  },
  videoLabel: {
    marginBottom: 8,
  },
  videoLabelText: {
    fontSize: 13,
    fontWeight: '600',
    color: '#FFF',
    marginBottom: 2,
  },
  videoDate: {
    fontSize: 11,
    color: '#999',
  },
  instructions: {
    backgroundColor: '#2A2A2A',
    borderRadius: 8,
    padding: 12,
  },
  instructionsText: {
    fontSize: 12,
    color: '#999',
    lineHeight: 16,
  },
});

export default VideoComparisonView;
