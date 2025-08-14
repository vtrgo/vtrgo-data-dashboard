"use client"

import * as React from "react"
import { cva, type VariantProps } from "class-variance-authority"
import * as ProgressPrimitive from "@radix-ui/react-progress"

import { cn } from "@/lib/utils"

const progressRootVariants = cva("relative h-4 w-full overflow-hidden rounded-full", {
  variants: {
    variant: {
      default: "bg-secondary",
      success: "bg-success/20",
      warning: "bg-warning/20",
      destructive: "bg-destructive/20",
    },
  },
  defaultVariants: {
    variant: "default",
  },
})

const progressIndicatorVariants = cva("h-full w-full flex-1 transition-all bg-gradient-to-r", {
  variants: {
    variant: {
      default: "bg-primary",
      success: "from-success/20 to-success",
      warning: "from-warning/20 to-warning",
      destructive: "from-destructive/20 to-destructive",
    },
  },
  defaultVariants: {
    variant: "default",
  },
})

export interface ProgressProps
  extends React.ComponentPropsWithoutRef<typeof ProgressPrimitive.Root>,
    VariantProps<typeof progressIndicatorVariants> {
      value?: number | null;
    }

const Progress = React.forwardRef<React.ComponentRef<typeof ProgressPrimitive.Root>, ProgressProps>(
  ({ className, value, variant, ...props }, ref) => (
  <ProgressPrimitive.Root
    ref={ref}
    className={cn(progressRootVariants({ variant, className }))}
    value={value}
    {...props}
  >
    <ProgressPrimitive.Indicator
      className={cn(progressIndicatorVariants({ variant }))}
      style={{ transform: `translateX(-${100 - (value || 0)}%)` }}
    />
  </ProgressPrimitive.Root>
))
Progress.displayName = ProgressPrimitive.Root.displayName

export { Progress }
