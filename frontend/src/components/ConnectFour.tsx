'use client';

import React, { useState } from 'react';

interface ConnectFourProps {
  board: number[][];
  currentPlayer: number;
  myPlayer: number;
  onMove: (column: number) => void;
  lastMove: number | null;
  disabled: boolean;
}

export default function ConnectFour({ board, currentPlayer, myPlayer, onMove, lastMove, disabled }: ConnectFourProps) {
  const [hoverCol, setHoverCol] = useState<number | null>(null);
  const isMyTurn = currentPlayer === myPlayer && !disabled;
  const rows = board.length;    // 6
  const cols = board[0]?.length || 7;

  // lastMove is column index for connect four
  // Find the row of the last move for highlighting
  let lastMoveRow: number | null = null;
  if (lastMove !== null) {
    for (let r = 0; r < rows; r++) {
      if (board[r][lastMove] !== 0) {
        lastMoveRow = r;
        break;
      }
    }
  }

  const canDropInCol = (col: number) => board[0][col] === 0;

  return (
    <div className="flex flex-col items-center gap-2">
      {/* Column hover indicators */}
      <div className="grid gap-1" style={{ gridTemplateColumns: `repeat(${cols}, 1fr)` }}>
        {Array.from({ length: cols }, (_, col) => {
          const canDrop = canDropInCol(col) && isMyTurn;
          return (
            <button
              key={col}
              onClick={() => canDrop && onMove(col)}
              onMouseEnter={() => setHoverCol(col)}
              onMouseLeave={() => setHoverCol(null)}
              disabled={!canDrop}
              className={`
                w-10 h-6 sm:w-12 sm:h-8 flex items-center justify-center
                rounded-t-md text-xs font-bold transition-all
                ${canDrop
                  ? 'hover:bg-gray-600 cursor-pointer text-gray-400 hover:text-white'
                  : 'cursor-not-allowed text-transparent'
                }
              `}
            >
              {canDrop ? '\u25BC' : ''}
            </button>
          );
        })}
      </div>

      {/* Board */}
      <div className="bg-blue-800 rounded-xl p-2 sm:p-3 shadow-2xl">
        <div className="grid gap-1" style={{ gridTemplateColumns: `repeat(${cols}, 1fr)` }}>
          {board.map((row, r) =>
            row.map((cell, c) => {
              const isLast = lastMove === c && lastMoveRow === r;
              const isHovered = hoverCol === c && cell === 0 && isMyTurn;

              return (
                <button
                  key={`${r}-${c}`}
                  onClick={() => canDropInCol(c) && isMyTurn && onMove(c)}
                  className={`
                    w-10 h-10 sm:w-12 sm:h-12 rounded-full
                    transition-all duration-200
                    ${cell === 0
                      ? isHovered
                        ? myPlayer === 1 ? 'bg-red-400/30' : 'bg-yellow-400/30'
                        : 'bg-gray-900'
                      : cell === 1
                        ? 'bg-red-500'
                        : 'bg-yellow-400'
                    }
                    ${isLast ? 'ring-2 ring-white/60' : ''}
                    ${cell !== 0 ? 'animate-drop-in' : ''}
                    ${cell === 0 && isMyTurn && canDropInCol(c) ? 'cursor-pointer hover:bg-gray-700' : 'cursor-default'}
                  `}
                  disabled={!isMyTurn || cell !== 0 || !canDropInCol(c)}
                />
              );
            })
          )}
        </div>
      </div>

      {/* Legend */}
      <div className="flex gap-4 text-sm text-gray-400 mt-1">
        <span className="flex items-center gap-1">
          <span className="w-3 h-3 rounded-full bg-red-500 inline-block" /> Player 1
        </span>
        <span className="flex items-center gap-1">
          <span className="w-3 h-3 rounded-full bg-yellow-400 inline-block" /> Player 2
        </span>
      </div>
    </div>
  );
}
