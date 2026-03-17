import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  ActivityIndicator,
} from 'react-native';
import NetInfo from '@react-native-community/netinfo';
import { syncEngine, SyncStatus } from '../../services/sync';

interface SyncStatusIndicatorProps {
  showDetails?: boolean;
  onPress?: () => void;
}

export const SyncStatusIndicator: React.FC<SyncStatusIndicatorProps> = ({
  showDetails = false,
  onPress,
}) => {
  const [syncStatus, setSyncStatus] = useState<SyncStatus>({
    isSyncing: false,
    pendingChanges: 0,
    errors: [],
  });
  const [isOnline, setIsOnline] = useState(true);

  useEffect(() => {
    // Subscribe to sync status changes
    const unsubscribe = syncEngine.subscribe(status => {
      setSyncStatus(status);
    });

    // Subscribe to network status
    const netInfoUnsubscribe = NetInfo.addEventListener(state => {
      setIsOnline(state.isConnected ?? false);
    });

    // Load initial status
    syncEngine.getStatus().then(setSyncStatus);
    NetInfo.fetch().then(state => setIsOnline(state.isConnected ?? false));

    return () => {
      unsubscribe();
      netInfoUnsubscribe();
    };
  }, []);

  const handlePress = () => {
    if (onPress) {
      onPress();
    } else {
      // Trigger manual sync
      syncEngine.sync();
    }
  };

  // Determine status color and icon
  const getStatusDisplay = () => {
    if (!isOnline) {
      return {
        color: '#FF9500',
        icon: '📴',
        label: 'Offline',
      };
    }

    if (syncStatus.isSyncing) {
      return {
        color: '#007AFF',
        icon: null,
        label: 'Syncing...',
      };
    }

    if (syncStatus.errors.length > 0) {
      return {
        color: '#FF3B30',
        icon: '⚠️',
        label: 'Sync Error',
      };
    }

    if (syncStatus.pendingChanges > 0) {
      return {
        color: '#FF9500',
        icon: '⏳',
        label: `${syncStatus.pendingChanges} pending`,
      };
    }

    return {
      color: '#34C759',
      icon: '✓',
      label: 'Synced',
    };
  };

  const status = getStatusDisplay();

  if (!showDetails) {
    // Compact view
    return (
      <TouchableOpacity
        style={styles.compactContainer}
        onPress={handlePress}
        activeOpacity={0.7}
      >
        {syncStatus.isSyncing ? (
          <ActivityIndicator size="small" color={status.color} />
        ) : (
          <Text style={[styles.icon, { color: status.color }]}>{status.icon}</Text>
        )}
      </TouchableOpacity>
    );
  }

  // Detailed view
  return (
    <TouchableOpacity
      style={[styles.detailedContainer, { borderColor: status.color }]}
      onPress={handlePress}
      activeOpacity={0.7}
    >
      <View style={styles.detailedContent}>
        <View style={styles.statusRow}>
          {syncStatus.isSyncing ? (
            <ActivityIndicator size="small" color={status.color} />
          ) : (
            <Text style={styles.statusIcon}>{status.icon}</Text>
          )}
          <Text style={[styles.statusLabel, { color: status.color }]}>
            {status.label}
          </Text>
        </View>

        {syncStatus.lastSyncedAt && !syncStatus.isSyncing && (
          <Text style={styles.timestamp}>
            Last synced: {syncStatus.lastSyncedAt.toLocaleTimeString()}
          </Text>
        )}

        {syncStatus.errors.length > 0 && (
          <Text style={styles.errorText} numberOfLines={1}>
            {syncStatus.errors[0]}
          </Text>
        )}
      </View>
    </TouchableOpacity>
  );
};

const styles = StyleSheet.create({
  compactContainer: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: '#1A1A1A',
    justifyContent: 'center',
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  icon: {
    fontSize: 16,
  },
  detailedContainer: {
    backgroundColor: '#1A1A1A',
    borderRadius: 8,
    padding: 12,
    borderWidth: 1,
  },
  detailedContent: {
    gap: 4,
  },
  statusRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
  },
  statusIcon: {
    fontSize: 16,
  },
  statusLabel: {
    fontSize: 13,
    fontWeight: '600',
  },
  timestamp: {
    fontSize: 11,
    color: '#999',
  },
  errorText: {
    fontSize: 11,
    color: '#FF3B30',
  },
});

export default SyncStatusIndicator;
