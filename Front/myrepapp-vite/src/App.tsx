import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import HomePage from '@/pages/HomePage';
import DownloadPage from '@/pages/DownloadPage';
import ZsWorkExperience from '@/pages/ZsWorkExperience';
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
      </Routes>
    </Router>
  );
}

export default App;
