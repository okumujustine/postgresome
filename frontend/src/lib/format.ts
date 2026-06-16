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
        badge: "border-red-200 bg-red-50 text-red-700",
        accent: "bg-red-500",
        text: "text-red-700",
      };
    case "warning":
      return {
        badge: "border-amber-200 bg-amber-50 text-amber-700",
        accent: "bg-amber-500",
        text: "text-amber-700",
      };
    default:
      return {
        badge: "border-blue-200 bg-blue-50 text-blue-700",
        accent: "bg-blue-500",
        text: "text-blue-700",
      };
  }
}

export function statusClasses(value?: string | null) {
  switch ((value || "").toLowerCase()) {
    case "critical":
    case "degraded":
      return "border-red-200 bg-red-50 text-red-700";
    case "warning":
      return "border-amber-200 bg-amber-50 text-amber-700";
    case "healthy":
      return "border-emerald-200 bg-emerald-50 text-emerald-700";
    default:
      return "border-slate-200 bg-slate-50 text-slate-600";
  }
}

