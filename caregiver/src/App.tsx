import { Routes, Route } from "react-router-dom";
import Home from "./pages/Home";
import Login from "./pages/Login";
import Settings from "./pages/Settings";
import AccountSetupPage from "./pages/AccountSetup";
import SelectCarepartner from "./pages/SelectCarepartner";
import RequireCarePartner from "./components/RequireCarepartner";

function App() {
  return (
    <Routes>
      <Route
        path="/"
        element={
          <RequireCarePartner>
            <Home />
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
