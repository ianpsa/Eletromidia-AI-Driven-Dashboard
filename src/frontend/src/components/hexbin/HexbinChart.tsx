import { HexagonLayer } from "@deck.gl/aggregation-layers";
import type { MapViewState } from "@deck.gl/core";
import DeckGL from "@deck.gl/react";
import { useEffect, useMemo, useState } from "react";
import type { StyleSpecification } from "maplibre-gl";
import MapView from "react-map-gl/maplibre";
import "maplibre-gl/dist/maplibre-gl.css";
import type { HexbinPoint } from "../../types/geo";
import {
  HEXBIN_DISTANCE_DEFAULT,
  HEXBIN_DISTANCE_MAX,
  HEXBIN_DISTANCE_MIN,
  MAP_STYLE,
} from "../../utils/hexbinOptions";
import "./HexbinChart.css";

type HexbinChartProps = {
  title: string;
  data: HexbinPoint[];
  height?: number;
  maxDistanceKm?: number;
  initialViewState?: MapViewState;
};

const DEFAULT_VIEW_STATE: MapViewState = {
  longitude: -46.6333,
  latitude: -23.5505,
  zoom: 10.8,
  pitch: 0,
  bearing: 0,
};

const MIN_HEXBIN_RADIUS_METERS = 120;
const MAX_HEXBIN_RADIUS_METERS = 3000;
const GRANULARITY_CURVE_EXPONENT = 1.25;
const SECONDARY_MAP_STYLE = "https://demotiles.maplibre.org/style.json";

const FALLBACK_MAP_STYLE: StyleSpecification = {
  version: 8,
  sources: {
    osm: {
      type: "raster",
      tiles: [
        "https://a.tile.openstreetmap.org/{z}/{x}/{y}.png",
        "https://b.tile.openstreetmap.org/{z}/{x}/{y}.png",
        "https://c.tile.openstreetmap.org/{z}/{x}/{y}.png",
      ],
      tileSize: 256,
      attribution: "© OpenStreetMap contributors",
    },
  },
  layers: [
    {
      id: "osm",
      type: "raster",
      source: "osm",
      minzoom: 0,
      maxzoom: 22,
    },
  ],
};

export function HexbinChart({
  title,
  data,
  height = 480,
  maxDistanceKm = HEXBIN_DISTANCE_DEFAULT,
  initialViewState = DEFAULT_VIEW_STATE,
}: HexbinChartProps) {
  const [viewState, setViewState] = useState<MapViewState>(initialViewState);
  const [densityDomain, setDensityDomain] = useState<[number, number]>([0, 0]);
  const [mapStyle, setMapStyle] = useState<string | StyleSpecification>(MAP_STYLE);

  const hexbinRadiusMeters = useMemo(() => {
    const boundedDistanceKm = Math.max(
      HEXBIN_DISTANCE_MIN,
      Math.min(HEXBIN_DISTANCE_MAX, maxDistanceKm),
    );

    const normalizedDistance =
      (boundedDistanceKm - HEXBIN_DISTANCE_MIN) /
      Math.max(1, HEXBIN_DISTANCE_MAX - HEXBIN_DISTANCE_MIN);

    const easedDistance = Math.pow(normalizedDistance, GRANULARITY_CURVE_EXPONENT);

    return (
      MIN_HEXBIN_RADIUS_METERS +
      easedDistance * (MAX_HEXBIN_RADIUS_METERS - MIN_HEXBIN_RADIUS_METERS)
    );
  }, [maxDistanceKm]);

  useEffect(() => {
    setDensityDomain([0, 0]);
  }, [data, hexbinRadiusMeters]);

  const { minDensityValue, maxDensityValue } = useMemo(() => {
    const [rawMin, rawMax] = densityDomain;
    const min = Number.isFinite(rawMin) ? rawMin : 0;
    const max = Number.isFinite(rawMax) ? rawMax : 0;

    if (max < min) {
      return { minDensityValue: 0, maxDensityValue: 0 };
    }

    return { minDensityValue: min, maxDensityValue: max };
  }, [densityDomain]);

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
        radius: hexbinRadiusMeters,
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
        onSetColorDomain: (domain) => {
          if (!domain || domain.length < 2) {
            setDensityDomain([0, 0]);
            return;
          }

          const nextMin = Number.isFinite(domain[0]) ? domain[0] : 0;
          const nextMax = Number.isFinite(domain[1]) ? domain[1] : 0;
          setDensityDomain([nextMin, nextMax]);
        },
      }),
    ];
  }, [data, hexbinRadiusMeters, title]);

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
              mapStyle={mapStyle}
              style={{
                position: "absolute",
                inset: 0,
                width: "100%",
                height: "100%",
              }}
              onError={() => {
                setMapStyle((current) =>
                  typeof current === "string"
                    ? current === SECONDARY_MAP_STYLE
                      ? FALLBACK_MAP_STYLE
                      : SECONDARY_MAP_STYLE
                    : current,
                );
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
