"use client";

import { PeerInfo } from "@/hooks/useTsunaSocket";

interface Props {
  peers: Map<string, PeerInfo>;
  localId: string;
  isHost: boolean;
}

export default function PeersPanel({ peers, localId, isHost }: Props) {
  return (
    <div className="panel">
      <div className="panel-header">PEERS</div>
      <div className="p-3 space-y-1.5">
        <div className="flex items-center gap-2 px-2 py-1 border border-border-dim">
          <span className="w-1.5 h-1.5 bg-success shrink-0" />
          <span className="text-xs text-text truncate">{localId.slice(0, 20)}</span>
          <span className="text-[10px] text-dim ml-auto uppercase tracking-wider">
            you, {isHost ? "host" : "peer"}
          </span>
        </div>

        {Array.from(peers.values()).map((p) => (
          <div key={p.PeerID} className="flex items-center gap-2 px-2 py-1 border border-border-dim animate-fade-in">
            <span
              className={`w-1.5 h-1.5 shrink-0 ${
                p.online ? (p.buffering ? "bg-warning" : "bg-success") : "bg-danger"
              }`}
            />
            <span className="text-xs text-text truncate">
              {(p.DisplayName || p.PeerID).slice(0, 20)}
            </span>
            <div className="ml-auto flex items-center gap-3 text-[10px] text-dim">
              {p.online && p.rtt > 0 && (
                <span className={p.rtt < 50 ? "text-success" : p.rtt < 200 ? "text-warning" : "text-danger"}>
                  {p.rtt.toFixed(0)}ms
                </span>
              )}
              {p.online && p.syncDelta !== 0 && (
                <span>Δ{p.syncDelta > 0 ? "+" : ""}{p.syncDelta}ms</span>
              )}
              {p.buffering && <span className="text-warning">buffering</span>}
              {!p.online && <span className="text-danger">offline</span>}
            </div>
          </div>
        ))}

        {peers.size === 0 && (
          <p className="text-[10px] text-dim px-2">waiting for peers...</p>
        )}
      </div>
    </div>
  );
}
