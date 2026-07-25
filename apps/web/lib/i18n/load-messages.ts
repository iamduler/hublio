import type { Locale } from "./config";
import {
  FEATURE_NAMESPACES,
  type FeatureNamespace,
  type MessageNamespace,
} from "./namespaces";

export type FeatureMessages = {
  [K in FeatureNamespace]: Record<string, unknown>;
};

export type Messages = FeatureMessages;

async function loadFeatureNamespace(
  locale: Locale,
  namespace: FeatureNamespace,
): Promise<Record<string, unknown>> {
  const mod = await import(`@/messages/${locale}/${namespace}.json`);
  return mod.default as Record<string, unknown>;
}

export async function loadMessages(locale: Locale): Promise<Messages> {
  const featureEntries = await Promise.all(
    FEATURE_NAMESPACES.map(async (namespace) => {
      const messages = await loadFeatureNamespace(locale, namespace);
      return [namespace, messages] as const;
    }),
  );

  return Object.fromEntries(featureEntries) as FeatureMessages;
}

export async function loadNamespaces(
  locale: Locale,
  namespaces: readonly MessageNamespace[],
): Promise<Partial<Messages>> {
  const result: Partial<Messages> = {};
  const featureEntries = await Promise.all(
    namespaces.map(async (namespace) => {
      const messages = await loadFeatureNamespace(locale, namespace);
      return [namespace, messages] as const;
    }),
  );

  for (const [namespace, messages] of featureEntries) {
    result[namespace] = messages;
  }

  return result;
}
