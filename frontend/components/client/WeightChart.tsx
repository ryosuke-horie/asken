'use client'

import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  ReferenceLine,
} from 'recharts'
import { WeightRecord } from '@/types/weight'
import { formatDateShort } from '@/lib/date'
import styles from './WeightChart.module.css'

interface WeightChartProps {
  records: WeightRecord[]
  targetWeight?: number
}

export default function WeightChart({ records, targetWeight }: WeightChartProps) {
  const chartData = records.map((record) => ({
    date: formatDateShort(record.recorded_at),
    weight: record.weight,
  }))

  const weights = records.map((r) => r.weight)
  const minWeight = weights.length > 0 ? Math.min(...weights) : 0
  const maxWeight = weights.length > 0 ? Math.max(...weights) : 100

  const yMin = Math.floor(Math.min(minWeight, targetWeight ?? minWeight) - 2)
  const yMax = Math.ceil(Math.max(maxWeight, targetWeight ?? maxWeight) + 2)

  if (records.length === 0) {
    return (
      <div className={styles.empty}>
        <p>まだ体重の記録がありません</p>
      </div>
    )
  }

  return (
    <div className={styles.container}>
      <ResponsiveContainer width="100%" height={250}>
        <LineChart data={chartData} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#eee" />
          <XAxis dataKey="date" tick={{ fontSize: 12 }} stroke="#666" />
          <YAxis domain={[yMin, yMax]} tick={{ fontSize: 12 }} stroke="#666" unit="kg" />
          <Tooltip
            formatter={(value) => [`${value} kg`, '体重']}
            labelFormatter={(label) => `${label}`}
            contentStyle={{ fontSize: 14 }}
          />
          <Line
            type="monotone"
            dataKey="weight"
            stroke="#f97316"
            strokeWidth={2}
            dot={{ fill: '#f97316', strokeWidth: 2, r: 4 }}
            activeDot={{ r: 6 }}
          />
          {targetWeight && (
            <ReferenceLine
              y={targetWeight}
              stroke="#22c55e"
              strokeDasharray="5 5"
              strokeWidth={2}
              label={{
                value: `目標: ${targetWeight}kg`,
                position: 'right',
                fill: '#22c55e',
                fontSize: 12,
              }}
            />
          )}
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}
