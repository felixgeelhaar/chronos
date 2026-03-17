import React, { Component, ErrorInfo, ReactNode } from 'react';
import { View, Text, StyleSheet, TouchableOpacity, ScrollView } from 'react-native';

interface Props {
  children: ReactNode;
  fallback?: (error: Error, errorInfo: ErrorInfo, resetError: () => void) => ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
  errorInfo: ErrorInfo | null;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
      errorInfo: null,
    };
  }

  static getDerivedStateFromError(error: Error): Partial<State> {
    return {
      hasError: true,
      error,
    };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    console.error('ErrorBoundary caught an error:', error, errorInfo);

    // Log to crash reporting service if analytics enabled
    this.logErrorToService(error, errorInfo);

    this.setState({
      error,
      errorInfo,
    });
  }

  private logErrorToService(error: Error, errorInfo: ErrorInfo): void {
    // In production, this would send to a crash reporting service like Sentry
    // For now, we just log to console
    console.error('Error logged to service:', {
      message: error.message,
      stack: error.stack,
      componentStack: errorInfo.componentStack,
    });
  }

  resetError = (): void => {
    this.setState({
      hasError: false,
      error: null,
      errorInfo: null,
    });
  };

  render(): ReactNode {
    const { hasError, error, errorInfo } = this.state;
    const { children, fallback } = this.props;

    if (hasError && error && errorInfo) {
      if (fallback) {
        return fallback(error, errorInfo, this.resetError);
      }

      return (
        <View style={styles.container}>
          <View style={styles.content}>
            <Text style={styles.icon}>⚠️</Text>
            <Text style={styles.title}>Something went wrong</Text>
            <Text style={styles.message}>
              We encountered an unexpected error. This has been logged and we'll look into it.
            </Text>

            {__DEV__ && (
              <ScrollView style={styles.errorDetails}>
                <Text style={styles.errorTitle}>Error Details (Development Only):</Text>
                <Text style={styles.errorText}>
                  {error.toString()}
                  {'\n\n'}
                  {errorInfo.componentStack}
                </Text>
              </ScrollView>
            )}

            <TouchableOpacity style={styles.button} onPress={this.resetError}>
              <Text style={styles.buttonText}>Try Again</Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={styles.buttonSecondary}
              onPress={() => {
                // In production, this could navigate to home or support
                this.resetError();
              }}
            >
              <Text style={styles.buttonSecondaryText}>Go to Home</Text>
            </TouchableOpacity>
          </View>
        </View>
      );
    }

    return children;
  }
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0A0A0A',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 24,
  },
  content: {
    width: '100%',
    maxWidth: 400,
    alignItems: 'center',
  },
  icon: {
    fontSize: 64,
    marginBottom: 24,
  },
  title: {
    fontSize: 24,
    fontWeight: '700',
    color: '#FFF',
    marginBottom: 16,
    textAlign: 'center',
  },
  message: {
    fontSize: 16,
    color: '#999',
    lineHeight: 24,
    textAlign: 'center',
    marginBottom: 32,
  },
  errorDetails: {
    width: '100%',
    maxHeight: 200,
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    marginBottom: 24,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  errorTitle: {
    fontSize: 14,
    fontWeight: '600',
    color: '#FF9500',
    marginBottom: 12,
  },
  errorText: {
    fontSize: 12,
    color: '#999',
    fontFamily: 'monospace',
    lineHeight: 18,
  },
  button: {
    width: '100%',
    backgroundColor: '#007AFF',
    borderRadius: 12,
    padding: 18,
    alignItems: 'center',
    marginBottom: 12,
  },
  buttonText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#FFF',
  },
  buttonSecondary: {
    width: '100%',
    backgroundColor: '#2A2A2A',
    borderRadius: 12,
    padding: 18,
    alignItems: 'center',
  },
  buttonSecondaryText: {
    fontSize: 16,
    fontWeight: '600',
    color: '#999',
  },
});

export default ErrorBoundary;
