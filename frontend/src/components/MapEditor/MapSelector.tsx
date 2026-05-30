import React from "react";
import type { Map as MapInfo } from "../../api/maps";

interface MapSelectorProps {
  mapsList: readonly MapInfo[];
  mapId: number;
  onChange: (targetMapId: number) => void;
}

export const MapSelector: React.FC<MapSelectorProps> = ({ mapsList, mapId, onChange }) => {
  return (
      <div style={{ marginBottom: "10px", display: "flex", gap: 5 }}>

        <select
            style={{ flex: 1 }}
            value={mapId}
            onChange={(e) => onChange(Number(e.target.value))}
        >
          {mapsList.map((m) => (
              <option key={m.id} value={m.id}>
                {m.name}
              </option>
          ))}
        </select>
      </div>
  );
};
