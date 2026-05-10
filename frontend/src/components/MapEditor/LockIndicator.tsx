import React from "react";
import styles from "./LockIndicator.module.css";

interface LockIndicatorProps {
  isLocked: boolean;
  lockedByMe: boolean;
  editorName?: string;
  canEdit: boolean;
  remainingSeconds?: number;
  onRefreshLock?: () => void;
  onReacquireLock?: () => void;
  showReacquireButton?: boolean;
}

function formatTime(seconds: number): string {
  const m = Math.floor(seconds / 60);
  const s = seconds % 60;
  return `${String(m).padStart(2, "0")}:${String(s).padStart(2, "0")}`;
}

export const LockIndicator: React.FC<LockIndicatorProps> = ({
  isLocked,
  lockedByMe,
  editorName,
  canEdit,
  remainingSeconds,
  onRefreshLock,
  onReacquireLock,
  showReacquireButton,
}) => {
  const timeStr = remainingSeconds != null ? formatTime(remainingSeconds) : null;

  return (
    <div className={styles.container}>
      {!canEdit && !isLocked && !showReacquireButton && (
        <span className={styles.expiringSoon}>Verrouillage requis pour éditer</span>
      )}

      {isLocked && !lockedByMe && editorName && timeStr && (
        <span className={styles.lockedByOther}>
          Verrouillé par {editorName} — Expire dans {timeStr}
        </span>
      )}

      {isLocked && !lockedByMe && editorName && !timeStr && (
        <span className={styles.lockedByOther}>
          Verrouillé par {editorName}
        </span>
      )}

      {isLocked && lockedByMe && timeStr && remainingSeconds != null && (
        <>
          {remainingSeconds > 60 && (
            <span className={styles.lockedByMe}>
              Édition en cours — Verrou expire dans {timeStr}
            </span>
          )}
          {remainingSeconds > 10 && remainingSeconds <= 60 && (
            <span className={styles.expiringSoon}>
              Édition en cours — Verrou expire dans {timeStr}
            </span>
          )}
          {remainingSeconds <= 10 && (
            <span className={styles.critical}>
              Verrou expire dans {timeStr}
            </span>
          )}
          {onRefreshLock && remainingSeconds < 300 && (
            <button className={styles.extendButton} onClick={onRefreshLock}>
              Prolonger
            </button>
          )}
        </>
      )}

      {isLocked && lockedByMe && !timeStr && (
        <span className={styles.lockedByMe}>Édition en cours</span>
      )}

      {showReacquireButton && !isLocked && (
        <>
          <span className={styles.lost}>Verrou d'édition perdu</span>
          {onReacquireLock && (
            <button className={styles.reacquireButton} onClick={onReacquireLock}>
              Ré-acquérir le verrou
            </button>
          )}
        </>
      )}
    </div>
  );
};
