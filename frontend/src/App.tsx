import { HashRouter, Routes, Route } from "react-router-dom";
import { Sidebar } from "@/components/sidebar";
import OverviewPage from "@/app/overview/page";
import ModelsPage from "@/app/models/page";
import SystemPage from "@/app/system/page";

export function App() {
  return (
    <HashRouter>
      <div className="h-screen flex bg-background text-foreground">
        <Sidebar />
        <main className="flex-1 overflow-auto">
          <Routes>
            <Route path="/" element={<OverviewPage />} />
            <Route path="/models" element={<ModelsPage />} />
            <Route path="/system" element={<SystemPage />} />
          </Routes>
        </main>
      </div>
    </HashRouter>
  );
}

export default App
