"use client";

import { useState, useRef, useEffect } from "react";

interface Message {
  id: number;
  sender: string;
  text: string;
  time: string;
}

interface Props {
  send: (type: string, data?: any) => void;
  localId: string;
}

export default function ChatPanel({ send, localId }: Props) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const handleSend = () => {
    const text = input.trim();
    if (!text) return;

    const time = new Date().toLocaleTimeString("en-US", {
      hour12: false,
      hour: "2-digit",
      minute: "2-digit",
    });

    setMessages((prev) => [
      ...prev,
      { id: Date.now(), sender: localId, text, time },
    ]);
    send("chat", { text });
    setInput("");
  };

  return (
    <div className="panel flex flex-col h-full">
      <div className="panel-header">CHAT</div>

      <div className="flex-1 overflow-y-auto p-3 space-y-2">
        {messages.map((msg) => (
          <div key={msg.id} className="animate-fade-in">
            <div className="flex items-baseline gap-2">
              <span className="text-[10px] text-muted">{msg.time}</span>
              <span className="text-xs text-accent font-bold">
                {msg.sender.slice(0, 12)}
              </span>
            </div>
            <p className="text-xs text-text pl-12">{msg.text}</p>
          </div>
        ))}
        {messages.length === 0 && (
          <p className="text-[10px] text-dim">no messages yet</p>
        )}
        <div ref={bottomRef} />
      </div>

      <div className="flex gap-0 border-t border-border" style={{ background: 'rgba(6, 4, 2, 0.6)' }}>
        <div className="flex items-center px-2 text-dim text-xs">
          ⌨
        </div>
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleSend()}
          placeholder="type a message..."
          className="flex-1 bg-transparent border-none px-2 py-2 text-xs text-text placeholder:text-muted outline-none"
        />
        <button
          onClick={handleSend}
          className="px-4 py-2 bg-accent text-bg text-xs font-bold uppercase tracking-wider hover:bg-text-bright transition-colors active:scale-95"
        >
          send
        </button>
      </div>
    </div>
  );
}
