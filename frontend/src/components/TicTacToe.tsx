'use client';

import React from 'react';

interface TicTacToeProps {
  board: number[];
  currentPlayer: number;
  myPlayer: number;
  onMove: (position: number) => void;
  lastMove: number | null;
  disabled: boolean;
}

const WINNING_LINES = [
  [0, 1, 2], [3, 4, 5], [6, 7, 8],
  [0, 3, 6], [1, 4, 7], [2, 5, 8],
  [0, 4, 8], [2, 4, 6],
];

function getWinningLine(board: number[]): number[] | null {
  for (const line of WINNING_LINES) {
    const [a, b, c] = line;
    if (board[a] !== 0 && board[a] === board[b] && board[a] === board[c]) {
      return line;
    }
  }
  return null;
}

export default function TicTacToe({ board, currentPlayer, myPlayer, onMove, lastMove, disabled }: TicTacToeProps) {
  const winLine = getWinningLine(board);
  const isMyTurn = currentPlayer === myPlayer && !disabled;

  return (
    <div className="flex flex-col items-center gap-4">
      <div className="grid grid-cols-3 gap-2 max-w-xs w-full aspect-square">
        {board.map((cell, idx) => {
          const isWinCell = winLine?.includes(idx);
          const isLast = lastMove === idx;
          const canClick = cell === 0 && isMyTurn;

          return (
            <button
              key={idx}
              onClick={() => canClick && onMove(idx)}
              disabled={!canClick}
              className={`
                aspect-square rounded-lg text-4xl sm:text-5xl font-bold
                flex items-center justify-center
                transition-all duration-200
                ${cell === 0
                  ? canClick
                    ? 'bg-gray-700 hover:bg-gray-600 cursor-pointer'
                    : 'bg-gray-700/50 cursor-not-allowed'
                  : 'bg-gray-700'
                }
                ${isWinCell ? 'ring-2 ring-emerald-400 bg-emerald-900/30' : ''}
                ${isLast && cell !== 0 ? 'ring-2 ring-yellow-400/60' : ''}
              `}
            >
              {cell !== 0 && (
                <span className={`animate-pop-in ${cell === 1 ? 'text-blue-400' : 'text-red-400'}`}>
                  {cell === 1 ? 'X' : 'O'}
                </span>
              )}
            </button>
          );
        })}
      </div>
    </div>
  );
}
