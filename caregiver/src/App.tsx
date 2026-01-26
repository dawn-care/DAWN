import { Routes, Route } from "react-router-dom";
import Home from "./pages/Home";
import UploadMedia from "./pages/UploadMedia";
import Header from "./components/Header";
import AgendaView from "./pages/AgendaView";

function App() {
  return (
    <>
      <Header/>

      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/upload" element={<UploadMedia />} />
        <Route path="/agenda/:date" element={<AgendaView />} />
      </Routes>
    </>
  );
}

export default App;
