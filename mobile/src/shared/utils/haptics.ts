import { Platform } from 'react-native';
import ReactNativeHapticFeedback from 'react-native-haptic-feedback';

/**
 * Haptic feedback utility for iOS and Android
 * Provides tactile feedback for user interactions
 */

const options = {
  enableVibrateFallback: true, // Fall back to vibration on Android
  ignoreAndroidSystemSettings: false,
};

export const haptics = {
  /**
   * Light impact feedback - for subtle interactions
   * Use for: list item selections, minor actions
   */
  light: () => {
    ReactNativeHapticFeedback.trigger('impactLight', options);
  },

  /**
   * Medium impact feedback - for standard interactions
   * Use for: button presses, toggles, swipes
   */
  medium: () => {
    ReactNativeHapticFeedback.trigger('impactMedium', options);
  },

  /**
   * Heavy impact feedback - for important interactions
   * Use for: completing actions, confirming destructive operations
   */
  heavy: () => {
    ReactNativeHapticFeedback.trigger('impactHeavy', options);
  },

  /**
   * Success feedback - for successful operations
   * Use for: saved data, completed tasks, successful submissions
   */
  success: () => {
    if (Platform.OS === 'ios') {
      ReactNativeHapticFeedback.trigger('notificationSuccess', options);
    } else {
      ReactNativeHapticFeedback.trigger('impactMedium', options);
    }
  },

  /**
   * Warning feedback - for warnings or cautionary actions
   * Use for: validation errors, important notices
   */
  warning: () => {
    if (Platform.OS === 'ios') {
      ReactNativeHapticFeedback.trigger('notificationWarning', options);
    } else {
      ReactNativeHapticFeedback.trigger('impactLight', options);
    }
  },

  /**
   * Error feedback - for errors and failures
   * Use for: form errors, failed operations, destructive confirmations
   */
  error: () => {
    if (Platform.OS === 'ios') {
      ReactNativeHapticFeedback.trigger('notificationError', options);
    } else {
      ReactNativeHapticFeedback.trigger('impactHeavy', options);
    }
  },

  /**
   * Selection feedback - for UI element selection
   * Use for: picker changes, slider movements, switching tabs
   */
  selection: () => {
    ReactNativeHapticFeedback.trigger('selection', options);
  },

  /**
   * Rigid feedback - for rigid stops and boundaries
   * Use for: reaching min/max values, end of scrollable content
   */
  rigid: () => {
    if (Platform.OS === 'ios') {
      ReactNativeHapticFeedback.trigger('rigid', options);
    } else {
      ReactNativeHapticFeedback.trigger('impactHeavy', options);
    }
  },

  /**
   * Soft feedback - for soft impacts
   * Use for: gentle interactions, subtle state changes
   */
  soft: () => {
    if (Platform.OS === 'ios') {
      ReactNativeHapticFeedback.trigger('soft', options);
    } else {
      ReactNativeHapticFeedback.trigger('impactLight', options);
    }
  },
};

export default haptics;
