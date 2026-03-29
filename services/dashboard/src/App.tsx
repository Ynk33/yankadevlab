import { BrowserRouter, Navigate, Route, Routes } from "react-router";
import { AuthProvider } from "@/contexts/auth";
import { useAuth } from "@/hooks/use-auth";
import LoginPage from "@/pages/login";
import HomePage from "@/pages/home";
import type { ReactNode } from "react";

function RequireAuth({ children }: { children: ReactNode }) {
  const { isAuthenticated } = useAuth();
  if (!isAuthenticated) return <Navigate to="/login" replace />;
  return children;
}

export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route
            path="/"
            element={
              <RequireAuth>
                <HomePage />
              </RequireAuth>
            }
          />
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
