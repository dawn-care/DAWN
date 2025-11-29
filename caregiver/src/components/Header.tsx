import { Link } from "react-router-dom";
import logo from "../assets/startup-logo.png"; // replace with your logo file

export default function Header() {
  return (
    <header className="w-full bg-[#d8f7f7] border-b border-gray-300 px-8 py-4 flex items-center justify-between">
      <div className="flex items-center gap-3">
        <img src={logo} alt="StartUp Logo" className="h-10 w-auto" />
      </div>

      <nav className="flex gap-10 text-xl font-medium text-gray-800">
        <Link to="/agenda" className="hover:underline">
          Agenda
        </Link>
        <Link to="/library" className="hover:underline">
          Library
        </Link>
        <Link to="/settings" className="hover:underline">
          Settings
        </Link>
      </nav>
    </header>
  );
}
