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



export function HexbinChart({
  title,
  data,
  height = 480,
  initialViewState = DEFAULT_VIEW_STATE,
}: HexbinChartProps) {
  const [viewState, setViewState] = useState<MapViewState>(initialViewState);

  const layers = useMemo(() => {
    return [
      new HexagonLayer<HexbinPoint>({
        id: `hexbin-${title}`,
        data,
        pickable: true,
        extruded: false,
        radius: 700,
        coverage: 0.82,
        upperPercentile: 100,
        opacity: 0.55,
        colorRange: [
          [255, 245, 240],
          [254, 224, 210],
          [252, 187, 161],
          [252, 146, 114],
          [251, 106, 74],
          [222, 45, 38],
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
                html: `<strong>${count} registros</strong>`,
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
        <span>
          <i className="hexbin-card__legend-hot" />
          Alta densidade
        </span>
        <span>
          <i className="hexbin-card__legend-cold" />
          Baixa densidade
        </span>
      </div>
    </section>
  );
}
