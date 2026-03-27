import { Routes, Route } from "react-router-dom";
import Home from "./pages/Home";
import MediaFormCreate from "./pages/MediaFormCreate";
import MediaFormUpdate from "./pages/MediaFormUpdate";
import MediaFormDelete from "./pages/MediaFormDelete"
import Header from "./components/Header";
import AgendaView from "./pages/AgendaView";
import Library from "./pages/Library";
import PromptForm from "./pages/PromptForm"
import CalendarView from "./pages/CalendarView";
import WidgetForm from "./pages/WidgetForm";
import Login from "./pages/Login"
import Settings from "./pages/Settings";
import AccountSetupPage from "./pages/AccountSetup";
import RequireCarePartner from "./components/RequireCarepartner";

function App() {
  return (
    <>
      <Header/>

      <Routes>
        <Route path="/" element={<Home />} />

        <Route
          path="/library"
          element={
            <RequireCarePartner>
              <Library />
            </RequireCarePartner>
          }
        />

        <Route
          path="/agenda"
          element={
            <RequireCarePartner>
              <CalendarView />
            </RequireCarePartner>
          }
        />

        <Route
          path="/agenda/:date"
          element={
            <RequireCarePartner>
              <AgendaView />
            </RequireCarePartner>
          }
        />

        <Route
          path="/media"
          element={
            <RequireCarePartner>
              <MediaFormCreate />
            </RequireCarePartner>
          }
        />

        <Route
          path="/media/update"
          element={
            <RequireCarePartner>
              <MediaFormUpdate />
            </RequireCarePartner>
          }
        />

        <Route
          path="/media/delete"
          element={
            <RequireCarePartner>
              <MediaFormDelete />
            </RequireCarePartner>
          }
        />

        <Route
          path="/prompt"
          element={
            <RequireCarePartner>
              <PromptForm />
            </RequireCarePartner>
          }
        />

        <Route
          path="/prompt/:moduleId"
          element={
            <RequireCarePartner>
              <PromptForm />
            </RequireCarePartner>
          }
        />

        <Route
          path="/widget"
          element={
            <RequireCarePartner>
              <WidgetForm />
            </RequireCarePartner>
          }
        />

        <Route
          path="/widget/:moduleId"
          element={
            <RequireCarePartner>
              <WidgetForm />
            </RequireCarePartner>
          }
        />

        <Route path="/login" element={<Login />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="/setup" element={<AccountSetupPage />} />
      </Routes>
    </>
  );
}

export default App;
