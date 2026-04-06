'use client';

import React, { useEffect, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';
import { getGameTypes, getLobbyGames, getGameHistory, joinGame } from '@/lib/api';
import type { GameType, LobbyGame, GameHistoryEntry } from '@/lib/api';
import NewGameDialog from '@/components/NewGameDialog';

function HeroSection() {
  return (
    <div className="flex-1 flex flex-col items-center justify-center px-4 py-20">
      <div className="text-center max-w-2xl">
        <h1 className="text-6xl sm:text-7xl font-bold text-white mb-4 tracking-tight">
          BAC<span className="text-emerald-400">OPA</span>
        </h1>
        <p className="text-xl sm:text-2xl text-gray-400 mb-8">
          Open Creative Multi-Game Playground
        </p>
        <p className="text-gray-500 mb-10 max-w-md mx-auto">
          Play classic board games against friends or AI. Track your Elo rating and climb the ranks.
        </p>
        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <Link
            href="/register"
            className="bg-emerald-600 hover:bg-emerald-500 text-white font-semibold
              px-8 py-3 rounded-lg transition-colors text-lg"
          >
            Get Started
          </Link>
          <Link
            href="/login"
            className="bg-gray-800 hover:bg-gray-700 text-gray-300 hover:text-white
              font-semibold px-8 py-3 rounded-lg transition-colors text-lg border border-gray-700"
          >
            Sign In
          </Link>
        </div>
      </div>
    </div>
  );
}

function Dashboard() {
  const router = useRouter();
  const [gameTypes, setGameTypes] = useState<GameType[]>([]);
  const [lobbyGames, setLobbyGames] = useState<LobbyGame[]>([]);
  const [history, setHistory] = useState<GameHistoryEntry[]>([]);
  const [selectedType, setSelectedType] = useState<GameType | null>(null);
  const [loading, setLoading] = useState(true);
  const [joiningId, setJoiningId] = useState<string | null>(null);

  useEffect(() => {
    Promise.all([
      getGameTypes().catch(() => []),
      getLobbyGames().catch(() => []),
      getGameHistory().catch(() => []),
    ]).then(([types, lobby, hist]) => {
      setGameTypes(types);
      setLobbyGames(lobby);
      setHistory(hist);
      setLoading(false);
    });
  }, []);

  const handleJoin = async (gameId: string) => {
    setJoiningId(gameId);
    try {
      await joinGame(gameId);
      router.push(`/game/${gameId}`);
    } catch {
      setJoiningId(null);
    }
  };

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="w-8 h-8 border-2 border-emerald-400 border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  return (
    <div className="max-w-6xl mx-auto w-full px-4 sm:px-6 lg:px-8 py-8 space-y-10">
      {/* New Game Section */}
      <section>
        <h2 className="text-2xl font-bold text-white mb-4">New Game</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {gameTypes.map((gt) => (
            <div
              key={gt.id}
              className="bg-gray-800 rounded-xl p-5 border border-gray-700
                hover:border-emerald-500/50 transition-all group"
            >
              <h3 className="text-lg font-semibold text-white group-hover:text-emerald-400 transition-colors">
                {gt.name}
              </h3>
              <p className="text-gray-400 text-sm mt-1 mb-4">{gt.description}</p>
              <p className="text-gray-500 text-xs mb-4">
                {gt.minPlayers}-{gt.maxPlayers} players
              </p>
              <button
                onClick={() => setSelectedType(gt)}
                className="w-full bg-emerald-600 hover:bg-emerald-500 text-white
                  font-medium py-2 rounded-lg transition-colors text-sm"
              >
                Play
              </button>
            </div>
          ))}
          {gameTypes.length === 0 && (
            <p className="text-gray-500 col-span-full">No game types available.</p>
          )}
        </div>
      </section>

      {/* Open Games Lobby */}
      <section>
        <h2 className="text-2xl font-bold text-white mb-4">Open Games</h2>
        {lobbyGames.length === 0 ? (
          <div className="bg-gray-800 rounded-xl p-6 border border-gray-700 text-center text-gray-500">
            No open games right now. Create one!
          </div>
        ) : (
          <div className="space-y-2">
            {lobbyGames.map((g) => (
              <div
                key={g.gameId}
                className="bg-gray-800 rounded-lg px-5 py-3 border border-gray-700
                  flex items-center justify-between"
              >
                <div className="flex items-center gap-4">
                  <span className="text-sm font-medium text-emerald-400 bg-emerald-400/10 px-2 py-0.5 rounded">
                    {g.gameType}
                  </span>
                  <span className="text-gray-300 text-sm">by {g.creator}</span>
                  <span className="text-gray-500 text-xs">
                    {new Date(g.createdAt).toLocaleTimeString()}
                  </span>
                </div>
                <button
                  onClick={() => handleJoin(g.gameId)}
                  disabled={joiningId === g.gameId}
                  className="bg-emerald-600 hover:bg-emerald-500 disabled:bg-gray-600
                    text-white text-sm font-medium px-4 py-1.5 rounded-md transition-colors"
                >
                  {joiningId === g.gameId ? 'Joining...' : 'Join'}
                </button>
              </div>
            ))}
          </div>
        )}
      </section>

      {/* Recent Games */}
      <section>
        <h2 className="text-2xl font-bold text-white mb-4">Recent Games</h2>
        {history.length === 0 ? (
          <div className="bg-gray-800 rounded-xl p-6 border border-gray-700 text-center text-gray-500">
            No games played yet. Start your first match!
          </div>
        ) : (
          <div className="space-y-2">
            {history.slice(0, 10).map((h, i) => (
              <div
                key={i}
                className="bg-gray-800 rounded-lg px-5 py-3 border border-gray-700
                  flex items-center justify-between"
              >
                <div className="flex items-center gap-4">
                  <span className="text-sm font-medium text-emerald-400 bg-emerald-400/10 px-2 py-0.5 rounded">
                    {h.gameType}
                  </span>
                  <span className="text-gray-300 text-sm">vs {h.opponent || 'AI'}</span>
                  <span className={`text-sm font-semibold ${
                    h.result === 'win' ? 'text-emerald-400' :
                    h.result === 'loss' ? 'text-red-400' : 'text-yellow-400'
                  }`}>
                    {h.result.toUpperCase()}
                  </span>
                </div>
                <div className="flex items-center gap-3">
                  <span className={`text-sm font-semibold ${
                    h.eloChange >= 0 ? 'text-emerald-400' : 'text-red-400'
                  }`}>
                    {h.eloChange >= 0 ? '+' : ''}{h.eloChange}
                  </span>
                  <span className="text-gray-500 text-xs">
                    {new Date(h.playedAt).toLocaleDateString()}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </section>

      {selectedType && (
        <NewGameDialog
          gameType={selectedType}
          onClose={() => setSelectedType(null)}
          onCreated={(gameId) => {
            setSelectedType(null);
            router.push(`/game/${gameId}`);
          }}
        />
      )}
    </div>
  );
}

export default function Home() {
  const { user, loading } = useAuth();

  if (loading) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="w-8 h-8 border-2 border-emerald-400 border-t-transparent rounded-full animate-spin" />
      </div>
    );
  }

  return user ? <Dashboard /> : <HeroSection />;
}
