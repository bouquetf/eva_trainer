import React from "react";
import type { ZoneDto } from "../../lib/api/types";

interface ZoneListProps {
  zones: ZoneDto[];
  selectedZoneIds: Set<string>;
  onZoneClick: (id: string, event: React.MouseEvent) => void;
  onDeleteZone: (id: string) => void;
}

export const ZoneList: React.FC<ZoneListProps> = ({
  zones,
  selectedZoneIds,
  onZoneClick,
  onDeleteZone,
}) => {
  return (
    <div className="zone-list">
      <h3>Zones ({zones.length})</h3>
      <div style={{ display: "flex", flexDirection: "column", gap: "5px" }}>
        {zones.map((zone) => {
          const isSelected = selectedZoneIds.has(zone.id);
          return (
            <div
              key={zone.id}
              onClick={(e) => onZoneClick(zone.id, e)}
              style={{
                padding: "8px",
                backgroundColor: isSelected ? "#2e7d32" : "#333",
                borderRadius: "4px",
                cursor: "pointer",
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                border: isSelected ? "1px solid #00e676" : "1px solid transparent",
              }}
            >
              <div>
                <div style={{ fontWeight: "bold" }}>{zone.label}</div>
                <div style={{ fontSize: "10px", color: "#aaa" }}>
                  {zone.id} ({zone.points.length} pts)
                </div>
              </div>
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  onDeleteZone(zone.id);
                }}
                style={{ padding: "2px 6px", fontSize: "12px" }}
              >
                ✕
              </button>
            </div>
          );
        })}
        {zones.length === 0 && <div style={{ fontStyle: "italic", color: "#666" }}>Aucune zone.</div>}
      </div>
    </div>
  );
};
