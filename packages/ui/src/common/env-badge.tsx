import { Badge, type BadgeProps } from "../ui/badge";

type Variant = NonNullable<BadgeProps["variant"]>;

const ENV_VARIANT: Record<string, Variant> = {
  production: "green",
  prod: "green",
  staging: "amber",
  development: "sky",
  dev: "sky",
  sandbox: "violet",
};

export function EnvBadge({
  environment,
  className,
}: {
  environment: string;
  className?: string;
}) {
  const key = environment.toLowerCase();
  const variant = ENV_VARIANT[key] ?? "gray";
  return (
    <Badge variant={variant} className={className}>
      {environment}
    </Badge>
  );
}
