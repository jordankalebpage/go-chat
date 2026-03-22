import { useCallback, useEffect, useMemo, useState } from "react";
import type { FormEvent } from "react";

import { LiveChat } from "./components/LiveChat";
import { RoomList } from "./components/RoomList";

import type { RoomSummary } from "./types";

const fallbackRooms: RoomSummary[] = [
  { name: "general", memberCount: 0 },
  { name: "golang", memberCount: 0 },
  { name: "streaming", memberCount: 0 },
];

function App() {
  const [draftUsername, setDraftUsername] = useState("gopher");
  const [roomError, setRoomError] = useState<string | null>(null);
  const [rooms, setRooms] = useState<RoomSummary[]>(fallbackRooms);
  const [selectedRoom, setSelectedRoom] = useState("general");
  const [username, setUsername] = useState("gopher");

  const loadRooms = useCallback(async () => {
    try {
      const response = await fetch("/api/rooms");
      if (!response.ok) {
        throw new Error(`Request failed with status ${response.status}`);
      }

      const data = (await response.json()) as RoomSummary[];
      setRoomError(null);

      if (data.length === 0) {
        setRooms(fallbackRooms);
        return;
      }

      setRooms(ensureRoomExists(data, selectedRoom));
    } catch (error) {
      const message =
        error instanceof Error ? error.message : "Unable to load rooms.";
      setRoomError(message);
      setRooms((currentRooms) => ensureRoomExists(currentRooms, selectedRoom));
    }
  }, [selectedRoom]);

  useEffect(() => {
    void loadRooms();

    const interval = window.setInterval(() => {
      void loadRooms();
    }, 15000);

    return () => {
      window.clearInterval(interval);
    };
  }, [loadRooms]);

  const handleUsernameSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const trimmedUsername = draftUsername.trim();
    if (trimmedUsername === "") {
      return;
    }

    setUsername(trimmedUsername);
  };

  const visibleRooms = useMemo(
    () => ensureRoomExists(rooms, selectedRoom),
    [rooms, selectedRoom],
  );

  return (
    <main className="min-h-screen bg-slate-950 text-slate-100">
      <div className="mx-auto flex min-h-screen max-w-7xl flex-col gap-6 px-4 py-6 sm:px-6 lg:px-8">
        <section className="overflow-hidden rounded-[2rem] border border-white/10 bg-[radial-gradient(circle_at_top_left,_rgba(139,92,246,0.35),_transparent_35%),linear-gradient(180deg,_rgba(15,23,42,0.95),_rgba(2,6,23,0.98))] p-6 shadow-2xl shadow-slate-950/40 sm:p-8">
          <div className="grid gap-6 lg:grid-cols-[minmax(0,1fr)_24rem] lg:items-end">
            <div>
              <p className="text-xs font-semibold uppercase tracking-[0.35em] text-violet-300">
                Go Real-Time Chat
              </p>
              <h1 className="mt-4 text-4xl font-semibold tracking-tight text-white sm:text-5xl">
                Learn goroutines, channels, and WebSockets by building something
                deployable.
              </h1>
              <p className="mt-4 max-w-3xl text-sm leading-7 text-slate-300 sm:text-base">
                This app is intentionally small enough to understand and rich
                enough to teach the concurrency patterns that make Go popular in
                chat, streaming, and other high-throughput systems.
              </p>
            </div>

            <form
              className="rounded-[1.5rem] border border-white/10 bg-slate-950/70 p-4"
              onSubmit={handleUsernameSubmit}
            >
              <p className="text-xs font-semibold uppercase tracking-[0.3em] text-slate-400">
                Identity
              </p>
              <label
                className="mt-3 block text-sm font-medium text-slate-200"
                htmlFor="username"
              >
                Username
              </label>
              <input
                className="mt-2 w-full rounded-2xl border border-white/10 bg-slate-900 px-4 py-3 text-sm text-slate-100 outline-none transition placeholder:text-slate-500 focus:border-violet-400"
                id="username"
                maxLength={24}
                onChange={(event) => setDraftUsername(event.target.value)}
                placeholder="Pick a username"
                type="text"
                value={draftUsername}
              />
              <button
                className="mt-4 w-full rounded-2xl bg-violet-500 px-4 py-3 text-sm font-semibold text-white transition hover:bg-violet-400 disabled:cursor-not-allowed disabled:bg-slate-700 disabled:text-slate-400"
                disabled={draftUsername.trim() === ""}
                type="submit"
              >
                Connect as {draftUsername.trim() || "guest"}
              </button>
            </form>
          </div>

          {roomError !== null ? (
            <div className="mt-4 rounded-2xl border border-rose-400/20 bg-rose-500/10 px-4 py-3 text-sm text-rose-100">
              {roomError}
            </div>
          ) : null}
        </section>

        <div className="grid gap-6 lg:grid-cols-[18rem_minmax(0,1fr)]">
          <RoomList
            activeRoom={selectedRoom}
            onSelectRoom={setSelectedRoom}
            rooms={visibleRooms}
          />
          <LiveChat
            key={`${selectedRoom}:${username}`}
            room={selectedRoom}
            username={username}
          />
        </div>
      </div>
    </main>
  );
}

function ensureRoomExists(
  rooms: RoomSummary[],
  selectedRoom: string,
): RoomSummary[] {
  const hasSelectedRoom = rooms.some((room) => room.name === selectedRoom);
  if (hasSelectedRoom) {
    return rooms;
  }

  return [...rooms, { name: selectedRoom, memberCount: 0 }];
}

export default App;
