"use client";

import { useQuery } from "@tanstack/react-query";
import { getGames } from "@/services/api";
import GameCard from "@/components/GameCard";
import withAuth from "@/components/withAuth";
import Link from "next/link";
import { Button } from "@/components/ui/button";

interface Game {
  id: string;
  name: string;
  creator_id: string;
  current_turn_id?: string;
  players?: { id: string; user_id: string; turn_order: number; user: { email: string } }[];
}

function DashboardPage() {
  const {
    data: games,
    isLoading,
    isError,
    error,
  } = useQuery({ queryKey: ["games"], queryFn: getGames });

  if (isLoading) return <div>Loading...</div>;
  if (isError) return <div>Error fetching games</div>;

  if (!games || games.length === 0) {
    return (
      <div className="container mx-auto p-4">
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <div className="mt-4">
          <p>You are not a part of any games yet.</p>
          <Link href="/create-game">
            <Button className="mt-2">Create a Game</Button>
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-4">
      <h1 className="text-2xl font-bold">Dashboard</h1>
      <div className="mt-4">
        <h2 className="text-xl font-semibold">Your Games</h2>
        <div className="grid grid-cols-1 gap-4 mt-2 md:grid-cols-2 lg:grid-cols-3">
          {games?.map((game: Game) => (
            <GameCard key={game.id} game={game} />
          ))}
        </div>
        {error ? <span>{(error as Error).message}</span> : null}
      </div>
    </div>
  );
}

export default withAuth(DashboardPage);
