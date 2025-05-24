'use client';

import { Suspense } from 'react';
import RegisterForm from '@/components/auth/RegisterForm';
import { Loader } from 'lucide-react';

export default function RegisterPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen flex items-center justify-center bg-gray-50">
        <Loader className="h-8 w-8 animate-spin text-indigo-600" />
      </div>
    }>
      <RegisterForm />
    </Suspense>
  );
}