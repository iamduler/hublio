"use client";

import { AppToaster } from "../lib/toast";

/** Global toast host — library-specific props stay here. */
export function Toaster() {
  return <AppToaster position="bottom-right" richColors />;
}
