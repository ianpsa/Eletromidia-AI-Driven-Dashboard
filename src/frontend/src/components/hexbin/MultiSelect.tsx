import { useMemo, useRef, useState } from "react";
import "./MultiSelect.css";
import { useDropdownPosition } from "../../hooks/useDropdownPosition";
import { useOutsideClick } from "../../hooks/useOutsideClick";
import type { MultiSelectProps } from "../../types/hexbin";
import {
  DEFAULT_MULTISELECT_MAX_VISIBLE_OPTIONS,
  formatMultiSelectDisplay,
  getMultiSelectDropdownMaxHeight,
  shouldEnableMultiSelectScroll,
} from "../../utils/multiSelect";
import { MultiSelectDropdown } from "./MultiSelectDropdown";

export function MultiSelect({
  label,
  options,
  selected,
  onChange,
  placeholder = "Todos",
  maxVisibleOptions = DEFAULT_MULTISELECT_MAX_VISIBLE_OPTIONS,
}: MultiSelectProps) {
  const [open, setOpen] = useState(false);
  const rootRef = useRef<HTMLDivElement>(null);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useOutsideClick(rootRef, () => setOpen(false), { ignoreRefs: [dropdownRef] });

  const displayValue = useMemo(
    () => formatMultiSelectDisplay(selected, placeholder),
    [selected, placeholder],
  );

  const dropdownPosition = useDropdownPosition({
    open,
    anchorRef: rootRef,
  });

  const isScrollable = useMemo(
    () => shouldEnableMultiSelectScroll(options.length, maxVisibleOptions),
    [options.length, maxVisibleOptions],
  );

  const dropdownMaxHeight = useMemo(
    () => getMultiSelectDropdownMaxHeight(options.length, maxVisibleOptions),
    [options.length, maxVisibleOptions],
  );

  const toggleOption = (value: string) => {
    onChange(
      selected.includes(value)
        ? selected.filter((item) => item !== value)
        : [...selected, value],
    );
  };

  return (
    <div ref={rootRef} className="multi-select">
      <span className="multi-select__label">{label}</span>

      <button
        type="button"
        className={`multi-select__control ${open ? "is-open" : ""}`}
        onClick={() => setOpen((prev) => !prev)}
      >
        <span className={selected.length === 0 ? "is-placeholder" : ""}>
          {displayValue}
        </span>
        <span className="multi-select__arrow">▾</span>
      </button>

      <MultiSelectDropdown
        open={open}
        options={options}
        selected={selected}
        onToggleOption={toggleOption}
        dropdownRef={dropdownRef}
        positionStyle={dropdownPosition}
        maxHeight={dropdownMaxHeight}
        isScrollable={isScrollable}
      />
    </div>
  );
}
