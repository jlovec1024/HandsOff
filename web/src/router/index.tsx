import { createBrowserRouter, Navigate } from "react-router-dom";
import { useAuthStore } from "../stores/auth";
import Login from "../pages/Login";
import Dashboard from "../pages/Dashboard";
import Settings from "../pages/Settings";
import RepositoryList from "../pages/Repository/List";
import { ReviewList, ReviewDetail } from "../pages/Review";
import MainLayout from "../components/Layout/MainLayout";

// Protected route wrapper
const ProtectedRoute = ({ children }: { children: React.ReactNode }) => {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated());

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
};

export const router = createBrowserRouter([
  {
    path: "/login",
    element: <Login />,
  },
  {
    path: "/",
    element: (
      <ProtectedRoute>
        <MainLayout />
      </ProtectedRoute>
    ),
    children: [
      {
        index: true,
        element: <Dashboard />,
      },
      {
        path: "settings",
        element: <Settings />,
      },
      {
        path: "repositories",
        element: <RepositoryList />,
      },
      {
        path: "reviews",
        element: <ReviewList />,
      },
      {
        path: "reviews/:id",
        element: <ReviewDetail />,
      },
    ],
  },
]);
