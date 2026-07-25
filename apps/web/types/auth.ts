export interface User {
  id: string;
  organization_id: string;
  email: string;
  full_name: string;
  status: string;
}

export interface AuthSession {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}
