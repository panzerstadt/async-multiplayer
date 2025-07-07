import axios from 'axios';

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
});

export const createGame = async (name: string, players: string[]) => {
  const response = await api.post('/create-game', { name, players });
  return response.data;
};

export const joinGame = async (name: string) => {
  const response = await api.post(`/join-game/${name}`);
  return response.data;
};

export const getGame = async (id: string) => {
  const response = await api.get(`/games/${id}`);
  return response.data;
};

export const uploadSave = async (id: string, file: File) => {
  const formData = new FormData();
  formData.append('file', file);
  const response = await api.post(`/games/${id}/saves`, formData, {
    headers: {
      'Content-Type': 'multipart/form-data',
    },
  });
  return response.data;
};

export const getLatestSave = async (id: string) => {
  const response = await api.get(`/games/${id}/saves/latest`, {
    responseType: "blob",
  });
  return response.data;
};

export const getGames = async () => {
  const response = await api.get("/games");
  return response.data;
};
