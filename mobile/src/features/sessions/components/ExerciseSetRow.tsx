import React from 'react';
import { View, Text, TextInput, TouchableOpacity, StyleSheet } from 'react-native';

export interface SetData {
  id: string;
  exercise_name: string;
  weight: string;
  reps: string;
  rpe: string;
  set_order: number;
}

interface ExerciseSetRowProps {
  set: SetData;
  setNumber: number;
  onUpdate: (id: string, field: keyof SetData, value: string) => void;
  onRemove: (id: string) => void;
  isLastSet: boolean;
}

export const ExerciseSetRow: React.FC<ExerciseSetRowProps> = ({
  set,
  setNumber,
  onUpdate,
  onRemove,
  isLastSet,
}) => {
  return (
    <View style={styles.container}>
      <View style={styles.setNumber}>
        <Text style={styles.setNumberText}>{setNumber}</Text>
      </View>

      <View style={styles.inputs}>
        <View style={styles.inputGroup}>
          <Text style={styles.inputLabel}>Weight (kg)</Text>
          <TextInput
            style={styles.input}
            value={set.weight}
            onChangeText={(value) => onUpdate(set.id, 'weight', value)}
            keyboardType="decimal-pad"
            placeholder="0"
            placeholderTextColor="#666"
          />
        </View>

        <View style={styles.inputGroup}>
          <Text style={styles.inputLabel}>Reps</Text>
          <TextInput
            style={styles.input}
            value={set.reps}
            onChangeText={(value) => onUpdate(set.id, 'reps', value)}
            keyboardType="number-pad"
            placeholder="0"
            placeholderTextColor="#666"
          />
        </View>

        <View style={styles.inputGroup}>
          <Text style={styles.inputLabel}>RPE</Text>
          <TextInput
            style={styles.input}
            value={set.rpe}
            onChangeText={(value) => onUpdate(set.id, 'rpe', value)}
            keyboardType="decimal-pad"
            placeholder="-"
            placeholderTextColor="#666"
          />
        </View>
      </View>

      <TouchableOpacity
        style={styles.removeButton}
        onPress={() => onRemove(set.id)}
      >
        <Text style={styles.removeButtonText}>×</Text>
      </TouchableOpacity>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 12,
    marginBottom: 8,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  setNumber: {
    width: 32,
    height: 32,
    borderRadius: 16,
    backgroundColor: '#2A2A2A',
    justifyContent: 'center',
    alignItems: 'center',
    marginRight: 12,
  },
  setNumberText: {
    color: '#999',
    fontSize: 14,
    fontWeight: '600',
  },
  inputs: {
    flex: 1,
    flexDirection: 'row',
    gap: 8,
  },
  inputGroup: {
    flex: 1,
  },
  inputLabel: {
    fontSize: 10,
    color: '#666',
    marginBottom: 4,
    textAlign: 'center',
  },
  input: {
    backgroundColor: '#0A0A0A',
    borderRadius: 8,
    padding: 8,
    fontSize: 16,
    color: '#FFF',
    textAlign: 'center',
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  removeButton: {
    width: 32,
    height: 32,
    justifyContent: 'center',
    alignItems: 'center',
    marginLeft: 8,
  },
  removeButtonText: {
    color: '#FF4444',
    fontSize: 28,
    fontWeight: '300',
  },
});

export default ExerciseSetRow;
