"use client";

import { useTsunaSocket } from "@/hooks/useTsunaSocket";
import { useState, useEffect } from "react";
import ChatPanel from "@/components/ChatPanel";
import SessionSidebar from "@/components/SessionSidebar";

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

  return (
    <div className="h-screen flex flex-col bg-bg overflow-hidden relative z-10 p-4">
      <div className="flex-1 flex flex-col w-full max-w-[1080px] mx-auto min-h-0 gap-4">
        <header className="flex items-center select-none font-mono shrink-0">
          <div className="flex items-baseline gap-4">
            <h1 className="text-xl font-bold tracking-widest text-text-bright">tsuna</h1>
            {init && (
              <span className="text-[10px] text-dim">
                [{init.room_code.toLowerCase()}] {init.is_host ? "host node" : "peer node"}
              </span>
            )}
          </div>
        </header>

        <div
          className="flex-1 min-h-0 panel grid"
          style={{ gridTemplateColumns: "1fr 280px" }}
        >
          <div className="border-r border-border h-full overflow-hidden">
            <ChatPanel
              send={send}
              localId={init?.local_id || "..."}
              reactions={reactions}
              addReaction={addReaction}
            />
          </div>

          <div className="h-full overflow-hidden">
            <SessionSidebar
              state={state}
              roomState={roomState}
              queue={queue}
              peers={peers}
              localId={init?.local_id || "..."}
              isHost={init?.is_host || false}
              logs={logs}
              clock={clock}
              uptime={uptime}
              connected={connected}
            />
          </div>
        </div>
      </div>
    </div>
  );
}
