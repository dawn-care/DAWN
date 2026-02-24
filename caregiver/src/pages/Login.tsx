import { GoogleLogin } from "@react-oauth/google";
import { useEffect, useState } from "react";
import { useNavigate } from 'react-router-dom';
import { useAuth } from "../context/AuthContext";

const Login = () => {
  const [status, setStatus] = useState<"idle" | "success" | "error">("idle");
  const [name, setName] = useState<string | null>(null);
  const navigate = useNavigate();
  const { setIsAuthenticated } = useAuth();

  const handleSuccess = async (credentialResponse: any) => {
    try {
      const idToken = credentialResponse.credential;

      const res = await fetch("http://localhost:8080/api/auth/google", {
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
      // Set the global auth status to true
      setIsAuthenticated(true);
      setStatus("success");
    } catch (err) {
      console.error(err);
      setStatus("error");
    }
  };

  // Navigate after the state is successfully set
  useEffect(() => {
    if (status === 'success') {
      navigate('/library');
    }
  }, [status, navigate]); // Automatically trigger when status changes

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
