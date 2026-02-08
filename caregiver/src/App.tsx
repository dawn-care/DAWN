import { Routes, Route } from "react-router-dom";
import Home from "./pages/Home";
import UploadMedia from "./pages/UploadMedia";
import Header from "./components/Header";
import AgendaView from "./pages/AgendaView";
import Library from "./pages/Library";
import CreatePrompt from "./pages/CreatePrompt"
import CreateWidget from "./pages/CreateWidget";
import CalendarView from "./pages/CalendarView";
import WidgetForm from "./pages/WidgetForm";

function App() {
  return (
    <>
      <Header/>

      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/upload" element={<UploadMedia />} />
        <Route path="/agenda" element={<CalendarView />} />
        <Route path="/agenda/:date" element={<AgendaView />} />
        <Route path="/library" element={<Library />} />
        <Route path="/CreatePrompt" element={<CreatePrompt />} />
        <Route path="/widget" element={<WidgetForm />} />
        <Route path="/widget/:moduleId" element={<WidgetForm />} />
      </Routes>
    </>
  );
}

export default App;
