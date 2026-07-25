"use client";

import { useState } from "react";
import { useTranslations } from "next-intl";
import { Button } from "@hublio/ui/ui/button";
import { Input } from "@hublio/ui/ui/input";
import { Label } from "@hublio/ui/ui/label";
import { Card, CardContent } from "@hublio/ui/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@hublio/ui/ui/select";
import { toast } from "@/lib/toast";
import { useApiErrorMessage } from "@/hooks/use-api-error";
import { useAuth } from "@/providers/auth-provider";
import { useWorkspace } from "@/providers/workspace-provider";
import { useRouter } from "@/i18n/navigation";
import { useCreateWorkspace } from "../hooks";

const ENVIRONMENTS = ["production", "staging", "development"];

export function WorkspaceForm() {
  const t = useTranslations("workspaces");
  const getError = useApiErrorMessage();
  const router = useRouter();
  const { user } = useAuth();
  const { setActiveWorkspace } = useWorkspace();
  const createWorkspace = useCreateWorkspace(user?.organization_id);

  const [name, setName] = useState("");
  const [environment, setEnvironment] = useState("production");

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim()) return;
    try {
      const workspace = await createWorkspace.mutateAsync({
        name: name.trim(),
        environment,
      });
      setActiveWorkspace(workspace.id);
      toast.success(t("created"));
      router.replace("/dashboard");
    } catch (err) {
      toast.error(getError(err));
    }
  }

  return (
    <Card>
      <CardContent className="p-6">
        <form className="space-y-5" onSubmit={(e) => void onSubmit(e)}>
          <div className="space-y-2">
            <Label htmlFor="ws-name">{t("form.name")}</Label>
            <Input
              id="ws-name"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>
          <div className="space-y-2">
            <Label>{t("form.environment")}</Label>
            <Select value={environment} onValueChange={setEnvironment}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {ENVIRONMENTS.map((env) => (
                  <SelectItem key={env} value={env}>
                    {env}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="flex justify-end gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => router.back()}
            >
              {t("form.cancel")}
            </Button>
            <Button
              type="submit"
              disabled={!name.trim() || createWorkspace.isPending}
            >
              {t("form.submit")}
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
