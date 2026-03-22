import type { RoomSummary } from "../types";

interface RoomListProps {
  activeRoom: string;
  onSelectRoom: (room: string) => void;
  rooms: RoomSummary[];
}

export function RoomList({ activeRoom, onSelectRoom, rooms }: RoomListProps) {
  return (
    <nav
      aria-label="Chat rooms"
      className="rounded-3xl border border-white/10 bg-slate-900/70 p-4 shadow-2xl shadow-slate-950/30"
    >
      <div className="mb-4 flex items-center justify-between">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.3em] text-violet-300">
            Rooms
          </p>
          <h2 className="mt-2 text-lg font-semibold text-white">
            Pick a channel
          </h2>
        </div>
      </div>

      <ul className="space-y-2">
        {rooms.map((room) => {
          const isActive = room.name === activeRoom;

          return (
            <li key={room.name}>
              <button
                aria-pressed={isActive}
                className={`flex w-full items-center justify-between rounded-2xl border px-4 py-3 text-left transition ${
                  isActive
                    ? "border-violet-400 bg-violet-500/15 text-white"
                    : "border-white/10 bg-slate-950/50 text-slate-300 hover:border-violet-400/40 hover:bg-slate-900"
                }`}
                onClick={() => onSelectRoom(room.name)}
                type="button"
              >
                <span className="font-medium">#{room.name}</span>
                <span className="rounded-full bg-white/8 px-2 py-1 text-xs text-slate-400">
                  {room.memberCount}
                </span>
              </button>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}
