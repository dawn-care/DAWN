import { Routes, Route } from "react-router-dom";
import Home from "./pages/Home";
import UploadMedia from "./pages/UploadMedia";
import Header from "./components/Header";
import CreateWidget from "./pages/CreateWidget";

function App() {
  return (
    <>
      <Header/>

      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/upload" element={<UploadMedia />} />
        <Route path="/widget" element={<CreateWidget />} />
      </Routes>
    </>
  );
}

export default App;
