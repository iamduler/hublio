"use client";

import { Suspense } from "react";
import { useSearchParams } from "next/navigation";
import { NavigationProgress as NavigationProgressUI } from "@hublio/ui/common/navigation-progress";
import { usePathname } from "@/i18n/navigation";

function NavigationProgressInner() {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const urlKey = `${pathname}?${searchParams.toString()}`;
  return <NavigationProgressUI urlKey={urlKey} />;
}

/** YouTube-style top progress bar during client navigations. */
export function NavigationProgress() {
  return (
    <Suspense fallback={null}>
      <NavigationProgressInner />
    </Suspense>
  );
}
