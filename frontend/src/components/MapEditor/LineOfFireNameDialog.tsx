import React, { useState, useEffect } from "react";

interface LineOfFireNameDialogProps {
  open: boolean;
  initialName?: string;
  onConfirm: (name: string) => void;
  onCancel: () => void;
}

export const LineOfFireNameDialog: React.FC<LineOfFireNameDialogProps> = ({
  open,
  initialName = "",
  onConfirm,
  onCancel,
}) => {
  const [name, setName] = useState(initialName);
  const [error, setError] = useState("");

  useEffect(() => {
    if (open) {
      setName(initialName);
      setError("");
    }
  }, [open, initialName]);

  if (!open) return null;

  const handleConfirm = () => {
    const trimmed = name.trim();
    if (!trimmed) {
      setError("Le nom ne peut pas être vide.");
      return;
    }
    if (trimmed.length > 250) {
      setError("Le nom ne peut pas dépasser 250 caractères.");
      return;
    }
    setError("");
    onConfirm(trimmed);
    setName("");
  };

  const handleCancel = () => {
    setError("");
    setName("");
    onCancel();
  };

  const charCount = name.length;
  const isValid = name.trim().length > 0 && name.trim().length <= 250;

  return (
    <div
      style={{
        position: "fixed",
        inset: 0,
        background: "rgba(0,0,0,0.5)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        zIndex: 2000,
      }}
      onClick={handleCancel}
    >
      <div
        style={{
          background: "#1e1e1e",
          border: "1px solid #333",
          borderRadius: "8px",
          padding: "24px",
          minWidth: "350px",
          maxWidth: "450px",
        }}
        onClick={(e) => e.stopPropagation()}
      >
        <h3 style={{ margin: "0 0 16px 0" }}>Nommer la ligne de tir</h3>

        <input
          type="text"
          value={name}
          onChange={(e) => {
            setName(e.target.value);
            setError("");
          }}
          placeholder="Nom de la ligne de tir..."
          maxLength={250}
          autoFocus
          style={{
            width: "100%",
            padding: "10px",
            fontSize: "14px",
            borderRadius: "4px",
            border: `1px solid ${error ? "#c62828" : isValid ? "#00e676" : "#444"}`,
            background: "#2a2a2a",
            color: "#e0e0e0",
            outline: "none",
            boxSizing: "border-box",
          }}
          onKeyDown={(e) => {
            if (e.key === "Enter") handleConfirm();
            if (e.key === "Escape") handleCancel();
          }}
        />

        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            marginTop: "8px",
            fontSize: "12px",
          }}
        >
          <span style={{ color: error ? "#c62828" : "#888" }}>
            {error || `${charCount}/250`}
          </span>
        </div>

        <div style={{ display: "flex", gap: "10px", justifyContent: "flex-end", marginTop: "16px" }}>
          <button
            onClick={handleCancel}
            style={{
              padding: "8px 20px",
              background: "#333",
              border: "1px solid #555",
              borderRadius: "4px",
              color: "#e0e0e0",
              cursor: "pointer",
            }}
          >
            Annuler
          </button>
          <button
            onClick={handleConfirm}
            disabled={!isValid}
            style={{
              padding: "8px 20px",
              background: isValid ? "#00e676" : "#444",
              border: "none",
              borderRadius: "4px",
              color: isValid ? "#000" : "#888",
              cursor: isValid ? "pointer" : "default",
              fontWeight: "bold",
            }}
          >
            Créer la ligne
          </button>
        </div>
      </div>
    </div>
  );
};
