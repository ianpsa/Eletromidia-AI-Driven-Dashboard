import type {
  GeodataFilterOptions,
  GeodataFilterOptionsResponse,
} from "../types/geodataApi";

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
