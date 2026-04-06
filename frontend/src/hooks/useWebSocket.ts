'use client';

import { useState, useEffect, useRef, useCallback } from 'react';

export interface PlayerInfo {
  username: string;
  elo: number;
}

export interface GameState {
  board: number[] | number[][];
  currentPlayer: 1 | 2;
  status: 'waiting' | 'playing' | 'won' | 'draw';
  winner: null | 1 | 2;
  players: {
    1?: PlayerInfo;
    2?: PlayerInfo;
  };
  gameType: string;
  lastMove: number | null;
}

export interface ChatMessage {
  from: string;
  message: string;
  timestamp: string;
}

export interface GameOverData {
  winner: string | null;
  reason: 'win' | 'draw';
  eloChanges: Record<string, number>;
}

export function useWebSocket(gameId: string | null, token: string | null) {
  const [gameState, setGameState] = useState<GameState | null>(null);
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([]);
  const [gameOver, setGameOver] = useState<GameOverData | null>(null);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const retriesRef = useRef(0);
  const maxRetries = 3;

  const connect = useCallback(() => {
    if (!gameId || !token) return;

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = process.env.NEXT_PUBLIC_WS_URL || `${protocol}//${window.location.hostname}:8080`;
    const url = `${host}/ws?token=${encodeURIComponent(token)}&gameId=${encodeURIComponent(gameId)}`;

    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      setConnected(true);
      setError(null);
      retriesRef.current = 0;
    };

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data);
        switch (msg.type) {
          case 'game_state':
            setGameState(msg.data);
            break;
          case 'chat':
            setChatMessages((prev) => [...prev, msg.data]);
            break;
          case 'error':
            setError(msg.data.message);
            break;
          case 'player_joined':
            // game_state will follow
            break;
          case 'game_over':
            setGameOver(msg.data);
            break;
        }
      } catch {
        // ignore malformed messages
      }
    };

    ws.onclose = () => {
      setConnected(false);
      wsRef.current = null;
      if (retriesRef.current < maxRetries) {
        retriesRef.current += 1;
        setTimeout(connect, 1000 * retriesRef.current);
      }
    };

    ws.onerror = () => {
      setError('WebSocket connection error');
    };
  }, [gameId, token]);

  useEffect(() => {
    connect();
    return () => {
      retriesRef.current = maxRetries; // prevent reconnect on unmount
      wsRef.current?.close();
    };
  }, [connect]);

  const sendMove = useCallback((position: number) => {
    wsRef.current?.send(JSON.stringify({ type: 'move', data: { position } }));
  }, []);

  const sendChat = useCallback((message: string) => {
    wsRef.current?.send(JSON.stringify({ type: 'chat', data: { message } }));
  }, []);

  return { gameState, chatMessages, gameOver, sendMove, sendChat, connected, error };
}
