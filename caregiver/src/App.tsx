import { Routes, Route } from "react-router-dom";
import Home from "./pages/Home";
import Calendar from "./pages/Calendar";
import Media from "./pages/Media";
import Family from "./pages/Family";
import Preview from "./pages/Preview";
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
        {/* Future implementation: reachable by URL only, not linked in the nav bar */}
        <Route path="/family" element={<Family />} />
      </Route>

      {/* Full screen, no nav bar. Future implementation: the button that opens it is added in a later ticket */}
      <Route
        path="/preview"
        element={
          <RequireCarePartner>
            <Preview />
          </RequireCarePartner>
        }
      />

      <Route path="/login" element={<Login />} />
      <Route path="/settings" element={<Settings />} />
      <Route path="/setup" element={<AccountSetupPage />} />
      <Route path="/select-carepartner" element={<SelectCarepartner />} />
    </Routes>
  );
}

export default App;
