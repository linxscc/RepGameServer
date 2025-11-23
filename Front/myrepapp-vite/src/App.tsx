import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import HomePage from '@/pages/HomePage';
import DownloadPage from '@/pages/DownloadPage';
import ZsWorkExperience from '@/pages/ZsWorkExperience';
import AnyProductsDocs from '@/pages/AnyProductsDocs';
import Navigation from '@/components/Navigation';
import './App.css';

function App() {
  return (
    <Router>
      <Navigation />
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/download" element={<DownloadPage />} />
        <Route path="/zsworkexperience" element={<ZsWorkExperience />} />
        <Route path="/anyproductsdocs" element={<AnyProductsDocs />} />
      </Routes>
    </Router>
  );
}

export default App;
