import { BrowserRouter, Route, Routes } from "react-router";
import { DashboardLayout } from "@/components/layout/dashboard-layout";
import HomePage from "@/pages/home";
import NotFoundPage from "@/pages/not-found";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<DashboardLayout />}>
          <Route index element={<HomePage />} />
        </Route>
        <Route path="*" element={<NotFoundPage />} />
      </Routes>
    </BrowserRouter>
  );
}
