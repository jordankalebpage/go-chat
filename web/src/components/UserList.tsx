interface UserListProps {
  users: string[];
}

export function UserList({ users }: UserListProps) {
  return (
    <aside className="rounded-3xl border border-white/10 bg-slate-950/70 p-4">
      <div className="mb-4">
        <p className="text-xs font-semibold uppercase tracking-[0.3em] text-emerald-300">
          Presence
        </p>
        <h2 className="mt-2 text-lg font-semibold text-white">Active users</h2>
      </div>

      <ul className="space-y-2">
        {users.map((user) => (
          <li
            className="flex items-center gap-3 rounded-2xl border border-white/8 bg-slate-900/80 px-3 py-2 text-sm text-slate-200"
            key={user}
          >
            <span
              aria-hidden="true"
              className="h-2.5 w-2.5 rounded-full bg-emerald-400 shadow-[0_0_16px_rgba(52,211,153,0.8)]"
            />
            <span>{user}</span>
          </li>
        ))}
      </ul>
    </aside>
  );
}
