import type { ServerMessage } from "../types";

interface MessageListProps {
  messages: ServerMessage[];
}

export function MessageList({ messages }: MessageListProps) {
  if (messages.length === 0) {
    return (
      <div className="flex min-h-80 items-center justify-center rounded-3xl border border-dashed border-white/10 bg-slate-950/50 p-6 text-center text-sm text-slate-400">
        Join the room and send the first message. Join and leave events will
        show up here too, so you can see the hub pattern in action.
      </div>
    );
  }

  return (
    <ol className="flex min-h-80 flex-col gap-3 overflow-y-auto rounded-3xl border border-white/10 bg-slate-950/50 p-4">
      {messages.map((message, index) => {
        const isSystemMessage = message.type !== "message";

        return (
          <li
            className={`rounded-2xl border px-4 py-3 ${
              isSystemMessage
                ? "border-emerald-400/15 bg-emerald-500/10 text-emerald-100"
                : "border-white/8 bg-slate-900/80 text-slate-100"
            }`}
            key={`${message.timestamp ?? "no-time"}-${message.username ?? "anon"}-${index}`}
          >
            <div className="flex flex-col gap-2 sm:flex-row sm:items-baseline sm:justify-between">
              <div className="flex items-center gap-2">
                <span className="font-semibold">
                  {isSystemMessage
                    ? "System"
                    : (message.username ?? "Anonymous")}
                </span>
                {message.room !== undefined ? (
                  <span className="text-xs uppercase tracking-[0.2em] text-slate-400">
                    #{message.room}
                  </span>
                ) : null}
              </div>

              {message.timestamp !== undefined ? (
                <time
                  className="text-xs text-slate-500"
                  dateTime={message.timestamp}
                >
                  {formatTimestamp(message.timestamp)}
                </time>
              ) : null}
            </div>

            <p className="mt-2 whitespace-pre-wrap text-sm leading-6 text-inherit">
              {message.content}
            </p>
          </li>
        );
      })}
    </ol>
  );
}

function formatTimestamp(value: string): string {
  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
  });
}
