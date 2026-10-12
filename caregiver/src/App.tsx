import { Routes, Route } from "react-router-dom";
import Home from "./pages/Home";
import Calendar from "./pages/Calendar";
import Media from "./pages/Media";
import Login from "./pages/Login";
import Settings from "./pages/Settings";
import AccountSetupPage from "./pages/AccountSetup";
import SelectCarepartner from "./pages/SelectCarepartner";
import RequireCarePartner from "./components/RequireCarepartner";
import AppShell from "./components/AppShell";

function App() {
  return (
    <Routes>
      {/* Pages with the left nav bar */}
      <Route
        element={
          <RequireCarePartner>
            <AppShell />
          </RequireCarePartner>
        }
      >
        <Route path="/" element={<Home />} />
        <Route path="/calendar" element={<Calendar />} />
        <Route path="/media" element={<Media />} />
      </Route>

      <Route path="/login" element={<Login />} />
      <Route path="/settings" element={<Settings />} />
      <Route path="/setup" element={<AccountSetupPage />} />
      <Route path="/select-carepartner" element={<SelectCarepartner />} />
    </Routes>
  );
}

export default App;
