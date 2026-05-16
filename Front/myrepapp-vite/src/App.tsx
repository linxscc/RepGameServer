import { BrowserRouter as Router, Routes, Route, useLocation } from 'react-router-dom';
import HomePage from '@/pages/HomePage';
import DownloadPage from '@/pages/DownloadPage';
import ZsWorkExperience from '@/pages/ZsWorkExperience';
import AnyProductsDocs from '@/pages/AnyProductsDocs';
import VoyaraApp from '@/Voyara/VoyaraApp';
import Navigation from '@/components/Navigation';
import './App.css';

function AppLayout() {
  const location = useLocation();

  return (
    <>
      {!location.pathname.startsWith('/voyara') && <Navigation />}
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/download" element={<DownloadPage />} />
        <Route path="/zsworkexperience" element={<ZsWorkExperience />} />
        <Route path="/anyproductsdocs" element={<AnyProductsDocs />} />
        <Route path="/voyara/*" element={<VoyaraApp />} />
      </Routes>
    </>
  );
}

function App() {
  return (
    <Router>
      <AppLayout />
    </Router>
  );
}

export default App;
