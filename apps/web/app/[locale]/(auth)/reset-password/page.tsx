import { Suspense } from "react";
import ResetPasswordForm from "./reset-password-form";

export default function ResetPasswordPage() {
  return (
    <Suspense
      fallback={
        <div className="w-full max-w-md rounded-md border border-[var(--line)] bg-[var(--white)] p-8 text-center text-sm text-[var(--muted-clr)]">
          …
        </div>
      }
    >
      <ResetPasswordForm />
    </Suspense>
  );
}
