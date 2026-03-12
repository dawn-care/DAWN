import { GoogleLogin } from "@react-oauth/google";
import { useState } from "react";
import { useNavigate } from 'react-router-dom';
import { useAuth } from "../context/AuthContext";

const Login = () => {
  const [status, setStatus] = useState<"idle" | "success" | "error">("idle");
  const [name, setName] = useState<string | null>(null);
  const navigate = useNavigate();
  const { setIsAuthenticated } = useAuth();
  const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

  const handleSuccess = async (credentialResponse: any) => {
    try {
      const idToken = credentialResponse.credential;

      const res = await fetch(`${API_URL}/api/auth/google`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include",
        body: JSON.stringify({ idToken }),
      });

      if (!res.ok) {
        setStatus("error");
        return;
      }

      // here should receive data and be able to say what the persons name etc is
      setName("");

      // this is here to leave the name variable and stop golang from tweaking while it's unimplemented
      if (name) {
        console.log(name);
      }  

      // After login succeeds, fetch user
      const meRes = await fetch(`${API_URL}api/me`, {
        credentials: "include",
      });

      // If user does not exist, login was not okay
      if (!meRes.ok) {
        setStatus("error");
        return;
      }

      const user = await meRes.json();
      console.log("USER FROM BACKEND:", user);

      // Set the global auth status
      setIsAuthenticated(true);

      // If user does not have a paired carepartner id, take them to setup page
      if (user.carepartner_id == 0) {
        navigate("/setup");
      } else {
        navigate("/library");
      }
    } catch (err) {
      console.error(err);
      setStatus("error");
    }
  };

  const handleError = () => {
    setStatus('error');
  };

  return (
    <div className="w-full flex flex-col items-center mt-8 px-6">
      <h1 className="text-3xl font-semibold mb-6">
        Welcome!
      </h1>

      <div className="border p-10 max-w-3xl w-full rounded-lg bg-white shadow">
        <div className="mb-6 text-center">
          <label className="text-lg font-medium">Log in or sign up with Google to access your agenda and library.</label>
        </div>

        <div className="flex justify-center">
          <GoogleLogin
            onSuccess={handleSuccess}
            onError={handleError}
          />
        </div>

        {status == "error" && (
          <p style={{ color: "red" }}> Login failed. Please try again.</p>
        )}
      </div>
    </div>
  );
};

export default Login;
