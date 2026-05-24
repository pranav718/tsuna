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
    <div className="h-screen flex flex-col p-3 gap-3 overflow-hidden relative z-10">
      <header className="flex items-center justify-between border border-border px-3 py-2 panel">
        <div className="flex items-center gap-4">
          <h1
            className="text-sm font-bold tracking-widest uppercase text-glow"
            style={{ fontFamily: "var(--font-jp)" }}
          >
            津波 <span className="text-text-bright">TSUNA</span>
          </h1>
          {init && (
            <span className="font-mono text-xs px-2 py-0.5 border border-border text-accent animate-pulse-glow">
              {init.room_code}
            </span>
          )}
        </div>

        <div className="flex items-center gap-4">
          <ReactionsOverlay reactions={reactions} onReact={addReaction} />
          <div className="flex items-center gap-2">
            <span
              className={`w-2 h-2 ${connected ? "bg-success" : "bg-danger"}`}
            />
            <span className="text-xs text-dim uppercase tracking-wider">
              {connected ? "● live" : "○ connecting..."}
            </span>
          </div>
        </div>
      </header>

      <div className="flex-1 grid grid-cols-3 gap-3 min-h-0">
        <div className="flex flex-col gap-3">
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

      <footer className="flex items-center justify-between text-xs text-muted border border-border-dim px-3 py-1.5 panel-dim" style={{ backdropFilter: 'blur(10px)' }}>
        <span>tsuna v0.1.0</span>
        <span>
          {init?.is_host ? "hosting" : "joined"} • {peers.size} peer
          {peers.size !== 1 ? "s" : ""}
        </span>
      </footer>
    </div>
  );
}
