import { useState } from "react";
import type { FormEvent } from "react";

interface MessageInputProps {
  disabled: boolean;
  onSendMessage: (content: string) => boolean;
}

export function MessageInput({ disabled, onSendMessage }: MessageInputProps) {
  const [content, setContent] = useState("");

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    const didSend = onSendMessage(content);
    if (!didSend) {
      return;
    }

    setContent("");
  };

  return (
    <form
      className="mt-4 flex flex-col gap-3 sm:flex-row"
      onSubmit={handleSubmit}
    >
      <label className="sr-only" htmlFor="message-input">
        Send a message
      </label>

      <input
        className="min-h-12 flex-1 rounded-2xl border border-white/10 bg-slate-950/80 px-4 py-3 text-sm text-slate-100 outline-none transition placeholder:text-slate-500 focus:border-violet-400"
        disabled={disabled}
        id="message-input"
        maxLength={400}
        onChange={(event) => setContent(event.target.value)}
        placeholder={
          disabled
            ? "Connect first to start chatting."
            : "Write a message for the room."
        }
        type="text"
        value={content}
      />

      <button
        className="min-h-12 rounded-2xl bg-violet-500 px-5 py-3 text-sm font-semibold text-white transition hover:bg-violet-400 disabled:cursor-not-allowed disabled:bg-slate-700 disabled:text-slate-400"
        disabled={disabled || content.trim() === ""}
        type="submit"
      >
        Send
      </button>
    </form>
  );
}
