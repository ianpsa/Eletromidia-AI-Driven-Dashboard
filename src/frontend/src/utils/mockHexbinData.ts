import type { HexbinPoint } from "../types/geo";

function randomBetween(min: number, max: number) {
  return min + Math.random() * (max - min);
}

function gaussianRandom() {
  let u = 0;
  let v = 0;

  while (u === 0) u = Math.random();
  while (v === 0) v = Math.random();

  return Math.sqrt(-2.0 * Math.log(u)) * Math.cos(2.0 * Math.PI * v);
}

type Cluster = {
  lng: number;
  lat: number;
  spreadLng: number;
  spreadLat: number;
  weight: number;
};

const CLUSTERS: Cluster[] = [
  {
    lng: -46.6333,
    lat: -23.5505,
    spreadLng: 0.035,
    spreadLat: 0.028,
    weight: 0.38,
  },
  {
    lng: -46.69,
    lat: -23.58,
    spreadLng: 0.022,
    spreadLat: 0.018,
    weight: 0.22,
  },
  {
    lng: -46.57,
    lat: -23.52,
    spreadLng: 0.018,
    spreadLat: 0.016,
    weight: 0.18,
  },
  { lng: -46.61, lat: -23.61, spreadLng: 0.02, spreadLat: 0.02, weight: 0.22 },
];

function pickClusterIndex() {
  const roll = Math.random();
  let acc = 0;

  for (let i = 0; i < CLUSTERS.length; i++) {
    acc += CLUSTERS[i].weight;
    if (roll <= acc) return i;
  }

  return CLUSTERS.length - 1;
}

export function generateMockHexbinData(count = 3000): HexbinPoint[] {
  return Array.from({ length: count }, (_, index) => {
    const cluster = CLUSTERS[pickClusterIndex()];

    const lngNoise = gaussianRandom() * cluster.spreadLng;
    const latNoise = gaussianRandom() * cluster.spreadLat;

    const driftLng = randomBetween(-0.004, 0.004);
    const driftLat = randomBetween(-0.004, 0.004);

    return {
      id: String(index + 1),
      longitude: cluster.lng + lngNoise + driftLng,
      latitude: cluster.lat + latNoise + driftLat,
      value: 1,
    };
  });
}
