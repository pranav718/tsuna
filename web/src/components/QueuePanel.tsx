"use client";

import { QueueItem } from "@/hooks/useTsunaSocket";

interface Props {
  items: QueueItem[];
  current: number;
}

export default function QueuePanel({ items, current }: Props) {
  if (items.length === 0) return null;

  return (
    <div className="panel">
      <div className="panel-header">QUEUE</div>
      <div className="p-3 space-y-1">
        {items.map((item, i) => {
          const isCurrent = i === current;
          const name = item.Filename.split("/").pop() || item.Filename;

          return (
            <div
              key={item.ID}
              className={`flex items-center gap-2 px-2 py-1 transition-colors ${
                isCurrent ? "border border-border bg-surface-hover" : "border border-transparent hover:border-border-dim"
              }`}
            >
              <span className="text-[10px] w-4 text-center">
                {isCurrent ? (
                  <span className="text-accent text-glow">▶</span>
                ) : (
                  <span className="text-dim">{i + 1}</span>
                )}
              </span>
              <span className={`text-xs truncate flex-1 ${isCurrent ? "text-text-bright" : "text-dim"}`}>
                {name.length > 35 ? name.slice(0, 34) + "…" : name}
              </span>
              <span className="text-[10px] text-muted">{item.AddedBy.slice(0, 12)}</span>
            </div>
          );
        })}
      </div>
    </div>
  );
}
