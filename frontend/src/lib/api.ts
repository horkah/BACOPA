const BASE_URL = process.env.NEXT_PUBLIC_API_URL || "";

function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("token");
}

async function request<T>(
  path: string,
  options: RequestInit = {}
): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  const res = await fetch(`${BASE_URL}${path}`, {
    ...options,
    headers,
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({ message: res.statusText }));
    throw new Error(body.message || body.error || `Request failed: ${res.status}`);
  }

  return res.json();
}

export interface User {
  id: string;
  username: string;
  email: string;
  elo: number;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface GameType {
  id: string;
  name: string;
  description: string;
  minPlayers: number;
  maxPlayers: number;
}

export interface LobbyGame {
  gameId: string;
  gameType: string;
  mode: string;
  creator: string;
  status: string;
  createdAt: string;
}

export interface GameHistoryEntry {
  gameId: string;
  gameType: string;
  mode: string;
  opponent: string;
  result: string;
  eloChange: number;
  playedAt: string;
}

export interface CreateGameResponse {
  gameId: string;
  gameType: string;
  mode: string;
  status: string;
  creator: string;
}

export interface JoinGameResponse {
  gameId: string;
  gameType: string;
  status: string;
}

export function register(
  username: string,
  email: string,
  password: string
): Promise<AuthResponse> {
  return request<AuthResponse>("/api/auth/register", {
    method: "POST",
    body: JSON.stringify({ username, email, password }),
  });
}

export function login(email: string, password: string): Promise<AuthResponse> {
  return request<AuthResponse>("/api/auth/login", {
    method: "POST",
    body: JSON.stringify({ email, password }),
  });
}

export function getMe(): Promise<User> {
  return request<User>("/api/auth/me");
}

export function getGameTypes(): Promise<GameType[]> {
  return request<GameType[]>("/api/games/types");
}

export function createGame(
  gameType: string,
  mode: "pvp" | "ai",
  aiDifficulty?: "easy" | "medium" | "hard"
): Promise<CreateGameResponse> {
  return request<CreateGameResponse>("/api/games", {
    method: "POST",
    body: JSON.stringify({ gameType, mode, aiDifficulty }),
  });
}

export function getLobbyGames(): Promise<LobbyGame[]> {
  return request<LobbyGame[]>("/api/games/lobby");
}

export function joinGame(gameId: string): Promise<JoinGameResponse> {
  return request<JoinGameResponse>(`/api/games/${gameId}/join`, {
    method: "POST",
  });
}

export function getGameHistory(): Promise<GameHistoryEntry[]> {
  return request<GameHistoryEntry[]>("/api/games/history");
}
