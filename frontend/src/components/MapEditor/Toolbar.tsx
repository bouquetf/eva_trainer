import React from "react";
import type { EditorTool } from "../../lib/api/types";

interface ToolbarProps {
  activeTool: EditorTool;
  setActiveTool: (tool: EditorTool) => void;
  onSave: () => void;
  onReload: () => void;
  onReset: () => void;
  onSelectAll: () => void;
  onClearSelection: () => void;
  onDeleteSelection: () => void;
  onCopy: () => void;
  onPaste: () => void;
  onFlipH: () => void;
  onFlipV: () => void;
  selectedCount: number;
  tempPointsCount: number;
}

export const Toolbar: React.FC<ToolbarProps> = ({
  activeTool,
  setActiveTool,
  onSave,
  onReload,
  onReset,
  onSelectAll,
  onClearSelection,
  onDeleteSelection,
  onCopy,
  onPaste,
  onFlipH,
  onFlipV,
  selectedCount,
  tempPointsCount,
}) => {
  return (
    <div className="toolbar">
      <div style={{ borderRight: "1px solid #444", paddingRight: "10px", display: "flex", gap: "5px" }}>
        <button className={activeTool === "select" ? "active" : ""} onClick={() => setActiveTool("select")}>Sélection</button>
        <button className={activeTool === "draw-polygon" ? "active" : ""} onClick={() => setActiveTool("draw-polygon")}>Dessiner</button>
        <button className={activeTool === "move-zone" ? "active" : ""} onClick={() => setActiveTool("move-zone")}>Déplacer Zone</button>
        <button className={activeTool === "edit-points" ? "active" : ""} onClick={() => setActiveTool("edit-points")}>Éditer Points</button>
      </div>

      <div style={{ display: "flex", gap: "5px" }}>
        <button onClick={onSave} title="Sauvegarder dans localStorage">💾 Sauvegarder</button>
        <button onClick={onReload} title="Recharger depuis localStorage">🔄 Recharger</button>
        <button onClick={onReset} title="Supprimer tout">🗑️ Reset</button>
      </div>

      <div style={{ display: "flex", gap: "5px" }}>
        <button onClick={onSelectAll}>Tout sélectionner</button>
        <button onClick={onClearSelection} disabled={selectedCount === 0}>Désélectionner</button>
        <button onClick={onDeleteSelection} disabled={selectedCount === 0}>Supprimer ({selectedCount})</button>
      </div>

      <div style={{ display: "flex", gap: "5px" }}>
        <button onClick={onCopy} disabled={selectedCount === 0} title="Copier la sélection">📋 Copier</button>
        <button onClick={onPaste} title="Coller la zone copiée">📋 Coller</button>
        <button onClick={onFlipH} disabled={selectedCount === 0} title="Miroir Horizontal">↕ Flip H</button>
        <button onClick={onFlipV} disabled={selectedCount === 0} title="Miroir Vertical">↔ Flip V</button>
      </div>

      <div className="help-text">
        {activeTool === "draw-polygon" && `Dessin: ${tempPointsCount} points. Clic G: ajouter, Ctrl+Clic: undo, Clic Milieu: terminer.`}
        {activeTool === "select" && "Sélection: Clic simple ou Ctrl/Cmd+Clic pour multiple."}
        {activeTool === "move-zone" && "Déplacement: Glissez une zone sélectionnée."}
        {activeTool === "edit-points" && "Édition: Glissez les poignées blanches."}
      </div>
    </div>
  );
};
