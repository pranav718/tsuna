"use client";

import { useEffect, useRef, useState } from "react";
import { StateInfo, QueueItem, PeerInfo, LogEntry } from "@/hooks/useTsunaSocket";

interface Props {
  state: StateInfo;
  roomState: string;
  queue: { items: QueueItem[]; current: number };
  peers: Map<string, PeerInfo>;
  localId: string;
  isHost: boolean;
  logs: LogEntry[];
  clock: string;
  uptime: number;
  connected: boolean;
}

function formatTime(seconds: number): string {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = Math.floor(seconds % 60);
  if (h > 0) {
    return `${h}:${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
  }
  return `${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}

const STATUS_TEXT: Record<string, string> = {
  PLAYING: "playing",
  PAUSED:  "paused",
  HOLDING: "buffering",
  IDLE:    "idle",
};

export default function SessionSidebar({
  state,
  roomState,
  queue,
  peers,
  localId,
  isHost,
  logs,
  uptime,
  connected,
}: Props) {
  const signalRef = useRef<HTMLDivElement>(null);
  const peerList = Array.from(peers.values());
  const localName = localId && localId !== "..."
    ? localId.slice(0, 12)
    : (connected ? "connecting..." : "offline");
  const currentItem = queue.current >= 0 ? queue.items[queue.current] : null;
  const filename = currentItem
    ? currentItem.Filename.split("/").pop() || currentItem.Filename
    : null;

  useEffect(() => {
    if (signalRef.current) {
      signalRef.current.scrollTop = signalRef.current.scrollHeight;
    }
  }, [logs]);

  const formatUptime = (s: number) => {
    const h = String(Math.floor(s / 3600)).padStart(2, "0");
    const m = String(Math.floor((s % 3600) / 60)).padStart(2, "0");
    const sec = String(s % 60).padStart(2, "0");
    return `${h}:${m}:${sec}`;
  };

  return (
    <div className="flex flex-col h-full select-none text-[11px] leading-relaxed">
      <div className="h-[38px] px-4 border-b border-border bg-[#141416] flex items-center shrink-0">
        <span className="text-text-bright font-bold text-[11px]">session</span>
      </div>

      <div className="p-3 border-b border-border flex flex-col gap-1 shrink-0">
        <div className="flex justify-between items-baseline">
          <span className="text-text-bright">status:</span>
          <span className="font-mono text-dim">
            {STATUS_TEXT[roomState] || "idle"}
          </span>
        </div>
        {filename && (
          <div className="flex justify-between items-baseline gap-2">
            <span className="text-text-bright shrink-0">track:</span>
            <span className="truncate font-mono text-dim" title={filename}>
              {filename}
            </span>
          </div>
        )}
        <div className="flex justify-between items-baseline">
          <span className="text-text-bright">time:</span>
          <span className="font-mono text-dim tabular-nums">
            {formatTime(state.Position)}
          </span>
        </div>
        <div className="flex justify-between items-baseline">
          <span className="text-text-bright">uptime:</span>
          <span className="font-mono text-dim tabular-nums">
            {formatUptime(uptime)}
          </span>
        </div>
        {state.SyncDelta !== 0 && (
          <div className="flex justify-between items-baseline">
            <span className="text-text-bright">sync delta:</span>
            <span className="font-mono text-dim tabular-nums">
              {state.SyncDelta > 0 ? "+" : ""}{state.SyncDelta}ms
            </span>
          </div>
        )}
      </div>

      <div className="p-3 border-b border-border flex flex-col gap-1.5 shrink-0">
        <div className="text-text-bright font-semibold">users:</div>
        <div className="flex flex-col gap-1 pl-1">
          <div className="flex items-center justify-between">
            <span className={`truncate ${connected ? "text-text-bright font-semibold" : "text-dim font-normal"}`}>
              • {localName} <span className="text-[9px] text-muted font-normal">(you, {isHost ? "host" : "peer"})</span>
            </span>
          </div>
          {peerList.map((p) => (
            <div key={p.PeerID} className="flex items-center justify-between">
              <span className={`truncate ${p.online ? "text-text" : "text-dim"}`}>
                • {(p.DisplayName || p.PeerID).slice(0, 12)}
              </span>
              {p.online && p.rtt > 0 && (
                <span className="text-[9px] text-dim font-mono">{p.rtt.toFixed(0)}ms</span>
              )}
            </div>
          ))}
        </div>
      </div>

      <div className="p-3 border-b border-border flex flex-col gap-1 shrink-0">
        <div className="text-text-bright font-semibold">signal:</div>
        <div 
          ref={signalRef}
          className="h-16 overflow-y-auto pl-1 font-mono text-[9px] text-dim space-y-0.5 scrollbar-none"
        >
          {logs.length === 0 ? (
            <div className="text-muted">(awaiting signal)</div>
          ) : (
            logs.map((entry, i) => (
              <div key={i} className="truncate">
                {entry.time.slice(0, 5)} · {entry.text}
              </div>
            ))
          )}
        </div>
      </div>

      <div className="p-3 flex-1 flex flex-col min-h-0">
        <div className="text-text-bright font-semibold mb-1">queue:</div>
        <div className="flex-1 overflow-y-auto min-h-0 pl-1 space-y-1">
          {queue.items.length === 0 ? (
            <div className="text-muted text-[10px]">(queue is empty)</div>
          ) : (
            queue.items.map((item, i) => {
              const isCurrent = i === queue.current;
              const name = item.Filename.split("/").pop() || item.Filename;
              return (
                <div
                  key={item.ID}
                  className={`truncate flex gap-1 ${isCurrent ? "text-text-bright font-bold" : "text-dim"}`}
                >
                  <span className="w-3 text-right select-none">{isCurrent ? ">" : i + 1}</span>
                  <span className="truncate">{name}</span>
                </div>
              );
            })
          )}
        </div>
      </div>

      <div className="shrink-0 border-t border-border px-4 text-[9px] text-dim flex flex-col justify-center bg-black/25 font-mono h-[66px] space-y-1">
        <div className="flex justify-between">
          <span>network</span>
          <span>p2p mesh</span>
        </div>
        <div className="flex justify-between">
          <span>bridge</span>
          <span>
            {connected ? "online" : "offline"}
          </span>
        </div>
      </div>
    </div>
  );
}
