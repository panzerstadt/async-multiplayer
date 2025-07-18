"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Label } from "./ui/label";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "./ui/card";
import { broadcastToPlayers, deleteGame, getLatestSave, uploadSave } from "@/services/api";
import Link from "next/link";
import { AxiosError } from "axios";
import { useAuth } from "@/context/AuthContext";

interface GameCardProps {
  game: {
    id: string;
    name: string;
    creator_id: string;
    current_turn_id?: string;
    players?: { id: string; user_id: string; turn_order: number; user: { email: string } }[];
  };
  showViewGameButton?: boolean;
}

export default function GameCard({ game, showViewGameButton = true }: GameCardProps) {
  const queryClient = useQueryClient();
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const { user } = useAuth();

  const uploadMutation = useMutation({
    mutationFn: (data: { gameId: string; file: File }) => uploadSave(data.gameId, data.file),
    onSuccess: () => {
      toast.success("Save file uploaded successfully!");
      queryClient.invalidateQueries({ queryKey: ["games"] });
    },
    onError: (error: AxiosError<{ error: string }>) => {
      const errorMessage = error.response?.data?.error || error.message;
      toast.error(`Error uploading save: ${errorMessage}`);
    },
  });

  const broadcastMutation = useMutation({
    mutationFn: (data: { gameId: string; message: string }) =>
      broadcastToPlayers(data.gameId, data.message),
    onSuccess: () => {
      console.log("Message sent successfully!");
    },
    onError: (error: AxiosError<{ error: string }>) => {
      const errorMessage = error.response?.data?.error || error.message;
      toast.error(`Error sending message: ${errorMessage}`);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: deleteGame,
    onSuccess: () => {
      toast.success("Game deleted successfully!");
      queryClient.invalidateQueries({ queryKey: ["games"] });
    },
    onError: (error: AxiosError<{ error: string }>) => {
      const errorMessage = error.response?.data?.error || error.message;
      toast.error(`Error deleting game: ${errorMessage}`);
    },
  });

  const handleFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files && event.target.files[0]) {
      setSelectedFile(event.target.files[0]);
    }
  };

  const handleUpload = () => {
    if (selectedFile) {
      uploadMutation.mutate({ gameId: game.id, file: selectedFile });
    }
  };
  const handleBroadcast = () => {
    broadcastMutation.mutate({ gameId: game.id, message: "yo" });
  };

  const handleDownload = async () => {
    try {
      const data = await getLatestSave(game.id);
      const url = window.URL.createObjectURL(new Blob([data]));
      const link = document.createElement("a");
      link.href = url;
      link.setAttribute("download", `${game.name}_latest.zip`);
      document.body.appendChild(link);
      link.click();
      link.remove();
      toast.success("Latest save downloaded!");
    } catch (error) {
      const axiosError = error as AxiosError<{ error: string }>;
      const errorMessage = axiosError.response?.data?.error || axiosError.message;
      toast.error(`Error downloading save: ${errorMessage}`);
    }
  };

  const handleDelete = () => {
    if (
      window.confirm("Are you sure you want to delete this game? This action cannot be undone.")
    ) {
      deleteMutation.mutate(game.id);
    }
  };

  return (
    <Card className="max-w-[350px]">
      <CardHeader>
        <CardTitle>{game.name}</CardTitle>
        <CardDescription>Game ID: {game.id}</CardDescription>
      </CardHeader>
      <CardContent>
        {game.players && (
          <div>
            <h4 className="text-md font-semibold">Players:</h4>
            <ul>
              {game.players?.map((player) => (
                <li key={player.id}>
                  Player {player.turn_order + 1}: {player.user.email}{" "}
                  {game.current_turn_id === player.id ? "(Current Turn)" : ""}
                </li>
              ))}
            </ul>
          </div>
        )}
        <div className="mt-4 flex flex-col gap-2">
          <Label htmlFor="save-file">Upload Save</Label>
          <Input id="save-file" type="file" onChange={handleFileChange} />
          <Button
            onClick={handleUpload}
            disabled={uploadMutation.isPending || !selectedFile}
            className="mt-2"
          >
            {uploadMutation.isPending ? "Uploading..." : "Upload Save"}
          </Button>
        </div>
        <div className="mt-2 gap-2 flex flex-col">
          <Button onClick={handleDownload} className="w-full">
            Download Latest Save
          </Button>
          <Button onClick={handleBroadcast} className="w-full">
            Ping
          </Button>
        </div>
      </CardContent>
      <CardFooter className="flex justify-between">
        {showViewGameButton && (
          <Link href={`/games/${game.id}`}>
            <Button>View Game</Button>
          </Link>
        )}
        {user?.id === game.creator_id && (
          <Button variant="destructive" onClick={handleDelete} disabled={deleteMutation.isPending}>
            {deleteMutation.isPending ? "Deleting..." : "Delete Game"}
          </Button>
        )}
      </CardFooter>
    </Card>
  );
}
