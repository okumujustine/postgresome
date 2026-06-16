import * as Dialog from "@radix-ui/react-dialog";
import { X } from "lucide-react";
import type { ReactNode } from "react";
import { cn } from "@/lib/utils";

export function Sheet({
  open,
  onOpenChange,
  trigger,
  children,
  title,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  trigger: ReactNode;
  children: ReactNode;
  title: string;
}) {
  return (
    <Dialog.Root open={open} onOpenChange={onOpenChange}>
      <Dialog.Trigger asChild>{trigger}</Dialog.Trigger>
      <Dialog.Portal>
        <Dialog.Overlay className="fixed inset-0 z-40 bg-slate-900/20 backdrop-blur-[1px]" />
        <Dialog.Content
          className={cn(
            "fixed inset-y-0 left-0 z-50 w-[88vw] max-w-sm border-r bg-white p-5 shadow-sheet outline-none",
          )}
        >
          <div className="mb-5 flex items-center justify-between">
            <Dialog.Title className="font-heading text-sm font-semibold text-foreground">
              {title}
            </Dialog.Title>
            <Dialog.Close className="rounded-md p-2 text-muted-foreground hover:bg-slate-100 hover:text-foreground">
              <X className="h-4 w-4" />
            </Dialog.Close>
          </div>
          {children}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
