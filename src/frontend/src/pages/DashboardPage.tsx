import { HexbinChart } from "../components/hexbin/HexbinChart";
import { generateMockHexbinData } from "../utils/mockHexbinData";

const hexbinData = generateMockHexbinData(4500);

export function DashboardPage() {
  return (
    <div className="dashboard-page">
      <HexbinChart
        title="Mapa de densidade de pessoas"
        data={hexbinData}
        height={520}
      />
    </div>
  );
}