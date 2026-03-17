import React, { useState, useRef } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  ActivityIndicator,
  Dimensions,
} from 'react-native';
import Video, { VideoRef } from 'react-native-video';

interface VideoPlayerProps {
  videoId: string;
  videoUrl: string;
  thumbnailUrl?: string;
  autoPlay?: boolean;
  controls?: boolean;
  height?: number;
}

export const VideoPlayer: React.FC<VideoPlayerProps> = ({
  videoId,
  videoUrl,
  thumbnailUrl,
  autoPlay = false,
  controls = true,
  height = 300,
}) => {
  const videoRef = useRef<VideoRef>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isPaused, setIsPaused] = useState(!autoPlay);
  const [hasError, setHasError] = useState(false);
  const [currentTime, setCurrentTime] = useState(0);
  const [duration, setDuration] = useState(0);
  const screenWidth = Dimensions.get('window').width;

  const handleLoad = (data: any) => {
    setIsLoading(false);
    setDuration(data.duration);
  };

  const handleError = (error: any) => {
    console.error('Video playback error:', error);
    setHasError(true);
    setIsLoading(false);
  };

  const handleProgress = (data: any) => {
    setCurrentTime(data.currentTime);
  };

  const togglePlayPause = () => {
    setIsPaused(!isPaused);
  };

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  if (hasError) {
    return (
      <View style={[styles.container, { height }]}>
        <View style={styles.errorContainer}>
          <Text style={styles.errorIcon}>⚠️</Text>
          <Text style={styles.errorText}>Failed to load video</Text>
        </View>
      </View>
    );
  }

  return (
    <View style={[styles.container, { height }]}>
      <Video
        ref={videoRef}
        source={{ uri: videoUrl }}
        style={[styles.video, { height }]}
        paused={isPaused}
        resizeMode="contain"
        onLoad={handleLoad}
        onError={handleError}
        onProgress={handleProgress}
        repeat={false}
      />

      {isLoading && (
        <View style={styles.loadingOverlay}>
          <ActivityIndicator size="large" color="#FFF" />
        </View>
      )}

      {controls && !isLoading && (
        <View style={styles.controls}>
          <TouchableOpacity
            style={styles.playButton}
            onPress={togglePlayPause}
          >
            <Text style={styles.playButtonText}>
              {isPaused ? '▶' : '❚❚'}
            </Text>
          </TouchableOpacity>

          <View style={styles.timeContainer}>
            <Text style={styles.timeText}>
              {formatTime(currentTime)} / {formatTime(duration)}
            </Text>
          </View>

          <View style={styles.progressBarContainer}>
            <View
              style={[
                styles.progressBar,
                { width: `${duration > 0 ? (currentTime / duration) * 100 : 0}%` },
              ]}
            />
          </View>
        </View>
      )}
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#000',
    borderRadius: 12,
    overflow: 'hidden',
    position: 'relative',
  },
  video: {
    width: '100%',
    backgroundColor: '#000',
  },
  loadingOverlay: {
    ...StyleSheet.absoluteFillObject,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: 'rgba(0, 0, 0, 0.5)',
  },
  controls: {
    position: 'absolute',
    bottom: 0,
    left: 0,
    right: 0,
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    padding: 12,
  },
  playButton: {
    width: 40,
    height: 40,
    borderRadius: 20,
    backgroundColor: 'rgba(255, 255, 255, 0.3)',
    justifyContent: 'center',
    alignItems: 'center',
    marginBottom: 8,
  },
  playButtonText: {
    color: '#FFF',
    fontSize: 16,
    fontWeight: '700',
  },
  timeContainer: {
    marginBottom: 8,
  },
  timeText: {
    color: '#FFF',
    fontSize: 12,
    fontWeight: '600',
  },
  progressBarContainer: {
    height: 4,
    backgroundColor: 'rgba(255, 255, 255, 0.3)',
    borderRadius: 2,
    overflow: 'hidden',
  },
  progressBar: {
    height: '100%',
    backgroundColor: '#007AFF',
  },
  errorContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#1A1A1A',
  },
  errorIcon: {
    fontSize: 48,
    marginBottom: 12,
  },
  errorText: {
    fontSize: 14,
    color: '#999',
  },
});

export default VideoPlayer;
