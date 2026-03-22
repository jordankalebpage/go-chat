import { useCallback, useEffect, useRef, useState } from "react";

import type { ConnectionStatus, ServerMessage } from "../types";

const reconnectDelays = [1000, 2000, 5000, 10000];
const maxMessages = 200;

interface UseWebSocketOptions {
  room: string;
  username: string;
}

interface UseWebSocketResult {
  error: string | null;
  messages: ServerMessage[];
  sendMessage: (content: string) => boolean;
  status: ConnectionStatus;
  users: string[];
}

export function useWebSocket({
  room,
  username,
}: UseWebSocketOptions): UseWebSocketResult {
  const [error, setError] = useState<string | null>(null);
  const [messages, setMessages] = useState<ServerMessage[]>([]);
  const [status, setStatus] = useState<ConnectionStatus>(() => {
    if (room === "" || username === "") {
      return "idle";
    }

    return "connecting";
  });
  const [users, setUsers] = useState<string[]>([]);

  const socketRef = useRef<WebSocket | null>(null);

  const sendMessage = useCallback((content: string) => {
    const socket = socketRef.current;
    const trimmedContent = content.trim();

    if (trimmedContent === "") {
      return false;
    }

    if (socket === null || socket.readyState !== WebSocket.OPEN) {
      return false;
    }

    socket.send(
      JSON.stringify({
        type: "message",
        content: trimmedContent,
      }),
    );

    return true;
  }, []);

  useEffect(() => {
    if (room === "" || username === "") {
      return undefined;
    }

    let closedByEffect = false;
    let reconnectAttempt = 0;
    let reconnectTimer: number | undefined;

    const clearTimer = () => {
      if (reconnectTimer === undefined) {
        return;
      }

      window.clearTimeout(reconnectTimer);
      reconnectTimer = undefined;
    };

    const connect = () => {
      clearTimer();

      const protocol = window.location.protocol === "https:" ? "wss" : "ws";
      const socketUrl = `${protocol}://${window.location.host}/ws?room=${encodeURIComponent(room)}&username=${encodeURIComponent(username)}`;

      setStatus(reconnectAttempt === 0 ? "connecting" : "reconnecting");

      const socket = new WebSocket(socketUrl);
      socketRef.current = socket;

      socket.addEventListener("open", () => {
        reconnectAttempt = 0;
        setError(null);
        setStatus("connected");
      });

      socket.addEventListener("message", (event) => {
        const parsed = parseServerMessage(event.data);
        if (parsed === null) {
          setError("Received an invalid message from the server.");
          return;
        }

        if (parsed.users !== undefined) {
          setUsers(parsed.users);
        }

        setMessages((currentMessages) => {
          const nextMessages = [...currentMessages, parsed];
          return nextMessages.slice(-maxMessages);
        });
      });

      socket.addEventListener("error", () => {
        setError("The WebSocket connection ran into an error.");
        setStatus("error");
      });

      socket.addEventListener("close", () => {
        if (closedByEffect) {
          setStatus("idle");
          return;
        }

        setStatus("disconnected");

        const delay =
          reconnectDelays[
            Math.min(reconnectAttempt, reconnectDelays.length - 1)
          ];
        reconnectAttempt += 1;

        reconnectTimer = window.setTimeout(() => {
          connect();
        }, delay);
      });
    };

    connect();

    return () => {
      closedByEffect = true;
      clearTimer();

      if (socketRef.current !== null) {
        socketRef.current.close();
        socketRef.current = null;
      }
    };
  }, [room, username]);

  return {
    error,
    messages,
    sendMessage,
    status,
    users,
  };
}

function parseServerMessage(value: unknown): ServerMessage | null {
  if (typeof value !== "string") {
    return null;
  }

  try {
    const parsed = JSON.parse(value) as Partial<ServerMessage>;

    if (
      parsed.type !== "join" &&
      parsed.type !== "leave" &&
      parsed.type !== "message"
    ) {
      return null;
    }

    return {
      type: parsed.type,
      room: typeof parsed.room === "string" ? parsed.room : undefined,
      username:
        typeof parsed.username === "string" ? parsed.username : undefined,
      content: typeof parsed.content === "string" ? parsed.content : undefined,
      users: Array.isArray(parsed.users)
        ? parsed.users.filter(
            (entry): entry is string => typeof entry === "string",
          )
        : undefined,
      timestamp:
        typeof parsed.timestamp === "string" ? parsed.timestamp : undefined,
    };
  } catch {
    return null;
  }
}
