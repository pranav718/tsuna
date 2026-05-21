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
    <div className="glass p-4 flex flex-col">
      <h2 className="text-xs font-semibold tracking-widest text-dim mb-3">CHAT</h2>

      <div className="flex-1 h-40 overflow-y-auto space-y-2 mb-3">
        {messages.map((msg) => (
          <div key={msg.id} className="animate-fade-in">
            <div className="flex items-baseline gap-2">
              <span className="text-[10px] text-muted">{msg.time}</span>
              <span className="text-xs text-accent font-medium">
                {msg.sender.slice(0, 12)}
              </span>
            </div>
            <p className="text-sm text-text pl-12">{msg.text}</p>
          </div>
        ))}
        {messages.length === 0 && (
          <p className="text-xs text-dim italic">no messages yet</p>
        )}
        <div ref={bottomRef} />
      </div>

      <div className="flex gap-2">
        <input
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleSend()}
          placeholder="type a message..."
          className="flex-1 bg-surface border border-border rounded-lg px-3 py-2 text-sm text-text placeholder:text-muted outline-none focus:border-accent/40 transition-colors"
        />
        <button
          onClick={handleSend}
          className="px-4 py-2 rounded-lg bg-accent/15 text-accent text-sm font-medium hover:bg-accent/25 transition-colors active:scale-95"
        >
          send
        </button>
      </div>
    </div>
  );
}
