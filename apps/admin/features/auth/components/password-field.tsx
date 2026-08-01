"use client";

import { useTranslations } from "next-intl";
import {
  PasswordField as PasswordFieldUI,
  type PasswordFieldProps as PasswordFieldUIProps,
} from "@hublio/ui/common/password-field";
import { forwardRef } from "react";

export type PasswordFieldProps = Omit<
  PasswordFieldUIProps,
  "showPasswordLabel" | "hidePasswordLabel"
>;

/** App wrapper: wires i18n show/hide labels onto shared PasswordField. */
export const PasswordField = forwardRef<HTMLInputElement, PasswordFieldProps>(
  function PasswordField(props, ref) {
    const t = useTranslations("auth.passwordField");
    return (
      <PasswordFieldUI
        ref={ref}
        showPasswordLabel={t("show")}
        hidePasswordLabel={t("hide")}
        {...props}
      />
    );
  },
);
