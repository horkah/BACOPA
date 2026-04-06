'use client';

import React, { useState } from 'react';
import { createGame } from '@/lib/api';

interface NewGameDialogProps {
  gameType: { id: string; name: string; description: string };
  onClose: () => void;
  onCreated: (gameId: string) => void;
}

export default function NewGameDialog({ gameType, onClose, onCreated }: NewGameDialogProps) {
  const [mode, setMode] = useState<'select' | 'ai'>('select');
  const [difficulty, setDifficulty] = useState<'easy' | 'medium' | 'hard'>('medium');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleCreate = async (gameMode: 'pvp' | 'ai', aiDiff?: 'easy' | 'medium' | 'hard') => {
    setLoading(true);
    setError(null);
    try {
      const res = await createGame(gameType.id, gameMode, aiDiff);
      onCreated(res.gameId);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create game');
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 animate-fade-in" onClick={onClose}>
      <div
        className="bg-gray-800 rounded-2xl p-6 max-w-md w-full mx-4 animate-slide-up shadow-2xl border border-gray-700"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 className="text-2xl font-bold text-white mb-1">{gameType.name}</h2>
        <p className="text-gray-400 text-sm mb-6">{gameType.description}</p>

        {error && (
          <div className="bg-red-900/40 border border-red-700 text-red-300 text-sm px-3 py-2 rounded-lg mb-4">
            {error}
          </div>
        )}

        {mode === 'select' ? (
          <div className="space-y-3">
            <button
              onClick={() => setMode('ai')}
              disabled={loading}
              className="w-full bg-gray-700 hover:bg-gray-600 text-white p-4 rounded-xl
                transition-colors text-left disabled:opacity-50"
            >
              <div className="font-semibold text-lg">Play vs AI</div>
              <div className="text-gray-400 text-sm mt-1">Challenge the computer at your skill level</div>
            </button>

            <button
              onClick={() => handleCreate('pvp')}
              disabled={loading}
              className="w-full bg-gray-700 hover:bg-gray-600 text-white p-4 rounded-xl
                transition-colors text-left disabled:opacity-50"
            >
              <div className="font-semibold text-lg">Play vs Human</div>
              <div className="text-gray-400 text-sm mt-1">Create a lobby game and wait for an opponent</div>
            </button>
          </div>
        ) : (
          <div className="space-y-4">
            <p className="text-gray-300 font-medium">Select Difficulty</p>
            <div className="grid grid-cols-3 gap-3">
              {(['easy', 'medium', 'hard'] as const).map((d) => (
                <button
                  key={d}
                  onClick={() => setDifficulty(d)}
                  className={`py-3 px-4 rounded-lg font-semibold capitalize transition-all
                    ${difficulty === d
                      ? 'bg-emerald-600 text-white ring-2 ring-emerald-400'
                      : 'bg-gray-700 text-gray-300 hover:bg-gray-600'
                    }`}
                >
                  {d}
                </button>
              ))}
            </div>
            <button
              onClick={() => handleCreate('ai', difficulty)}
              disabled={loading}
              className="w-full bg-emerald-600 hover:bg-emerald-500 text-white font-semibold
                py-3 rounded-lg transition-colors disabled:opacity-50"
            >
              {loading ? 'Creating...' : 'Start Game'}
            </button>
            <button
              onClick={() => setMode('select')}
              className="w-full text-gray-400 hover:text-white text-sm transition-colors"
            >
              Back
            </button>
          </div>
        )}

        <button
          onClick={onClose}
          className="mt-4 w-full text-gray-500 hover:text-gray-300 text-sm transition-colors"
        >
          Cancel
        </button>
      </div>
    </div>
  );
}
