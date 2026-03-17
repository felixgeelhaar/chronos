import React, { useState, useCallback } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Alert,
  ActivityIndicator,
} from 'react-native';
import { videoService } from '../../services/api';
import { VideoFile, VideoPicker } from './VideoPicker';
import { VideoUploadResponse } from '../../services/api/types';

interface VideoUploadProps {
  setId?: string;
  onUploadComplete?: (video: VideoUploadResponse) => void;
  onUploadError?: (error: Error) => void;
  buttonText?: string;
  showProgress?: boolean;
}

export const VideoUpload: React.FC<VideoUploadProps> = ({
  setId,
  onUploadComplete,
  onUploadError,
  buttonText = 'Add Video',
  showProgress = true,
}) => {
  const [showPicker, setShowPicker] = useState(false);
  const [isUploading, setIsUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [selectedVideo, setSelectedVideo] = useState<VideoFile | null>(null);

  const handleVideoSelected = useCallback(async (video: VideoFile) => {
    setSelectedVideo(video);
    setIsUploading(true);
    setUploadProgress(0);

    try {
      const response = await videoService.uploadVideo(
        {
          uri: video.uri,
          type: video.type,
          name: video.name,
        },
        setId,
        (progress) => {
          setUploadProgress(progress);
        }
      );

      setIsUploading(false);
      setUploadProgress(100);
      onUploadComplete?.(response);

      // Show success message
      Alert.alert(
        'Upload Complete',
        'Your video has been uploaded and is being processed.'
      );

      // Reset after delay
      setTimeout(() => {
        setSelectedVideo(null);
        setUploadProgress(0);
      }, 2000);
    } catch (error: any) {
      console.error('Upload error:', error);
      setIsUploading(false);
      setUploadProgress(0);

      const errorMessage = error.response?.data?.error || 'Failed to upload video';
      Alert.alert('Upload Failed', errorMessage);

      onUploadError?.(error);
      setSelectedVideo(null);
    }
  }, [setId, onUploadComplete, onUploadError]);

  const handleCancel = useCallback(() => {
    if (isUploading) {
      Alert.alert(
        'Cancel Upload',
        'Are you sure you want to cancel the upload?',
        [
          { text: 'Continue Upload', style: 'cancel' },
          {
            text: 'Cancel',
            style: 'destructive',
            onPress: () => {
              setIsUploading(false);
              setUploadProgress(0);
              setSelectedVideo(null);
            },
          },
        ]
      );
    }
  }, [isUploading]);

  if (isUploading) {
    return (
      <View style={styles.uploadingContainer}>
        <View style={styles.uploadHeader}>
          <Text style={styles.uploadTitle}>Uploading Video...</Text>
          {showProgress && (
            <Text style={styles.uploadPercentage}>{uploadProgress}%</Text>
          )}
        </View>

        {selectedVideo && (
          <Text style={styles.fileName} numberOfLines={1}>
            {selectedVideo.name}
          </Text>
        )}

        {showProgress && (
          <View style={styles.progressBarContainer}>
            <View
              style={[
                styles.progressBar,
                { width: `${uploadProgress}%` },
              ]}
            />
          </View>
        )}

        <TouchableOpacity
          style={styles.cancelUploadButton}
          onPress={handleCancel}
        >
          <Text style={styles.cancelUploadText}>Cancel</Text>
        </TouchableOpacity>
      </View>
    );
  }

  if (uploadProgress === 100) {
    return (
      <View style={styles.successContainer}>
        <Text style={styles.successIcon}>✓</Text>
        <Text style={styles.successText}>Upload Complete</Text>
      </View>
    );
  }

  return (
    <>
      <TouchableOpacity
        style={styles.addButton}
        onPress={() => setShowPicker(true)}
      >
        <Text style={styles.addButtonIcon}>📹</Text>
        <Text style={styles.addButtonText}>{buttonText}</Text>
      </TouchableOpacity>

      <VideoPicker
        visible={showPicker}
        onClose={() => setShowPicker(false)}
        onVideoSelected={handleVideoSelected}
        onCancel={() => setShowPicker(false)}
        maxDuration={60}
      />
    </>
  );
};

const styles = StyleSheet.create({
  addButton: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#2A2A2A',
    borderWidth: 1,
    borderColor: '#3A3A3A',
    borderRadius: 8,
    padding: 12,
    borderStyle: 'dashed',
  },
  addButtonIcon: {
    fontSize: 20,
    marginRight: 8,
  },
  addButtonText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#007AFF',
  },
  uploadingContainer: {
    backgroundColor: '#1A1A1A',
    borderRadius: 8,
    padding: 16,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  uploadHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  uploadTitle: {
    fontSize: 14,
    fontWeight: '600',
    color: '#FFF',
  },
  uploadPercentage: {
    fontSize: 14,
    fontWeight: '700',
    color: '#007AFF',
  },
  fileName: {
    fontSize: 12,
    color: '#999',
    marginBottom: 12,
  },
  progressBarContainer: {
    height: 4,
    backgroundColor: '#2A2A2A',
    borderRadius: 2,
    overflow: 'hidden',
    marginBottom: 12,
  },
  progressBar: {
    height: '100%',
    backgroundColor: '#007AFF',
    borderRadius: 2,
  },
  cancelUploadButton: {
    alignItems: 'center',
    padding: 8,
  },
  cancelUploadText: {
    fontSize: 12,
    fontWeight: '600',
    color: '#FF3B30',
  },
  successContainer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#1A4D2E',
    borderRadius: 8,
    padding: 12,
    borderWidth: 1,
    borderColor: '#34C759',
  },
  successIcon: {
    fontSize: 20,
    marginRight: 8,
    color: '#34C759',
  },
  successText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#34C759',
  },
});

export default VideoUpload;
