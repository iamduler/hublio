import { Suspense } from "react";
import { LoginForm } from "./login-form";

export default function LoginPage() {
  return (
    <Suspense
      fallback={
        <div className="w-full max-w-md rounded-md border border-[var(--line)] bg-[var(--white)] p-8 text-center text-sm text-[var(--muted-clr)]">
          …
        </div>
      }
    >
      <LoginForm />
    </Suspense>
  );
}
