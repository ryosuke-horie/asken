'use client'

import styles from './QuantityStepper.module.css'

interface QuantityStepperProps {
  value: number
  onChange: (value: number) => void
  min?: number
  max?: number
  step?: number
}

export default function QuantityStepper({ value, onChange, min = 0.5, max = 10, step = 0.5 }: QuantityStepperProps) {
  const handleDecrement = () => {
    const newValue = Math.max(min, value - step)
    onChange(Math.round(newValue * 10) / 10)
  }

  const handleIncrement = () => {
    const newValue = Math.min(max, value + step)
    onChange(Math.round(newValue * 10) / 10)
  }

  return (
    <div className={styles.container}>
      <button type="button" onClick={handleDecrement} disabled={value <= min} className={styles.button}>
        −
      </button>
      <span className={styles.value}>{value}</span>
      <button type="button" onClick={handleIncrement} disabled={value >= max} className={styles.button}>
        +
      </button>
    </div>
  )
}
