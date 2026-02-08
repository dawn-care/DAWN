import { Routes, Route } from "react-router-dom";
import Home from "./pages/Home";
import UploadMedia from "./pages/UploadMedia";
import Header from "./components/Header";
import AgendaView from "./pages/AgendaView";
import Library from "./pages/Library";
import PromptForm from "./pages/PromptForm"
import CreateWidget from "./pages/CreateWidget";

function App() {
  return (
    <>
      <Header/>

      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/upload" element={<UploadMedia />} />
        <Route path="/agenda/:date" element={<AgendaView />} />
        <Route path="/library" element={<Library />} />
        <Route path="/prompt" element={<PromptForm />} />
        <Route path="/prompt/:moduleId" element={<PromptForm />} />
        <Route path="/widget" element={<CreateWidget />} />
      </Routes>
    </>
  );
}

export default App;
