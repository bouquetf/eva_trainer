import React, { useState, useRef, useEffect } from "react";
import type { Zone } from "../../lib/api/types";

interface ZoneListProps {
  zones: Zone[];
  editingZoneId: number | null;
  onSelectZone: (id: number) => void;
  onDeleteZone: (id: number) => void;
  onCreateZone: (name: string) => void;
  onRenameZone: (id: number, name: string) => void;
}

function nameInUse(zones: Zone[], name: string, excludeId?: number): boolean {
  const trimmed = name.trim().toLowerCase();
  if (!trimmed) return false;
  return zones.some((z) => {
    if (excludeId != null && z.id === excludeId) return false;
    return (z.name || "").toLowerCase() === trimmed;
  });
}

export const ZoneList: React.FC<ZoneListProps> = ({
  zones,
  editingZoneId,
  onSelectZone,
  onDeleteZone,
  onCreateZone,
  onRenameZone,
}) => {
  const [newZoneName, setNewZoneName] = useState("");
  const [renamingZoneId, setRenamingZoneId] = useState<number | null>(null);
  const [renameValue, setRenameValue] = useState("");
  const renameInputRef = useRef<HTMLInputElement>(null);

  const trimmedNewName = newZoneName.trim();
  const newNameDuplicate = trimmedNewName.length > 0 && nameInUse(zones, trimmedNewName);
  const canCreate = trimmedNewName.length > 0 && !newNameDuplicate;

  const trimmedRename = renameValue.trim();
  const renameDuplicate = trimmedRename.length > 0 && nameInUse(zones, trimmedRename, renamingZoneId ?? undefined);

  useEffect(() => {
    if (renamingZoneId !== null) {
      renameInputRef.current?.focus();
      renameInputRef.current?.select();
    }
  }, [renamingZoneId]);

  const handleCreate = () => {
    if (!canCreate) return;
    onCreateZone(trimmedNewName);
    setNewZoneName("");
  };

  const handleDoubleClick = (zone: Zone) => {
    if (zone.id !== editingZoneId) return;
    setRenamingZoneId(zone.id);
    setRenameValue(zone.name || `Zone #${zone.id}`);
  };

  const commitRename = () => {
    if (renamingZoneId === null) return;
    if (!trimmedRename || renameDuplicate) {
      setRenamingZoneId(null);
      setRenameValue("");
      return;
    }
    const currentName = zones.find((z) => z.id === renamingZoneId)?.name ?? "";
    if (trimmedRename !== currentName) {
      onRenameZone(renamingZoneId, trimmedRename);
    }
    setRenamingZoneId(null);
    setRenameValue("");
  };

  const cancelRename = () => {
    setRenamingZoneId(null);
    setRenameValue("");
  };

  return (
    <div className="zone-list">
      <h3>Zones ({zones.length})</h3>

      <div style={{ display: "flex", gap: "5px", marginBottom: "10px" }}>
        <div style={{ flex: 1, position: "relative" }}>
          <input
            type="text"
            value={newZoneName}
            onChange={(e) => setNewZoneName(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") handleCreate();
            }}
            placeholder="Nouvelle zone..."
            style={{
              width: "100%",
              padding: "6px 8px",
              fontSize: "13px",
              borderRadius: "4px",
              border: `1px solid ${newNameDuplicate ? "#c62828" : "#555"}`,
              background: "#2a2a2a",
              color: "#e0e0e0",
              outline: "none",
              boxSizing: "border-box",
            }}
          />
          {newNameDuplicate && (
            <div style={{ fontSize: "11px", color: "#c62828", marginTop: "2px" }}>
              Ce nom existe déjà
            </div>
          )}
        </div>
        <button
          onClick={handleCreate}
          disabled={!canCreate}
          style={{ padding: "6px 12px", fontSize: "13px", alignSelf: "flex-start" }}
        >
          +
        </button>
      </div>

      <div style={{ display: "flex", flexDirection: "column", gap: "5px" }}>
        {zones.map((zone) => {
          const isEditing = editingZoneId === zone.id;
          const isRenaming = renamingZoneId === zone.id;
          const polygonCount = zone.polygons.length;
          const totalPoints = zone.polygons.reduce((s, p) => s + p.points.length, 0);
          return (
            <div
              key={zone.id}
              onClick={() => { if (!isRenaming) onSelectZone(zone.id); }}
              style={{
                padding: "8px",
                backgroundColor: isEditing ? "#2e7d32" : "#333",
                borderRadius: "4px",
                cursor: "pointer",
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                border: isEditing ? "1px solid #00e676" : "1px solid transparent",
              }}
            >
              <div style={{ flex: 1 }} onDoubleClick={(e) => { e.stopPropagation(); handleDoubleClick(zone); }}>
                {isRenaming ? (
                  <div>
                    <input
                      ref={renameInputRef}
                      type="text"
                      value={renameValue}
                      onChange={(e) => setRenameValue(e.target.value)}
                      onKeyDown={(e) => {
                        e.stopPropagation();
                        if (e.key === "Enter") commitRename();
                        if (e.key === "Escape") cancelRename();
                      }}
                      onBlur={commitRename}
                      onClick={(e) => e.stopPropagation()}
                      maxLength={250}
                      style={{
                        width: "100%",
                        padding: "2px 4px",
                        fontSize: "13px",
                        fontWeight: "bold",
                        borderRadius: "2px",
                        border: `1px solid ${renameDuplicate ? "#c62828" : "#00e676"}`,
                        background: "#1e1e1e",
                        color: "#e0e0e0",
                        outline: "none",
                        boxSizing: "border-box",
                      }}
                    />
                    {renameDuplicate && (
                      <div style={{ fontSize: "11px", color: "#c62828", marginTop: "1px" }}>
                        Ce nom existe déjà
                      </div>
                    )}
                  </div>
                ) : (
                  <div style={{ fontWeight: "bold" }}>{zone.name || `Zone #${zone.id}`}</div>
                )}
                <div style={{ fontSize: "10px", color: "#aaa" }}>
                  {polygonCount} polygone{polygonCount > 1 ? "s" : ""}, {totalPoints} pts
                </div>
              </div>
              <div style={{ display: "flex", gap: "4px" }}>
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    onDeleteZone(zone.id);
                  }}
                  style={{ padding: "2px 6px", fontSize: "12px" }}
                >
                  X
                </button>
              </div>
            </div>
          );
        })}
        {zones.length === 0 && <div style={{ fontStyle: "italic", color: "#666" }}>Aucune zone.</div>}
      </div>
    </div>
  );
};
