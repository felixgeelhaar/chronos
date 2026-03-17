import React from 'react';
import { View, Text, StyleSheet, Dimensions } from 'react-native';
import { VictoryBar, VictoryChart, VictoryAxis, VictoryTheme } from 'victory-native';

interface DataPoint {
  x: number | Date | string;
  y: number;
}

interface BarChartProps {
  data: DataPoint[];
  title?: string;
  yLabel?: string;
  color?: string;
  height?: number;
}

export const BarChart: React.FC<BarChartProps> = ({
  data,
  title,
  yLabel,
  color = '#007AFF',
  height = 200,
}) => {
  const screenWidth = Dimensions.get('window').width;

  if (data.length === 0) {
    return (
      <View style={[styles.container, { height }]}>
        {title && <Text style={styles.title}>{title}</Text>}
        <View style={styles.emptyContainer}>
          <Text style={styles.emptyText}>No data available</Text>
        </View>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      {title && <Text style={styles.title}>{title}</Text>}
      <VictoryChart
        width={screenWidth - 48}
        height={height}
        theme={VictoryTheme.material}
        padding={{ top: 20, bottom: 40, left: 50, right: 20 }}
        domainPadding={{ x: 20 }}
      >
        <VictoryAxis
          style={{
            axis: { stroke: '#2A2A2A' },
            tickLabels: { fill: '#999', fontSize: 10 },
            grid: { stroke: '#1A1A1A' },
          }}
        />
        <VictoryAxis
          dependentAxis
          style={{
            axis: { stroke: '#2A2A2A' },
            tickLabels: { fill: '#999', fontSize: 10 },
            grid: { stroke: '#1A1A1A' },
          }}
          label={yLabel}
          axisLabelComponent={
            <Text style={{ fill: '#999', fontSize: 12, padding: 10 }} />
          }
        />
        <VictoryBar
          data={data}
          style={{
            data: { fill: color },
          }}
          cornerRadius={{ top: 4 }}
        />
      </VictoryChart>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    backgroundColor: '#1A1A1A',
    borderRadius: 12,
    padding: 16,
    borderWidth: 1,
    borderColor: '#2A2A2A',
  },
  title: {
    fontSize: 16,
    fontWeight: '600',
    color: '#FFF',
    marginBottom: 12,
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
  },
  emptyText: {
    color: '#666',
    fontSize: 14,
  },
});

export default BarChart;
