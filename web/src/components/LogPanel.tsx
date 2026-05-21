"use client";

import { useEffect, useRef } from "react";
import { LogEntry } from "@/hooks/useTsunaSocket";

interface Props {
  logs: LogEntry[];
}

export default function LogPanel({ logs }: Props) {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [logs]);

  return (
    <div className="glass p-4">
      <h2 className="text-xs font-semibold tracking-widest text-dim mb-3">LOG</h2>
      <div className="h-36 overflow-y-auto space-y-1 font-mono text-xs">
        {logs.map((entry, i) => (
          <div key={i} className="flex gap-2 animate-fade-in">
            <span className="text-muted shrink-0">{entry.time}</span>
            <span className="text-dim">{entry.text}</span>
          </div>
        ))}
        {logs.length === 0 && (
          <p className="text-dim italic">waiting for events...</p>
        )}
        <div ref={bottomRef} />
      </div>
    </div>
  );
}
