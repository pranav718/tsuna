"use client";

import { useTsunaSocket } from "@/hooks/useTsunaSocket";
import { useState, useEffect } from "react";
import PeersPanel from "@/components/PeersPanel";
import LogPanel from "@/components/LogPanel";
import ChatPanel from "@/components/ChatPanel";
import ReactionsOverlay from "@/components/ReactionsOverlay";
import NowPlayingDock from "@/components/NowPlayingDock";

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

  const [clock, setClock] = useState("");
  const [uptime, setUptime] = useState(0);

  useEffect(() => {
    const tick = () => {
      setClock(
        new Date().toLocaleTimeString("en-US", {
          hour12: false,
          hour: "2-digit",
          minute: "2-digit",
          second: "2-digit",
        })
      );
    };
    tick();
    const id = setInterval(tick, 1000);
    return () => clearInterval(id);
  }, []);

  useEffect(() => {
    if (!connected) {
      setUptime(0);
      return;
    }
    const start = Date.now();
    const id = setInterval(() => setUptime(Math.floor((Date.now() - start) / 1000)), 1000);
    return () => clearInterval(id);
  }, [connected]);

  const formatUptime = (s: number) => {
    const h = String(Math.floor(s / 3600)).padStart(2, "0");
    const m = String(Math.floor((s % 3600) / 60)).padStart(2, "0");
    const sec = String(s % 60).padStart(2, "0");
    return `${h}:${m}:${sec}`;
  };

  const totalUsers = peers.size + 1;

  return (
    <div className="h-screen flex flex-col overflow-hidden relative z-10">
      <header
        className="shrink-0 flex items-center justify-between panel px-4"
        style={{ minHeight: "54px" }}
      >
        <div className="flex flex-col gap-0">
          <div className="flex items-center gap-3">
            <h1
              className="text-[1.6rem] tracking-widest uppercase text-glow"
              style={{ fontFamily: "var(--font-display)", lineHeight: 1 }}
            >
              津波 TSUNA
            </h1>
            {init && (
              <span className="font-mono text-xs px-2 py-0.5 border border-border text-accent animate-pulse-glow tracking-widest">
                {init.room_code}
              </span>
            )}
            <span className="text-[10px] text-muted tracking-widest uppercase hidden sm:block">
              {init?.is_host ? "// HOST NODE" : "// PEER NODE"}
            </span>
          </div>
          <div className="text-[9px] text-muted tracking-widest uppercase leading-tight">
            {connected
              ? `NODE ONLINE · P2P MESH · ${totalUsers} USER${totalUsers !== 1 ? "S" : ""} CONNECTED`
              : "CONNECTING TO MESH..."}
          </div>
        </div>

        <div className="flex items-center gap-4">
          <ReactionsOverlay reactions={reactions} onReact={addReaction} />
          <div className="border-l border-border-dim pl-4 flex flex-col items-end gap-0 font-mono">
            <span className="text-[11px] text-dim tracking-wider tabular-nums">{clock}</span>
            <span
              className="text-[9px] tracking-wider"
              style={{ color: connected ? "var(--color-success)" : "var(--color-danger)" }}
            >
              {connected ? `↑ up ${formatUptime(uptime)}` : "○ reconnecting"}
            </span>
          </div>
        </div>
      </header>

      <div
        className="flex-1 min-h-0"
        style={{ display: "grid", gridTemplateColumns: "200px 1fr 220px" }}
      >
        <PeersPanel
          peers={peers}
          localId={init?.local_id || "..."}
          isHost={init?.is_host || false}
        />
        <ChatPanel send={send} localId={init?.local_id || "..."} />
        <LogPanel logs={logs} />
      </div>

      <NowPlayingDock state={state} roomState={roomState} queue={queue} />
    </div>
  );
}
