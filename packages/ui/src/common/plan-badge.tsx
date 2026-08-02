import { Badge, type BadgeProps } from "../ui/badge";

type Variant = NonNullable<BadgeProps["variant"]>;

const PLAN_VARIANT: Record<string, Variant> = {
  free: "gray",
  starter: "sky",
  growth: "violet",
  scale: "amber",
  enterprise: "green",
};

export function PlanBadge({
  plan,
  className,
}: {
  plan: string;
  className?: string;
}) {
  const key = plan.toLowerCase();
  const variant = PLAN_VARIANT[key] ?? "gray";
  return (
    <Badge variant={variant} className={className}>
      {plan}
    </Badge>
  );
}
