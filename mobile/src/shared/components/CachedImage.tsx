import React, { useState } from 'react';
import {
  Image,
  ImageProps,
  ImageStyle,
  StyleProp,
  View,
  StyleSheet,
  ActivityIndicator,
} from 'react-native';
import FastImage, { FastImageProps, Priority, ResizeMode } from 'react-native-fast-image';

interface CachedImageProps extends Omit<FastImageProps, 'source'> {
  source: { uri: string } | number;
  fallbackSource?: number;
  showLoading?: boolean;
  loadingSize?: 'small' | 'large';
  onLoadStart?: () => void;
  onLoadEnd?: () => void;
  onError?: (error: any) => void;
}

/**
 * Optimized image component with caching
 * Uses FastImage for better performance and memory management
 */
export const CachedImage: React.FC<CachedImageProps> = ({
  source,
  fallbackSource,
  showLoading = true,
  loadingSize = 'small',
  style,
  onLoadStart,
  onLoadEnd,
  onError,
  ...props
}) => {
  const [isLoading, setIsLoading] = useState(true);
  const [hasError, setHasError] = useState(false);

  // Handle local images (require() statements)
  if (typeof source === 'number') {
    return <Image source={source} style={style} {...(props as ImageProps)} />;
  }

  const handleLoadStart = () => {
    setIsLoading(true);
    onLoadStart?.();
  };

  const handleLoadEnd = () => {
    setIsLoading(false);
    onLoadEnd?.();
  };

  const handleError = (error: any) => {
    setIsLoading(false);
    setHasError(true);
    onError?.(error);
  };

  // If error and fallback exists, show fallback
  if (hasError && fallbackSource) {
    return <Image source={fallbackSource} style={style} />;
  }

  return (
    <View style={[styles.container, style]}>
      <FastImage
        source={{
          uri: source.uri,
          priority: FastImage.priority.normal,
          cache: FastImage.cacheControl.immutable,
        }}
        style={[styles.image, style]}
        onLoadStart={handleLoadStart}
        onLoadEnd={handleLoadEnd}
        onError={handleError}
        resizeMode={props.resizeMode || FastImage.resizeMode.cover}
        {...props}
      />
      {showLoading && isLoading && (
        <View style={styles.loadingOverlay}>
          <ActivityIndicator size={loadingSize} color="#007AFF" />
        </View>
      )}
    </View>
  );
};

/**
 * Preload images for better UX
 */
export const preloadImages = (uris: string[]): void => {
  const sources = uris.map((uri) => ({
    uri,
    priority: FastImage.priority.low,
    cache: FastImage.cacheControl.immutable,
  }));
  FastImage.preload(sources);
};

/**
 * Clear image cache
 */
export const clearImageCache = async (): Promise<void> => {
  await FastImage.clearMemoryCache();
  await FastImage.clearDiskCache();
};

const styles = StyleSheet.create({
  container: {
    position: 'relative',
    overflow: 'hidden',
  },
  image: {
    width: '100%',
    height: '100%',
  },
  loadingOverlay: {
    ...StyleSheet.absoluteFillObject,
    backgroundColor: 'rgba(0, 0, 0, 0.3)',
    justifyContent: 'center',
    alignItems: 'center',
  },
});

export default CachedImage;
