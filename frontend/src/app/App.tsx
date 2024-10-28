import { EditConfig } from '../components/EditConfig'
import { MembersTable } from '../components/Table';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Login from '../pages/auth/Login';
import Register from '../pages/auth/Signup';

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route path="/register" element={<Register />} />
        <Route path="/" element={<MembersTable />} />
        <Route path="/config/:agentId" element={<EditConfig />} />
      </Routes>
    </Router>
  )
}

export default App