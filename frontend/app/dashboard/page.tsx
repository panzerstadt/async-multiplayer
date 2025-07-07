"use client";

import { useQuery } from "@tanstack/react-query";
import { getGames } from "@/services/api"; // Assuming a getGames function exists in your API service

export default function DashboardPage() {
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
            <div key={game.id} className="p-4 border rounded-lg">
              <h3 className="font-bold">{game.name}</h3>
              <p>Next turn: {game.current_turn_id}</p>
              <button className="mt-2 px-4 py-2 text-white bg-blue-500 rounded">View Game</button>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}