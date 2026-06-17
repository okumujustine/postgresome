import { cva, type VariantProps } from "class-variance-authority";
import type { HTMLAttributes } from "react";
import { cn } from "@/lib/utils";

const badgeVariants = cva(
  "inline-flex items-center rounded-md border px-2.5 py-1 text-[11px] font-semibold tracking-[0.01em]",
  {
    variants: {
      variant: {
        default: "border-[#111111] bg-white text-slate-800",
        critical: "border-[#111111] bg-[#ffd7d2] text-[#a31616]",
        warning: "border-[#111111] bg-[#fff1b8] text-[#8a4b00]",
        info: "border-[#111111] bg-[#dce8ff] text-[#254fd2]",
        success: "border-[#111111] bg-[#d9f7d8] text-[#166534]",
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
