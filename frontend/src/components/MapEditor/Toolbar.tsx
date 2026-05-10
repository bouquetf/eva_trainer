import React from "react";
import { LockIndicator } from "./LockIndicator";
import { PublishButton } from "./PublishButton";

interface ToolbarProps {
  canEdit: boolean;
  isLocked: boolean;
  lockedByMe: boolean;
  editorName?: string;
  remainingSeconds?: number;
  onRefreshLock?: () => void;
  onReacquireLock?: () => void;
  showReacquireButton?: boolean;
  teamId: number;
  mapId: number;
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
  teamId,
  mapId,
}) => {
  return (
    <div className="toolbar">
      <div style={{ display: "flex", gap: "5px", alignItems: "center", marginLeft: "auto" }}>
        <PublishButton teamId={teamId} mapId={mapId} />
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
