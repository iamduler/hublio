"use client";

import { Suspense, useEffect, useRef, useState } from "react";
import { useSearchParams } from "next/navigation";
import { usePathname } from "@/i18n/navigation";

type ProgressPhase = "idle" | "start" | "loading" | "complete";

const SAFETY_MS = 8_000;
const POPSTATE_SETTLE_MS = 150;

function isModifiedClick(event: MouseEvent) {
  return (
    event.metaKey ||
    event.ctrlKey ||
    event.shiftKey ||
    event.altKey ||
    event.button !== 0
  );
}

function sameOriginInternalHref(anchor: HTMLAnchorElement): string | null {
  if (anchor.target && anchor.target !== "_self") return null;
  if (anchor.hasAttribute("download")) return null;
  const hrefAttr = anchor.getAttribute("href");
  if (!hrefAttr || hrefAttr.startsWith("#")) return null;

  let url: URL;
  try {
    url = new URL(anchor.href);
  } catch {
    return null;
  }

  if (url.origin !== window.location.origin) return null;

  return `${url.pathname}${url.search}`;
}

function NavigationProgressInner() {
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const [phase, setPhase] = useState<ProgressPhase>("idle");
  const [visible, setVisible] = useState(false);
  const activeRef = useRef(false);
  const hideTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const safetyTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const settleTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
  const urlKey = `${pathname}?${searchParams.toString()}`;

  function clearHideTimer() {
    if (hideTimerRef.current) {
      clearTimeout(hideTimerRef.current);
      hideTimerRef.current = null;
    }
  }

  function clearSafetyTimer() {
    if (safetyTimerRef.current) {
      clearTimeout(safetyTimerRef.current);
      safetyTimerRef.current = null;
    }
  }

  function clearSettleTimer() {
    if (settleTimerRef.current) {
      clearTimeout(settleTimerRef.current);
      settleTimerRef.current = null;
    }
  }

  function start() {
    if (activeRef.current) return;
    activeRef.current = true;
    clearHideTimer();
    clearSafetyTimer();
    clearSettleTimer();
    safetyTimerRef.current = setTimeout(() => {
      if (activeRef.current) done();
    }, SAFETY_MS);
    // Defer setState — Next.js may call history APIs inside useInsertionEffect,
    // which forbids scheduling React updates synchronously.
    queueMicrotask(() => {
      if (!activeRef.current) return;
      setVisible(true);
      setPhase("start");
      requestAnimationFrame(() => {
        requestAnimationFrame(() => {
          if (activeRef.current) setPhase("loading");
        });
      });
    });
  }

  function done() {
    if (!activeRef.current) return;
    activeRef.current = false;
    clearSafetyTimer();
    clearSettleTimer();
    setPhase("complete");
    hideTimerRef.current = setTimeout(() => {
      setVisible(false);
      setPhase("idle");
    }, 280);
  }

  useEffect(() => {
    done();
    // eslint-disable-next-line react-hooks/exhaustive-deps -- settle when URL changes
  }, [urlKey]);

  useEffect(() => {
    function onClick(event: MouseEvent) {
      if (isModifiedClick(event)) return;
      const target = event.target;
      if (!(target instanceof Element)) return;
      const anchor = target.closest("a");
      if (!(anchor instanceof HTMLAnchorElement)) return;
      if (anchor.dataset.progress === "false") return;

      const next = sameOriginInternalHref(anchor);
      if (!next) return;

      const current = `${window.location.pathname}${window.location.search}`;
      if (next === current) return;

      start();
    }

    function onPopState() {
      // browser back/forward — URL is already applied when popstate fires.
      // Path hooks may not re-emit the same urlKey, so force a short settle.
      start();
      clearSettleTimer();
      settleTimerRef.current = setTimeout(() => {
        if (activeRef.current) done();
      }, POPSTATE_SETTLE_MS);
    }

    const originalPush = history.pushState.bind(history);

    history.pushState = function (...args) {
      const url = args[2];
      if (typeof url === "string") {
        try {
          const next = new URL(url, window.location.href);
          const current = `${window.location.pathname}${window.location.search}`;
          const target = `${next.pathname}${next.search}`;
          if (target !== current) start();
        } catch {
          start();
        }
      } else if (url != null) {
        start();
      }
      return originalPush(...args);
    };

    // Intentionally do not wrap replaceState — App Router syncs URLs via
    // replaceState inside useInsertionEffect; starting progress there warns
    // ("useInsertionEffect must not schedule updates") and flashes the bar.

    document.addEventListener("click", onClick, true);
    window.addEventListener("popstate", onPopState);

    return () => {
      clearHideTimer();
      clearSafetyTimer();
      clearSettleTimer();
      document.removeEventListener("click", onClick, true);
      window.removeEventListener("popstate", onPopState);
      history.pushState = originalPush;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  if (!visible && phase === "idle") return null;

  const width =
    phase === "start"
      ? "0%"
      : phase === "loading"
        ? "72%"
        : phase === "complete"
          ? "100%"
          : "0%";

  return (
    <div
      className="nav-progress"
      data-phase={phase}
      aria-hidden
      style={{ opacity: visible ? 1 : 0 }}
    >
      <div className="nav-progress__bar" style={{ width }} />
    </div>
  );
}

/** YouTube-style top progress bar during client navigations. */
export function NavigationProgress() {
  return (
    <Suspense fallback={null}>
      <NavigationProgressInner />
    </Suspense>
  );
}
