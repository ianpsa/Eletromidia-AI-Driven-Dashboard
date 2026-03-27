import { HexagonLayer } from "@deck.gl/aggregation-layers";
import type { MapViewState } from "@deck.gl/core";
import DeckGL from "@deck.gl/react";
import { useMemo, useState } from "react";
import MapView from "react-map-gl/maplibre";
import "maplibre-gl/dist/maplibre-gl.css";
import type { HexbinPoint } from "../../types/geo";
import { MAP_STYLE } from "../../utils/hexbinOptions";
import "./HexbinChart.css";

type HexbinChartProps = {
  title: string;
  data: HexbinPoint[];
  height?: number;
  initialViewState?: MapViewState;
};

const DEFAULT_VIEW_STATE: MapViewState = {
  longitude: -46.6333,
  latitude: -23.5505,
  zoom: 10.8,
  pitch: 0,
  bearing: 0,
};

const HEXBIN_RADIUS_METERS = 700;



export function HexbinChart({
  title,
  data,
  height = 480,
  initialViewState = DEFAULT_VIEW_STATE,
}: HexbinChartProps) {
  const [viewState, setViewState] = useState<MapViewState>(initialViewState);

  const { minDensityValue, maxDensityValue } = useMemo(() => {
    if (data.length === 0) {
      return { minDensityValue: 0, maxDensityValue: 0 };
    }

    // Aproxima a agregação espacial para manter a legenda coerente
    // com a métrica visual do hexbin (registros por área).
    const latRef =
      data.reduce((acc, point) => acc + point.latitude, 0) / data.length;
    const metersPerDegLat = 110_540;
    const metersPerDegLon = 111_320 * Math.cos((latRef * Math.PI) / 180);

    const buckets = new Map<string, number>();

    for (const point of data) {
      const xMeters = point.longitude * metersPerDegLon;
      const yMeters = point.latitude * metersPerDegLat;

      const cellX = Math.floor(xMeters / HEXBIN_RADIUS_METERS);
      const cellY = Math.floor(yMeters / HEXBIN_RADIUS_METERS);
      const key = `${cellX}:${cellY}`;

      buckets.set(key, (buckets.get(key) ?? 0) + 1);
    }

    let min = Number.POSITIVE_INFINITY;
    let max = Number.NEGATIVE_INFINITY;

    for (const count of buckets.values()) {
      if (count < min) min = count;
      if (count > max) max = count;
    }

    return { minDensityValue: min, maxDensityValue: max };
  }, [data]);

  const formatDensityLabel = (value: number) =>
    new Intl.NumberFormat("pt-BR", {
      maximumFractionDigits: 0,
    }).format(value);

  const layers = useMemo(() => {
    return [
      new HexagonLayer<HexbinPoint>({
        id: `hexbin-${title}`,
        data,
        pickable: true,
        extruded: false,
  radius: HEXBIN_RADIUS_METERS,
        coverage: 0.82,
        upperPercentile: 100,
        opacity: 0.55,
        colorRange: [
          [252, 187, 161],
          [252, 146, 114],
          [251, 106, 74],
          [239, 85, 59],
          [222, 45, 38],
          [165, 15, 21],
        ],
        getPosition: (d: HexbinPoint) => [d.longitude, d.latitude],
      }),
    ];
  }, [data, title]);

  return (
    <section className="hexbin-card">
      <div className="hexbin-card__header">
        <h3>{title}</h3>
      </div>

      <div className="hexbin-card__stage" style={{ height }}>
        <div className="hexbin-map-wrap">
          <DeckGL
            style={{ position: "absolute", inset: "0" }}
            viewState={viewState}
            controller
            layers={layers}
            onViewStateChange={({ viewState: next }) =>
              setViewState(next as MapViewState)
            }
            getTooltip={({ object }) => {
              if (!object) return null;

              const points = (object as Record<string, unknown>).points as
                | unknown[]
                | undefined;
              const count =
                points?.length ??
                ((object as Record<string, unknown>).count as number) ??
                0;

              return {
                html: `<strong>${count} registros na área</strong>`,
              };
            }}
          >
            <MapView
              reuseMaps
              mapStyle={MAP_STYLE}
              style={{
                position: "absolute",
                inset: 0,
                width: "100%",
                height: "100%",
              }}
            />
          </DeckGL>
        </div>
      </div>

      <div className="hexbin-card__legend">
        <span className="hexbin-card__legend-title">Escala de densidade</span>
        <div
          className="hexbin-card__legend-scale"
          role="img"
          aria-label="Escala de cor de densidade, do mínimo ao máximo"
        />
        <div className="hexbin-card__legend-labels">
          <span>Mín: {formatDensityLabel(minDensityValue)}</span>
          <span>Máx: {formatDensityLabel(maxDensityValue)}</span>
        </div>
      </div>
    </section>
  );
}
