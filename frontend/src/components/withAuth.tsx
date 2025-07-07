"use client";

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/context/AuthContext';

const withAuth = (WrappedComponent: React.ComponentType) => {
  const Auth = (props: any) => {
    const { user } = useAuth();
    const router = useRouter();

    useEffect(() => {
      if (!user) {
        router.replace('/'); // Redirect to login page if not authenticated
      }
    }, [user, router]);

    if (!user) {
      return null; // Or a loading spinner
    }

    return <WrappedComponent {...props} />;
  };

  return Auth;
};

export default withAuth;
