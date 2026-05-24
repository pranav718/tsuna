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
      ? "border-success text-success"
      : roomState === "HOLDING"
      ? "border-warning text-warning"
      : "border-muted text-dim";

  return (
    <div className="panel">
      <div className="panel-header flex items-center justify-between">
        <span>PLAYBACK</span>
        <span className={`text-[10px] font-mono px-1.5 py-0 border ${badgeColor} bg-bg`}>
          {roomState}
        </span>
      </div>
      <div className="p-3">
        <div className="flex items-center gap-3">
          <div className="text-xl text-glow">
            {state.Paused ? "⏸" : "▶"}
          </div>
          <div className="flex-1">
            <p className="text-lg font-mono text-text-bright text-glow">{formatTime(state.Position)}</p>
            {state.SyncDelta !== 0 && (
              <p className="text-[10px] text-dim">
                sync delta:{" "}
                <span className={Math.abs(state.SyncDelta) <= 40 ? "text-success" : "text-warning"}>
                  {state.SyncDelta > 0 ? "+" : ""}{state.SyncDelta}ms
                </span>
              </p>
            )}
          </div>
        </div>

        <div className="mt-3 h-0.5 bg-muted overflow-hidden">
          <div
            className="h-full bg-accent transition-all duration-500"
            style={{ width: "0%" }}
          />
        </div>
      </div>
    </div>
  );
}
