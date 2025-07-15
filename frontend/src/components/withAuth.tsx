"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { useAuth } from "@/context/AuthContext";

const withAuth = <P extends object>(WrappedComponent: React.ComponentType<P>) => {
  const Auth = (props: P) => {
    const { user, isReady } = useAuth();
    const router = useRouter();

    useEffect(() => {
      if (isReady && !user) {
        router.replace("/"); // Redirect to login page if not authenticated
      }
    }, [user, isReady, router]);

    if (!user) {
      return null; // Or a loading spinner
    }

    return <WrappedComponent {...props} />;
  };

  return Auth;
};

export default withAuth;
