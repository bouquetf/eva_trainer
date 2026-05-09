import React from "react";

interface LockIndicatorProps {
  isLocked: boolean;
  lockedByMe: boolean;
  editorName?: string;
  canEdit: boolean;
}

export const LockIndicator: React.FC<LockIndicatorProps> = ({
  isLocked,
  lockedByMe,
  editorName,
  canEdit,
}) => {
  return (
    <div style={{ display: "flex", gap: "12px", alignItems: "center", fontSize: "12px" }}>
      {!canEdit && !isLocked && (
        <span style={{ color: "#ff9800" }}>Verrouillage requis pour éditer</span>
      )}

      {isLocked && !lockedByMe && editorName && (
        <span style={{ color: "#f44336" }}>
          Verrouillé par {editorName}
        </span>
      )}

      {isLocked && lockedByMe && (
        <span style={{ color: "#00e676" }}>Édition en cours</span>
      )}
    </div>
  );
};
