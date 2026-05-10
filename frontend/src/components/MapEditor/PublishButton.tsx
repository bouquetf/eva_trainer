import React, { useState, useEffect, useCallback } from "react";
import { getPublishStatus, publishMap, type PublishStatus } from "../../api/templates";
import { listMyTeams } from "../../api/teams";
import { ApiError } from "../../api/client";
import ConfirmDialog from "../ConfirmDialog";
import styles from "./PublishButton.module.css";

interface PublishButtonProps {
  teamId: number;
  mapId: number;
}

export const PublishButton: React.FC<PublishButtonProps> = ({ teamId, mapId }) => {
  const [status, setStatus] = useState<PublishStatus | null>(null);
  const [role, setRole] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [publishing, setPublishing] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);

  useEffect(() => {
    let cancelled = false;
    const fetch = async () => {
      try {
        const ps = await getPublishStatus(teamId, mapId);
        if (!cancelled) setStatus(ps);
      } catch {
        // ignore
      }
      try {
        const teams = await listMyTeams();
        if (!cancelled) {
          const team = teams.find((t) => t.team.id === teamId);
          if (team) setRole(team.my_role);
        }
      } catch {
        // ignore
      }
      if (!cancelled) setLoading(false);
    };
    fetch();
    return () => { cancelled = true; };
  }, [teamId, mapId]);

  const handlePublish = useCallback(async () => {
    setShowConfirm(false);
    setPublishing(true);
    try {
      const result = await publishMap(teamId, mapId);
      setStatus({ published: true, template: result });
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        const ps = await getPublishStatus(teamId, mapId);
        setStatus(ps);
      } else {
        alert("Erreur lors de la publication.");
      }
    } finally {
      setPublishing(false);
    }
  }, [teamId, mapId]);

  if (loading) return null;

  if (status?.published) {
    const date = status.template?.created_at
      ? new Date(status.template.created_at).toLocaleDateString()
      : null;
    return (
      <div style={{ display: "inline-flex", alignItems: "center", gap: "6px" }}>
        <span className={styles.badgePublic} aria-label="Carte publique">
          Publique
        </span>
        {date && (
          <span style={{ fontSize: "11px", color: "#999" }}>
            depuis le {date}
          </span>
        )}
      </div>
    );
  }

  const canPublish = role === "captain" || role === "cocaptain";

  if (!canPublish) {
    return (
      <span className={styles.badgePrivate} aria-label="Carte privée">
        Privée
      </span>
    );
  }

  return (
    <>
      <button
        className={styles.publishButton}
        onClick={() => setShowConfirm(true)}
        disabled={publishing}
        aria-label="Rendre la carte publique"
      >
        {publishing ? "Publication..." : "Rendre publique"}
      </button>
      <ConfirmDialog
        open={showConfirm}
        title="Rendre la carte publique ?"
        message="Cette action est irréversible. Toutes les zones de cette carte deviendront publiques et pourront être dupliquées par d'autres équipes. Vous pourrez continuer à modifier vos zones en privé après publication."
        confirmLabel="Rendre publique"
        cancelLabel="Annuler"
        danger
        onConfirm={handlePublish}
        onCancel={() => setShowConfirm(false)}
      />
    </>
  );
};
