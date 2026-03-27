import { useMemo, useState } from "react";
import { MultiSelect } from "./MultiSelect";
import "./HexbinFilters.css";

type HexbinFiltersState = {
  states: string[];
  cities: string[];
  addresses: string[];
  hours: string[];
  genders: string[];
  ages: string[];
  socialClasses: string[];
  maxDistance: number;
};

type HexbinFiltersProps = {
  onApplyFilters?: (filters: HexbinFiltersState) => void;
  onClearFilters?: () => void;
  initialFilters?: Partial<HexbinFiltersState>;
};

const STATE_OPTIONS = ["SP", "RJ", "MG", "PR"];
const CITY_OPTIONS = ["São Paulo", "Campinas", "Santos", "Guarulhos", "Osasco"];
const ADDRESS_OPTIONS = ["Rua M.M.D.C", "Avenida Paulista", "Rua da Consolação", "Rua Vergueiro"];
const HOUR_OPTIONS = ["00:00 - 06:00", "06:00 - 12:00", "12:00 - 18:00", "18:00 - 24:00"];
const GENDER_OPTIONS = ["Feminino", "Masculino", "Outro"];
const AGE_OPTIONS = ["18-19", "20-29", "30-39", "40-49", "50+"];
const CLASS_OPTIONS = ["Classe A/B", "Classe C", "Classe D/E"];

const DEFAULT_STATE: HexbinFiltersState = {
  states: [],
  cities: [],
  addresses: [],
  hours: [],
  genders: [],
  ages: [],
  socialClasses: [],
  maxDistance: 10,
};

export function HexbinFilters({
  onApplyFilters,
  onClearFilters,
  initialFilters,
}: HexbinFiltersProps) {
  const [filters, setFilters] = useState<HexbinFiltersState>({
    ...DEFAULT_STATE,
    ...initialFilters,
  });

  const hasAnySelection = useMemo(
    () =>
      filters.states.length > 0 ||
      filters.cities.length > 0 ||
      filters.addresses.length > 0 ||
      filters.hours.length > 0 ||
      filters.genders.length > 0 ||
      filters.ages.length > 0 ||
      filters.socialClasses.length > 0 ||
      filters.maxDistance !== DEFAULT_STATE.maxDistance,
    [filters]
  );

  const setField = <K extends keyof HexbinFiltersState>(key: K, value: HexbinFiltersState[K]) => {
    setFilters((prev) => ({ ...prev, [key]: value }));
  };

  const handleApply = () => {
    if (onApplyFilters) {
      onApplyFilters(filters);
      return;
    }

    console.log("Filtros aplicados:", filters);
  };

  const handleClear = () => {
    const cleared = { ...DEFAULT_STATE };
    setFilters(cleared);

    if (onClearFilters) {
      onClearFilters();
      return;
    }

    console.log("Filtros limpos");
  };

  return (
    <aside className="hexbin-filters">
      <h3 className="hexbin-filters__title">Filtros</h3>

      <MultiSelect
        label="Estado"
        options={STATE_OPTIONS}
        selected={filters.states}
        onChange={(values) => setField("states", values)}
      />

      <MultiSelect
        label="Cidade"
        options={CITY_OPTIONS}
        selected={filters.cities}
        onChange={(values) => setField("cities", values)}
      />

      <MultiSelect
        label="Endereço"
        options={ADDRESS_OPTIONS}
        selected={filters.addresses}
        onChange={(values) => setField("addresses", values)}
      />

      <label className="hexbin-filters__field">
        <span>Distância máxima</span>
        <input
          type="range"
          min={2}
          max={15}
          value={filters.maxDistance}
          onChange={(e) => setField("maxDistance", Number(e.target.value))}
        />
        <strong>{filters.maxDistance} km</strong>
      </label>

      <MultiSelect
        label="Horário"
        options={HOUR_OPTIONS}
        selected={filters.hours}
        onChange={(values) => setField("hours", values)}
      />

      <MultiSelect
        label="Gênero"
        options={GENDER_OPTIONS}
        selected={filters.genders}
        onChange={(values) => setField("genders", values)}
      />

      <MultiSelect
        label="Faixa etária"
        options={AGE_OPTIONS}
        selected={filters.ages}
        onChange={(values) => setField("ages", values)}
      />

      <MultiSelect
        label="Classe social"
        options={CLASS_OPTIONS}
        selected={filters.socialClasses}
        onChange={(values) => setField("socialClasses", values)}
      />

      <div className="hexbin-filters__actions">
        <button type="button" className="hexbin-filters__primary" onClick={handleApply}>
          Aplicar filtros
        </button>
        <button
          type="button"
          className="hexbin-filters__secondary"
          onClick={handleClear}
          disabled={!hasAnySelection}
        >
          Limpar filtros
        </button>
      </div>
    </aside>
  );
}