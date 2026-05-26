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
  const bottomRef = useRef<HTMLDivElement>(null);
  const localName = localId && localId !== "..." ? localId.slice(0, 12) : "local-user";

  useEffect(() => {
    setMessages([
      {
        id: 0,
        sender: "system",
        text: "tsuna watchroom active · welcome node user",
        time: ts(),
        type: "system",
      },
    ]);
  }, []);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const addMsg = (partial: Omit<Message, "id">) =>
    setMessages((prev) => {
      const nextId = prev.length > 0 ? Math.max(...prev.map((m) => m.id)) + 1 : 1;
      return [...prev, { ...partial, id: nextId }];
    });

  const handleSend = () => {
    const text = input.trim();
    if (!text) return;
    setInput("");
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
            text: "available commands:",
            time,
            type: "command",
          });
          addMsg({
            sender: "system",
            text: "/me <action> · show action status to peers",
            time,
            type: "system",
          });
          addMsg({
            sender: "system",
            text: "/react <emoji> · trigger screen reaction",
            time,
            type: "system",
          });
          addMsg({
            sender: "system",
            text: "/status · show node & sync telemetry",
            time,
            type: "system",
          });
          addMsg({
            sender: "system",
            text: "/whois · display your unique node id",
            time,
            type: "system",
          });
          addMsg({
            sender: "system",
            text: "/clear · purge local message log",
            time,
            type: "system",
          });
          return;
        case "/me":
          if (arg) {
            addMsg({ sender: localName, text: `* ${arg}`, time, type: "event" });
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
            text: `you are ${localId === "..." ? "offline / local-user" : localId}`,
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

    addMsg({ sender: localName, text, time, type: "self" });
    send("chat", { text });
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInput(e.target.value);
  };

  const triggerReaction = (emoji: string) => {
    addReaction(emoji);
    addMsg({
      sender: "system",
      text: `you reacted with ${emoji}`,
      time: ts(),
      type: "event",
    });
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === "Enter") {
      handleSend();
    }
  };

  const isCmd = input.startsWith("/");

  return (
    <div className="flex flex-col h-full relative bg-surface">
      <div className="h-[38px] px-4 border-b border-border bg-[#141416] flex items-center shrink-0">
        <span className="text-text-bright font-bold text-[11px]">chatroom</span>
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

      <div className="shrink-0 border-t border-border bg-black/40 h-[66px] flex flex-col justify-between">
        <div className="px-4 text-[10px] text-dim border-b border-border flex items-center justify-between select-none tracking-wider h-[20px]">
          <span className="text-muted font-bold font-mono">reactions:</span>
          <div className="flex gap-3">
            {EMOJI_LIST.map((emoji) => (
              <button
                key={emoji}
                onClick={() => triggerReaction(emoji)}
                className="hover:scale-125 transition-transform duration-100 cursor-pointer active:scale-95 text-xs px-0.5"
              >
                {emoji}
              </button>
            ))}
          </div>
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
            send
          </button>
        </div>
      </div>
    </div>
  );
}
