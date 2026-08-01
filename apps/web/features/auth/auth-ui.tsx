"use client";

import { useRef, useState } from "react";
import Image from "next/image";
import { Check, Eye, EyeOff } from "lucide-react";
import { twMerge } from "tailwind-merge";

export function cx(...classes: Array<string | undefined | false | null>) {
  return twMerge(classes.filter(Boolean).join(" "));
}

export const inputBase =
  "w-full h-9 px-3 rounded-lg text-sm text-(--ink) placeholder:text-(--faint) border bg-(--white) transition-colors duration-150 focus:outline-none focus:ring-2 focus:ring-blue-600/20 focus:border-blue-500 disabled:opacity-50 disabled:cursor-not-allowed";

const LOGO_HEIGHT: Record<"sm" | "md" | "lg" | "xl" | "xxl", number> = {
  sm: 28,
  md: 36,
  lg: 52,
  xl: 64,
  xxl: 105,
};

/** Full wordmark + tagline. Transparent PNG — always on a white plate. */
export function Logo({
  size = "md",
  className,
  priority = false,
}: {
  size?: "sm" | "md" | "lg" | "xl" | "xxl";
  className?: string;
  priority?: boolean;
}) {
  const height = LOGO_HEIGHT[size];
  const width = Math.round(height * (564 / 182));
  const paddingClass = {
    sm: "px-2 py-1.5",
    md: "px-4 py-3",
    lg: "px-6 py-4",
    xl: "px-8 py-6",
    xxl: "px-10 py-8",
  }

  return (
    <div
      className={cx(
        "inline-flex items-center justify-center rounded-sm bg-white",
        paddingClass[size],
        className,
      )}
    >
      <Image
        src="/logo/logo.png"
        alt="Hublio"
        width={width}
        height={height}
        className="h-auto w-auto max-w-full"
        style={{ height, width: "auto" }}
        priority={priority}
      />
    </div>
  );
}

/** Inner auth card only — page background comes from auth layout (theme-aware). */
export function AuthCard({ children }: { children: React.ReactNode }) {
  return (
    <div className="w-full max-w-105">
      <div
        className="w-full rounded-xl border border-(--line) bg-(--white) p-8"
        style={{
          boxShadow:
            "0 1px 3px rgba(0,0,0,0.07), 0 4px 20px rgba(0,0,0,0.05)",
        }}
      >
        {children}
      </div>
      <p className="mt-5 text-center text-[11px] text-(--faint)">
        © {new Date().getFullYear()} Hublio
      </p>
    </div>
  );
}

export function OnboardingShell({
  children,
  step,
}: {
  children: React.ReactNode;
  step: number;
}) {
  const steps = ["Organization", "Workspace", "Invite team", "Complete"];
  return (
    <div className="flex w-full flex-col items-center justify-center px-4 py-8">
      <div className="w-full max-w-lg">
        <div className="mb-8 flex justify-center">
          <Logo size="lg" priority />
        </div>
        <div className="rounded-2xl border border-(--line) bg-(--white) px-8 py-8 shadow-sm">
          <div className="mb-8 flex items-center gap-0">
            {steps.map((label, i) => {
              const done = i < step;
              const active = i === step;
              const last = i === steps.length - 1;
              return (
                <div key={label} className="flex flex-1 items-center last:flex-none">
                  <div className="flex flex-shrink-0 flex-col items-center gap-1.5">
                    <div
                      className={cx(
                        "flex h-7 w-7 items-center justify-center rounded-full border-2 text-[11px] font-bold transition-all",
                        done
                          ? "border-blue-600 bg-blue-600 text-white"
                          : active
                            ? "border-blue-600 bg-(--white) text-blue-600"
                            : "border-(--line) bg-(--white) text-(--faint)",
                      )}
                    >
                      {done ? "✓" : i + 1}
                    </div>
                    <span
                      className={cx(
                        "whitespace-nowrap text-[11px] font-medium",
                        active
                          ? "text-blue-600"
                          : done
                            ? "text-(--ink-2)"
                            : "text-(--faint)",
                      )}
                    >
                      {label}
                    </span>
                  </div>
                  {!last && (
                    <div
                      className={cx(
                        "mx-2 mt-[-14px] h-px flex-1",
                        done ? "bg-blue-600" : "bg-(--line)",
                      )}
                    />
                  )}
                </div>
              );
            })}
          </div>
          {children}
        </div>
      </div>
    </div>
  );
}

export function GoogleIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" aria-hidden="true">
      <path
        d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
        fill="#4285F4"
      />
      <path
        d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
        fill="#34A853"
      />
      <path
        d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z"
        fill="#FBBC05"
      />
      <path
        d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
        fill="#EA4335"
      />
    </svg>
  );
}

