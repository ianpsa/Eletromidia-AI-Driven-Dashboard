import { useMemo, useState } from "react";
import { Login } from "./components/Login";
import { useBucketFolder } from "./hooks/useBucketFolder";
import { computeSortedItems } from "./utils/sort";
import Sidebar from "./components/Sidebar";
import AgentFullPage from "./pages/AgentFullPage";
import { TopBar } from "./components/TopBar";
import { Breadcrumb } from "./components/Breadcrumb";
import { SummaryCards } from "./components/SummaryCards";
import { StatusMessage } from "./components/StatusMessage";
import { FilesTable } from "./components/FilesTable";

function Dashboard() {
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
        { !bucket.loading && !bucket.error && sortedItems.length > 0 && (
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

function App() {
  const [loggedIn, setLoggedIn] = useState(false);
  const [route, setRoute] = useState<"cloud" | "agent">("cloud");

  if (!loggedIn) return <Login onLogin={() => setLoggedIn(true)} />;

  return (
    <div className="drive-shell">
      <Sidebar onNavigate={(r) => setRoute(r as any)} active={route} />
      <main className="content">
        {route === "cloud" ? <Dashboard /> : <AgentFullPage />}
      </main>
    </div>
  );
}

export default App;

