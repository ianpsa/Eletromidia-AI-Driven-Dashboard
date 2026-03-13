import { useMemo, useState } from "react";
import { Login } from "./components/Login";
import { LookerDashboardPage } from "./components/LookerDashboardPage";
import { useBucketFolder } from "./hooks/useBucketFolder";
import { computeSortedItems } from "./utils/sort";
import Sidebar from "./components/Sidebar";
import AgentFullPage from "./pages/AgentFullPage";
import { TopBar } from "./components/TopBar";
import { Breadcrumb } from "./components/Breadcrumb";
import { SummaryCards } from "./components/SummaryCards";
import { StatusMessage } from "./components/StatusMessage";
import { FilesTable } from "./components/FilesTable";

type Page = "dashboard" | "cloud" | "agent";

function CloudPage({ onNavigate }: { onNavigate: (page: Page) => void }) {
  const bucket = useBucketFolder();
  const [query, setQuery] = useState("");
  const [sortField, setSortField] = useState<"name" | "size">("name");
  const [sortAsc, setSortAsc] = useState(true);

  const sortedItems = useMemo(
    () =>
      computeSortedItems(
        bucket.folders,
        bucket.files,
        query,
        sortField,
        sortAsc,
      ),
    [bucket.folders, bucket.files, query, sortField, sortAsc],
  );

  function navigateTo(folder: string) {
    setQuery("");
    bucket.navigateTo(folder);
  }

  function handleSort(field: "name" | "size") {
    if (sortField === field) {
      setSortAsc(!sortAsc);
    } else {
      setSortField(field);
      setSortAsc(true);
    }
  }

  const isEmpty = !bucket.loading && !bucket.error && sortedItems.length === 0;

  return (
    <>
      <Sidebar activePage="cloud" onNavigate={onNavigate} />
      <main className="content">
        <TopBar
          bucketName={bucket.bucketName}
          query={query}
          onQueryChange={setQuery}
          onRefresh={bucket.refresh}
        />
        <Breadcrumb
          currentFolder={bucket.currentFolder}
          breadcrumbs={bucket.breadcrumbs}
          onNavigate={navigateTo}
        />
        <SummaryCards folders={bucket.folders} files={bucket.files} />
        <StatusMessage
          loading={bucket.loading}
          error={bucket.error}
          empty={isEmpty}
          downloadError={bucket.downloadError}
        />
        {!bucket.loading && !bucket.error && sortedItems.length > 0 && (
          <FilesTable
            items={sortedItems}
            currentFolder={bucket.currentFolder}
            sortField={sortField}
            sortAsc={sortAsc}
            downloadingId={bucket.downloadingId}
            onSort={handleSort}
            onNavigate={navigateTo}
            onNavigateUp={bucket.navigateUp}
            onDownload={bucket.handleDownload}
          />
        )}
      </main>
    </>
  );
}

function AppShell() {
  const [page, setPage] = useState<Page>("dashboard");

  return (
    <div className="drive-shell">
      {page === "dashboard" ? (
        <>
          <Sidebar activePage="dashboard" onNavigate={setPage} />
          <LookerDashboardPage />
        </>
      ) : page === "agent" ? (
        <>
          <Sidebar activePage="agent" onNavigate={setPage} />
          <main className="content">
            <AgentFullPage />
          </main>
        </>
      ) : (
        <CloudPage onNavigate={setPage} />
      )}
    </div>
  );
}

function App() {
  const [loggedIn, setLoggedIn] = useState(false);

  if (!loggedIn) return <Login onLogin={() => setLoggedIn(true)} />;

  return <AppShell />;
}

export default App;
