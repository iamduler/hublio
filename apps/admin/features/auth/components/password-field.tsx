"use client";

import { forwardRef, useState } from "react";
import { Eye, EyeOff } from "lucide-react";
import { useTranslations } from "next-intl";
import { Input } from "@hublio/ui/ui/input";
import { Label } from "@hublio/ui/ui/label";
import { cn } from "@/lib/utils";

export type PasswordFieldProps = Omit<
  React.InputHTMLAttributes<HTMLInputElement>,
  "type"
> & {
  label?: string;
  error?: string;
};

/** Shared password input with show/hide toggle. Works with react-hook-form via `register`. */
export const PasswordField = forwardRef<HTMLInputElement, PasswordFieldProps>(
  function PasswordField(
    {
      id,
      label,
      error,
      className,
      autoComplete = "current-password",
      disabled,
      ...props
    },
    ref,
  ) {
    const t = useTranslations("auth.passwordField");
    const [show, setShow] = useState(false);

    return (
      <div className="space-y-2">
        {label ? (
          <Label htmlFor={id}>
            {label}
            <span className="ml-0.5 text-destructive">*</span>
          </Label>
        ) : null}
        <div className="relative">
          <Input
            id={id}
            ref={ref}
            type={show ? "text" : "password"}
            autoComplete={autoComplete}
            disabled={disabled}
            placeholder="••••••••"
            aria-invalid={Boolean(error) || undefined}
            className={cn("pr-10", error && "border-destructive", className)}
            {...props}
          />
          <button
            type="button"
            tabIndex={-1}
            disabled={disabled}
            aria-label={show ? t("hide") : t("show")}
            onClick={() => setShow((v) => !v)}
            className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground disabled:opacity-50"
          >
            {show ? <EyeOff size={14} /> : <Eye size={14} />}
          </button>
        </div>
        {error ? (
          <p className="text-sm text-destructive" role="alert">
            {error}
          </p>
        ) : null}
      </div>
    );
  },
);
