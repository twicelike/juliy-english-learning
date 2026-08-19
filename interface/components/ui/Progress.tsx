import * as React from "react"

export interface ProgressProps extends React.HTMLAttributes<HTMLDivElement> {
  value?: number
  max?: number
}

const Progress = React.forwardRef<HTMLDivElement, ProgressProps>(
  ({ className = "", value = 0, max = 100, ...props }, ref) => {
    // Ensure value is between 0 and max
    const safeValue = Math.min(Math.max(value, 0), max)
    const percentage = max > 0 ? (safeValue / max) * 100 : 0

    return (
      <div
        ref={ref}
        className={`progress-track relative h-4 w-full overflow-hidden rounded-full bg-white/10 ${className}`}
        {...props}
      >
        <div
          className="progress-fill h-full w-full flex-1 bg-indigo-500 transition-all duration-300 ease-in-out"
          style={{ width: `${percentage}%` }}
        />
      </div>
    )
  }
)
Progress.displayName = "Progress"

export { Progress }
