import * as React from "react";
import { cn } from "@/lib/utils";

const Input = React.forwardRef<HTMLInputElement, React.ComponentProps<"input">>(
  ({ className, ...props }, ref) => {
    return (
        <input
        className={cn(
          "flex h-11 w-full rounded-md border border-[#111111] bg-white px-3.5 py-2 text-sm text-foreground shadow-[1px_1px_0_#111111] outline-none transition-all placeholder:text-muted-foreground focus-visible:-translate-x-[1px] focus-visible:-translate-y-[1px] focus-visible:bg-[#fffdf7] focus-visible:ring-0 focus-visible:shadow-[2px_2px_0_#111111]",
          className,
        )}
        ref={ref}
        {...props}
      />
    );
  },
);
Input.displayName = "Input";

export { Input };
