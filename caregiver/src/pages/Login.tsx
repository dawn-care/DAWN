/**
 * Login.tsx
 * 
 * User: Caregiver
 * 
 * This component handles user authentication via Google OAuth.
 * 
 * Responsibilities:
 *  - Initiate Google login flow
 *  - Send ID token to backend for verification
 *  - Fetch authenticated user data
 *  - Update global AuthContext
 *  - Redirect user based on account state
 * 
 * Redirect logic:
 *  - No carepartners → Account setup page
 *  - Has carepartners → Home page (`/`)
 */

import { GoogleLogin } from "@react-oauth/google";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

/* ============================
   Component
============================ */

/**
 * Login Component
 * 
 * Handles:
 *  - Google authentication
 *  - Backend session creation
 *  - User state initialization
 */
const Login = () => {
  
  /* ============================
     Routing / Context
  ============================ */

  const navigate = useNavigate();
  const { setUser, setIsAuthenticated, setActiveCarePartnerId } = useAuth();

  /**
   * Base API URL
   */
  const API_URL = import.meta.env.VITE_API_URL;

  /* ============================
     State
  ============================ */

  /**
   * Status of login process
   */
  const [status, setStatus] = useState<"idle" | "success" | "error">("idle");

  /**
   * Loading indicator for login request
   */
  const [loading, setLoading] = useState(false);

  /* ============================
     Handlers
  ============================ */

  /**
   * Handles successful Google authentication
   * 
   * Flow:
   *  1. Send Google ID token to backend
   *  2. Backend verifies and creates session
   *  3. Fetch authenticated user data
   *  4. Store user in global context
   *  5. Redirect based on carepartner state
   */
  const handleSuccess = async (credentialResponse: any) => {
    setStatus("idle");
    setLoading(true);

    try {
      const idToken = credentialResponse.credential;

      /* ============================
         Step 1: Backend Login
      ============================ */
      const loginRes = await fetch(`${API_URL}/auth/google`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ idToken }),
      });

      if (!loginRes.ok) {
        setStatus("error");
        setLoading(false);
        return;
      }

      /* ============================
         Step 2: Fetch User
      ============================ */
      const meRes = await fetch(`${API_URL}/me`, {
        credentials: "include",
      });

      if (!meRes.ok) {
        setStatus("error");
        setLoading(false);
        return;
      }

      const user = await meRes.json();

      /* ============================
         Step 3: Update Context
      ============================ */
      setUser(user);
      setIsAuthenticated(true);

      /* ============================
         Step 4: Redirect Logic
      ============================ */
      if (!user.carepartners || user.carepartners.length === 0) {
        navigate("/setup");
      } else {
        setActiveCarePartnerId(user.carepartners[0].id);
        navigate("/");
      }

      setStatus("success");
    } catch (err) {
      console.error(err);
      setStatus("error");
    } finally {
      setLoading(false);
    }
  };

  /**
   * Handles Google login error
   */
  const handleError = () => {
    setStatus("error");
  };

  /* ============================
     Render
  ============================ */
  return (
    <div className="w-full flex flex-col items-center mt-8 px-6">
      <h1 className="text-3xl font-semibold mb-6">Welcome!</h1>

      <div className="border p-10 max-w-3xl w-full rounded-lg bg-white shadow">

        {/* Description */}
        <div className="mb-6 text-center">
          <label className="text-lg font-medium">
            Log in or sign up with Google to access your agenda and library.
          </label>
        </div>

        {/* Google Login Button */}
        <div className="flex justify-center">
          <GoogleLogin onSuccess={handleSuccess} onError={handleError} />
        </div>

        {/* Error Message */}
        {status === "error" && (
          <p style={{ color: "red", marginTop: 12 }}>
            Login failed. Please try again.
          </p>
        )}

        {/* Loading Indicator */}
        {loading && (
          <p style={{ marginTop: 12 }}>Logging in… please wait.</p>
        )}
      </div>
    </div>
  );
};

export default Login;