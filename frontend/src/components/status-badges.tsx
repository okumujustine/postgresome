import { Badge } from './ui/badge';

export function SeverityBadge({ severity }: { severity: string }) {
  const key = severity.toLowerCase();
  if (key === 'critical') return <Badge variant="critical">Critical</Badge>;
  if (key === 'warning') return <Badge variant="warning">Warning</Badge>;
  if (key === 'info') return <Badge variant="info">Info</Badge>;
  return <Badge variant="neutral">{severity}</Badge>;
}

export function HealthBadge({ status }: { status: string }) {
  const key = status.toLowerCase();
  if (key === 'critical') return <Badge variant="critical">Attention Required</Badge>;
  if (key === 'warning') return <Badge variant="warning">Watch Closely</Badge>;
  if (key === 'healthy') return <Badge variant="success">Healthy</Badge>;
  return <Badge variant="neutral">Unknown</Badge>;
}

export function VerificationBadge({ status }: { status: string }) {
  switch (status) {
    case 'verified_fixed':
      return <Badge variant="success">Fix verified</Badge>;
    case 'improving':
      return <Badge variant="info">Improving</Badge>;
    case 'regressed':
      return <Badge variant="critical">Regressed</Badge>;
    default:
      return <Badge variant="neutral">Active</Badge>;
  }
}