export function MicrosoftIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 23 23" aria-hidden="true">
      <rect x="1" y="1" width="10" height="10" fill="#F25022" />
      <rect x="12" y="1" width="10" height="10" fill="#7FBA00" />
      <rect x="1" y="12" width="10" height="10" fill="#00A4EF" />
      <rect x="12" y="12" width="10" height="10" fill="#FFB900" />
    </svg>
  );
}

export function GitHubIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
      <path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z" />
    </svg>
  );
}

export function SSOButton({
  provider,
  label,
  onClick,
  disabled,
}: {
  provider: "google" | "microsoft" | "github";
  label: string;
  onClick: () => void;
  disabled?: boolean;
}) {
  const icon =
    provider === "google" ? (
      <GoogleIcon />
    ) : provider === "microsoft" ? (
      <MicrosoftIcon />
    ) : (
      <GitHubIcon />
    );
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={disabled}
      className="flex h-9 w-full items-center justify-center gap-2.5 rounded-lg border border-(--line) bg-(--white) text-[13px] font-medium text-(--ink-2) transition-colors duration-150 hover:border-(--line-3) hover:bg-(--line-2) focus:outline-none focus:ring-2 focus:ring-slate-400/20 disabled:cursor-not-allowed disabled:opacity-50"
    >
      {icon}
      {label}
    </button>
  );
}

export function AuthDivider({ label }: { label: string }) {
  return (
    <div className="my-1 flex items-center gap-3">
      <div className="h-px flex-1 bg-(--line)" />
      <span className="text-[11px] font-medium text-(--faint)">{label}</span>
      <div className="h-px flex-1 bg-(--line)" />
    </div>
  );
}

export function AuthFieldLabel({
  htmlFor,
  children,
  required,
}: {
  htmlFor?: string;
  children: React.ReactNode;
  required?: boolean;
}) {
  return (
    <label
      htmlFor={htmlFor}
      className="mb-1.5 block text-[13px] font-medium text-(--ink-2)"
    >
      {children}
      {required ? <span className="ml-0.5 text-red-500">*</span> : null}
    </label>
  );
}

export function PasswordField({
  id,
  label,
  error,
  autoComplete = "current-password",
  ...props
}: React.InputHTMLAttributes<HTMLInputElement> & {
  label?: string;
  error?: string;
}) {
  const [show, setShow] = useState(false);
  return (
    <div>
      {label ? (
        <AuthFieldLabel htmlFor={id} required>
          {label}
        </AuthFieldLabel>
      ) : null}
      <div className="relative">
        <input
          id={id}
          type={show ? "text" : "password"}
          autoComplete={autoComplete}
          placeholder="••••••••"
          className={cx(
            inputBase,
            "pr-10",
            error
              ? "border-red-300 bg-red-50/40"
              : "border-(--line) hover:border-(--line-3)",
          )}
          {...props}
        />
        <button
          type="button"
          tabIndex={-1}
          onClick={() => setShow((v) => !v)}
          className="absolute right-2.5 top-1/2 -translate-y-1/2 text-(--faint) hover:text-(--ink-2)"
        >
          {show ? <EyeOff size={14} /> : <Eye size={14} />}
        </button>
      </div>
      {error ? <p className="mt-1 text-[12px] text-red-600">{error}</p> : null}
    </div>
  );
}

function passwordStrength(pw: string) {
  if (!pw) return null;
  const score = [
    pw.length >= 8,
    /[A-Z]/.test(pw),
    /[a-z]/.test(pw),
    /[0-9]/.test(pw),
    /[^A-Za-z0-9]/.test(pw),
  ].filter(Boolean).length;
  if (score <= 1) return { score: 1, label: "Weak", color: "bg-red-500" };
  if (score === 2) return { score: 2, label: "Fair", color: "bg-amber-400" };
  if (score === 3) return { score: 3, label: "Good", color: "bg-yellow-400" };
  if (score === 4) return { score: 4, label: "Strong", color: "bg-blue-500" };
  return { score: 5, label: "Very strong", color: "bg-green-500" };
}

export function PasswordStrengthMeter({ password }: { password: string }) {
  const s = passwordStrength(password);
  if (!s) return null;
  return (
    <div className="mt-2">
      <div className="mb-1 flex gap-1">
        {[1, 2, 3, 4, 5].map((n) => (
          <div
            key={n}
            className={cx(
              "h-1 flex-1 rounded-full transition-all duration-300",
              n <= s.score ? s.color : "bg-(--line)",
            )}
          />
        ))}
      </div>
      <p className="text-[11px] text-(--muted-clr)">{s.label}</p>
    </div>
  );
}

