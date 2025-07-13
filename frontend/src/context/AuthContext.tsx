"use client";

import { createContext, useContext, useState, useEffect, ReactNode, useCallback } from "react";
import { useRouter } from "next/navigation";

interface User {
  id: string;
  email: string;
  token: string;
}

interface AuthContextType {
  user: User | null;
  login: (userData: User) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const router = useRouter();

  const login = useCallback(
    (userData: User) => {
      setUser(userData);
      localStorage.setItem("user", JSON.stringify(userData));
      router.push("/dashboard");
    },
    [router]
  );

  const logout = useCallback(() => {
    setUser(null);
    localStorage.removeItem("user");
    router.push("/");
  }, [router]);

  useEffect(() => {
    const storedUser = localStorage.getItem("user");
    if (storedUser) {
      setUser(JSON.parse(storedUser));
    } else {
      const urlParams = new URLSearchParams(window.location.search);
      const token = urlParams.get("token");
      if (token) {
        try {
          const decodedToken = JSON.parse(atob(token.split(".")[1]));
          const userData = { id: decodedToken.sub, email: decodedToken.email || "", token };
          login(userData);
          window.history.replaceState({}, document.title, window.location.pathname);
        } catch (e) {
          console.error("Failed to decode token or parse user data:", e);
        }
      }
    }
  }, [login]);

  return <AuthContext.Provider value={{ user, login, logout }}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
