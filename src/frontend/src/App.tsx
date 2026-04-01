import { useState } from "react";
import { useAuth } from "./AuthContext";
import { ChatSidebar } from "./components/ChatSidebar";
import { Login } from "./components/Login";
import { MobileBottomBar } from "./components/MobileBottomBar";
import { Sidebar } from "./components/Sidebar";
import { AnalyticsPage } from "./pages/AnalyticsPage";
import { CloudPage } from "./pages/CloudPage";
import { DashboardPage } from "./pages/DashboardPage";

type Page = "dashboard" | "analytics" | "cloud";

function AppShell() {
  const [page, setPage] = useState<Page>("dashboard");
  const [chatOpen, setChatOpen] = useState(false);
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);
  const [lookerUrl, setLookerUrl] = useState<string | undefined>(undefined);

  function handleLookerUrl(url: string) {
    setLookerUrl(url);
    setPage("analytics");
  }

  return (
    <div className={`drive-shell ${sidebarCollapsed ? "sidebar-is-collapsed" : ""}`}>
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
        {page === "analytics" && <AnalyticsPage src={lookerUrl} />}
        {page === "cloud" && <CloudPage />}
      </main>
      <ChatSidebar
        open={chatOpen}
        onClose={() => setChatOpen(false)}
        onLookerUrl={handleLookerUrl}
      />
      <MobileBottomBar
        activePage={page}
        onNavigate={setPage}
        onToggleChat={() => setChatOpen((v) => !v)}
        chatOpen={chatOpen}
      />
    </div>
  );
}

function App() {
  const { user, loading } = useAuth();

  // While Firebase is resolving the persisted session, show nothing
  if (loading) return null;

  if (!user) return <Login />;

  return <AppShell />;
}

export default App;