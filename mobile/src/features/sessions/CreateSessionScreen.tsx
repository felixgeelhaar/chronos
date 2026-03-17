import React, { useState } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  ScrollView,
  StyleSheet,
  Alert,
  ActivityIndicator,
  KeyboardAvoidingView,
  Platform,
} from 'react-native';
import { format } from 'date-fns';
import { sessionService, CreateSessionRequest, CreateSetInput } from '../../services/api';
import { ExerciseSetRow, SetData } from './components/ExerciseSetRow';

interface CreateSessionScreenProps {
  navigation: any;
}

// Common exercises for quick selection
const COMMON_EXERCISES = [
  'Squat',
  'Bench Press',
  'Deadlift',
  'Overhead Press',
  'Barbell Row',
  'Pull-up',
];

export const CreateSessionScreen: React.FC<CreateSessionScreenProps> = ({ navigation }) => {
  const [date, setDate] = useState(new Date());
  const [notes, setNotes] = useState('');
  const [currentExercise, setCurrentExercise] = useState('');
  const [sets, setSets] = useState<SetData[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const addSet = (exerciseName?: string) => {
    const exercise = exerciseName || currentExercise.trim();
    if (!exercise) {
      Alert.alert('Error', 'Please enter an exercise name');
      return;
    }

    const newSet: SetData = {
      id: Date.now().toString(),
      exercise_name: exercise,
      weight: '',
      reps: '',
      rpe: '',
      set_order: sets.length + 1,
    };

    setSets([...sets, newSet]);
    if (!exerciseName) {
      setCurrentExercise('');
    }
  };

  const updateSet = (id: string, field: keyof SetData, value: string) => {
    setSets(sets.map(set =>
      set.id === id ? { ...set, [field]: value } : set
    ));
  };

  const removeSet = (id: string) => {
    const updatedSets = sets.filter(set => set.id !== id);
    // Reorder remaining sets
    setSets(updatedSets.map((set, index) => ({
      ...set,
      set_order: index + 1,
    })));
  };

  const validateAndSubmit = async () => {
    if (sets.length === 0) {
      Alert.alert('Error', 'Please add at least one set');
      return;
    }

    // Validate that all sets have weight and reps
    const invalidSets = sets.filter(set => !set.weight || !set.reps);
    if (invalidSets.length > 0) {
      Alert.alert('Error', 'All sets must have weight and reps');
      return;
    }

    try {
      setIsLoading(true);

      const sessionSets: CreateSetInput[] = sets.map(set => ({
        exercise_name: set.exercise_name,
        weight: parseFloat(set.weight),
        reps: parseInt(set.reps, 10),
        rpe: set.rpe ? parseFloat(set.rpe) : undefined,
        set_order: set.set_order,
      }));

      const request: CreateSessionRequest = {
        date: format(date, 'yyyy-MM-dd'),
        notes: notes.trim() || undefined,
        sets: sessionSets,
      };

      await sessionService.createSession(request);

      Alert.alert('Success', 'Workout logged successfully', [
        {
          text: 'OK',
          onPress: () => navigation.goBack(),
        },
      ]);
    } catch (error: any) {
      console.error('Failed to create session:', error);
      Alert.alert(
        'Error',
        error.response?.data?.error?.message || 'Failed to log workout. Please try again.'
      );
    } finally {
      setIsLoading(false);
    }
  };

  // Group sets by exercise
  const groupedSets = sets.reduce((acc, set) => {
    if (!acc[set.exercise_name]) {
      acc[set.exercise_name] = [];
    }
    acc[set.exercise_name].push(set);
    return acc;
  }, {} as Record<string, SetData[]>);

  return (
    <KeyboardAvoidingView
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      style={styles.container}
    >
      <View style={styles.header}>
        <TouchableOpacity onPress={() => navigation.goBack()}>
          <Text style={styles.cancelButton}>Cancel</Text>
        </TouchableOpacity>
        <Text style={styles.title}>Log Workout</Text>
        <TouchableOpacity onPress={validateAndSubmit} disabled={isLoading}>
          {isLoading ? (
            <ActivityIndicator color="#007AFF" />
          ) : (
            <Text style={styles.saveButton}>Save</Text>
          )}
        </TouchableOpacity>
      </View>

      <ScrollView style={styles.content} showsVerticalScrollIndicator={false}>
        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Date</Text>
          <View style={styles.dateContainer}>
            <Text style={styles.dateText}>{format(date, 'EEEE, MMMM d, yyyy')}</Text>
          </View>
        </View>

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Exercise</Text>
          <TextInput
            style={styles.input}
            placeholder="Enter exercise name"
            placeholderTextColor="#666"
            value={currentExercise}
            onChangeText={setCurrentExercise}
            onSubmitEditing={() => addSet()}
            autoCapitalize="words"
          />

          <ScrollView
            horizontal
            showsHorizontalScrollIndicator={false}
            style={styles.quickExercises}
          >
            {COMMON_EXERCISES.map((exercise) => (
              <TouchableOpacity
                key={exercise}
                style={styles.quickExerciseButton}
                onPress={() => addSet(exercise)}
              >
                <Text style={styles.quickExerciseText}>{exercise}</Text>
              </TouchableOpacity>
            ))}
          </ScrollView>
        </View>

        {Object.keys(groupedSets).length > 0 && (
          <View style={styles.section}>
            <Text style={styles.sectionTitle}>Sets</Text>
            {Object.entries(groupedSets).map(([exerciseName, exerciseSets]) => (
              <View key={exerciseName} style={styles.exerciseGroup}>
                <Text style={styles.exerciseName}>{exerciseName}</Text>
                {exerciseSets.map((set, index) => (
                  <ExerciseSetRow
                    key={set.id}
                    set={set}
                    setNumber={index + 1}
                    onUpdate={updateSet}
                    onRemove={removeSet}
                    isLastSet={index === exerciseSets.length - 1}
                  />
                ))}
                <TouchableOpacity
                  style={styles.addSetButton}
                  onPress={() => addSet(exerciseName)}
                >
                  <Text style={styles.addSetButtonText}>+ Add Set</Text>
                </TouchableOpacity>
              </View>
            ))}
          </View>
        )}

        <View style={styles.section}>
          <Text style={styles.sectionTitle}>Notes (Optional)</Text>
          <TextInput
            style={styles.textArea}
            placeholder="How did the workout feel?"
            placeholderTextColor="#666"
            value={notes}
            onChangeText={setNotes}
            multiline
            numberOfLines={4}
            textAlignVertical="top"
          />
        </View>

        <View style={{ height: 40 }} />
      </ScrollView>
    </KeyboardAvoidingView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0A0A0A',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 16,
    paddingTop: 60,
    borderBottomWidth: 1,
    borderBottomColor: '#1A1A1A',
  },
  title: {
    fontSize: 18,
    fontWeight: '600',
    color: '#FFF',
  },
  cancelButton: {
    fontSize: 16,
    color: '#999',
  },
  saveButton: {
    fontSize: 16,
    color: '#007AFF',
    fontWeight: '600',
  },
  content: {
    flex: 1,
  },
  section: {
    padding: 24,
    borderBottomWidth: 1,
    borderBottomColor: '#1A1A1A',
  },
  sectionTitle: {
    fontSize: 14,
    fontWeight: '600',
    color: '#999',
    marginBottom: 12,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  dateContainer: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  dateText: {
    fontSize: 16,
    color: '#FFF',
  },
  input: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    fontSize: 16,
    color: '#FFF',
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  quickExercises: {
    marginTop: 12,
  },
  quickExerciseButton: {
    backgroundColor: '#1A1A1A',
    borderRadius: 20,
    paddingHorizontal: 16,
    paddingVertical: 8,
    marginRight: 8,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  quickExerciseText: {
    color: '#FFF',
    fontSize: 14,
  },
  exerciseGroup: {
    marginBottom: 24,
  },
  exerciseName: {
    fontSize: 18,
    fontWeight: '600',
    color: '#FFF',
    marginBottom: 12,
  },
  addSetButton: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 12,
    alignItems: 'center',
    borderWidth: 1,
    borderColor: '#2A2A2A',
    borderStyle: 'dashed',
  },
  addSetButtonText: {
    color: '#007AFF',
    fontSize: 14,
    fontWeight: '600',
  },
  textArea: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    fontSize: 16,
    color: '#FFF',
    borderWidth: 1,
    borderColor: '#2A2A2A',
    minHeight: 100,
  },
});

export default CreateSessionScreen;
