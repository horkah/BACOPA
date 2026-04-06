import type { Metadata } from 'next';
import './globals.css';
import { AuthProvider } from '@/context/AuthContext';
import NavBar from '@/components/NavBar';

export const metadata: Metadata = {
  title: 'BACOPA - Multi-Game Platform',
  description: 'Open Creative Multi-Game Playground',
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="h-full antialiased">
      <body className="min-h-full flex flex-col bg-gray-900 text-gray-100 font-sans">
        <AuthProvider>
          <NavBar />
          <main className="flex-1 flex flex-col">{children}</main>
        </AuthProvider>
      </body>
    </html>
  );
}
