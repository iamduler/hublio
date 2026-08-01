import { Suspense } from "react";
import { ResetPasswordForm } from "@/features/auth/components/reset-password-form";
import { LoadingState } from "@hublio/ui/ui/loading-state";

export default function ResetPasswordPage() {
  return (
    <Suspense
      fallback={
        <div className="w-full max-w-md">
          <LoadingState rows={4} />
        </div>
      }
    >
      <ResetPasswordForm />
    </Suspense>
  );
}
