"use client";

import { StateInfo } from "@/hooks/useTsunaSocket";

interface Props {
  state: StateInfo;
  roomState: string;
}

function formatTime(seconds: number): string {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = Math.floor(seconds % 60);
  if (h > 0) return `${h}:${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
  return `${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}

export default function PlaybackPanel({ state, roomState }: Props) {
  const badgeColor =
    roomState === "PLAYING"
      ? "bg-success/20 text-success"
      : roomState === "HOLDING"
      ? "bg-warning/20 text-warning"
      : "bg-muted/30 text-dim";

  return (
    <div className="glass p-4">
      <div className="flex items-center justify-between mb-3">
        <h2 className="text-xs font-semibold tracking-widest text-dim">PLAYBACK</h2>
        <span className={`text-[10px] font-mono px-2 py-0.5 rounded-full ${badgeColor}`}>
          {roomState}
        </span>
      </div>

      <div className="flex items-center gap-3">
        <div className="text-2xl">
          {state.Paused ? "⏸" : "▶"}
        </div>
        <div className="flex-1">
          <p className="text-lg font-mono text-text">{formatTime(state.Position)}</p>
          {state.SyncDelta !== 0 && (
            <p className="text-xs text-dim">
              sync delta: <span className={Math.abs(state.SyncDelta) <= 40 ? "text-success" : "text-warning"}>
                {state.SyncDelta > 0 ? "+" : ""}{state.SyncDelta}ms
              </span>
            </p>
          )}
        </div>
      </div>

      <div className="mt-3 h-1 rounded-full bg-surface-hover overflow-hidden">
        <div
          className="h-full rounded-full bg-accent transition-all duration-500"
          style={{ width: "0%" }}
        />
      </div>
    </div>
  );
}
