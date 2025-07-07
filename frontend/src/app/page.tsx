"use client";

import Link from "next/link";
import { Button } from "@/components/ui/button";
import { useAuth } from "@/context/AuthContext";

export default function Home() {
  const { user, logout } = useAuth();

  return (
    <div className="flex flex-col items-center justify-center min-h-screen py-2">
      <main className="flex flex-col items-center justify-center flex-1 px-20 text-center">
        <h1 className="text-4xl font-bold mb-6">
          Welcome to Async Multiplayer Management
        </h1>

        <div className="flex flex-col space-y-4">
          {user ? (
            <Button size="lg" onClick={logout}>
              Logout ({user.email})
            </Button>
          ) : (
            <a href={`${process.env.NEXT_PUBLIC_API_URL}/auth/google/login`}>
              <Button size="lg">Login with Google</Button>
            </a>
          )}
          <Link href="/create-game">
            <Button variant="outline" size="lg">Create Game</Button>
          </Link>
          <Link href="/join-game">
            <Button variant="outline" size="lg">Join Game</Button>
          </Link>
          <Link href="/dashboard">
            <Button variant="outline" size="lg">Dashboard</Button>
          </Link>
        </div>
      </main>
    </div>
  );
}
