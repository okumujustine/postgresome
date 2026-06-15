import { Search } from 'lucide-react';
import { useMemo, useState } from 'react';
import { Input } from './input';

interface CommandItemType {
  id: string;
  title: string;
  description?: string;
  onSelect: () => void;
}

export function CommandMenu({
  open,
  onOpenChange,
  title,
  items,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  items: CommandItemType[];
}) {
  const [query, setQuery] = useState('');

  const filtered = useMemo(() => {
    const needle = query.trim().toLowerCase();
    if (!needle) return items;
    return items.filter((item) => `${item.title} ${item.description ?? ''}`.toLowerCase().includes(needle));
  }, [items, query]);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center bg-[rgba(9,30,66,0.28)] px-4 pt-24">
      <div className="w-full max-w-2xl rounded-xl border border-[var(--border)] bg-[var(--panel)] shadow-[var(--shadow-lg)]">
        <div className="border-b border-[var(--border)] px-4 py-4">
          <div className="mb-3 text-sm font-semibold text-[var(--foreground)]">{title}</div>
          <div className="relative">
            <Search size={15} className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-[var(--muted-foreground)]" />
            <Input value={query} onChange={(event) => setQuery(event.target.value)} className="pl-9" placeholder="Search pages and findings" />
          </div>
        </div>
        <div className="max-h-[60vh] overflow-y-auto p-2">
          {filtered.map((item) => (
            <button
              key={item.id}
              onClick={() => {
                item.onSelect();
                onOpenChange(false);
              }}
              className="flex w-full flex-col items-start rounded-lg px-3 py-3 text-left hover:bg-[var(--muted)]"
            >
              <div className="text-sm font-medium text-[var(--foreground)]">{item.title}</div>
              {item.description ? <div className="mt-1 text-sm text-[var(--muted-foreground)]">{item.description}</div> : null}
            </button>
          ))}
          {filtered.length === 0 ? <div className="px-3 py-8 text-sm text-[var(--muted-foreground)]">No matching actions.</div> : null}
        </div>
      </div>
    </div>
  );
}
