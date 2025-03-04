"use client";

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { login, isLoggedIn } from '../utils/auth';
import { NextPage } from 'next';

const Login: NextPage = () => {
  const [username, setUsername] = useState<string>('');
  const [password, setPassword] = useState<string>('');
  const [error, setError] = useState<string>('');
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const router = useRouter();

  useEffect(() => {
    if (isLoggedIn()) {
      router.push('/dashboard');
    }
  }, [router]);

  const handleSubmit = (e: React.FormEvent): void => {
    e.preventDefault();
    setIsLoading(true);
    setError('');
    
    // Simulating network latency
    setTimeout(() => {
      if (login(username, password)) {
        router.push('/dashboard');
      } else {
        setError('Invalid credentials. Use admin/admin to login.');
        setIsLoading(false);
      }
    }, 500);
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-background py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h1 className="mt-6 text-center text-3xl font-extrabold">
            MinorV3 Platform
          </h1>
          <p className="mt-2 text-center text-sm text-muted-foreground">
            Sign in to your account
          </p>
        </div>
        <div className="card">
          <form className="space-y-6" onSubmit={handleSubmit}>
            <div className="space-y-4">
              <div>
                <label htmlFor="username" className="block text-sm font-medium mb-1">
                  Username
                </label>
                <input
                  id="username"
                  name="username"
                  type="text"
                  required
                  className="input"
                  placeholder="Username"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                />
              </div>
              <div>
                <label htmlFor="password" className="block text-sm font-medium mb-1">
                  Password
                </label>
                <input
                  id="password"
                  name="password"
                  type="password"
                  required
                  className="input"
                  placeholder="Password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                />
              </div>
            </div>

            {error && (
              <div className="py-2 px-3 bg-destructive/10 border border-destructive/20 text-destructive rounded-md text-sm">
                {error}
              </div>
            )}

            <div className="pt-2">
              <button
                type="submit"
                className="btn btn-primary w-full"
                disabled={isLoading}
              >
                {isLoading ? 'Signing in...' : 'Sign in'}
              </button>
            </div>
            
            <div className="text-center text-xs text-muted-foreground">
              <p>Demo credentials: admin/admin</p>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};

export default Login;
