import React, { useState, useEffect, ReactNode } from 'react';
import { View, StyleSheet } from 'react-native';
import Toast, { toast, ToastType } from './Toast';

interface ToastConfig {
  id: string;
  message: string;
  type: ToastType;
  duration?: number;
  action?: {
    label: string;
    onPress: () => void;
  };
}

interface ToastProviderProps {
  children: ReactNode;
}

export const ToastProvider: React.FC<ToastProviderProps> = ({ children }) => {
  const [toasts, setToasts] = useState<ToastConfig[]>([]);

  useEffect(() => {
    toast.setListener((props) => {
      const id = Date.now().toString();
      setToasts((prev) => [
        ...prev,
        {
          id,
          message: props.message,
          type: props.type || 'info',
          duration: props.duration,
          action: props.action,
        },
      ]);
    });

    return () => {
      toast.removeListener();
    };
  }, []);

  const dismissToast = (id: string) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  };

  return (
    <>
      {children}
      <View style={styles.toastContainer} pointerEvents="box-none">
        {toasts.map((toastConfig) => (
          <Toast
            key={toastConfig.id}
            message={toastConfig.message}
            type={toastConfig.type}
            duration={toastConfig.duration}
            action={toastConfig.action}
            onDismiss={() => dismissToast(toastConfig.id)}
          />
        ))}
      </View>
    </>
  );
};

const styles = StyleSheet.create({
  toastContainer: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
  },
});

export default ToastProvider;
