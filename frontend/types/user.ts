export interface CreateUserRequest {
  login: string;
  password: string;
}

export interface CreateUserResponse {
  id: string;
}

export interface UpdateUserRequest {
  password: string;
  name: string;
  avatar: string | null;
}

export interface AuthResponse {
  userId: string;
}

export interface User {
  id: string;
  name: string;
  avatar: string | null;
}

export interface DbUser extends User {
  login: string;
  password: string;
}

export interface Host extends User {
  isConnected: boolean;
}

export interface Player extends User {
  score: number;
  betAmount: number | null;
  isConnected: boolean;
}
