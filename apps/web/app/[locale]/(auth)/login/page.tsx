import { Suspense } from "react";
import { LoginForm } from "./login-form";

export default function LoginPage() {
  return (
    <Suspense
      fallback={
        <div className="w-full max-w-md rounded-md border border-(--line) bg-(--white) p-8 text-center text-sm text-(--muted-clr)">
          …
        </div>
      }
    >
      <LoginForm />
    </Suspense>
  );
}
