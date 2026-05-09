import React from "react";
import { LockIndicator } from "./LockIndicator";

interface ToolbarProps {
  canEdit: boolean;
  isLocked: boolean;
  lockedByMe: boolean;
  editorName?: string;
}

export const Toolbar: React.FC<ToolbarProps> = ({
  canEdit,
  isLocked,
  lockedByMe,
  editorName,
}) => {
  return (
    <div className="toolbar">
      <div style={{ display: "flex", gap: "5px", alignItems: "center", marginLeft: "auto" }}>
        <LockIndicator
          isLocked={isLocked}
          lockedByMe={lockedByMe}
          editorName={editorName}
          canEdit={canEdit}
        />
      </div>
    </div>
  );
};
