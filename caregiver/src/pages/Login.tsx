import { GoogleLogin } from "@react-oauth/google";
import { useState } from "react";

const Login = () => {
  const [status, setStatus] = useState<"idle" | "success" | "error">("idle");
  const [name, setName] = useState<string | null>(null);

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
      setStatus("success");
    } catch (err) {
      console.error(err);
      setStatus("error");
    }
  };

  const handleError = () => {
    setStatus("error");
  };

  return (
    <div>
      <h1>Login</h1>

      {status === "idle" && (
        <GoogleLogin onSuccess={handleSuccess} onError={handleError} />
      )}

      {status === "success" && (
        <p style={{ color: "green" }}>Login Success! Welcome!</p>
      )}

      {status === "error" && (
        <p style={{ color: "red" }}> Login Failed. Please try again.</p>
      )}
    </div>
  );
};

export default Login;
