"use client";

import { useTsunaSocket } from "@/hooks/useTsunaSocket";
import PeersPanel from "@/components/PeersPanel";
import PlaybackPanel from "@/components/PlaybackPanel";
import QueuePanel from "@/components/QueuePanel";
import LogPanel from "@/components/LogPanel";
import ChatPanel from "@/components/ChatPanel";
import ReactionsOverlay from "@/components/ReactionsOverlay";

export default function Dashboard() {
  const {
    connected,
    init,
    peers,
    state,
    queue,
    logs,
    roomState,
    reactions,
    send,
    addReaction,
  } = useTsunaSocket("ws://localhost:9090/ws");

  return (
    <div className="h-screen flex flex-col p-4 gap-4 overflow-hidden">
      <header className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <h1 className="text-xl font-bold tracking-tight" style={{ fontFamily: "var(--font-jp)" }}>
            津波 <span className="text-text">TSUNA</span>
          </h1>
          {init && (
            <span className="font-mono text-sm px-3 py-1 rounded-lg glass-bright text-accent animate-pulse-glow">
              {init.room_code}
            </span>
          )}
        </div>

        <div className="flex items-center gap-3">
          <ReactionsOverlay reactions={reactions} onReact={addReaction} />
          <div className="flex items-center gap-2">
            <span
              className={`w-2 h-2 rounded-full ${connected ? "bg-success" : "bg-danger"}`}
            />
            <span className="text-xs text-dim">
              {connected ? "live" : "connecting..."}
            </span>
          </div>
        </div>
      </header>

      <div className="flex-1 grid grid-cols-3 gap-4 min-h-0">
        <div className="flex flex-col gap-4">
          <PeersPanel
            peers={peers}
            localId={init?.local_id || "..."}
            isHost={init?.is_host || false}
          />
          <PlaybackPanel state={state} roomState={roomState} />
          <QueuePanel items={queue.items} current={queue.current} />
        </div>

        <div className="flex flex-col">
          <ChatPanel send={send} localId={init?.local_id || "..."} />
        </div>
        <div className="flex flex-col">
          <LogPanel logs={logs} />
        </div>
      </div>

      <footer className="flex items-center justify-between text-xs text-muted">
        <span>tsuna v0.1.0</span>
        <span>{init?.is_host ? "hosting" : "joined"} • {peers.size} peer{peers.size !== 1 ? "s" : ""}</span>
      </footer>
    </div>
  );
}
