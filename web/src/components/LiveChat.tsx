import { ChatRoom } from "./ChatRoom";
import { useWebSocket } from "../hooks/useWebSocket";

interface LiveChatProps {
  room: string;
  username: string;
}

export function LiveChat({ room, username }: LiveChatProps) {
  const { error, messages, sendMessage, status, users } = useWebSocket({
    room,
    username,
  });

  return (
    <div className="space-y-4">
      {error !== null ? (
        <div className="rounded-2xl border border-rose-400/20 bg-rose-500/10 px-4 py-3 text-sm text-rose-100">
          {error}
        </div>
      ) : null}

      <ChatRoom
        messages={messages}
        onSendMessage={sendMessage}
        room={room}
        status={status}
        users={users}
        username={username}
      />
    </div>
  );
}
