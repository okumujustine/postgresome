export function formatRelativeTime(iso: string): string {
  const diffMs = Date.now() - new Date(iso).getTime();
  const diffMin = Math.round(diffMs / 60000);
  if (diffMin < 1) return 'just now';
  if (diffMin < 60) return `${diffMin}m ago`;
  const diffHr = Math.round(diffMin / 60);
  if (diffHr < 24) return `${diffHr}h ago`;
  return `${Math.round(diffHr / 24)}d ago`;
}

// formatRelativeTimeShort is like formatRelativeTime but reports
// second-level granularity for very recent timestamps (e.g. "Last checked").
export function formatRelativeTimeShort(iso: string): string {
  const diffMs = Date.now() - new Date(iso).getTime();
  const diffSec = Math.max(0, Math.round(diffMs / 1000));
  if (diffSec < 60) return `${diffSec}s ago`;
  return formatRelativeTime(iso);
}

export function formatDuration(ms: number): string {
  if (ms >= 1000) {
    return `${(ms / 1000).toFixed(2)} s`;
  }
  return `${ms.toFixed(2)} ms`;
}

const PG_BLOCK_SIZE_BYTES = 8192;

// formatBytes converts a count of PostgreSQL 8KB blocks into a human-readable
// size (KB/MB/GB).
export function formatBytes(blocks: number): string {
  const bytes = blocks * PG_BLOCK_SIZE_BYTES;
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  if (bytes < 1024 * 1024 * 1024) return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  return `${(bytes / (1024 * 1024 * 1024)).toFixed(2)} GB`;
}

const SEVERITY_EMOJI: Record<string, string> = {
  critical: '🔴',
  warning: '🟡',
  info: '🔵',
};

export function severityEmoji(severity: string): string {
  return SEVERITY_EMOJI[severity.toLowerCase()] ?? '⚪';
}
