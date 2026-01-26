import { Routes, Route } from "react-router-dom";
import Home from "./pages/Home";
import UploadMedia from "./pages/UploadMedia";
import Header from "./components/Header";
import AgendaView from "./pages/AgendaView";
import Library from "./pages/Library";

function App() {
  return (
    <>
      <Header/>

      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/upload" element={<UploadMedia />} />
        <Route path="/agenda/:date" element={<AgendaView />} />
        <Route path="/library" element={<Library />} />
      </Routes>
    </>
  );
}

export default App;
