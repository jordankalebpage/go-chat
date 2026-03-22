import { MessageInput } from "./MessageInput";
import { MessageList } from "./MessageList";
import { UserList } from "./UserList";

import type { ConnectionStatus, ServerMessage } from "../types";

interface ChatRoomProps {
  messages: ServerMessage[];
  onSendMessage: (content: string) => boolean;
  room: string;
  status: ConnectionStatus;
  users: string[];
  username: string;
}

const statusLabels: Record<ConnectionStatus, string> = {
  connected: "Connected",
  connecting: "Connecting",
  disconnected: "Disconnected",
  error: "Connection error",
  idle: "Idle",
  reconnecting: "Reconnecting",
};

export function ChatRoom({
  messages,
  onSendMessage,
  room,
  status,
  users,
  username,
}: ChatRoomProps) {
  const isReady =
    username !== "" && (status === "connected" || status === "reconnecting");

  return (
    <section className="rounded-[2rem] border border-white/10 bg-slate-900/70 p-4 shadow-2xl shadow-slate-950/30 sm:p-6">
      <div className="mb-4 flex flex-col gap-4 border-b border-white/10 pb-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.3em] text-violet-300">
            Live room
          </p>
          <h2 className="mt-2 text-2xl font-semibold text-white">#{room}</h2>
          <p className="mt-2 max-w-2xl text-sm leading-6 text-slate-400">
            This room makes Go&apos;s concurrency model visible: one hub
            goroutine routes events while each browser connection gets a
            dedicated read and write pump.
          </p>
        </div>

        <div className="flex items-center gap-3 text-sm">
          <span
            aria-hidden="true"
            className={`h-3 w-3 rounded-full ${
              status === "connected"
                ? "bg-emerald-400 shadow-[0_0_18px_rgba(52,211,153,0.8)]"
                : "bg-amber-400 shadow-[0_0_18px_rgba(251,191,36,0.8)]"
            }`}
          />
          <span className="font-medium text-slate-200">
            {statusLabels[status]}
          </span>
          <span className="text-slate-500">as {username || "guest"}</span>
        </div>
      </div>

      <div className="grid gap-4 lg:grid-cols-[minmax(0,1fr)_18rem]">
        <MessageList messages={messages} />
        <UserList users={users} />
      </div>

      <MessageInput disabled={!isReady} onSendMessage={onSendMessage} />
    </section>
  );
}
