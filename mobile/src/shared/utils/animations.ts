import { Animated, Easing } from 'react-native';

/**
 * Animation utility functions for smooth, performant animations
 */

export const animations = {
  /**
   * Fade in animation
   */
  fadeIn: (
    animatedValue: Animated.Value,
    duration: number = 300,
    delay: number = 0
  ): Animated.CompositeAnimation => {
    return Animated.timing(animatedValue, {
      toValue: 1,
      duration,
      delay,
      easing: Easing.out(Easing.ease),
      useNativeDriver: true,
    });
  },

  /**
   * Fade out animation
   */
  fadeOut: (
    animatedValue: Animated.Value,
    duration: number = 300,
    delay: number = 0
  ): Animated.CompositeAnimation => {
    return Animated.timing(animatedValue, {
      toValue: 0,
      duration,
      delay,
      easing: Easing.in(Easing.ease),
      useNativeDriver: true,
    });
  },

  /**
   * Slide in from bottom
   */
  slideInFromBottom: (
    animatedValue: Animated.Value,
    duration: number = 400
  ): Animated.CompositeAnimation => {
    return Animated.spring(animatedValue, {
      toValue: 0,
      tension: 65,
      friction: 10,
      useNativeDriver: true,
    });
  },

  /**
   * Slide out to bottom
   */
  slideOutToBottom: (
    animatedValue: Animated.Value,
    distance: number = 300,
    duration: number = 300
  ): Animated.CompositeAnimation => {
    return Animated.timing(animatedValue, {
      toValue: distance,
      duration,
      easing: Easing.in(Easing.ease),
      useNativeDriver: true,
    });
  },

  /**
   * Scale in animation (pop effect)
   */
  scaleIn: (
    animatedValue: Animated.Value,
    duration: number = 300
  ): Animated.CompositeAnimation => {
    return Animated.spring(animatedValue, {
      toValue: 1,
      tension: 80,
      friction: 7,
      useNativeDriver: true,
    });
  },

  /**
   * Scale out animation
   */
  scaleOut: (
    animatedValue: Animated.Value,
    duration: number = 200
  ): Animated.CompositeAnimation => {
    return Animated.timing(animatedValue, {
      toValue: 0,
      duration,
      easing: Easing.in(Easing.ease),
      useNativeDriver: true,
    });
  },

  /**
   * Bounce animation
   */
  bounce: (animatedValue: Animated.Value): Animated.CompositeAnimation => {
    return Animated.sequence([
      Animated.timing(animatedValue, {
        toValue: 1.1,
        duration: 100,
        easing: Easing.out(Easing.ease),
        useNativeDriver: true,
      }),
      Animated.timing(animatedValue, {
        toValue: 1,
        duration: 100,
        easing: Easing.in(Easing.ease),
        useNativeDriver: true,
      }),
    ]);
  },

  /**
   * Shake animation (for errors)
   */
  shake: (animatedValue: Animated.Value): Animated.CompositeAnimation => {
    return Animated.sequence([
      Animated.timing(animatedValue, {
        toValue: 10,
        duration: 50,
        useNativeDriver: true,
      }),
      Animated.timing(animatedValue, {
        toValue: -10,
        duration: 50,
        useNativeDriver: true,
      }),
      Animated.timing(animatedValue, {
        toValue: 10,
        duration: 50,
        useNativeDriver: true,
      }),
      Animated.timing(animatedValue, {
        toValue: 0,
        duration: 50,
        useNativeDriver: true,
      }),
    ]);
  },

  /**
   * Pulse animation (continuous)
   */
  pulse: (animatedValue: Animated.Value): Animated.CompositeAnimation => {
    return Animated.loop(
      Animated.sequence([
        Animated.timing(animatedValue, {
          toValue: 1.05,
          duration: 500,
          easing: Easing.inOut(Easing.ease),
          useNativeDriver: true,
        }),
        Animated.timing(animatedValue, {
          toValue: 1,
          duration: 500,
          easing: Easing.inOut(Easing.ease),
          useNativeDriver: true,
        }),
      ])
    );
  },

  /**
   * Rotate animation (continuous)
   */
  rotate: (animatedValue: Animated.Value): Animated.CompositeAnimation => {
    return Animated.loop(
      Animated.timing(animatedValue, {
        toValue: 1,
        duration: 1000,
        easing: Easing.linear,
        useNativeDriver: true,
      })
    );
  },

  /**
   * Staggered animation for list items
   */
  stagger: (
    animations: Animated.CompositeAnimation[],
    delay: number = 50
  ): Animated.CompositeAnimation => {
    return Animated.stagger(delay, animations);
  },

  /**
   * Parallel animations
   */
  parallel: (animations: Animated.CompositeAnimation[]): Animated.CompositeAnimation => {
    return Animated.parallel(animations);
  },

  /**
   * Sequential animations
   */
  sequence: (animations: Animated.CompositeAnimation[]): Animated.CompositeAnimation => {
    return Animated.sequence(animations);
  },
};

/**
 * Common easing functions
 */
export const easings = {
  easeIn: Easing.in(Easing.ease),
  easeOut: Easing.out(Easing.ease),
  easeInOut: Easing.inOut(Easing.ease),
  linear: Easing.linear,
  bounce: Easing.bounce,
  elastic: Easing.elastic(1),
  bezier: Easing.bezier(0.25, 0.1, 0.25, 1),
};

/**
 * Animation presets for common use cases
 */
export const animationPresets = {
  modal: {
    duration: 300,
    easing: easings.easeOut,
  },
  toast: {
    duration: 250,
    easing: easings.easeInOut,
  },
  button: {
    duration: 150,
    easing: easings.easeOut,
  },
  page: {
    duration: 400,
    easing: easings.easeInOut,
  },
};

export default animations;
