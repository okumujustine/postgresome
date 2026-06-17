import { X } from "lucide-react";
import { useEffect, useState, type ReactNode } from "react";

export function DismissibleAlert({
  children,
}: {
  children: ReactNode;
}) {
  const [dismissed, setDismissed] = useState(false);

  useEffect(() => {
    setDismissed(false);
  }, [children]);

  if (dismissed) {
    return null;
  }

  return (
    <div className="inline-alert flex items-start justify-between gap-3">
      <div className="min-w-0 flex-1">{children}</div>
      <button
        type="button"
        onClick={() => setDismissed(true)}
        className="rounded-md p-1 text-slate-700 transition-colors hover:bg-black/5 hover:text-slate-950"
        aria-label="Dismiss alert"
      >
        <X className="h-4 w-4" />
      </button>
    </div>
  );
}
