import React, { useState } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Alert,
  Modal,
  ActivityIndicator,
  Platform,
} from 'react-native';
import { launchCamera, launchImageLibrary, Asset } from 'react-native-image-picker';

export interface VideoFile {
  uri: string;
  type: string;
  name: string;
  fileSize?: number;
  duration?: number;
}

interface VideoPickerProps {
  onVideoSelected: (video: VideoFile) => void;
  onCancel?: () => void;
  maxDuration?: number; // in seconds
  visible: boolean;
  onClose: () => void;
}

export const VideoPicker: React.FC<VideoPickerProps> = ({
  onVideoSelected,
  onCancel,
  maxDuration = 60,
  visible,
  onClose,
}) => {
  const [isProcessing, setIsProcessing] = useState(false);

  const handleVideoSelected = async (asset: Asset) => {
    if (!asset.uri || !asset.fileName) {
      Alert.alert('Error', 'Invalid video file');
      return;
    }

    // Check duration if available
    if (asset.duration && asset.duration > maxDuration) {
      Alert.alert(
        'Video Too Long',
        `Please select a video shorter than ${maxDuration} seconds. This video is ${Math.round(asset.duration)} seconds.`
      );
      return;
    }

    // Check file size (limit to 100MB)
    const maxSize = 100 * 1024 * 1024; // 100MB
    if (asset.fileSize && asset.fileSize > maxSize) {
      Alert.alert(
        'File Too Large',
        'Please select a video smaller than 100MB.'
      );
      return;
    }

    const videoFile: VideoFile = {
      uri: asset.uri,
      type: asset.type || 'video/mp4',
      name: asset.fileName,
      fileSize: asset.fileSize,
      duration: asset.duration,
    };

    onVideoSelected(videoFile);
    onClose();
  };

  const handleRecordVideo = async () => {
    setIsProcessing(true);
    try {
      const result = await launchCamera({
        mediaType: 'video',
        videoQuality: 'high',
        durationLimit: maxDuration,
        saveToPhotos: true,
      });

      if (result.didCancel) {
        onCancel?.();
      } else if (result.errorCode) {
        Alert.alert('Error', result.errorMessage || 'Failed to record video');
      } else if (result.assets && result.assets[0]) {
        await handleVideoSelected(result.assets[0]);
      }
    } catch (error: any) {
      console.error('Error recording video:', error);
      Alert.alert('Error', 'Failed to record video');
    } finally {
      setIsProcessing(false);
    }
  };

  const handleSelectFromLibrary = async () => {
    setIsProcessing(true);
    try {
      const result = await launchImageLibrary({
        mediaType: 'video',
        videoQuality: 'high',
      });

      if (result.didCancel) {
        onCancel?.();
      } else if (result.errorCode) {
        Alert.alert('Error', result.errorMessage || 'Failed to select video');
      } else if (result.assets && result.assets[0]) {
        await handleVideoSelected(result.assets[0]);
      }
    } catch (error: any) {
      console.error('Error selecting video:', error);
      Alert.alert('Error', 'Failed to select video');
    } finally {
      setIsProcessing(false);
    }
  };

  return (
    <Modal
      visible={visible}
      transparent
      animationType="fade"
      onRequestClose={onClose}
    >
      <View style={styles.overlay}>
        <View style={styles.modalContainer}>
          <Text style={styles.title}>Add Video</Text>
          <Text style={styles.subtitle}>
            Record or select a video of your set (max {maxDuration}s)
          </Text>

          {isProcessing ? (
            <View style={styles.loadingContainer}>
              <ActivityIndicator size="large" color="#007AFF" />
              <Text style={styles.loadingText}>Processing...</Text>
            </View>
          ) : (
            <>
              <TouchableOpacity
                style={styles.option}
                onPress={handleRecordVideo}
                disabled={isProcessing}
              >
                <Text style={styles.optionIcon}>📹</Text>
                <View style={styles.optionContent}>
                  <Text style={styles.optionTitle}>Record Video</Text>
                  <Text style={styles.optionDescription}>
                    Record a new video with your camera
                  </Text>
                </View>
              </TouchableOpacity>

              <TouchableOpacity
                style={styles.option}
                onPress={handleSelectFromLibrary}
                disabled={isProcessing}
              >
                <Text style={styles.optionIcon}>🎬</Text>
                <View style={styles.optionContent}>
                  <Text style={styles.optionTitle}>Choose from Library</Text>
                  <Text style={styles.optionDescription}>
                    Select an existing video
                  </Text>
                </View>
              </TouchableOpacity>

              <TouchableOpacity
                style={styles.cancelButton}
                onPress={onClose}
              >
                <Text style={styles.cancelButtonText}>Cancel</Text>
              </TouchableOpacity>
            </>
          )}
        </View>
      </View>
    </Modal>
  );
};

const styles = StyleSheet.create({
  overlay: {
    flex: 1,
    backgroundColor: 'rgba(0, 0, 0, 0.7)',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 20,
  },
  modalContainer: {
    backgroundColor: '#1A1A1A',
    borderRadius: 16,
    padding: 24,
    width: '100%',
    maxWidth: 400,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  title: {
    fontSize: 24,
    fontWeight: '700',
    color: '#FFF',
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 14,
    color: '#999',
    marginBottom: 24,
  },
  option: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#2A2A2A',
    borderRadius: 12,
    padding: 16,
    marginBottom: 12,
  },
  optionIcon: {
    fontSize: 32,
    marginRight: 16,
  },
  optionContent: {
    flex: 1,
  },
  optionTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#FFF',
    marginBottom: 4,
  },
  optionDescription: {
    fontSize: 13,
    color: '#999',
  },
  cancelButton: {
    backgroundColor: 'transparent',
    borderWidth: 1,
    borderColor: '#2A2A2A',
    borderRadius: 12,
    padding: 16,
    alignItems: 'center',
    marginTop: 8,
  },
  cancelButtonText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#999',
  },
  loadingContainer: {
    alignItems: 'center',
    paddingVertical: 32,
  },
  loadingText: {
    fontSize: 14,
    color: '#999',
    marginTop: 12,
  },
});

export default VideoPicker;
