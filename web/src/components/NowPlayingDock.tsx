"use client";

import { useState } from "react";
import { StateInfo, QueueItem } from "@/hooks/useTsunaSocket";

interface Props {
  state: StateInfo;
  roomState: string;
  queue: { items: QueueItem[]; current: number };
}

function formatTime(seconds: number): string {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = Math.floor(seconds % 60);
  if (h > 0)
    return `${h}:${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
  return `${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}

const SYNC_CFG: Record<string, { text: string; color: string; pulse: boolean }> = {
  PLAYING: { text: "synchronized", color: "var(--color-success)", pulse: false },
  PAUSED:  { text: "paused",       color: "var(--color-dim)",     pulse: false },
  HOLDING: { text: "buffering...", color: "var(--color-warning)",  pulse: true  },
  IDLE:    { text: "idle",         color: "var(--color-muted)",   pulse: false },
};

export default function NowPlayingDock({ state, roomState, queue }: Props) {
  const [queueOpen, setQueueOpen] = useState(false);

  const currentItem = queue.current >= 0 ? queue.items[queue.current] : null;
  const filename = currentItem
    ? currentItem.Filename.split("/").pop() || currentItem.Filename
    : null;

  const cfg = SYNC_CFG[roomState] ?? SYNC_CFG.IDLE;
  const isPlaying = roomState === "PLAYING";

  return (
    <div className="shrink-0 relative">
      {queueOpen && queue.items.length > 0 && (
        <div
          className="absolute bottom-full left-0 right-0 panel-dim border border-border border-b-0 max-h-48 overflow-y-auto"
          style={{ zIndex: 10 }}
        >
          <div className="panel-header flex items-center justify-between">
            <span>QUEUE</span>
            <span className="text-[10px] opacity-50 font-mono font-normal normal-case tracking-normal">
              {queue.items.length} tracks
            </span>
          </div>
          {queue.items.map((item, i) => {
            const isCurrent = i === queue.current;
            const name = item.Filename.split("/").pop() || item.Filename;
            return (
              <div
                key={item.ID}
                className={`flex items-center gap-2 px-3 py-1.5 text-[11px] font-mono border-b border-border-dim ${
                  isCurrent ? "text-text-bright" : "text-dim"
                }`}
              >
                <span className="w-4 text-center shrink-0">
                  {isCurrent ? (
                    <span style={{ color: "var(--color-accent)" }}>▶</span>
                  ) : (
                    <span className="text-muted">{i + 1}</span>
                  )}
                </span>
                <span className="flex-1 truncate">{name}</span>
                <span className="text-muted text-[9px] shrink-0">
                  {item.AddedBy.slice(0, 10)}
                </span>
              </div>
            );
          })}
        </div>
      )}

      <div className="dock-panel flex items-center gap-5 px-5 py-2.5">
        <span
          className="text-[1.6rem] text-glow shrink-0"
          style={{ fontFamily: "var(--font-display)", lineHeight: 1 }}
        >
          {state.Paused ? "⏸" : "▶"}
        </span>

        <div className="flex flex-col shrink-0" style={{ width: "220px" }}>
          <span className="text-[9px] text-muted uppercase tracking-widest">NOW PLAYING</span>
          <span className="text-[11px] text-text-bright truncate font-mono">
            {filename || "—"}
          </span>
        </div>

        <div className="flex items-center gap-3 flex-1 min-w-0">
          <span className="text-sm font-mono text-text-bright text-glow shrink-0 tabular-nums">
            {formatTime(state.Position)}
          </span>

          <div
            className="flex-1 relative overflow-hidden"
            style={{ height: "2px", background: "rgba(80, 55, 15, 0.45)" }}
          >
            {isPlaying && (
              <div
                className="absolute inset-0 opacity-60"
                style={{
                  background:
                    "linear-gradient(90deg, transparent 0%, var(--color-accent) 50%, transparent 100%)",
                  animation: "scan 2.5s linear infinite",
                }}
              />
            )}
            {!isPlaying && (
              <div
                className="absolute inset-y-0 left-0"
                style={{ width: "35%", background: "rgba(200, 125, 16, 0.25)" }}
              />
            )}
          </div>

          {state.SyncDelta !== 0 && (
            <span
              className="text-[10px] font-mono shrink-0 tabular-nums"
              style={{
                color:
                  Math.abs(state.SyncDelta) <= 40
                    ? "var(--color-success)"
                    : "var(--color-warning)",
              }}
            >
              Δ{state.SyncDelta > 0 ? "+" : ""}
              {state.SyncDelta}ms
            </span>
          )}
        </div>

        <span
          className={`shrink-0 text-[10px] font-mono px-2 py-0.5 border uppercase tracking-wider ${
            cfg.pulse ? "animate-pulse-glow" : ""
          }`}
          style={{ color: cfg.color, borderColor: cfg.color + "55" }}
        >
          {cfg.text}
        </span>

        <button
          onClick={() => setQueueOpen((o) => !o)}
          className={`shrink-0 text-[10px] font-mono px-2 py-1 border uppercase tracking-wider transition-colors ${
            queueOpen
              ? "text-accent border-border"
              : "text-muted border-border-dim hover:text-dim hover:border-border"
          }`}
        >
          QUEUE {queue.items.length > 0 ? `${queueOpen ? "▾" : "▴"} ${queue.items.length}` : "·"}
        </button>
      </div>
    </div>
  );
}
