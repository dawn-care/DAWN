import { GoogleLogin } from "@react-oauth/google";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

const Login = () => {
  const [status, setStatus] = useState<"idle" | "success" | "error">("idle");
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const { setUser, setIsAuthenticated } = useAuth();
  const API_URL = import.meta.env.VITE_API_URL + "/api";

  const handleSuccess = async (credentialResponse: any) => {
    setStatus("idle");
    setLoading(true);

    try {
      const idToken = credentialResponse.credential;

      // Login with Google 
      const loginRes = await fetch(`${API_URL}/api/auth/google`, {
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

      // Fetch logged-in user 
      const meRes = await fetch(`${API_URL}/api/me`, {
        credentials: "include",
      });

      if (!meRes.ok) {
        setStatus("error");
        setLoading(false);
        return;
      }

      const user = await meRes.json();

      // Update AuthContext so user info can be accessed globally
      setUser(user);
      setIsAuthenticated(true);

      // Redirect to pairing if carepartner id has not been setup, or library if it has
      if (!user.carepartner_id || user.carepartner_id === 0) {
        navigate("/setup");
      } else {
        navigate("/library");
      }

      setStatus("success");
    } catch (err) {
      console.error(err);
      setStatus("error");
    } finally {
      setLoading(false);
    }
  };

  const handleError = () => {
    setStatus("error");
  };

  return (
    <div className="w-full flex flex-col items-center mt-8 px-6">
      <h1 className="text-3xl font-semibold mb-6">Welcome!</h1>

      <div className="border p-10 max-w-3xl w-full rounded-lg bg-white shadow">
        <div className="mb-6 text-center">
          <label className="text-lg font-medium">
            Log in or sign up with Google to access your agenda and library.
          </label>
        </div>

        <div className="flex justify-center">
          <GoogleLogin onSuccess={handleSuccess} onError={handleError} />
        </div>

        {status === "error" && (
          <p style={{ color: "red", marginTop: 12 }}>
            Login failed. Please try again.
          </p>
        )}

        {loading && (
          <p style={{ marginTop: 12 }}>Logging in… please wait.</p>
        )}
      </div>
    </div>
  );
};

export default Login;