import { type CSSProperties, useCallback, useEffect, useState } from "react";
import type { MultiSelectDropdownPositionParams } from "../types/multiSelect";

export function useDropdownPosition({
  open,
  anchorRef,
  offset = 4,
}: MultiSelectDropdownPositionParams) {
  const [positionStyle, setPositionStyle] = useState<CSSProperties>({});

  const updatePosition = useCallback(() => {
    if (!open || !anchorRef.current) return;

    const rect = anchorRef.current.getBoundingClientRect();

    setPositionStyle({
      position: "fixed",
      top: rect.bottom + offset,
      left: rect.left,
      width: rect.width,
      zIndex: 1200,
    });
  }, [anchorRef, open, offset]);

  useEffect(() => {
    if (!open) return;

    updatePosition();

    window.addEventListener("resize", updatePosition);
    window.addEventListener("scroll", updatePosition, true);

    return () => {
      window.removeEventListener("resize", updatePosition);
      window.removeEventListener("scroll", updatePosition, true);
    };
  }, [open, updatePosition]);

  return positionStyle;
}
