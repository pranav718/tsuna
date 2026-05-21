"use client";

interface Props {
  reactions: { id: number; emoji: string }[];
  onReact: (emoji: string) => void;
}

const EMOJI_LIST = ["❤️", "🔥", "😂", "👀", "🎉", "💀", "✨", "🥺"];

export default function ReactionsOverlay({ reactions, onReact }: Props) {
  return (
    <>
      <div className="fixed bottom-24 right-8 pointer-events-none z-50">
        {reactions.map((r) => (
          <div
            key={r.id}
            className="animate-float-up text-3xl absolute bottom-0"
            style={{ right: `${Math.random() * 60}px` }}
          >
            {r.emoji}
          </div>
        ))}
      </div>

      <div className="flex gap-1.5">
        {EMOJI_LIST.map((emoji) => (
          <button
            key={emoji}
            onClick={() => onReact(emoji)}
            className="w-8 h-8 rounded-lg bg-surface hover:bg-surface-hover border border-border hover:border-border-bright transition-all duration-200 flex items-center justify-center text-sm hover:scale-110 active:scale-95"
          >
            {emoji}
          </button>
        ))}
      </div>
    </>
  );
}
