export interface User {
  id: string;
  organization_id: string;
  email: string;
  full_name: string;
  status: string;
  is_platform_admin?: boolean;
}

export interface Organization {
  id: string;
  name: string;
  status: string;
  created_at?: string;
  updated_at?: string;
}

export interface AuthSession {
  user: User | null;
  organization: Organization | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}
