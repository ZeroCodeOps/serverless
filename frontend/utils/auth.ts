"use client";
import { useEffect } from 'react';

// Simple auth utility functions
export function login(username: string, password: string): boolean {
  // In a real app, you would make an API call here
  if (username === 'admin' && password === 'admin') {
    localStorage.setItem('isLoggedIn', 'true');
    localStorage.setItem('username', username);
    return true;
  }
  return false;
}

export function logout(): void {
  localStorage.removeItem('isLoggedIn');
  localStorage.removeItem('username');
}

export function getUsername(): string | null {
  return typeof window !== 'undefined' ? localStorage.getItem('username') : null;
}

export function isLoggedIn(): boolean {
  return typeof window !== 'undefined' ? localStorage.getItem('isLoggedIn') === 'true' : false;
}

// Auth protection hook for pages
export function useAuth(): boolean {  
  useEffect(() => {
    if (!isLoggedIn() && window.location.href !== '/') {
      window.location.href = "/"
    }
  }, [window.location.href]);

  return isLoggedIn();
}