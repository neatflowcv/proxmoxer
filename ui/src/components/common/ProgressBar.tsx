interface ProgressBarProps {
  value: number // 0-100
  max?: number
  label?: string
  showPercentage?: boolean
  size?: 'sm' | 'md' | 'lg'
  color?: 'blue' | 'green' | 'yellow' | 'red'
  className?: string
}

function getColorClass(color: string, value: number): string {
  if (color !== 'auto') {
    const colors: Record<string, string> = {
      blue: 'bg-blue-500',
      green: 'bg-green-500',
      yellow: 'bg-yellow-500',
      red: 'bg-red-500',
    }
    return colors[color] || 'bg-blue-500'
  }

  // Auto color based on value
  if (value < 50) return 'bg-green-500'
  if (value < 75) return 'bg-yellow-500'
  if (value < 90) return 'bg-orange-500'
  return 'bg-red-500'
}

function getSizeClass(size: string): string {
  const sizes: Record<string, string> = {
    sm: 'h-1.5',
    md: 'h-2.5',
    lg: 'h-4',
  }
  return sizes[size] || 'h-2.5'
}

export function ProgressBar({
  value,
  max = 100,
  label,
  showPercentage = true,
  size = 'md',
  color = 'auto' as 'blue' | 'green' | 'yellow' | 'red',
  className = '',
}: ProgressBarProps & { color?: 'blue' | 'green' | 'yellow' | 'red' | 'auto' }) {
  const percentage = Math.min(100, Math.max(0, (value / max) * 100))
  const colorClass = getColorClass(color, percentage)
  const sizeClass = getSizeClass(size)

  return (
    <div className={className}>
      {(label || showPercentage) && (
        <div className="flex justify-between mb-1">
          {label && <span className="text-sm font-medium text-gray-700">{label}</span>}
          {showPercentage && (
            <span className="text-sm font-medium text-gray-500">{percentage.toFixed(1)}%</span>
          )}
        </div>
      )}
      <div className={`w-full bg-gray-200 rounded-full ${sizeClass}`}>
        <div
          className={`${colorClass} ${sizeClass} rounded-full transition-all duration-300`}
          style={{ width: `${percentage}%` }}
        />
      </div>
    </div>
  )
}
