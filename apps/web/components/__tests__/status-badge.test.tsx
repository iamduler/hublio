import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import { StatusBadge } from "@hublio/ui/common/status-badge";

describe("StatusBadge", () => {
  const cases: Array<{ status: string; label: string }> = [
    { status: "active", label: "Active" },
    { status: "verification_failed", label: "Verification Failed" },
    { status: "dead_letter", label: "Dead Letter" },
  ];

  for (const { status, label } of cases) {
    it(`humanizes ${status}`, () => {
      render(<StatusBadge status={status} />);
      expect(screen.getByText(label)).toBeInTheDocument();
    });
  }
});
