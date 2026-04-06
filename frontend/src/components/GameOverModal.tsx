'use client';

import React from 'react';

interface GameOverModalProps {
  winner: string | null;
  reason: 'win' | 'draw';
  eloChanges: Record<string, number>;
  myUsername: string;
  onClose: () => void;
}

export default function GameOverModal({ winner, reason, eloChanges, myUsername, onClose }: GameOverModalProps) {
  const isWinner = winner === myUsername;
  const isDraw = reason === 'draw';
  const myEloChange = eloChanges[myUsername] ?? 0;

  let title: string;
  let titleColor: string;
  if (isDraw) {
    title = "It's a Draw!";
    titleColor = 'text-yellow-400';
  } else if (isWinner) {
    title = 'You Win!';
    titleColor = 'text-emerald-400';
  } else {
    title = 'You Lose!';
    titleColor = 'text-red-400';
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 animate-fade-in">
      <div className="bg-gray-800 rounded-2xl p-8 max-w-sm w-full mx-4 text-center animate-slide-up shadow-2xl border border-gray-700">
        <h2 className={`text-4xl font-bold mb-4 ${titleColor}`}>{title}</h2>

        {!isDraw && winner && (
          <p className="text-gray-300 mb-2">
            Winner: <span className="font-semibold text-white">{winner}</span>
          </p>
        )}

        <div className="my-6 p-4 bg-gray-900 rounded-lg">
          <p className="text-sm text-gray-400 mb-2">Elo Change</p>
          <p className={`text-3xl font-bold ${myEloChange >= 0 ? 'text-emerald-400' : 'text-red-400'}`}>
            {myEloChange >= 0 ? '+' : ''}{myEloChange}
          </p>
        </div>

        {Object.keys(eloChanges).length > 1 && (
          <div className="mb-4 space-y-1">
            {Object.entries(eloChanges).map(([name, change]) => (
              <div key={name} className="flex justify-between text-sm text-gray-400">
                <span>{name}</span>
                <span className={change >= 0 ? 'text-emerald-400' : 'text-red-400'}>
                  {change >= 0 ? '+' : ''}{change}
                </span>
              </div>
            ))}
          </div>
        )}

        <button
          onClick={onClose}
          className="w-full bg-emerald-600 hover:bg-emerald-500 text-white font-semibold
            py-3 px-6 rounded-lg transition-colors text-lg"
        >
          Back to Lobby
        </button>
      </div>
    </div>
  );
}
