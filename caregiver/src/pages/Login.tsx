import { GoogleLogin } from '@react-oauth/google';

const Login = () => {
  const handleSuccess = async (credentialResponse: any) => {
    const idToken = credentialResponse.credential;

    // Send token to backend
    const res = await fetch("http://localhost:8080/api/auth/google", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({ idToken }),
    });

    if (!res.ok) {
      console.error("Login failed");
      return;
    }
    // LOGIN SUCCESS !! :)
    // here is where it will redirect to some page other than the login one
  };

  const handleError = () => {
    console.error("Google login failed");
  };

  return (
    <div>
      <h1>Login</h1>

      <GoogleLogin onSuccess={handleSuccess} onError={handleError} />
    </div>
  );
};

export default Login;
