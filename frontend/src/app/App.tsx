import React from 'react';
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
  useLocation,
  Location
} from 'react-router-dom';
import { ControlPlaneLanding } from '../components/ControlPlaneLanding';
import { EditConfig } from '../components/EditConfig';
import Login from '../pages/auth/Login';
import Register from '../pages/auth/Signup';
import { ROUTES } from '../constants/routes';

interface LocationState {
  from: Location;
}

const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const token = localStorage.getItem('authToken');
  const location = useLocation();

  if (!token) {
    return (
      <Navigate
        to={ROUTES.LOGIN}
        state={{ from: location }}
        replace
      />
    );
  }

  return <>{children}</>;
};

const PublicRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const token = localStorage.getItem('authToken');
  const location = useLocation();
  const from = (location.state as LocationState)?.from?.pathname || ROUTES.HOME;

  if (token) {
    // Redirect to the attempted protected route or default to members
    return <Navigate to={from} replace />;
  }

  return <>{children}</>;
};

function App() {
  return (
    <Router>
      <Routes>
        {/* Public Routes */}
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
        {/* Protected Routes */}
        <Route
          path={ROUTES.HOME}
          element={
        <ProtectedRoute>
          <ControlPlaneLanding />
        </ProtectedRoute>
          }
        />
        <Route
          path={`${ROUTES.CONFIG}/:agentId`}
          element={
        <ProtectedRoute>
          <EditConfig />
        </ProtectedRoute>
          }
        />
        {/* Redirects */}
        <Route
          path="/"
          element={
        localStorage.getItem('authToken') ? (
          <Navigate to={ROUTES.HOME} replace />
        ) : (
          <Navigate to={ROUTES.LOGIN} replace />
        )
          }
        />
        <Route
          path="*"
          element={<Navigate to={ROUTES.LOGIN} replace />}
        />
      </Routes>
    </Router>
  );
}

export default App;