import React from "react";
import type { ZoneDto } from "../../lib/api/types";

interface ZoneDetailsProps {
  selectedZones: ZoneDto[];
  onDeleteSelection: () => void;
}

export const ZoneDetails: React.FC<ZoneDetailsProps> = ({
  selectedZones,
  onDeleteSelection,
}) => {
  if (selectedZones.length === 0) {
    return (
      <div className="zone-details" style={{ marginTop: "20px" }}>
        <h3>Détails</h3>
        <p style={{ fontStyle: "italic", color: "#888" }}>Aucune zone sélectionnée.</p>
      </div>
    );
  }

  if (selectedZones.length === 1) {
    const zone = selectedZones[0];
    return (
      <div className="zone-details" style={{ marginTop: "20px" }}>
        <h3>Détails : {zone.label}</h3>
        <div style={{ marginBottom: "10px" }}>
          <strong>ID :</strong> {zone.id} <br />
          <strong>Points :</strong> {zone.points.length} <br />
          <strong>Créé le :</strong> {new Date(zone.createdAt).toLocaleString()} <br />
          <strong>Modifié le :</strong> {new Date(zone.updatedAt).toLocaleString()}
        </div>
        <pre>{JSON.stringify(zone, null, 2)}</pre>
        <button onClick={onDeleteSelection} style={{ width: "100%", backgroundColor: "#c62828" }}>
          Supprimer la zone
        </button>
      </div>
    );
  }

  return (
    <div className="zone-details" style={{ marginTop: "20px" }}>
      <h3>Détails : Multi-sélection</h3>
      <p>{selectedZones.length} zones sélectionnées.</p>
      <ul style={{ maxHeight: "150px", overflow: "auto", paddingLeft: "20px" }}>
        {selectedZones.map((z) => (
          <li key={z.id}>
            {z.label} ({z.id})
          </li>
        ))}
      </ul>
      <pre>{JSON.stringify(selectedZones, null, 2)}</pre>
      <button onClick={onDeleteSelection} style={{ width: "100%", backgroundColor: "#c62828" }}>
        Supprimer la sélection ({selectedZones.length})
      </button>
    </div>
  );
};
