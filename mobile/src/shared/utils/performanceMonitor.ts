import { InteractionManager, PerformanceObserver } from 'react-native';

/**
 * Performance monitoring utility
 * Tracks app performance metrics and provides insights
 */

interface PerformanceMetric {
  name: string;
  duration: number;
  timestamp: number;
}

class PerformanceMonitor {
  private metrics: PerformanceMetric[] = [];
  private observers: Map<string, PerformanceObserver> = new Map();
  private timers: Map<string, number> = new Map();

  /**
   * Start timing an operation
   */
  startTimer(name: string): void {
    this.timers.set(name, Date.now());
  }

  /**
   * End timing an operation and record the metric
   */
  endTimer(name: string): number | null {
    const startTime = this.timers.get(name);
    if (!startTime) {
      console.warn(`Timer '${name}' was not started`);
      return null;
    }

    const duration = Date.now() - startTime;
    this.recordMetric(name, duration);
    this.timers.delete(name);

    return duration;
  }

  /**
   * Record a performance metric
   */
  private recordMetric(name: string, duration: number): void {
    const metric: PerformanceMetric = {
      name,
      duration,
      timestamp: Date.now(),
    };

    this.metrics.push(metric);

    // Log slow operations in development
    if (__DEV__ && duration > 1000) {
      console.warn(`Slow operation detected: ${name} took ${duration}ms`);
    }

    // Keep only last 100 metrics to prevent memory issues
    if (this.metrics.length > 100) {
      this.metrics.shift();
    }
  }

  /**
   * Measure async operation with automatic timing
   */
  async measure<T>(name: string, operation: () => Promise<T>): Promise<T> {
    this.startTimer(name);
    try {
      const result = await operation();
      return result;
    } finally {
      this.endTimer(name);
    }
  }

  /**
   * Measure sync operation with automatic timing
   */
  measureSync<T>(name: string, operation: () => T): T {
    this.startTimer(name);
    try {
      return operation();
    } finally {
      this.endTimer(name);
    }
  }

  /**
   * Wait for interactions to complete before executing
   * Useful for non-critical operations that can wait
   */
  async runAfterInteractions<T>(operation: () => Promise<T>): Promise<T> {
    await new Promise((resolve) => {
      InteractionManager.runAfterInteractions(resolve);
    });
    return operation();
  }

  /**
   * Get all recorded metrics
   */
  getMetrics(): PerformanceMetric[] {
    return [...this.metrics];
  }

  /**
   * Get metrics for a specific operation
   */
  getMetricsByName(name: string): PerformanceMetric[] {
    return this.metrics.filter((m) => m.name === name);
  }

  /**
   * Get average duration for an operation
   */
  getAverageDuration(name: string): number | null {
    const operationMetrics = this.getMetricsByName(name);
    if (operationMetrics.length === 0) return null;

    const totalDuration = operationMetrics.reduce((sum, m) => sum + m.duration, 0);
    return totalDuration / operationMetrics.length;
  }

  /**
   * Get slowest operations
   */
  getSlowestOperations(limit: number = 10): PerformanceMetric[] {
    return [...this.metrics]
      .sort((a, b) => b.duration - a.duration)
      .slice(0, limit);
  }

  /**
   * Clear all metrics
   */
  clearMetrics(): void {
    this.metrics = [];
    this.timers.clear();
  }

  /**
   * Generate performance report
   */
  generateReport(): {
    totalOperations: number;
    averageDuration: number;
    slowestOperations: PerformanceMetric[];
    operationBreakdown: Record<string, { count: number; avgDuration: number }>;
  } {
    const totalOperations = this.metrics.length;
    const totalDuration = this.metrics.reduce((sum, m) => sum + m.duration, 0);
    const averageDuration = totalOperations > 0 ? totalDuration / totalOperations : 0;

    // Group by operation name
    const operationBreakdown: Record<string, { count: number; avgDuration: number }> = {};
    const operationTotals: Record<string, { count: number; totalDuration: number }> = {};

    this.metrics.forEach((metric) => {
      if (!operationTotals[metric.name]) {
        operationTotals[metric.name] = { count: 0, totalDuration: 0 };
      }
      operationTotals[metric.name].count++;
      operationTotals[metric.name].totalDuration += metric.duration;
    });

    Object.entries(operationTotals).forEach(([name, data]) => {
      operationBreakdown[name] = {
        count: data.count,
        avgDuration: data.totalDuration / data.count,
      };
    });

    return {
      totalOperations,
      averageDuration,
      slowestOperations: this.getSlowestOperations(5),
      operationBreakdown,
    };
  }

  /**
   * Log performance report to console
   */
  logReport(): void {
    const report = this.generateReport();
    console.log('=== Performance Report ===');
    console.log(`Total Operations: ${report.totalOperations}`);
    console.log(`Average Duration: ${report.averageDuration.toFixed(2)}ms`);
    console.log('\nSlowest Operations:');
    report.slowestOperations.forEach((op, index) => {
      console.log(`${index + 1}. ${op.name}: ${op.duration}ms`);
    });
    console.log('\nOperation Breakdown:');
    Object.entries(report.operationBreakdown)
      .sort((a, b) => b[1].avgDuration - a[1].avgDuration)
      .forEach(([name, data]) => {
        console.log(`  ${name}: ${data.count} calls, avg ${data.avgDuration.toFixed(2)}ms`);
      });
    console.log('========================');
  }
}

export const performanceMonitor = new PerformanceMonitor();

/**
 * Decorator for measuring method performance
 */
export function measurePerformance(target: any, propertyKey: string, descriptor: PropertyDescriptor) {
  const originalMethod = descriptor.value;

  descriptor.value = async function (...args: any[]) {
    const operationName = `${target.constructor.name}.${propertyKey}`;
    return performanceMonitor.measure(operationName, () => originalMethod.apply(this, args));
  };

  return descriptor;
}

export default performanceMonitor;
