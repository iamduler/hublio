import { Suspense } from "react";
import { MFAForm } from "./mfa-form";

export default function MFAPage() {
  return (
    <Suspense
      fallback={
        <div className="w-full max-w-md rounded-md border border-[var(--line)] bg-[var(--white)] p-8 text-center text-sm text-[var(--muted-clr)]">
          …
        </div>
      }
    >
      <MFAForm />
    </Suspense>
  );
}
