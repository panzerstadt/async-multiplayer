"use client";

import { useQuery } from "@tanstack/react-query";
import { getGames } from "@/services/api";
import GameCard from "@/components/GameCard";
import withAuth from "@/components/withAuth";

function DashboardPage() {
  const { data: games, isLoading, isError } = useQuery({ queryKey: ["games"], queryFn: getGames });

  if (isLoading) return <div>Loading...</div>;
  if (isError) return <div>Error fetching games</div>;

  return (
    <div className="container mx-auto p-4">
      <h1 className="text-2xl font-bold">Dashboard</h1>
      <div className="mt-4">
        <h2 className="text-xl font-semibold">Your Games</h2>
        <div className="grid grid-cols-1 gap-4 mt-2 md:grid-cols-2 lg:grid-cols-3">
          {games?.map((game: any) => (
            <GameCard key={game.id} game={game} />
          ))}
        </div>
      </div>
    </div>
  );
}

export default withAuth(DashboardPage);
