export type GeodataFilterOptionsResponse = {
  estados?: string[];
  cidades?: string[];
  enderecos?: string[];
  horarios?: string[];
  generos?: string[];
  faixas_etarias?: string[];
  classes_sociais?: string[];
};

export type GeodataFilterOptions = {
  states: string[];
  cities: string[];
  addresses: string[];
  hours: string[];
  genders: string[];
  ages: string[];
  socialClasses: string[];
};

export type GeodataPointResponse = {
  id: string;
  latitude: number;
  longitude: number;
  value?: number;
};

export type GeodataPointsResponse = {
  count?: number;
  points?: GeodataPointResponse[];
};
