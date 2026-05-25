"use client";

import { useEffect, useRef } from "react";
import { LogEntry } from "@/hooks/useTsunaSocket";

interface Props {
  logs: LogEntry[];
}

function classify(text: string): { glyph: string; color: string } {
  const t = text.toLowerCase();
  if (t.includes("joined") || t.includes("connected") || t.includes("dashboard"))
    return { glyph: "→", color: "var(--color-success)" };
  if (t.includes("disconnected") || t.includes("bye") || t.includes("offline"))
    return { glyph: "←", color: "var(--color-danger)" };
  if (t.includes("buffer") || t.includes("holding") || t.includes("warning"))
    return { glyph: "△", color: "var(--color-warning)" };
  if (
    t.includes("sync") ||
    t.includes("correction") ||
    t.includes("resolved") ||
    t.includes("delta")
  )
    return { glyph: "⟳", color: "var(--color-accent)" };
  if (t.includes("play") || t.includes("queue") || t.includes("track"))
    return { glyph: "▶", color: "var(--color-text)" };
  return { glyph: "·", color: "var(--color-dim)" };
}

export default function LogPanel({ logs }: Props) {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [logs]);

  return (
    <div className="panel flex flex-col h-full">
      <div className="panel-header flex items-center justify-between">
        <span>SIGNAL</span>
        <span className="text-[10px] opacity-50 font-mono font-normal normal-case tracking-normal">
          {logs.length}
        </span>
      </div>

      <div className="flex-1 overflow-y-auto py-1 font-mono">
        {logs.length === 0 && (
          <p className="text-[10px] text-muted px-2 py-1 italic">
            awaiting signal<span className="animate-blink">_</span>
          </p>
        )}
        {logs.map((entry, i) => {
          const { glyph, color } = classify(entry.text);
          return (
            <div
              key={i}
              className="flex gap-1.5 animate-fade-in px-2 py-0.5 hover:bg-white/[0.015] transition-colors"
            >
              <span className="text-[9px] text-muted shrink-0 tabular-nums leading-tight pt-px">
                {entry.time.slice(0, 5)}
              </span>
              <span
                className="text-[10px] shrink-0 pt-px"
                style={{ color }}
              >
                {glyph}
              </span>
              <span className="text-[10px] text-dim leading-tight break-all">
                {entry.text}
              </span>
            </div>
          );
        })}
        <div ref={bottomRef} />
      </div>
    </div>
  );
}
