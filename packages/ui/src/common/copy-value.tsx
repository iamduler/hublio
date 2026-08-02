"use client";

import * as React from "react";
import { Check, Copy } from "lucide-react";
import { Button } from "../ui/button";
import { cn } from "../lib/utils";

/**
 * Monospaced value with a copy-to-clipboard affordance.
 * Used for API keys, IDs, webhook secrets, correlation ids.
 */
export function CopyValue({
  value,
  masked,
  className,
}: {
  value: string;
  masked?: boolean;
  className?: string;
}) {
  const [copied, setCopied] = React.useState(false);

  async function copy() {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {
      // clipboard unavailable — ignore
    }
  }

  const display = masked ? maskValue(value) : value;

  return (
    <div
      className={cn(
        "flex items-center gap-2 rounded-md border border-(--line) bg-(--line-2) px-2.5 py-1.5",
        className,
      )}
    >
      <code className="flex-1 truncate text-xs text-(--ink-2)">
        {display}
      </code>
      <Button
        type="button"
        variant="ghost"
        size="icon-sm"
        onClick={() => void copy()}
        aria-label="Copy"
      >
        {copied ? <Check size={14} /> : <Copy size={14} />}
      </Button>
    </div>
  );
}

function maskValue(value: string): string {
  if (value.length <= 8) return "••••••••";
  return `${value.slice(0, 4)}••••${value.slice(-4)}`;
}