export function PasswordRequirements({
  password,
  labels,
}: {
  password: string;
  labels: {
    minLength: string;
    uppercase: string;
    lowercase: string;
    number: string;
    special: string;
  };
}) {
  const reqs = [
    { label: labels.minLength, met: password.length >= 8 },
    { label: labels.uppercase, met: /[A-Z]/.test(password) },
    { label: labels.lowercase, met: /[a-z]/.test(password) },
    { label: labels.number, met: /[0-9]/.test(password) },
    { label: labels.special, met: /[^A-Za-z0-9]/.test(password) },
  ];
  return (
    <ul className="mt-3 space-y-1.5">
      {reqs.map((r) => (
        <li key={r.label} className="flex items-center gap-2">
          <div
            className={cx(
              "flex h-3.5 w-3.5 flex-shrink-0 items-center justify-center rounded-full transition-colors",
              r.met ? "bg-green-100" : "bg-(--line-2)",
            )}
          >
            <Check
              size={8}
              strokeWidth={3}
              className={r.met ? "text-green-600" : "text-(--faint)"}
            />
          </div>
          <span
            className={cx(
              "text-[12px]",
              r.met ? "text-(--ink-2)" : "text-(--faint)",
            )}
          >
            {r.label}
          </span>
        </li>
      ))}
    </ul>
  );
}

export function OTPInput({
  value,
  onChange,
}: {
  value: string[];
  onChange: (v: string[]) => void;
}) {
  const refs = useRef<(HTMLInputElement | null)[]>([]);

  function handleChange(idx: number, raw: string) {
    const digit = raw.replace(/\D/g, "").slice(-1);
    const next = [...value];
    next[idx] = digit;
    onChange(next);
    if (digit && idx < 5) refs.current[idx + 1]?.focus();
  }

  function handleKeyDown(idx: number, e: React.KeyboardEvent) {
    if (e.key === "Backspace" && !value[idx] && idx > 0) {
      refs.current[idx - 1]?.focus();
    }
  }

  function handlePaste(e: React.ClipboardEvent) {
    e.preventDefault();
    const pasted = e.clipboardData.getData("text").replace(/\D/g, "").slice(0, 6);
    const next = [...value];
    pasted.split("").forEach((d, i) => {
      next[i] = d;
    });
    onChange(next);
    refs.current[Math.min(pasted.length, 5)]?.focus();
  }

  return (
    <div className="flex justify-center gap-2.5">
      {Array.from({ length: 6 }).map((_, i) => (
        <input
          key={i}
          ref={(el) => {
            refs.current[i] = el;
          }}
          type="text"
          inputMode="numeric"
          maxLength={1}
          value={value[i] || ""}
          onChange={(e) => handleChange(i, e.target.value)}
          onKeyDown={(e) => handleKeyDown(i, e)}
          onPaste={handlePaste}
          className={cx(
            "w-12 rounded-xl border bg-(--white) text-center font-mono text-[18px] font-semibold text-(--ink) transition-all focus:border-blue-500 focus:outline-none focus:ring-2 focus:ring-blue-600/20",
            value[i]
              ? "border-blue-300 bg-blue-50/60"
              : "border-(--line) hover:border-(--line-3)",
          )}
          style={{ height: 52 }}
        />
      ))}
    </div>
  );
}

export function CheckboxField({
  id,
  checked,
  onChange,
  label,
  error,
}: {
  id: string;
  checked: boolean;
  onChange: (v: boolean) => void;
  label: React.ReactNode;
  error?: string;
}) {
  return (
    <div>
      <label htmlFor={id} className="group flex cursor-pointer items-start gap-2.5">
        <div className="relative mt-0.5 flex-shrink-0">
          <input
            id={id}
            type="checkbox"
            checked={checked}
            onChange={(e) => onChange(e.target.checked)}
            className="sr-only"
          />
          <div
            className={cx(
              "flex h-4 w-4 items-center justify-center rounded-[4px] border-[1.5px] transition-all",
              checked
                ? "border-blue-600 bg-blue-600"
                : "border-(--line-3) bg-(--white) group-hover:border-(--faint)",
            )}
          >
            {checked ? <Check size={9} strokeWidth={3} className="text-white" /> : null}
          </div>
        </div>
        <span className="text-[13px] leading-snug text-(--ink-2) select-none">
          {label}
        </span>
      </label>
      {error ? <p className="mt-1 text-[12px] text-red-600">{error}</p> : null}
    </div>
  );
}
