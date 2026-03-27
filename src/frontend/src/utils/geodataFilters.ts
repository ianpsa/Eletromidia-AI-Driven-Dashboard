import type {
  GeodataFilterOptions,
  GeodataFilterOptionsResponse,
  GeodataPointsResponse,
} from "../types/geodataApi";
import type { HexbinPoint } from "../types/geo";
import type { HexbinFiltersState } from "../types/hexbin";

export function toGeoFilterOptions(
  payload: GeodataFilterOptionsResponse | null | undefined,
): GeodataFilterOptions {
  return {
    states: Array.isArray(payload?.estados) ? payload.estados : [],
    cities: Array.isArray(payload?.cidades) ? payload.cidades : [],
    addresses: Array.isArray(payload?.enderecos) ? payload.enderecos : [],
    hours: Array.isArray(payload?.horarios) ? payload.horarios : [],
    genders: Array.isArray(payload?.generos) ? payload.generos : [],
    ages: Array.isArray(payload?.faixas_etarias) ? payload.faixas_etarias : [],
    socialClasses: Array.isArray(payload?.classes_sociais)
      ? payload.classes_sociais
      : [],
  };
}

export function buildGeoFilterOptionsQuery(selectedState?: string, selectedCity?: string) {
  return {
    uf_estado: selectedState,
    cidade: selectedCity,
  };
}

export function buildGeoPointsQuery(filters: HexbinFiltersState) {
  return {
    uf_estado: filters.states[0],
    cidade: filters.cities[0],
    endereco: filters.addresses[0],
    genero: filters.genders[0],
    faixa_etaria: filters.ages[0],
    classe_social: filters.socialClasses[0],
  };
}

export function toHexbinPoints(payload: GeodataPointsResponse | null | undefined): HexbinPoint[] {
  if (!Array.isArray(payload?.points)) {
    return [];
  }

  return payload.points
    .filter(
      (point) =>
        point &&
        Number.isFinite(point.latitude) &&
        Number.isFinite(point.longitude),
    )
    .map((point, index) => ({
      id: point.id || String(index + 1),
      latitude: point.latitude,
      longitude: point.longitude,
      value: point.value,
    }));
}
