"use client";

import { useMutation } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { createGame } from "@/services/api";

import { toast } from "sonner";

const createGameSchema = z.object({
  name: z.string().min(1, "Game name is required"),
  players: z.string().min(1, "At least one player is required"),
});

type CreateGameValues = z.infer<typeof createGameSchema>;

export default function CreateGamePage() {
  const mutation = useMutation({
    mutationFn: (data: { name: string; players: string[] }) => createGame(data.name, data.players),
    onSuccess: () => {
      toast.success("Game created successfully!");
    },
    onError: (error) => {
      toast.error(`An error occurred: ${error.message}`);
    },
  });

  const { register, handleSubmit, formState: { errors } } = useForm<CreateGameValues>({
    resolver: zodResolver(createGameSchema),
  });

  const onSubmit = (data: CreateGameValues) => {
    const players = data.players.split(",").map((email) => email.trim());
    mutation.mutate({ name: data.name, players });
  };

  return (
    <div className="flex items-center justify-center h-screen">
      <form onSubmit={handleSubmit(onSubmit)} className="w-full max-w-md p-8 space-y-8 bg-white rounded-lg shadow-md">
        <h1 className="text-2xl font-bold text-center">Create a New Game</h1>
        <div className="space-y-4">
          <div>
            <Label htmlFor="name">Game Name</Label>
            <Input id="name" placeholder="Enter game name" {...register("name")} />
            {errors.name && <p className="text-red-500">{errors.name.message}</p>}
          </div>
          <div>
            <Label htmlFor="players">Players</Label>
            <Input id="players" placeholder="Enter player emails, separated by commas" {...register("players")} />
            {errors.players && <p className="text-red-500">{errors.players.message}</p>}
          </div>
          <Button type="submit" className="w-full" disabled={mutation.isPending}>
            {mutation.isPending ? "Creating..." : "Create Game"}
          </Button>
        </div>
      </form>
    </div>
  );
}
