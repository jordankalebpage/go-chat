export type MessageType = "join" | "leave" | "message";

export type ConnectionStatus =
  | "idle"
  | "connecting"
  | "connected"
  | "reconnecting"
  | "disconnected"
  | "error";

export interface ServerMessage {
  type: MessageType;
  room?: string;
  username?: string;
  content?: string;
  users?: string[];
  timestamp?: string;
}

export interface RoomSummary {
  name: string;
  memberCount: number;
}

export interface SessionState {
  requiresPassword: boolean;
  unlocked: boolean;
}
