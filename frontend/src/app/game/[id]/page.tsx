'use client';

import React, { use } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';
import { useWebSocket } from '@/hooks/useWebSocket';
import TicTacToe from '@/components/TicTacToe';
import ConnectFour from '@/components/ConnectFour';
import Chat from '@/components/Chat';
import GameOverModal from '@/components/GameOverModal';

export default function GamePage({ params }: { params: Promise<{ id: string }> }) {
  const { id: gameId } = use(params);
  const router = useRouter();
  const { user, token } = useAuth();
  const { gameState, chatMessages, gameOver, sendMove, sendChat, connected, error } = useWebSocket(gameId, token);

  // Determine which player number we are
  const myPlayer: 1 | 2 = (() => {
    if (!gameState?.players || !user) return 1;
    if (gameState.players[1]?.username === user.username) return 1;
    if (gameState.players[2]?.username === user.username) return 2;
    return 1;
  })();

  if (!user) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center">
          <p className="text-gray-400 mb-4">You need to be logged in to play.</p>
          <button
            onClick={() => router.push('/login')}
            className="bg-emerald-600 hover:bg-emerald-500 text-white px-6 py-2 rounded-lg transition-colors"
          >
            Sign In
          </button>
        </div>
      </div>
    );
  }

  if (!connected && !gameState) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center">
          <div className="w-8 h-8 border-2 border-emerald-400 border-t-transparent rounded-full animate-spin mx-auto mb-4" />
          <p className="text-gray-400">Connecting to game...</p>
          {error && <p className="text-red-400 text-sm mt-2">{error}</p>}
        </div>
      </div>
    );
  }

  const isWaiting = gameState?.status === 'waiting';
  const isGameOver = gameState?.status === 'won' || gameState?.status === 'draw';
  const isMyTurn = gameState?.currentPlayer === myPlayer;

  const gameTypeLower = gameState?.gameType?.toLowerCase() || '';
  const isTicTacToe = gameTypeLower.includes('tic') || gameTypeLower.includes('ttt');
  const isConnectFour = gameTypeLower.includes('connect') || gameTypeLower.includes('c4');

  return (
    <div className="flex-1 flex flex-col lg:flex-row max-w-7xl mx-auto w-full px-4 py-6 gap-4">
      {/* Main game area */}
      <div className="flex-1 flex flex-col items-center">
        {/* Turn indicator */}
        {gameState && gameState.status === 'playing' && (
          <div className={`mb-4 px-4 py-2 rounded-lg text-sm font-semibold ${
            isMyTurn
              ? 'bg-emerald-600/20 text-emerald-400 animate-pulse-glow'
              : 'bg-gray-800 text-gray-400'
          }`}>
            {isMyTurn ? 'Your turn!' : `Waiting for ${
              gameState.players[gameState.currentPlayer]?.username || 'opponent'
            }...`}
          </div>
        )}

        {/* Waiting state */}
        {isWaiting && (
          <div className="flex-1 flex items-center justify-center">
            <div className="bg-gray-800 rounded-xl p-8 border border-gray-700 text-center max-w-md">
              <div className="w-10 h-10 border-2 border-emerald-400 border-t-transparent rounded-full animate-spin mx-auto mb-4" />
              <h2 className="text-xl font-bold text-white mb-2">Waiting for opponent...</h2>
              <p className="text-gray-400 text-sm mb-4">Share this game ID with a friend:</p>
              <div className="bg-gray-900 rounded-lg px-4 py-2 font-mono text-emerald-400 text-sm select-all">
                {gameId}
              </div>
            </div>
          </div>
        )}

        {/* Game board */}
        {gameState && !isWaiting && (
          <div className="flex-1 flex items-center justify-center w-full">
            {isTicTacToe && Array.isArray(gameState.board) && !Array.isArray(gameState.board[0]) && (
              <div className="max-w-md w-full">
                <TicTacToe
                  board={gameState.board as number[]}
                  currentPlayer={gameState.currentPlayer}
                  myPlayer={myPlayer}
                  onMove={sendMove}
                  lastMove={gameState.lastMove}
                  disabled={isGameOver || !isMyTurn}
                />
              </div>
            )}
            {isConnectFour && Array.isArray(gameState.board) && Array.isArray(gameState.board[0]) && (
              <div className="max-w-lg w-full flex justify-center">
                <ConnectFour
                  board={gameState.board as number[][]}
                  currentPlayer={gameState.currentPlayer}
                  myPlayer={myPlayer}
                  onMove={sendMove}
                  lastMove={gameState.lastMove}
                  disabled={isGameOver || !isMyTurn}
                />
              </div>
            )}
            {!isTicTacToe && !isConnectFour && (
              <div className="text-gray-400">
                Unknown game type: {gameState.gameType}
              </div>
            )}
          </div>
        )}
      </div>

      {/* Sidebar */}
      <div className="w-full lg:w-80 flex flex-col gap-4 lg:min-h-0">
        {/* Player info */}
        <div className="bg-gray-800 rounded-lg p-4 border border-gray-700">
          <h3 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-3">Players</h3>
          <div className="space-y-3">
            {([1, 2] as const).map((pNum) => {
              const player = gameState?.players[pNum];
              const isCurrent = gameState?.currentPlayer === pNum;
              const isMe = pNum === myPlayer;
              return (
                <div
                  key={pNum}
                  className={`flex items-center justify-between p-2 rounded-lg transition-colors ${
                    isCurrent && gameState?.status === 'playing'
                      ? 'bg-emerald-900/20 border border-emerald-700/40'
                      : 'bg-gray-900/50'
                  }`}
                >
                  <div className="flex items-center gap-2">
                    <div className={`w-3 h-3 rounded-full ${
                      isTicTacToe
                        ? pNum === 1 ? 'bg-blue-400' : 'bg-red-400'
                        : pNum === 1 ? 'bg-red-500' : 'bg-yellow-400'
                    }`} />
                    <span className="text-sm font-medium text-gray-200">
                      {player?.username || (isWaiting && pNum === 2 ? 'Waiting...' : `Player ${pNum}`)}
                    </span>
                    {isMe && (
                      <span className="text-xs text-emerald-400 bg-emerald-400/10 px-1.5 py-0.5 rounded">
                        you
                      </span>
                    )}
                  </div>
                  {player?.elo !== undefined && (
                    <span className="text-xs text-gray-400">{player.elo} ELO</span>
                  )}
                </div>
              );
            })}
          </div>
        </div>

        {/* Connection status */}
        {!connected && (
          <div className="bg-yellow-900/30 border border-yellow-700/50 text-yellow-400 text-xs px-3 py-2 rounded-lg">
            Reconnecting...
          </div>
        )}
        {error && (
          <div className="bg-red-900/30 border border-red-700/50 text-red-400 text-xs px-3 py-2 rounded-lg">
            {error}
          </div>
        )}

        {/* Chat */}
        <div className="flex-1 min-h-[200px] lg:min-h-0 flex flex-col">
          <Chat
            messages={chatMessages}
            onSend={sendChat}
            disabled={!connected}
          />
        </div>
      </div>

      {/* Game Over Modal */}
      {gameOver && (
        <GameOverModal
          winner={gameOver.winner}
          reason={gameOver.reason}
          eloChanges={gameOver.eloChanges}
          myUsername={user.username}
          onClose={() => router.push('/')}
        />
      )}
    </div>
  );
}
