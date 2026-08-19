import * as React from "react"

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: "primary" | "secondary" | "outline" | "ghost"
  size?: "default" | "sm" | "lg" | "icon"
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className = "", variant = "primary", size = "default", ...props }, ref) => {
    
    // Base Tailwind styles for the component (structure & fallback)
    const baseClasses = "inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50"
    
    // Semantic class for theme overrides (e.g. Sketch theme)
    const semanticClass = `btn btn-${variant}`
    
    // Default Tailwind variants (applied unless overridden)
    let variantClasses = ""
    switch (variant) {
      case "primary":
        variantClasses = "bg-indigo-500 text-white hover:bg-indigo-600"
        break
      case "secondary":
        variantClasses = "bg-white/10 text-white hover:bg-white/20"
        break
      case "outline":
        variantClasses = "border border-white/20 hover:bg-white/10"
        break
      case "ghost":
        variantClasses = "hover:bg-white/10 hover:text-white"
        break
    }
    
    let sizeClasses = ""
    switch (size) {
      case "default":
        sizeClasses = "h-10 px-4 py-2"
        break
      case "sm":
        sizeClasses = "h-9 rounded-md px-3"
        break
      case "lg":
        sizeClasses = "h-12 rounded-xl px-8"
        break
      case "icon":
        sizeClasses = "h-10 w-10"
        break
    }

    return (
      <button
        className={`${baseClasses} ${variantClasses} ${sizeClasses} ${semanticClass} ${className}`}
        ref={ref}
        {...props}
      />
    )
  }
)
Button.displayName = "Button"

export { Button }
