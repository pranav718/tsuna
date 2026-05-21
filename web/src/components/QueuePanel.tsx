"use client";

import { QueueItem } from "@/hooks/useTsunaSocket";

interface Props {
  items: QueueItem[];
  current: number;
}

export default function QueuePanel({ items, current }: Props) {
  if (items.length === 0) return null;

  return (
    <div className="glass p-4">
      <h2 className="text-xs font-semibold tracking-widest text-dim mb-3">QUEUE</h2>
      <div className="space-y-1.5">
        {items.map((item, i) => {
          const isCurrent = i === current;
          const name = item.Filename.split("/").pop() || item.Filename;

          return (
            <div
              key={item.ID}
              className={`flex items-center gap-2 px-2 py-1.5 rounded-lg transition-colors ${
                isCurrent ? "glass-bright" : "hover:bg-surface-hover"
              }`}
            >
              <span className="text-xs w-4 text-center">
                {isCurrent ? (
                  <span className="text-accent">▶</span>
                ) : (
                  <span className="text-dim">{i + 1}</span>
                )}
              </span>
              <span className={`text-sm truncate flex-1 ${isCurrent ? "text-text" : "text-dim"}`}>
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
