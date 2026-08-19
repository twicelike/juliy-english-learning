import * as React from "react"

export interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: "default" | "primary" | "new"
}

function Badge({ className = "", variant = "default", ...props }: BadgeProps) {
  let variantClasses = ""
  switch (variant) {
    case "default":
      variantClasses = "bg-white/10 text-white border-white/20"
      break
    case "primary":
      variantClasses = "bg-indigo-500/10 text-indigo-400 border-indigo-500/20"
      break
    case "new":
      variantClasses = "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
      break
  }

  // Add the semantic class matching the variant for Sketch overrides
  const semanticClass = `badge badge-${variant}`

  return (
    <div
      className={`inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 ${variantClasses} ${semanticClass} ${className}`}
      {...props}
    />
  )
}

export { Badge }
