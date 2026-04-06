'use client';

import React, { useState, useRef, useEffect } from 'react';

interface ChatProps {
  messages: { from: string; message: string; timestamp: string }[];
  onSend: (message: string) => void;
  disabled: boolean;
}

export default function Chat({ messages, onSend, disabled }: ChatProps) {
  const [input, setInput] = useState('');
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSend = () => {
    const trimmed = input.trim();
    if (!trimmed) return;
    onSend(trimmed);
    setInput('');
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const formatTime = (ts: string) => {
    try {
      return new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch {
      return '';
    }
  };

  return (
    <div className="flex flex-col h-full bg-gray-800 rounded-lg overflow-hidden">
      <div className="px-3 py-2 border-b border-gray-700 text-sm font-semibold text-gray-300">
        Chat
      </div>
      <div className="flex-1 overflow-y-auto p-3 space-y-2 custom-scrollbar min-h-0">
        {messages.length === 0 && (
          <p className="text-gray-500 text-sm text-center mt-4">No messages yet</p>
        )}
        {messages.map((msg, i) => (
          <div key={i} className="text-sm">
            <span className="font-semibold text-emerald-400">{msg.from}</span>
            <span className="text-gray-500 text-xs ml-2">{formatTime(msg.timestamp)}</span>
            <p className="text-gray-300 mt-0.5">{msg.message}</p>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>
      <div className="p-2 border-t border-gray-700">
        <div className="flex gap-2">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            disabled={disabled}
            placeholder={disabled ? 'Chat unavailable' : 'Type a message...'}
            className="flex-1 bg-gray-700 text-gray-200 rounded-md px-3 py-1.5 text-sm
              placeholder-gray-500 focus:outline-none focus:ring-1 focus:ring-emerald-500
              disabled:opacity-50"
          />
          <button
            onClick={handleSend}
            disabled={disabled || !input.trim()}
            className="bg-emerald-600 hover:bg-emerald-500 disabled:bg-gray-600
              disabled:cursor-not-allowed text-white text-sm px-3 py-1.5 rounded-md
              transition-colors"
          >
            Send
          </button>
        </div>
      </div>
    </div>
  );
}
