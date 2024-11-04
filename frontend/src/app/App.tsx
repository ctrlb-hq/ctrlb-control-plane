import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { EditConfig } from '../components/EditConfig';
import { MembersTable } from '../components/Table';
import Login from '../pages/auth/Login';
import Register from '../pages/auth/Signup';

function getSessionIdCookie(): string | null {
  const cookieName = 'session_id';
  const cookies = document.cookie.split('; ');

  for (const cookie of cookies) {
    const [name, value] = cookie.split('=');
    if (name === cookieName) {
      return decodeURIComponent(value);
    }
  }

  return null;
}

function App() {
  const isAuthenticated = Boolean(getSessionIdCookie());

  return (
    <Router>
      <Routes>
        <Route
          path="/"
          element={isAuthenticated ? <MembersTable /> : <Navigate to="/login" replace />}
        />
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route
          path="/config/:agentId"
          element={isAuthenticated ? <EditConfig /> : <Navigate to="/login" replace />}
        />
      </Routes>
    </Router>
  );
}

export default App;
