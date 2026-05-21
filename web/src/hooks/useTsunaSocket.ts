"use client";

import { useEffect, useRef, useCallback, useState } from "react";

export interface WSEvent {
  type: string;
  data: any;
}

export interface RoomInit {
  room_code: string;
  local_id: string;
  is_host: boolean;
}

export interface PeerInfo {
  PeerID: string;
  DisplayName: string;
  online: boolean;
  rtt: number;
  offset: number;
  syncDelta: number;
  buffering: boolean;
}

export interface StateInfo {
  Position: number;
  Paused: boolean;
  SyncDelta: number;
}

export interface QueueItem {
  ID: string;
  Filename: string;
  AddedBy: string;
}

export interface LogEntry {
  time: string;
  text: string;
}

export function useTsunaSocket(url: string) {
  const wsRef = useRef<WebSocket | null>(null);
  const [connected, setConnected] = useState(false);
  const [init, setInit] = useState<RoomInit | null>(null);
  const [peers, setPeers] = useState<Map<string, PeerInfo>>(new Map());
  const [state, setState] = useState<StateInfo>({ Position: 0, Paused: true, SyncDelta: 0 });
  const [queue, setQueue] = useState<{ items: QueueItem[]; current: number }>({ items: [], current: -1 });
  const [logs, setLogs] = useState<LogEntry[]>([]);
  const [roomState, setRoomState] = useState("IDLE");
  const [reactions, setReactions] = useState<{ id: number; emoji: string }[]>([]);

  const addLog = useCallback((text: string) => {
    const time = new Date().toLocaleTimeString("en-US", { hour12: false, hour: "2-digit", minute: "2-digit", second: "2-digit" });
    setLogs((prev) => [...prev.slice(-50), { time, text }]);
  }, []);

  useEffect(() => {
    let ws: WebSocket;
    let retryTimeout: NodeJS.Timeout;

    function connect() {
      ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        setConnected(true);
        addLog("dashboard connected");
      };

      ws.onclose = () => {
        setConnected(false);
        retryTimeout = setTimeout(connect, 2000);
      };

      ws.onmessage = (e) => {
        try {
          const ev: WSEvent = JSON.parse(e.data);
          handleEvent(ev);
        } catch {}
      };
    }

    function handleEvent(ev: WSEvent) {
      switch (ev.type) {
        case "init":
          setInit(ev.data);
          addLog(`joined room ${ev.data.room_code}`);
          break;

        case "peer_hello":
          if (ev.data) {
            setPeers((prev) => {
              const next = new Map(prev);
              next.set(ev.data.PeerID, {
                PeerID: ev.data.PeerID,
                DisplayName: ev.data.DisplayName || ev.data.PeerID,
                online: true,
                rtt: 0,
                offset: 0,
                syncDelta: 0,
                buffering: false,
              });
              return next;
            });
            addLog(`peer connected: ${ev.data.DisplayName || ev.data.PeerID}`);
          }
          break;

        case "peer_bye":
          if (ev.data?.PeerID) {
            setPeers((prev) => {
              const next = new Map(prev);
              const p = next.get(ev.data.PeerID);
              if (p) next.set(ev.data.PeerID, { ...p, online: false });
              return next;
            });
            addLog("peer disconnected");
          }
          break;

        case "clock_sync":
          if (ev.data) {
            setPeers((prev) => {
              const next = new Map(prev);
              const p = next.get(ev.data.PeerID);
              if (p) {
                next.set(ev.data.PeerID, {
                  ...p,
                  rtt: ev.data.RTT / 1_000_000,
                  offset: ev.data.Offset / 1_000_000,
                });
              }
              return next;
            });
          }
          break;

        case "state_update":
          if (ev.data) {
            setState({
              Position: ev.data.Position / 1_000_000_000,
              Paused: ev.data.Paused,
              SyncDelta: ev.data.SyncDelta,
            });
            setRoomState(ev.data.Paused ? "PAUSED" : "PLAYING");
          }
          break;

        case "buffering_start":
          setRoomState("HOLDING");
          addLog("buffering detected");
          break;

        case "buffering_stop":
          setRoomState("PLAYING");
          addLog("buffering resolved");
          break;

        case "queue_update":
          if (ev.data) {
            setQueue({ items: ev.data.Items || [], current: ev.data.Current });
          }
          break;

        case "correction":
          if (ev.data) {
            addLog(`${ev.data.CorrType}: Δ${(ev.data.Delta / 1_000_000).toFixed(0)}ms`);
          }
          break;

        case "log":
          if (typeof ev.data === "string") {
            addLog(ev.data);
          }
          break;
      }
    }

    connect();
    return () => {
      clearTimeout(retryTimeout);
      ws?.close();
    };
  }, [url, addLog]);

  const send = useCallback((type: string, data?: any) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type, data }));
    }
  }, []);

  const addReaction = useCallback((emoji: string) => {
    const id = Date.now();
    setReactions((prev) => [...prev, { id, emoji }]);
    setTimeout(() => {
      setReactions((prev) => prev.filter((r) => r.id !== id));
    }, 2500);
    send("reaction", { emoji });
  }, [send]);

  return {
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
  };
}
