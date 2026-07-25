import { Badge, type BadgeProps } from "../ui/badge";

type Variant = NonNullable<BadgeProps["variant"]>;

const ROLE_VARIANT: Record<string, Variant> = {
  owner: "violet",
  admin: "sky",
  member: "gray",
};

export function RoleBadge({
  role,
  className,
}: {
  role: string;
  className?: string;
}) {
  const key = role.toLowerCase();
  const variant = ROLE_VARIANT[key] ?? "gray";
  const label = role.charAt(0).toUpperCase() + role.slice(1);
  return (
    <Badge variant={variant} className={className}>
      {label}
    </Badge>
  );
}
