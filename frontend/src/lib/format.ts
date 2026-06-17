const numberFormatter = new Intl.NumberFormat("en-US", {
  maximumFractionDigits: 1,
});

export function formatTimestamp(value?: string | Date | null) {
  if (!value) {
    return "Not available";
  }

  const date = value instanceof Date ? value : new Date(value);

  return new Intl.DateTimeFormat("en-US", {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  }).format(date);
}

export function formatRelativeTime(value?: string | Date | null) {
  if (!value) {
    return "unknown";
  }

  const date = value instanceof Date ? value : new Date(value);
  const diff = date.getTime() - Date.now();
  const minutes = Math.round(diff / 60000);
  const absMinutes = Math.abs(minutes);

  if (absMinutes < 60) {
    return `${absMinutes}m ${minutes <= 0 ? "ago" : "from now"}`;
  }

  const hours = Math.round(absMinutes / 60);
  if (hours < 24) {
    return `${hours}h ${minutes <= 0 ? "ago" : "from now"}`;
  }

  const days = Math.round(hours / 24);
  return `${days}d ${minutes <= 0 ? "ago" : "from now"}`;
}

export function formatNumber(value?: number | null) {
  if (value === undefined || value === null || Number.isNaN(value)) {
    return "0";
  }

  return numberFormatter.format(value);
}

export function formatPercent(value?: number | null) {
  if (value === undefined || value === null || Number.isNaN(value)) {
    return "0%";
  }

  return `${numberFormatter.format(value)}%`;
}

export function formatDurationMs(value?: number | null) {
  if (value === undefined || value === null || Number.isNaN(value)) {
    return "0 ms";
  }

  if (value >= 1000) {
    return `${numberFormatter.format(value / 1000)} s`;
  }

  return `${numberFormatter.format(value)} ms`;
}

export function severityLabel(value?: string | null) {
  if (!value) {
    return "Info";
  }

  return value.charAt(0).toUpperCase() + value.slice(1).toLowerCase();
}

export function severityClasses(value?: string | null) {
  switch ((value || "").toLowerCase()) {
    case "critical":
      return {
        badge: "border-[#111111] bg-[#ffd7d2] text-[#a31616]",
        accent: "bg-[#dc2626]",
        text: "text-[#a31616]",
      };
    case "warning":
      return {
        badge: "border-[#111111] bg-[#fff1b8] text-[#8a4b00]",
        accent: "bg-[#b45309]",
        text: "text-[#8a4b00]",
      };
    default:
      return {
        badge: "border-[#111111] bg-[#dce8ff] text-[#254fd2]",
        accent: "bg-[#254fd2]",
        text: "text-[#254fd2]",
      };
  }
}

export function statusClasses(value?: string | null) {
  switch ((value || "").toLowerCase()) {
    case "running":
    case "open":
    case "active":
    case "pending":
    case "scheduled":
      return "border-[#111111] bg-[#EAF2FF] text-[#2F5FD0]";
    case "critical":
    case "degraded":
      return "border-[#111111] bg-[#ffd7d2] text-[#a31616]";
    case "warning":
      return "border-[#111111] bg-[#fff1b8] text-[#8a4b00]";
    case "healthy":
      return "border-[#111111] bg-[#d9f7d8] text-[#166534]";
    default:
      return "border-[#111111] bg-white text-slate-700";
  }
}
