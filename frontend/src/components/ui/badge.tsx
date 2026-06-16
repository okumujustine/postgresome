import { cva, type VariantProps } from "class-variance-authority";
import type { HTMLAttributes } from "react";
import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center rounded border px-2.5 py-1 font-mono text-[11px] uppercase tracking-[0.06em]",
  {
    variants: {
      variant: {
        default: "border-slate-200 bg-slate-50 text-slate-700",
        critical: "border-red-300 bg-[rgba(220,38,38,0.05)] text-red-700",
        warning: "border-amber-300 bg-[rgba(217,119,6,0.05)] text-amber-700",
        info: "border-blue-300 bg-[rgba(37,99,235,0.05)] text-blue-700",
        success: "border-emerald-300 bg-[rgba(22,163,74,0.05)] text-emerald-700",
      },
    },
    defaultVariants: {
      variant: "default",
    },
  },
);

export interface BadgeProps
  extends HTMLAttributes<HTMLDivElement>,
    VariantProps<typeof badgeVariants> {}

export function Badge({ className, variant, ...props }: BadgeProps) {
  return <div className={cn(badgeVariants({ variant }), className)} {...props} />;
}
