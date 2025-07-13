"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { Label } from "./ui/label";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "./ui/card";
import { getLatestSave, uploadSave } from "@/services/api";
import Link from "next/link";
import { AxiosError } from "axios";

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
        <div className="mt-4">
          <Button onClick={handleDownload} className="w-full">
            Download Latest Save
          </Button>
        </div>
      </CardContent>
      <CardFooter>
        {showViewGameButton && (
          <Link href={`/games/${game.id}`}>
            <Button>View Game</Button>
          </Link>
        )}
      </CardFooter>
    </Card>
  );
}
