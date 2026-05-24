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
    <div className="panel flex flex-col h-full">
      <div className="panel-header">LOG</div>
      <div className="flex-1 overflow-y-auto p-3 space-y-0.5 font-mono text-[11px]">
        {logs.map((entry, i) => (
          <div key={i} className="flex gap-2 animate-fade-in">
            <span className="text-muted shrink-0">{entry.time}</span>
            <span className="text-dim">{entry.text}</span>
          </div>
        ))}
        {logs.length === 0 && (
          <p className="text-dim">waiting for events...</p>
        )}
        <div ref={bottomRef} />
      </div>
    </div>
  );
}
