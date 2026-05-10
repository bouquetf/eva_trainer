import React from "react";
import { LockIndicator } from "./LockIndicator";

interface ToolbarProps {
  canEdit: boolean;
  isLocked: boolean;
  lockedByMe: boolean;
  editorName?: string;
  remainingSeconds?: number;
  onRefreshLock?: () => void;
  onReacquireLock?: () => void;
  showReacquireButton?: boolean;
}

export const Toolbar: React.FC<ToolbarProps> = ({
  canEdit,
  isLocked,
  lockedByMe,
  editorName,
  remainingSeconds,
  onRefreshLock,
  onReacquireLock,
  showReacquireButton,
}) => {
  return (
    <div className="toolbar">
      <div style={{ display: "flex", gap: "5px", alignItems: "center", marginLeft: "auto" }}>
        <LockIndicator
          isLocked={isLocked}
          lockedByMe={lockedByMe}
          editorName={editorName}
          canEdit={canEdit}
          remainingSeconds={remainingSeconds}
          onRefreshLock={onRefreshLock}
          onReacquireLock={onReacquireLock}
          showReacquireButton={showReacquireButton}
        />
      </div>
    </div>
  );
};
