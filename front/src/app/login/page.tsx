'use client';

import { LoginForm } from '@/components/auth/LoginForm';
import { useEffect } from 'react';

export default function LoginPage() {
  useEffect(() => {
    console.log('API URL from env:', process.env.NEXT_PUBLIC_API_URL);
  }, []);

  return <LoginForm />;
}
