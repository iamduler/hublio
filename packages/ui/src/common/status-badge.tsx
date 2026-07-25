import { Badge, type BadgeProps } from "../ui/badge";

type Variant = NonNullable<BadgeProps["variant"]>;

/**
 * Maps the backend status enums (workspace, connector, connection, intent,
 * execution, sync-route, api-key, etc.) to a badge tone.
 */
const STATUS_VARIANT: Record<string, Variant> = {
  // healthy / positive
  active: "green",
  enabled: "green",
  succeeded: "green",
  success: "green",
  accepted: "green",
  verified: "green",
  published: "green",
  // in-progress / neutral-positive
  running: "sky",
  queued: "sky",
  verifying: "sky",
  created: "sky",
  submitted: "sky",
  pending: "amber",
  // warning
  draft: "gray",
  disabled: "gray",
  inactive: "gray",
  archived: "gray",
  expired: "amber",
  // negative
  failed: "danger",
  verification_failed: "danger",
  rejected: "danger",
  cancelled: "gray",
  dead_letter: "danger",
  suspended: "danger",
  revoked: "danger",
};

function humanize(status: string): string {
  return status.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

export function StatusBadge({
  status,
  className,
}: {
  status: string;
  className?: string;
}) {
  const key = status.toLowerCase();
  const variant = STATUS_VARIANT[key] ?? "gray";
  return (
    <Badge variant={variant} dot className={className}>
      {humanize(status)}
    </Badge>
  );
}
