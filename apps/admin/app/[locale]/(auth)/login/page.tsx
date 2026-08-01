import { Suspense } from "react";
import { LoginForm } from "@/features/auth/components/login-form";
import { LoadingState } from "@hublio/ui/ui/loading-state";

export default function LoginPage() {
  return (
    <Suspense
      fallback={
        <div className="w-full max-w-md">
          <LoadingState rows={4} />
        </div>
      }
    >
      <LoginForm />
    </Suspense>
  );
}
