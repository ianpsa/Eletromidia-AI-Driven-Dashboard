import { useState } from "react";
import { Login } from "./components/Login";
import { Sidebar } from "./components/Sidebar";
import { AgentPage } from "./pages/AgentPage";
import { CloudPage } from "./pages/CloudPage";
import { DashboardPage } from "./pages/DashboardPage";

type Page = "dashboard" | "cloud" | "agent";

function AppShell() {
  const [page, setPage] = useState<Page>("dashboard");

  return (
    <div className="drive-shell">
      <Sidebar activePage={page} onNavigate={setPage} />
      <main className="content">
        {page === "dashboard" && <DashboardPage />}
        {page === "cloud" && <CloudPage />}
        {page === "agent" && <AgentPage />}
      </main>
    </div>
  );
}

function App() {
  const [loggedIn, setLoggedIn] = useState(false);

  if (!loggedIn) return <Login onLogin={() => setLoggedIn(true)} />;

  return <AppShell />;
}

export default App;
