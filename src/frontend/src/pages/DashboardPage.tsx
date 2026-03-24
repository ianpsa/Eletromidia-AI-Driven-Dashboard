import { HexbinChart } from "../components/hexbin/HexbinChart";
import { HexbinFilters } from "../components/hexbin/HexbinFilters";
import { generateMockHexbinData } from "../utils/mockHexbinData";

const data = generateMockHexbinData(4500);

export function DashboardPage() {
  return (
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
  );
}