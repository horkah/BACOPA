'use client';

import React from 'react';
import Link from 'next/link';
import { useAuth } from '@/context/AuthContext';

export default function NavBar() {
  const { user, loading, logout } = useAuth();

  return (
    <nav className="bg-gray-900 border-b border-gray-800 sticky top-0 z-40">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-14">
          <Link href="/" className="flex items-center gap-2 group">
            <span className="text-2xl font-bold text-emerald-400 group-hover:text-emerald-300 transition-colors tracking-tight">
              BACOPA
            </span>
          </Link>

          <div className="flex items-center gap-4">
            {loading ? (
              <div className="w-20 h-5 bg-gray-700 rounded animate-pulse" />
            ) : user ? (
              <>
                <div className="hidden sm:flex items-center gap-2 text-sm">
                  <span className="text-gray-300 font-medium">{user.username}</span>
                  <span className="text-emerald-400 font-semibold bg-emerald-400/10 px-2 py-0.5 rounded-full text-xs">
                    {user.elo} ELO
                  </span>
                </div>
                <button
                  onClick={logout}
                  className="text-sm text-gray-400 hover:text-white transition-colors
                    bg-gray-800 hover:bg-gray-700 px-3 py-1.5 rounded-md"
                >
                  Logout
                </button>
              </>
            ) : (
              <>
                <Link
                  href="/login"
                  className="text-sm text-gray-300 hover:text-white transition-colors"
                >
                  Login
                </Link>
                <Link
                  href="/register"
                  className="text-sm bg-emerald-600 hover:bg-emerald-500 text-white
                    px-4 py-1.5 rounded-md transition-colors font-medium"
                >
                  Register
                </Link>
              </>
            )}
          </div>
        </div>
      </div>
    </nav>
  );
}
