export type Organization = {
  id: string;
  name: string;
  status: "active" | "suspended" | "archived" | string;
  created_at?: string;
  updated_at?: string;
};
