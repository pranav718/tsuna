"use client";

import React, { useState, useRef, useEffect } from "react";

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
  reactions: { id: number; emoji: string }[];
  addReaction: (emoji: string) => void;
}

const EMOJI_LIST = ["❤️", "😭", "😂", "💀", "👀"];

function ts(): string {
  return new Date().toLocaleTimeString("en-US", {
    hour12: false,
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
}

export default function ChatPanel({ send, localId, reactions, addReaction }: Props) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [input, setInput] = useState("");
  const [isTyping, setIsTyping] = useState(false);
  const bottomRef = useRef<HTMLDivElement>(null);
  const typingRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    setMessages([
      {
        id: 0,
        sender: "system",
        text: "tsuna watchroom active · welcome node user",
        time: ts(),
        type: "system",
      },
      {
        id: 1,
        sender: "system",
        text: "press Ctrl+1 to Ctrl+5 to trigger reactions",
        time: ts(),
        type: "system",
      },
    ]);
  }, []);

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
            text: "/clear · /me <action> · /status · /whois · /react <emoji> · /help",
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
        case "/react":
          if (arg && EMOJI_LIST.includes(arg)) {
            addReaction(arg);
            addMsg({
              sender: "system",
              text: `you reacted with ${arg}`,
              time,
              type: "event",
            });
          } else {
            addMsg({
              sender: "system",
              text: `invalid emoji. use one of: ${EMOJI_LIST.join(" ")}`,
              time,
              type: "system",
            });
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
            text: `unknown command: ${cmd} · try /help`,
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

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if ((e.ctrlKey || e.metaKey) && e.key >= "1" && e.key <= "5") {
      e.preventDefault();
      const idx = parseInt(e.key) - 1;
      const emoji = EMOJI_LIST[idx];
      if (emoji) {
        addReaction(emoji);
        addMsg({
          sender: "system",
          text: `you reacted with ${emoji}`,
          time: ts(),
          type: "event",
        });
      }
    } else if (e.key === "Enter") {
      handleSend();
    }
  };

  const isCmd = input.startsWith("/");

  return (
    <div className="flex flex-col h-full relative bg-surface">
      <div className="h-[38px] px-4 border-b border-border bg-[#141416] flex items-center font-semibold shrink-0">
        <span className="text-text-bright font-bold">chatroom</span>
      </div>

      <div className="flex-1 overflow-y-auto px-4 py-3 space-y-2 relative">
        {messages.map((msg) => (
          <div key={msg.id} className="animate-fade-in flex gap-3 text-xs leading-relaxed">
            <span className="text-[10px] text-muted shrink-0 select-none tabular-nums pt-0.5 font-mono">
              [{msg.time}]
            </span>
            <div className="flex-1 min-w-0">
              {msg.type === "system" && (
                <p className="text-dim italic">
                  * {msg.text}
                </p>
              )}
              {msg.type === "command" && (
                <p
                  className="font-semibold"
                  style={{ color: "var(--color-command)" }}
                >
                  $ {msg.text}
                </p>
              )}
              {msg.type === "event" && (
                <p className="font-bold" style={{ color: "var(--color-warning)" }}>
                  &lt;react&gt; {msg.sender}: {msg.text}
                </p>
              )}
              {msg.type === "self" && (
                <div>
                  <div className="flex items-baseline gap-1.5 select-none">
                    <span className="text-text-bright font-bold">
                      &lt;{msg.sender}&gt;
                    </span>
                    <span className="text-[9px] text-muted">(you)</span>
                  </div>
                  <p className="text-text-bright break-all mt-0.5">
                    {msg.text}
                  </p>
                </div>
              )}
              {msg.type === "user" && (
                <div>
                  <span className="text-accent font-bold">
                    &lt;{msg.sender}&gt;
                  </span>
                  <p className="text-text break-all mt-0.5">
                    {msg.text}
                  </p>
                </div>
              )}
            </div>
          </div>
        ))}
        {messages.length === 0 && (
          <p className="text-[10px] text-muted italic select-none">
            console buffer cleared
          </p>
        )}
        <div ref={bottomRef} />

        <div className="absolute bottom-4 right-4 pointer-events-none z-30 select-none">
          {reactions.map((r) => (
            <div
              key={r.id}
              className="animate-float-up text-3xl absolute bottom-0 right-0 font-sans"
              style={{
                right: `${Math.random() * 60}px`,
              }}
            >
              {r.emoji}
            </div>
          ))}
        </div>
      </div>

      {isTyping && (
        <div className="px-4 py-1 flex items-center gap-1.5 text-[10px] text-dim border-t border-border select-none font-mono">
          <span className="animate-blink">▮</span>
          <span>writing command buffer...</span>
        </div>
      )}

      <div className="shrink-0 border-t border-border bg-black/40 h-[66px] flex flex-col justify-between">
        <div className="px-4 text-[9px] text-dim border-b border-border flex flex-wrap gap-x-2 select-none tracking-wider h-[20px] items-center">
          <span className="text-muted mr-1 font-bold">hotkeys:</span>
          {EMOJI_LIST.map((emoji, idx) => (
            <span key={emoji}>
              ctrl+{idx + 1}:<span className="text-text-bright ml-0.5">{emoji}</span>
            </span>
          ))}
        </div>

        <div className="flex items-center flex-1 h-[45px]">
          <span
            className="pl-4 pr-2 text-sm font-mono shrink-0 select-none font-bold"
            style={{ color: isCmd ? "var(--color-command)" : "var(--color-accent)" }}
          >
            {isCmd ? "$" : "▸"}
          </span>
          <input
            type="text"
            value={input}
            onChange={handleInputChange}
            onKeyDown={handleKeyDown}
            placeholder="type message or command /help"
            className="flex-1 bg-transparent border-none text-xs text-text-bright placeholder:text-dim outline-none"
          />
          <button
            onClick={handleSend}
            className="px-4 h-full text-[10px] font-bold lowercase tracking-wider text-bg bg-accent hover:bg-text-bright transition-colors active:scale-95 shrink-0 select-none"
          >
            exec
          </button>
        </div>
      </div>
    </div>
  );
}
