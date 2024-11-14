import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { MembersTable } from '../components/Table';
import { EditConfig } from '../components/EditConfig';
import Login from '../pages/auth/Login';
import Register from '../pages/auth/Signup';
import { ROUTES } from '../constants/routes';

const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const token = localStorage.getItem('authToken');
  const location = useLocation();

  if (!token) {
    return <Navigate to={ROUTES.LOGIN} state={{ from: location }} replace />;
  }

  return <>{children}</>;
};

const PublicRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const token = localStorage.getItem('authToken');
  const location = useLocation();

  if (token) {
    return <Navigate to={ROUTES.MEMBERS} state={{ from: location }} replace />;
  }

  return <>{children}</>;
};

function App() {
  return (
    <Router>
      <Routes>
        <Route
          path={ROUTES.LOGIN}
          element={
            <PublicRoute>
              <Login />
            </PublicRoute>
          }
        />
        <Route
          path={ROUTES.REGISTER}
          element={
            <PublicRoute>
              <Register />
            </PublicRoute>
          }
        />
        <Route
          path={ROUTES.MEMBERS}
          element={
            <ProtectedRoute>
              <MembersTable />
            </ProtectedRoute>
          }
        />
        <Route
          path="/"
          element={<Navigate to={ROUTES.MEMBERS} replace />}
        />
        <Route 
          path={`${ROUTES.CONFIG}/:agentId`}
          element={
            <ProtectedRoute>
              <EditConfig />
            </ProtectedRoute>
          }
        />
        <Route path="*" element={<Navigate to={ROUTES.LOGIN} replace />} />
      </Routes>
    </Router>
  );
}

export default App;