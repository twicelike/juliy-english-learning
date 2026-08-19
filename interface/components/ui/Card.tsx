import * as React from "react"

export interface CardProps extends React.HTMLAttributes<HTMLDivElement> {}

const Card = React.forwardRef<HTMLDivElement, CardProps>(
  ({ className = "", ...props }, ref) => {
    return (
      <div
        ref={ref}
        className={`card rounded-2xl border border-white/10 bg-white/5 p-6 shadow-sm ${className}`}
        {...props}
      />
    )
  }
)
Card.displayName = "Card"

export { Card }
