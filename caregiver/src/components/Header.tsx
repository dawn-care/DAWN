import { Link, useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import logo from "../assets/startup-logo.png";

export default function Header() {
  const { isAuthenticated } = useAuth(); 
  const navigate = useNavigate();

  return (
    <header className="w-full bg-[#d8f7f7] border-b border-gray-300 px-8 py-4 flex items-center justify-between">
      {/* Logo always routes home */}
      <div className="flex items-center gap-3 cursor-pointer" onClick={() => navigate("/")}>
        <img src={logo} alt="StartUp Logo" className="h-10 w-auto" />
      </div>

      <nav className="flex gap-10 text-xl font-medium text-gray-800 items-center">
        {isAuthenticated ? (
          <>
            {/* Only show when logged in */}
            <Link to="/agenda" className="hover:underline">
              Agenda
            </Link>
            <Link to="/library" className="hover:underline">
              Library
            </Link>

            {/* Profile icon */}
            <div
              onClick={() => navigate("/settings")}
              className="h-10 w-10 rounded-full bg-gray-300 cursor-pointer"
            >
              <span className="text-xl flex justify-center items-center h-full w-full">👤</span>
            </div>
          </>
        ) : (
          <>
            <button
              onClick={() => navigate("/login")}
              className="px-4 py-2 bg-black text-white rounded-lg"
            >
              Log In
            </button>
          </>
        )}
      </nav>
    </header>
  );
}