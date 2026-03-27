import { useState } from "react";
import { HexbinChart } from "../components/hexbin/HexbinChart";
import { HexbinFilters } from "../components/hexbin/HexbinFilters";
import { CompareChartsModal } from "../components/hexbin/CompareChartsModal";
import { generateMockHexbinData } from "../utils/mockHexbinData";

const data = generateMockHexbinData(4500);

export function DashboardPage() {
  const [compareOpen, setCompareOpen] = useState(false);

  return (
    <>
      <div className="dashboard-grid dashboard-grid--hexbin">
        <div className="main-column">
          <HexbinChart
            title="Mapa de densidade de pessoas"
            data={data}
            height={520}
          />
        </div>

        <HexbinFilters />
      </div>

      <div className="hexbin-compare-bar">
        <button className="hexbin-compare-button" onClick={() => setCompareOpen(true)}>
          Comparar gráficos
        </button>
      </div>

      <CompareChartsModal
        open={compareOpen}
        onClose={() => setCompareOpen(false)}
        onConfirm={(config) => {
          console.log("config de comparação:", config);
        }}
      />
    </>
  );
}