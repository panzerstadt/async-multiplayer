"use client";

import { useMutation } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { joinGame } from "@/services/api";

import { toast } from "sonner";

const joinGameSchema = z.object({
  name: z.string().min(1, "Game name is required"),
});

type JoinGameValues = z.infer<typeof joinGameSchema>;

export default function JoinGamePage() {
  const mutation = useMutation({
    mutationFn: (data: { name: string }) => joinGame(data.name),
    onSuccess: () => {
      toast.success("Joined game successfully!");
    },
    onError: (error) => {
      toast.error(`An error occurred: ${error.message}`);
    },
  });

  const { register, handleSubmit, formState: { errors } } = useForm<JoinGameValues>({
    resolver: zodResolver(joinGameSchema),
  });

  const onSubmit = (data: JoinGameValues) => {
    mutation.mutate({ name: data.name });
  };

  return (
    <div className="flex items-center justify-center h-screen">
      <form onSubmit={handleSubmit(onSubmit)} className="w-full max-w-md p-8 space-y-8 bg-white rounded-lg shadow-md">
        <h1 className="text-2xl font-bold text-center">Join a Game</h1>
        <div className="space-y-4">
          <div>
            <Label htmlFor="name">Game Name</Label>
            <Input id="name" placeholder="Enter game name" {...register("name")} />
            {errors.name && <p className="text-red-500">{errors.name.message}</p>}
          </div>
          <Button type="submit" className="w-full" disabled={mutation.isPending}>
            {mutation.isPending ? "Joining..." : "Join Game"}
          </Button>
        </div>
      </form>
    </div>
  );
}
