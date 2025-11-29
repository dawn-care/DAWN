import { Routes, Route, Link } from "react-router-dom";
import Home from "./pages/Home";
import UploadMedia from "./pages/UploadMedia";
import Header from "./components/Header";

function App() {
  return (
    <>
      <Header/>

      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/upload" element={<UploadMedia />} />
      </Routes>
    </>
  );
}

export default App;
