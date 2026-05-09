import React from "react";
import type { Zone } from "../../lib/api/types";

interface ZoneDetailsProps {
  selectedZones: Zone[];
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
    const polygonCount = zone.polygons.length;
    const totalPoints = zone.polygons.reduce((s, p) => s + p.points.length, 0);

    return (
      <div className="zone-details" style={{ marginTop: "20px" }}>
        <h3>Détails : {zone.name || `Zone #${zone.id}`}</h3>
        <div style={{ marginBottom: "10px" }}>
          <strong>ID :</strong> {zone.id} <br />
          <strong>Polygones :</strong> {polygonCount} <br />
          <strong>Points total :</strong> {totalPoints} <br />
          {zone.polygons.map((poly, i) => (
            <div key={i} style={{ marginTop: "4px", fontSize: "12px", color: "#aaa" }}>
              Polygone {i + 1} : {poly.points.length} pts {poly.id > 0 ? `(id: ${poly.id})` : "(non sauvegardé)"}
            </div>
          ))}
          <strong>Créé le :</strong> {new Date(zone.created_at).toLocaleString()} <br />
          <strong>Modifié le :</strong> {new Date(zone.updated_at).toLocaleString()}
        </div>
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
            {z.name || `Zone #${z.id}`} ({z.polygons.length} polygones)
          </li>
        ))}
      </ul>
      <button onClick={onDeleteSelection} style={{ width: "100%", backgroundColor: "#c62828" }}>
        Supprimer la sélection ({selectedZones.length})
      </button>
    </div>
  );
};
