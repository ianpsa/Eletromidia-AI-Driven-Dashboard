import { useState } from "react";
import { ChatSidebar } from "./components/ChatSidebar";
import { Login } from "./components/Login";
import { Sidebar } from "./components/Sidebar";
import { AnalyticsPage } from "./pages/AnalyticsPage";
import { CloudPage } from "./pages/CloudPage";
import { DashboardPage } from "./pages/DashboardPage";

type Page = "dashboard" | "analytics" | "cloud";

function AppShell() {
  const [page, setPage] = useState<Page>("dashboard");
  const [chatOpen, setChatOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

  return (
    <div
      className={`drive-shell ${sidebarCollapsed ? "sidebar-is-collapsed" : ""}`}
    >
      <Sidebar
        activePage={page}
        onNavigate={setPage}
        onToggleChat={() => setChatOpen((v) => !v)}
        chatOpen={chatOpen}
        collapsed={sidebarCollapsed}
        onToggleCollapse={() => setSidebarCollapsed((v) => !v)}
      />
      <main className="content">
        {page === "dashboard" && <DashboardPage />}
        {page === "analytics" && <AnalyticsPage />}
        {page === "cloud" && <CloudPage />}
      </main>
      <ChatSidebar open={chatOpen} onClose={() => setChatOpen(false)} />
    </div>
  );
}

function App() {
  const [loggedIn, setLoggedIn] = useState(false);

  if (!loggedIn) return <Login onLogin={() => setLoggedIn(true)} />;

  return <AppShell />;
}

export default App;
