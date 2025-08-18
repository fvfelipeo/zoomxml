'use client';

import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { User, AuthTokens } from '@/types/api';

interface AuthContextType {
  user: User | null;
  adminToken: string | null;
  userToken: string | null;
  isAuthenticated: boolean;
  isAdmin: boolean;
  isLoading: boolean;
  login: (tokens: AuthTokens, userData?: User) => void;
  logout: () => void;
  setUser: (user: User | null) => void;
  setAdminToken: (token: string | null) => void;
  setUserToken: (token: string | null) => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [adminToken, setAdminToken] = useState<string | null>(null);
  const [userToken, setUserToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Carregar dados do localStorage na inicialização
  useEffect(() => {
    const storedUser = localStorage.getItem('zoomxml_user');
    const storedAdminToken = localStorage.getItem('zoomxml_admin_token');
    const storedUserToken = localStorage.getItem('zoomxml_user_token');

    if (storedUser) {
      try {
        setUser(JSON.parse(storedUser));
      } catch (error) {
        console.error('Erro ao carregar dados do usuário:', error);
        localStorage.removeItem('zoomxml_user');
      }
    }

    if (storedAdminToken) {
      setAdminToken(storedAdminToken);
    }

    if (storedUserToken) {
      setUserToken(storedUserToken);
    }

    // Marcar como carregado após verificar o localStorage
    setIsLoading(false);
  }, []);

  // Salvar dados no localStorage quando mudarem
  useEffect(() => {
    if (user) {
      localStorage.setItem('zoomxml_user', JSON.stringify(user));
    } else {
      localStorage.removeItem('zoomxml_user');
    }
  }, [user]);

  useEffect(() => {
    if (adminToken) {
      localStorage.setItem('zoomxml_admin_token', adminToken);
    } else {
      localStorage.removeItem('zoomxml_admin_token');
    }
  }, [adminToken]);

  useEffect(() => {
    if (userToken) {
      localStorage.setItem('zoomxml_user_token', userToken);
    } else {
      localStorage.removeItem('zoomxml_user_token');
    }
  }, [userToken]);

  const login = (tokens: AuthTokens, userData?: User) => {
    if (tokens.adminToken) {
      setAdminToken(tokens.adminToken);
    }
    if (tokens.userToken) {
      setUserToken(tokens.userToken);
    }
    if (userData) {
      setUser(userData);
    }
  };

  const logout = () => {
    setUser(null);
    setAdminToken(null);
    setUserToken(null);
    localStorage.removeItem('zoomxml_user');
    localStorage.removeItem('zoomxml_admin_token');
    localStorage.removeItem('zoomxml_user_token');
  };

  const isAuthenticated = !!(adminToken || userToken);
  const isAdmin = user?.role === 'admin' || !!adminToken;

  const value: AuthContextType = {
    user,
    adminToken,
    userToken,
    isAuthenticated,
    isAdmin,
    isLoading,
    login,
    logout,
    setUser,
    setAdminToken,
    setUserToken,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth deve ser usado dentro de um AuthProvider');
  }
  return context;
}
