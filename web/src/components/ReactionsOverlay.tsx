"use client";

interface Props {
  reactions: { id: number; emoji: string }[];
  onReact: (emoji: string) => void;
}

const EMOJI_LIST = ["❤️", "🔥", "😂", "👀", "🎉", "💀", "✨", "🥺"];

export default function ReactionsOverlay({ reactions, onReact }: Props) {
  return (
    <>
      <div className="fixed bottom-16 right-6 pointer-events-none z-50">
        {reactions.map((r) => (
          <div
            key={r.id}
            className="animate-float-up text-2xl absolute bottom-0"
            style={{ right: `${Math.random() * 48}px` }}
          >
            {r.emoji}
          </div>
        ))}
      </div>

      <div
        className="flex items-center gap-px border border-border-dim px-1 py-0.5"
        style={{ background: "rgba(6, 4, 2, 0.7)" }}
      >
        {EMOJI_LIST.map((emoji) => (
          <button
            key={emoji}
            onClick={() => onReact(emoji)}
            className="w-5 h-5 flex items-center justify-center text-xs hover:scale-125 active:scale-90 transition-transform duration-100 hover:bg-white/10"
            title={emoji}
          >
            {emoji}
          </button>
        ))}
      </div>
    </>
  );
}
