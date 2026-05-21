"use client";

import { PeerInfo } from "@/hooks/useTsunaSocket";

interface Props {
  peers: Map<string, PeerInfo>;
  localId: string;
  isHost: boolean;
}

export default function PeersPanel({ peers, localId, isHost }: Props) {
  return (
    <div className="glass p-4">
      <h2 className="text-xs font-semibold tracking-widest text-dim mb-3">PEERS</h2>
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-success shrink-0" />
          <span className="text-sm text-text truncate">{localId.slice(0, 20)}</span>
          <span className="text-xs text-dim ml-auto">you, {isHost ? "host" : "peer"}</span>
        </div>

        {Array.from(peers.values()).map((p) => (
          <div key={p.PeerID} className="flex items-center gap-2 animate-fade-in">
            <span
              className={`w-2 h-2 rounded-full shrink-0 ${
                p.online ? (p.buffering ? "bg-warning" : "bg-success") : "bg-danger"
              }`}
            />
            <span className="text-sm text-text truncate">
              {(p.DisplayName || p.PeerID).slice(0, 20)}
            </span>
            <div className="ml-auto flex items-center gap-3 text-xs text-dim">
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
          <p className="text-xs text-dim italic">waiting for peers...</p>
        )}
      </div>
    </div>
  );
}
