export type CompareMode = "gender" | "age" | "socialClass";

export type CompareFilterKey =
  | "location"
  | "hour"
  | "distance"
  | "gender"
  | "age"
  | "socialClass";

export type LocationValue = {
  state: string[];
  city: string[];
  address: string[];
};

export type CompareFilter =
  | {
      key: "location";
      label: string;
      enabled: boolean;
      value: LocationValue;
    }
  | {
      key: "hour";
      label: string;
      enabled: boolean;
      value: string[];
    }
  | {
      key: "distance";
      label: string;
      enabled: boolean;
      value: number;
    }
  | {
      key: "gender" | "age" | "socialClass";
      label: string;
      enabled: boolean;
      value: string[];
    };

export type HexbinFiltersState = {
  states: string[];
  cities: string[];
  addresses: string[];
  hours: string[];
  genders: string[];
  ages: string[];
  socialClasses: string[];
  maxDistance: number;
};

export type MultiSelectProps = {
  label: string;
  options: readonly string[];
  selected: string[];
  onChange: (values: string[]) => void;
  placeholder?: string;
  maxVisibleOptions?: number;
};

export type CompareChartsConfig = {
  compareMode: CompareMode | null;
  filters: CompareFilter[];
};
