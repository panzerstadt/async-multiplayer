"use client";

import { useQuery } from "@tanstack/react-query";
import { getGame } from "@/services/api";
import GameCard from "@/components/GameCard";
import withAuth from "@/components/withAuth";
import { use } from "react";

function GamePage({ params }: { params: Promise<{ id: string }> }) {
  const prms = use(params);
  const {
    data: game,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ["game", prms.id],
    queryFn: () => getGame(prms.id),
  });

  if (isLoading) return <div>Loading...</div>;
  if (isError) return <div>Error fetching game data</div>;
  if (!game) return <div>Game not found</div>;

  return (
    <div className="container mx-auto p-4">
      <GameCard game={game} showViewGameButton={false} />
    </div>
  );
}

export default withAuth(GamePage);
