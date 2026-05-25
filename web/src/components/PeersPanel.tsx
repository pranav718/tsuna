"use client";

import { useEffect, useState } from "react";
import { PeerInfo } from "@/hooks/useTsunaSocket";

interface Props {
  peers: Map<string, PeerInfo>;
  localId: string;
  isHost: boolean;
}

function SignalBars({ rtt }: { rtt: number }) {
  const strength = rtt === 0 ? 0 : rtt < 50 ? 4 : rtt < 100 ? 3 : rtt < 200 ? 2 : 1;
  return (
    <span className="font-mono text-[9px] tracking-tighter">
      {[1, 2, 3, 4].map((i) => (
        <span
          key={i}
          style={{
            color: i <= strength ? "var(--color-success)" : "var(--color-muted)",
          }}
        >
          ▪
        </span>
      ))}
    </span>
  );
}

function PeerStatus({ p }: { p: PeerInfo }) {
  if (!p.online) return <span style={{ color: "var(--color-danger)", fontSize: "9px" }}>offline</span>;
  if (p.buffering) return <span style={{ color: "var(--color-warning)", fontSize: "9px" }}>buffering</span>;
  if (Math.abs(p.syncDelta) > 300)
    return <span style={{ color: "var(--color-warning)", fontSize: "9px" }}>drifting</span>;
  return <span style={{ color: "var(--color-dim)", fontSize: "9px" }}>listening</span>;
}

export default function PeersPanel({ peers, localId, isHost }: Props) {
  const [, setTick] = useState(0);
  useEffect(() => {
    const id = setInterval(() => setTick((t) => t + 1), 1000);
    return () => clearInterval(id);
  }, []);

  const peerList = Array.from(peers.values());
  const onlineCount = peerList.filter((p) => p.online).length + 1;

  return (
    <div className="panel flex flex-col h-full">
      <div className="panel-header flex items-center justify-between">
        <span>PEERS</span>
        <span className="text-[10px] opacity-50 font-mono font-normal normal-case tracking-normal">
          {onlineCount}/{peerList.length + 1}
        </span>
      </div>

      <div className="flex-1 overflow-y-auto">
        <div className="px-2 py-1.5 border-b border-border-dim">
          <div className="flex items-center gap-1.5">
            <span className="text-[11px] text-success animate-peer-online">◉</span>
            <span className="text-[11px] text-text-bright font-mono truncate flex-1">
              {localId.slice(0, 16)}
            </span>
          </div>
          <div className="flex items-center justify-between pl-[18px] mt-0.5">
            <span
              className="text-[9px] uppercase tracking-wider font-mono"
              style={{ color: "var(--color-accent)" }}
            >
              {isHost ? "HOST" : "peer"}
            </span>
            <span className="text-[9px] text-muted">you</span>
          </div>
        </div>

        {peerList.length === 0 ? (
          <div className="px-2 py-2 text-[10px] text-muted font-mono">
            waiting for peers<span className="animate-blink">_</span>
          </div>
        ) : (
          peerList.map((p) => (
            <div
              key={p.PeerID}
              className="px-2 py-1.5 border-b border-border-dim animate-fade-in hover:bg-white/[0.02] transition-colors"
            >
              <div className="flex items-center gap-1.5">
                <span
                  className={`text-[11px] ${
                    p.online ? "text-success animate-peer-online" : "text-muted"
                  }`}
                >
                  {p.online ? "◉" : "◌"}
                </span>
                <span
                  className={`text-[11px] font-mono truncate flex-1 ${
                    p.online ? "text-text" : "text-dim"
                  }`}
                >
                  {(p.DisplayName || p.PeerID).slice(0, 16)}
                </span>
              </div>
              <div className="flex items-center justify-between pl-[18px] mt-0.5">
                <PeerStatus p={p} />
                {p.online && p.rtt > 0 && (
                  <div className="flex items-center gap-1">
                    <SignalBars rtt={p.rtt} />
                    <span className="text-[9px] text-muted tabular-nums">
                      {p.rtt.toFixed(0)}ms
                    </span>
                  </div>
                )}
              </div>
            </div>
          ))
        )}
      </div>

      <div className="shrink-0 border-t border-border-dim px-2 py-1.5 text-[9px] font-mono text-muted space-y-0.5">
        <div className="flex justify-between">
          <span>mesh</span>
          <span style={{ color: "var(--color-success)" }}>online</span>
        </div>
        <div className="flex justify-between">
          <span>transport</span>
          <span>p2p</span>
        </div>
      </div>
    </div>
  );
}
