import { Suspense } from "react";
import VerifyEmailForm from "./verify-email-form";

export default function VerifyEmailPage() {
  return (
    <Suspense
      fallback={
        <div className="w-full max-w-md rounded-md border border-(--line) bg-(--white) p-8 text-center text-sm text-(--muted-clr)">
          …
        </div>
      }
    >
      <VerifyEmailForm />
    </Suspense>
  );
}
