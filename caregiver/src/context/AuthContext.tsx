import React, { createContext, useState, useContext, useEffect, type JSX } from "react";

// Define the user type
export interface User {
  id: number;    // Caregiver ID
  name?: string;
  role?: string;

  carepartners?: {
    id: number;
    name: string;
  }[];
}

// Context type includes isAuthenticated, user, and their setters
interface AuthContextType {
  isAuthenticated: boolean;
  setIsAuthenticated: React.Dispatch<React.SetStateAction<boolean>>;
  user: User | null;
  activeCarePartnerId: number | null;
  setUser: React.Dispatch<React.SetStateAction<User | null>>;
  setActiveCarePartnerId: React.Dispatch<React.SetStateAction<number | null>>;
}

// Create context
const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Hook for components
// eslint-disable-next-line react-refresh/only-export-components
export const useAuth = (): AuthContextType => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};

// Provider component
export const AuthProvider = ({ children }: { children: JSX.Element | JSX.Element[] }) => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [user, setUser] = useState<User | null>(null);

  const [activeCarePartnerId, setActiveCarePartnerId] = useState<number | null>(() => {
    try {
      const saved = localStorage.getItem("activeCarePartnerId");
      return saved ? Number(saved) : null;
    } catch {
      // Storage can be blocked; RequireLogin then picks the first carepartner
      return null;
    }
  });

  useEffect(() => {
    try {
      if (activeCarePartnerId !== null) {
        localStorage.setItem("activeCarePartnerId", activeCarePartnerId.toString());
      } else {
        localStorage.removeItem("activeCarePartnerId");
      }
    } catch {
      // Storage can be blocked; the choice then lasts until the page closes
    }
  }, [activeCarePartnerId]);

  return (
    <AuthContext.Provider value={{ isAuthenticated, setIsAuthenticated, user, setUser, activeCarePartnerId, setActiveCarePartnerId }}>
      {children}
    </AuthContext.Provider>
  );
};