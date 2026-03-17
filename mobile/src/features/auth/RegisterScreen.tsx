import React, { useState } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  Alert,
  KeyboardAvoidingView,
  Platform,
  ScrollView,
  ActivityIndicator,
} from 'react-native';
import { useForm, Controller } from 'react-hook-form';
import { useAuth } from '../../shared/contexts';
import { RegisterRequest } from '../../services/api';

interface RegisterFormData {
  name: string;
  email: string;
  password: string;
  confirmPassword: string;
  bodyWeight?: string;
}

interface RegisterScreenProps {
  navigation: any;
}

export const RegisterScreen: React.FC<RegisterScreenProps> = ({ navigation }) => {
  const { register } = useAuth();
  const [isLoading, setIsLoading] = useState(false);

  const {
    control,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<RegisterFormData>({
    defaultValues: {
      name: '',
      email: '',
      password: '',
      confirmPassword: '',
      bodyWeight: '',
    },
  });

  const password = watch('password');

  const onSubmit = async (data: RegisterFormData) => {
    try {
      setIsLoading(true);

      const registerData: RegisterRequest = {
        name: data.name,
        email: data.email,
        password: data.password,
      };

      if (data.bodyWeight && data.bodyWeight.trim() !== '') {
        registerData.body_weight = parseFloat(data.bodyWeight);
      }

      await register(registerData);
      // Navigation to home screen will be handled by navigation structure
    } catch (error: any) {
      console.error('Registration error:', error);
      Alert.alert(
        'Registration Failed',
        error.response?.data?.error?.message || 'Failed to create account. Please try again.'
      );
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <KeyboardAvoidingView
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
      style={styles.container}
    >
      <ScrollView contentContainerStyle={styles.scrollContent} showsVerticalScrollIndicator={false}>
        <View style={styles.content}>
          <View style={styles.header}>
            <Text style={styles.title}>Create Account</Text>
            <Text style={styles.subtitle}>Start your strength journey</Text>
          </View>

          <View style={styles.form}>
            <View style={styles.inputContainer}>
              <Text style={styles.label}>Name</Text>
              <Controller
                control={control}
                name="name"
                rules={{
                  required: 'Name is required',
                  minLength: {
                    value: 2,
                    message: 'Name must be at least 2 characters',
                  },
                }}
                render={({ field: { onChange, onBlur, value } }) => (
                  <TextInput
                    style={[styles.input, errors.name && styles.inputError]}
                    placeholder="John Doe"
                    placeholderTextColor="#999"
                    value={value}
                    onChangeText={onChange}
                    onBlur={onBlur}
                    autoCapitalize="words"
                    autoCorrect={false}
                    editable={!isLoading}
                  />
                )}
              />
              {errors.name && <Text style={styles.errorText}>{errors.name.message}</Text>}
            </View>

            <View style={styles.inputContainer}>
              <Text style={styles.label}>Email</Text>
              <Controller
                control={control}
                name="email"
                rules={{
                  required: 'Email is required',
                  pattern: {
                    value: /^[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}$/i,
                    message: 'Invalid email address',
                  },
                }}
                render={({ field: { onChange, onBlur, value } }) => (
                  <TextInput
                    style={[styles.input, errors.email && styles.inputError]}
                    placeholder="your@email.com"
                    placeholderTextColor="#999"
                    value={value}
                    onChangeText={onChange}
                    onBlur={onBlur}
                    keyboardType="email-address"
                    autoCapitalize="none"
                    autoCorrect={false}
                    editable={!isLoading}
                  />
                )}
              />
              {errors.email && <Text style={styles.errorText}>{errors.email.message}</Text>}
            </View>

            <View style={styles.inputContainer}>
              <Text style={styles.label}>Password</Text>
              <Controller
                control={control}
                name="password"
                rules={{
                  required: 'Password is required',
                  minLength: {
                    value: 8,
                    message: 'Password must be at least 8 characters',
                  },
                }}
                render={({ field: { onChange, onBlur, value } }) => (
                  <TextInput
                    style={[styles.input, errors.password && styles.inputError]}
                    placeholder="At least 8 characters"
                    placeholderTextColor="#999"
                    value={value}
                    onChangeText={onChange}
                    onBlur={onBlur}
                    secureTextEntry
                    autoCapitalize="none"
                    autoCorrect={false}
                    editable={!isLoading}
                  />
                )}
              />
              {errors.password && <Text style={styles.errorText}>{errors.password.message}</Text>}
            </View>

            <View style={styles.inputContainer}>
              <Text style={styles.label}>Confirm Password</Text>
              <Controller
                control={control}
                name="confirmPassword"
                rules={{
                  required: 'Please confirm your password',
                  validate: (value) => value === password || 'Passwords do not match',
                }}
                render={({ field: { onChange, onBlur, value } }) => (
                  <TextInput
                    style={[styles.input, errors.confirmPassword && styles.inputError]}
                    placeholder="Re-enter password"
                    placeholderTextColor="#999"
                    value={value}
                    onChangeText={onChange}
                    onBlur={onBlur}
                    secureTextEntry
                    autoCapitalize="none"
                    autoCorrect={false}
                    editable={!isLoading}
                  />
                )}
              />
              {errors.confirmPassword && (
                <Text style={styles.errorText}>{errors.confirmPassword.message}</Text>
              )}
            </View>

            <View style={styles.inputContainer}>
              <Text style={styles.label}>Body Weight (kg) - Optional</Text>
              <Controller
                control={control}
                name="bodyWeight"
                rules={{
                  pattern: {
                    value: /^\d+(\.\d{1,2})?$/,
                    message: 'Please enter a valid weight',
                  },
                }}
                render={({ field: { onChange, onBlur, value } }) => (
                  <TextInput
                    style={[styles.input, errors.bodyWeight && styles.inputError]}
                    placeholder="e.g., 75.5"
                    placeholderTextColor="#999"
                    value={value}
                    onChangeText={onChange}
                    onBlur={onBlur}
                    keyboardType="decimal-pad"
                    editable={!isLoading}
                  />
                )}
              />
              {errors.bodyWeight && (
                <Text style={styles.errorText}>{errors.bodyWeight.message}</Text>
              )}
            </View>

            <TouchableOpacity
              style={[styles.button, isLoading && styles.buttonDisabled]}
              onPress={handleSubmit(onSubmit)}
              disabled={isLoading}
            >
              {isLoading ? (
                <ActivityIndicator color="#FFF" />
              ) : (
                <Text style={styles.buttonText}>Create Account</Text>
              )}
            </TouchableOpacity>

            <View style={styles.footer}>
              <Text style={styles.footerText}>Already have an account? </Text>
              <TouchableOpacity onPress={() => navigation.navigate('Login')} disabled={isLoading}>
                <Text style={styles.linkText}>Login</Text>
              </TouchableOpacity>
            </View>
          </View>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#0A0A0A',
  },
  scrollContent: {
    flexGrow: 1,
  },
  content: {
    flex: 1,
    justifyContent: 'center',
    padding: 24,
  },
  header: {
    alignItems: 'center',
    marginBottom: 32,
  },
  title: {
    fontSize: 32,
    fontWeight: 'bold',
    color: '#FFF',
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 16,
    color: '#999',
  },
  form: {
    width: '100%',
  },
  inputContainer: {
    marginBottom: 16,
  },
  label: {
    fontSize: 14,
    fontWeight: '600',
    color: '#FFF',
    marginBottom: 8,
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
  inputError: {
    borderColor: '#FF4444',
  },
  errorText: {
    color: '#FF4444',
    fontSize: 12,
    marginTop: 4,
  },
  button: {
    backgroundColor: '#007AFF',
    borderRadius: 12,
    padding: 16,
    alignItems: 'center',
    marginTop: 8,
  },
  buttonDisabled: {
    opacity: 0.6,
  },
  buttonText: {
    color: '#FFF',
    fontSize: 16,
    fontWeight: '600',
  },
  footer: {
    flexDirection: 'row',
    justifyContent: 'center',
    marginTop: 24,
  },
  footerText: {
    color: '#999',
    fontSize: 14,
  },
  linkText: {
    color: '#007AFF',
    fontSize: 14,
    fontWeight: '600',
  },
});

export default RegisterScreen;
