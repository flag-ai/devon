import { BrowserRouter, Route, Routes } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import Layout from "./components/Layout";
import AuthGate from "./components/AuthGate";
import Dashboard from "./pages/Dashboard";
import Search from "./pages/Search";
import Models from "./pages/Models";
import Downloads from "./pages/Downloads";
import Agents from "./pages/Agents";
import Settings from "./pages/Settings";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      retry: 1,
    },
  },
});

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AuthGate>
          <Routes>
            <Route element={<Layout />}>
              <Route index element={<Dashboard />} />
              <Route path="search" element={<Search />} />
              <Route path="models" element={<Models />} />
              <Route path="downloads" element={<Downloads />} />
              <Route path="agents" element={<Agents />} />
              <Route path="settings" element={<Settings />} />
            </Route>
          </Routes>
        </AuthGate>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
