import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";
import { cn } from "@/lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md border border-[#111111] text-sm font-semibold transition-all duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 ring-offset-background",
  {
    variants: {
      variant: {
        default:
          "bg-primary text-primary-foreground shadow-[1px_1px_0_#111111] hover:-translate-x-[1px] hover:-translate-y-[1px] hover:shadow-[2px_2px_0_#111111]",
        secondary:
          "bg-[#f4f1e6] text-secondary-foreground shadow-[1px_1px_0_#111111] hover:-translate-x-[1px] hover:-translate-y-[1px] hover:bg-[#ece6d2] hover:shadow-[2px_2px_0_#111111]",
        ghost:
          "border-transparent bg-transparent text-foreground shadow-none hover:border-[#111111] hover:bg-[#f4f1e6] hover:shadow-[1px_1px_0_#111111]",
        outline:
          "bg-white text-slate-800 shadow-[1px_1px_0_#111111] hover:-translate-x-[1px] hover:-translate-y-[1px] hover:bg-[#f4f1e6] hover:shadow-[2px_2px_0_#111111]",
      },
      size: {
        default: "h-11 px-4 py-2",
        sm: "h-8 px-3 text-[13px]",
        lg: "h-11 px-5",
        icon: "h-10 w-10",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  },
);

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean;
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button";
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    );
  },
);
Button.displayName = "Button";

export { Button, buttonVariants };
