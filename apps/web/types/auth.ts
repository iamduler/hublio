export interface User {
  id: string;
  organization_id: string;
  email: string;
  full_name: string;
  status: string;
  is_platform_admin?: boolean;
}

export interface AuthSession {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}
