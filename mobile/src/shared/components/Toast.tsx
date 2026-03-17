import React, { useEffect, useRef } from 'react';
import {
  View,
  Text,
  StyleSheet,
  Animated,
  Dimensions,
  TouchableOpacity,
} from 'react-native';
import { haptics } from '../utils/haptics';

export type ToastType = 'success' | 'error' | 'warning' | 'info';

interface ToastProps {
  message: string;
  type?: ToastType;
  duration?: number;
  onDismiss?: () => void;
  action?: {
    label: string;
    onPress: () => void;
  };
}

export const Toast: React.FC<ToastProps> = ({
  message,
  type = 'info',
  duration = 3000,
  onDismiss,
  action,
}) => {
  const translateY = useRef(new Animated.Value(100)).current;
  const opacity = useRef(new Animated.Value(0)).current;

  useEffect(() => {
    // Trigger haptic feedback based on type
    switch (type) {
      case 'success':
        haptics.success();
        break;
      case 'error':
        haptics.error();
        break;
      case 'warning':
        haptics.warning();
        break;
      default:
        haptics.light();
    }

    // Slide in animation
    Animated.parallel([
      Animated.spring(translateY, {
        toValue: 0,
        tension: 65,
        friction: 10,
        useNativeDriver: true,
      }),
      Animated.timing(opacity, {
        toValue: 1,
        duration: 250,
        useNativeDriver: true,
      }),
    ]).start();

    // Auto dismiss after duration
    const timer = setTimeout(() => {
      dismiss();
    }, duration);

    return () => clearTimeout(timer);
  }, []);

  const dismiss = () => {
    Animated.parallel([
      Animated.timing(translateY, {
        toValue: 100,
        duration: 200,
        useNativeDriver: true,
      }),
      Animated.timing(opacity, {
        toValue: 0,
        duration: 200,
        useNativeDriver: true,
      }),
    ]).start(() => {
      onDismiss?.();
    });
  };

  const getBackgroundColor = () => {
    switch (type) {
      case 'success':
        return '#34C759';
      case 'error':
        return '#FF3B30';
      case 'warning':
        return '#FF9500';
      case 'info':
        return '#007AFF';
      default:
        return '#1A1A1A';
    }
  };

  const getIcon = () => {
    switch (type) {
      case 'success':
        return '✓';
      case 'error':
        return '✕';
      case 'warning':
        return '⚠';
      case 'info':
        return 'ℹ';
      default:
        return '';
    }
  };

  return (
    <Animated.View
      style={[
        styles.container,
        {
          backgroundColor: getBackgroundColor(),
          transform: [{ translateY }],
          opacity,
        },
      ]}
    >
      <View style={styles.content}>
        <Text style={styles.icon}>{getIcon()}</Text>
        <Text style={styles.message}>{message}</Text>
      </View>
      {action && (
        <TouchableOpacity
          style={styles.actionButton}
          onPress={() => {
            action.onPress();
            dismiss();
          }}
        >
          <Text style={styles.actionText}>{action.label}</Text>
        </TouchableOpacity>
      )}
      <TouchableOpacity style={styles.closeButton} onPress={dismiss}>
        <Text style={styles.closeIcon}>✕</Text>
      </TouchableOpacity>
    </Animated.View>
  );
};

const styles = StyleSheet.create({
  container: {
    position: 'absolute',
    bottom: 20,
    left: 20,
    right: 20,
    borderRadius: 12,
    padding: 16,
    flexDirection: 'row',
    alignItems: 'center',
    shadowColor: '#000',
    shadowOffset: {
      width: 0,
      height: 4,
    },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 8,
  },
  content: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
  },
  icon: {
    fontSize: 20,
    color: '#FFF',
    marginRight: 12,
  },
  message: {
    flex: 1,
    fontSize: 15,
    color: '#FFF',
    fontWeight: '500',
    lineHeight: 20,
  },
  actionButton: {
    marginLeft: 12,
    paddingHorizontal: 12,
    paddingVertical: 6,
    backgroundColor: 'rgba(255, 255, 255, 0.2)',
    borderRadius: 6,
  },
  actionText: {
    fontSize: 14,
    color: '#FFF',
    fontWeight: '600',
  },
  closeButton: {
    marginLeft: 12,
    padding: 4,
  },
  closeIcon: {
    fontSize: 18,
    color: '#FFF',
    opacity: 0.8,
  },
});

// Toast manager for showing toasts globally
class ToastManager {
  private listener: ((props: ToastProps) => void) | null = null;

  setListener(listener: (props: ToastProps) => void) {
    this.listener = listener;
  }

  removeListener() {
    this.listener = null;
  }

  show(props: ToastProps) {
    if (this.listener) {
      this.listener(props);
    }
  }

  success(message: string, duration?: number) {
    this.show({ message, type: 'success', duration });
  }

  error(message: string, duration?: number) {
    this.show({ message, type: 'error', duration });
  }

  warning(message: string, duration?: number) {
    this.show({ message, type: 'warning', duration });
  }

  info(message: string, duration?: number) {
    this.show({ message, type: 'info', duration });
  }
}

export const toast = new ToastManager();

export default Toast;
