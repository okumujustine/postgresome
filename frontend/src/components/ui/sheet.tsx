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
        <Dialog.Overlay className="fixed inset-0 z-40 bg-[#111111]/15" />
        <Dialog.Content
          className={cn(
            "fixed inset-y-0 left-0 z-50 w-[88vw] max-w-sm border-r-2 border-r-[#111111] border-t-2 border-t-[#111111] bg-[#fffdf7] p-5 shadow-sheet outline-none",
          )}
        >
          <div className="mb-5 flex items-center justify-between">
            <Dialog.Title className="font-heading text-sm font-semibold text-foreground">
              {title}
            </Dialog.Title>
            <Dialog.Close className="rounded-md border-2 border-[#111111] bg-white p-2 text-foreground shadow-[2px_2px_0_#111111] transition-all hover:-translate-x-[1px] hover:-translate-y-[1px] hover:bg-[#f4f1e6] hover:shadow-[3px_3px_0_#111111]">
              <X className="h-4 w-4" />
            </Dialog.Close>
          </div>
          {children}
        </Dialog.Content>
      </Dialog.Portal>
    </Dialog.Root>
  );
}
