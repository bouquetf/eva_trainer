import React from "react";
import { LockIndicator } from "./LockIndicator";
import { PublishButton } from "./PublishButton";
import styles from "./Toolbar.module.css";

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
  onHelpClick?: () => void;
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
  onHelpClick,
}) => {
  return (
    <div className="toolbar">
      {onHelpClick && (
        <button
          className={styles.helpBtn}
          onClick={onHelpClick}
          title="Aide / Raccourcis (F1)"
          aria-label="Aide"
          data-testid="help-btn"
        >
          ?
        </button>
      )}
      <div style={{ display: "flex", gap: "5px", alignItems: "center" }}>
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
