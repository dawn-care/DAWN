import { Routes, Route } from "react-router-dom";
import Home from "./pages/Home";
import MediaForm from "./pages/MediaForm";
import Header from "./components/Header";
import AgendaView from "./pages/AgendaView";
import Library from "./pages/Library";
import PromptForm from "./pages/PromptForm"
import CalendarView from "./pages/CalendarView";
import WidgetForm from "./pages/WidgetForm";

function App() {
  return (
    <>
      <Header/>

      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/library" element={<Library />} />
        
        <Route path="/agenda" element={<CalendarView />} />
        <Route path="/agenda/:date" element={<AgendaView />} />

        <Route path="/media" element={<MediaForm />} />
        <Route path="/media/:mediaId" element={<MediaForm />} />

        <Route path="/prompt" element={<PromptForm />} />
        <Route path="/prompt/:promptId" element={<PromptForm />} />

        <Route path="/widget" element={<WidgetForm />} />
        <Route path="/widget/:moduleId" element={<WidgetForm />} />
      </Routes>
    </>
  );
}

export default App;
