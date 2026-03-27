import type { CSSProperties, RefObject } from "react";

export type MultiSelectDropdownPositionParams = {
  open: boolean;
  anchorRef: RefObject<HTMLElement>;
  offset?: number;
};

export type MultiSelectDropdownProps = {
  open: boolean;
  options: readonly string[];
  selected: string[];
  onToggleOption: (value: string) => void;
  dropdownRef: RefObject<HTMLDivElement>;
  positionStyle: CSSProperties;
  maxHeight: number;
  isScrollable: boolean;
};
