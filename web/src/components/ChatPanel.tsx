"use client";

import { useState, useRef, useEffect } from "react";

type MessageType = "system" | "self" | "user" | "command" | "event";

interface Message {
  id: number;
  sender: string;
  text: string;
  time: string;
  type: MessageType;
}

interface Props {
  send: (type: string, data?: any) => void;
  localId: string;
}

function ts(): string {
  return new Date().toLocaleTimeString("en-US", {
    hour12: false,
    hour: "2-digit",
    minute: "2-digit",
  });
}

export default function ChatPanel({ send, localId }: Props) {
  const [messages, setMessages] = useState<Message[]>([
    {
      id: 0,
      sender: "system",
      text: "welcome to tsuna · type /help for commands",
      time: ts(),
      type: "system",
    },
  ]);
  const [input, setInput] = useState("");
  const [isTyping, setIsTyping] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);
  const typingRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const addMsg = (partial: Omit<Message, "id">) =>
    setMessages((prev) => [...prev, { ...partial, id: Date.now() }]);

  const handleSend = () => {
    const text = input.trim();
    if (!text) return;
    setInput("");
    setIsTyping(false);
    const time = ts();

    if (text.startsWith("/")) {
      const [cmd, ...args] = text.split(" ");
      const arg = args.join(" ");
      switch (cmd.toLowerCase()) {
        case "/clear":
          setMessages([]);
          return;
        case "/help":
          addMsg({
            sender: "system",
            text: "/clear · /me <action> · /status · /whois · /help",
            time,
            type: "command",
          });
          return;
        case "/me":
          if (arg) {
            addMsg({ sender: localId.slice(0, 12), text: `* ${arg}`, time, type: "event" });
            send("chat", { text: `* ${arg}` });
          }
          return;
        case "/status":
          addMsg({
            sender: "system",
            text: "node: online · transport: p2p · sync: active",
            time,
            type: "command",
          });
          return;
        case "/whois":
          addMsg({
            sender: "system",
            text: `you are ${localId}`,
            time,
            type: "command",
          });
          return;
        default:
          addMsg({
            sender: "system",
            text: `unknown: ${cmd} · try /help`,
            time,
            type: "command",
          });
          return;
      }
    }

    addMsg({ sender: localId.slice(0, 12), text, time, type: "self" });
    send("chat", { text });
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInput(e.target.value);
    setIsTyping(true);
    if (typingRef.current) clearTimeout(typingRef.current);
    typingRef.current = setTimeout(() => setIsTyping(false), 1500);
  };

  const isCmd = input.startsWith("/");

  return (
    <div
      className="panel flex flex-col h-full"
      style={{ background: "rgba(4, 3, 1, 0.82)" }}
    >
      <div className="panel-header flex items-center justify-between">
        <span>CHAT</span>
        <span className="text-[10px] opacity-50 font-mono font-normal normal-case tracking-normal">
          {messages.length} msgs
        </span>
      </div>

      <div className="flex-1 overflow-y-auto px-3 py-2 space-y-1.5">
        {messages.map((msg) => (
          <div key={msg.id} className="animate-fade-in flex gap-2.5">
            <span className="text-[9px] text-muted shrink-0 mt-px select-none tabular-nums leading-tight pt-0.5">
              {msg.time}
            </span>
            <div className="flex-1 min-w-0">
              {msg.type === "system" && (
                <p className="text-[11px] text-dim italic leading-snug">
                  · {msg.text}
                </p>
              )}
              {msg.type === "command" && (
                <p
                  className="text-[11px] font-mono leading-snug"
                  style={{ color: "var(--color-command)" }}
                >
                  $ {msg.text}
                </p>
              )}
              {msg.type === "event" && (
                <p className="text-[11px] text-accent italic leading-snug">
                  * {msg.text}
                </p>
              )}
              {msg.type === "self" && (
                <div>
                  <div className="flex items-baseline gap-1.5">
                    <span className="text-[10px] text-text-bright font-bold">
                      {msg.sender}
                    </span>
                    <span className="text-[9px] text-muted">you</span>
                  </div>
                  <p className="text-[11px] text-text-bright leading-snug">
                    {msg.text}
                  </p>
                </div>
              )}
              {msg.type === "user" && (
                <div>
                  <span className="text-[10px] text-accent font-bold">
                    {msg.sender}
                  </span>
                  <p className="text-[11px] text-text leading-snug">
                    {msg.text}
                  </p>
                </div>
              )}
            </div>
          </div>
        ))}
        {messages.length === 0 && (
          <p className="text-[10px] text-muted italic">
            no messages · type /help
          </p>
        )}
        <div ref={bottomRef} />
      </div>

      {isTyping && (
        <div className="px-3 py-0.5 flex items-center gap-1.5 text-[10px] text-muted border-t border-border-dim">
          <span className="animate-blink">▮</span>
          <span>composing...</span>
        </div>
      )}

      <div
        className="shrink-0 flex items-center border-t border-border-dim"
        style={{ background: "rgba(2, 1, 0, 0.9)" }}
      >
        <span
          className="px-3 text-sm font-mono shrink-0"
          style={{ color: isCmd ? "var(--color-command)" : "var(--color-accent)" }}
        >
          ▸
        </span>
        <input
          type="text"
          value={input}
          onChange={handleInputChange}
          onKeyDown={(e) => e.key === "Enter" && handleSend()}
          placeholder="message or /command"
          className="flex-1 bg-transparent border-none py-2 pr-2 text-xs text-text placeholder:text-muted outline-none font-mono"
        />
        <button
          onClick={handleSend}
          className="px-3 py-2 text-[10px] font-bold uppercase tracking-wider text-bg bg-accent hover:bg-text-bright transition-colors active:scale-95 shrink-0"
        >
          SEND
        </button>
      </div>
    </div>
  );
}
