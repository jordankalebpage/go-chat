import { useMemo, useState } from "react";
import type { SubmitEvent } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";

import { LiveChat } from "./components/LiveChat";
import { RoomList } from "./components/RoomList";

import type { RoomSummary, SessionState } from "./types";

const fallbackRooms: RoomSummary[] = [
  { name: "general", memberCount: 0 },
  { name: "golang", memberCount: 0 },
  { name: "streaming", memberCount: 0 },
];

const sessionQueryKey = ["session"] as const;
const roomsQueryKey = ["rooms"] as const;

function App() {
  const [accessError, setAccessError] = useState<string | null>(null);
  const [accessPassword, setAccessPassword] = useState("");
  const [draftUsername, setDraftUsername] = useState("gopher");
  const [selectedRoom, setSelectedRoom] = useState("general");
  const [username, setUsername] = useState("gopher");
  const queryClient = useQueryClient();

  const sessionQuery = useQuery({
    queryKey: sessionQueryKey,
    queryFn: fetchSessionState,
    refetchInterval: 30000,
  });

  const sessionState = sessionQuery.data ?? null;
  const requiresPassword = sessionState?.requiresPassword ?? false;
  const isUnlocked = sessionState?.unlocked ?? false;
  const canLoadRooms =
    sessionState !== null && (!requiresPassword || isUnlocked);

  const roomsQuery = useQuery({
    queryKey: roomsQueryKey,
    queryFn: fetchRooms,
    enabled: canLoadRooms,
    refetchInterval: 15000,
  });

  const handleUsernameSubmit = (event: SubmitEvent<HTMLFormElement>) => {
    event.preventDefault();

    const trimmedUsername = draftUsername.trim();
    if (trimmedUsername === "") {
      return;
    }

    setUsername(trimmedUsername);
  };

  const handleAccessSubmit = async (event: SubmitEvent<HTMLFormElement>) => {
    event.preventDefault();

    const trimmedPassword = accessPassword.trim();
    if (trimmedPassword === "") {
      return;
    }

    try {
      const data = await unlockSession(trimmedPassword);

      queryClient.setQueryData<SessionState>(
        sessionQueryKey,
        (currentState) => ({
          requiresPassword: currentState?.requiresPassword ?? true,
          unlocked: data.unlocked,
        }),
      );

      setAccessError(null);
      setAccessPassword("");

      await Promise.all([
        queryClient.invalidateQueries({ queryKey: sessionQueryKey }),
        queryClient.invalidateQueries({ queryKey: roomsQueryKey }),
      ]);
    } catch (error) {
      setAccessError(getErrorMessage(error, "Unable to unlock demo."));
    }
  };

  const visibleRooms = useMemo(() => {
    if (!canLoadRooms) {
      return ensureRoomExists(fallbackRooms, selectedRoom);
    }

    const rooms = roomsQuery.data;
    if (rooms === undefined || rooms.length === 0) {
      return ensureRoomExists(fallbackRooms, selectedRoom);
    }

    return ensureRoomExists(rooms, selectedRoom);
  }, [canLoadRooms, roomsQuery.data, selectedRoom]);

  const visibleAccessError =
    accessError ??
    getQueryErrorMessage(sessionQuery.error, "Unable to load session.");
  const roomError = canLoadRooms
    ? getQueryErrorMessage(roomsQuery.error, "Unable to load rooms.")
    : null;
  const showAccessGate =
    sessionState === null || (requiresPassword && !isUnlocked);

  return (
    <main className="min-h-screen bg-slate-950 text-slate-100">
      <div className="mx-auto flex min-h-screen max-w-7xl flex-col gap-6 px-4 py-6 sm:px-6 lg:px-8">
        <section className="overflow-hidden rounded-4xl border border-white/10 bg-[radial-gradient(circle_at_top_left,rgba(139,92,246,0.35),transparent_35%),linear-gradient(180deg,rgba(15,23,42,0.95),rgba(2,6,23,0.98))] p-6 shadow-2xl shadow-slate-950/40 sm:p-8">
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

            <div className="space-y-4">
              <form
                className="rounded-3xl border border-white/10 bg-slate-950/70 p-4"
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

              {requiresPassword ? (
                <form
                  className="rounded-3xl border border-amber-300/20 bg-amber-500/10 p-4"
                  onSubmit={handleAccessSubmit}
                >
                  <p className="text-xs font-semibold uppercase tracking-[0.3em] text-amber-200">
                    Demo Access
                  </p>
                  <p className="mt-3 text-sm leading-6 text-amber-50/90">
                    This portfolio deployment is password-protected to keep the
                    demo private and control abuse.
                  </p>
                  <label
                    className="mt-3 block text-sm font-medium text-amber-50"
                    htmlFor="demo-password"
                  >
                    Demo password
                  </label>
                  <input
                    className="mt-2 w-full rounded-2xl border border-amber-100/15 bg-slate-950/80 px-4 py-3 text-sm text-slate-100 outline-none transition placeholder:text-slate-500 focus:border-amber-300"
                    id="demo-password"
                    onChange={(event) => setAccessPassword(event.target.value)}
                    placeholder="Enter access password"
                    type="password"
                    value={accessPassword}
                  />
                  <button
                    className="mt-4 w-full rounded-2xl bg-amber-300 px-4 py-3 text-sm font-semibold text-slate-950 transition hover:bg-amber-200 disabled:cursor-not-allowed disabled:bg-slate-700 disabled:text-slate-400"
                    disabled={accessPassword.trim() === ""}
                    type="submit"
                  >
                    {isUnlocked ? "Refresh access" : "Unlock demo"}
                  </button>
                </form>
              ) : null}
            </div>
          </div>

          {visibleAccessError !== null ? (
            <div className="mt-4 rounded-2xl border border-rose-400/20 bg-rose-500/10 px-4 py-3 text-sm text-rose-100">
              {visibleAccessError}
            </div>
          ) : null}

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
          {showAccessGate ? (
            <section className="rounded-4xl border border-white/10 bg-slate-900/70 p-6 shadow-2xl shadow-slate-950/30">
              <p className="text-xs font-semibold uppercase tracking-[0.3em] text-amber-200">
                Demo Locked
              </p>
              <h2 className="mt-3 text-2xl font-semibold text-white">
                Enter the demo password to open the live chat.
              </h2>
              <p className="mt-3 max-w-2xl text-sm leading-7 text-slate-300">
                The app is intentionally gated so a public link can still be
                shared with employers without leaving anonymous WebSocket access
                open.
              </p>
            </section>
          ) : (
            <LiveChat
              enabled={!showAccessGate}
              key={`${selectedRoom}:${username}:${isUnlocked}`}
              room={selectedRoom}
              username={username}
            />
          )}
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

async function fetchSessionState(): Promise<SessionState> {
  return fetchJson<SessionState>("/api/session", "Unable to load session.");
}

async function fetchRooms(): Promise<RoomSummary[]> {
  return fetchJson<RoomSummary[]>("/api/rooms", "Unable to load rooms.");
}

async function unlockSession(
  password: string,
): Promise<Pick<SessionState, "unlocked">> {
  const response = await fetch("/api/session", {
    body: JSON.stringify({
      password,
    }),
    headers: {
      "Content-Type": "application/json",
    },
    method: "POST",
  });

  if (!response.ok) {
    throw new Error(
      response.status === 401
        ? "Incorrect demo password."
        : `Request failed with status ${response.status}`,
    );
  }

  return (await response.json()) as Pick<SessionState, "unlocked">;
}

async function fetchJson<T>(url: string, fallbackMessage: string): Promise<T> {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`Request failed with status ${response.status}`);
  }

  try {
    return (await response.json()) as T;
  } catch (error) {
    throw new Error(getErrorMessage(error, fallbackMessage));
  }
}

function getQueryErrorMessage(
  error: Error | null,
  fallbackMessage: string,
): string | null {
  if (error === null) {
    return null;
  }

  return getErrorMessage(error, fallbackMessage);
}

function getErrorMessage(error: unknown, fallbackMessage: string): string {
  if (error instanceof Error) {
    return error.message;
  }

  return fallbackMessage;
}
