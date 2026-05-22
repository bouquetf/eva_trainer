import React, { useState, useEffect, useCallback } from "react";
import { getPublishStatus, publishMap, unpublishMap, type PublishStatus } from "../../api/templates";
import { listMyTeams } from "../../api/teams";
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
  const [actionLoading, setActionLoading] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);
  const [successMsg, setSuccessMsg] = useState<string | null>(null);

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

  const refreshStatus = useCallback(async () => {
    try {
      const ps = await getPublishStatus(teamId, mapId);
      setStatus(ps);
    } catch {
      // ignore
    }
  }, [teamId, mapId]);

  const handlePublish = useCallback(async () => {
    setShowConfirm(false);
    setActionLoading(true);
    setSuccessMsg(null);
    try {
      const result = await publishMap(teamId, mapId);
      if (result.created) {
        setSuccessMsg("Carte publiée avec succès !");
      } else {
        setSuccessMsg("Le template a été remplacé avec les zones actuelles.");
      }
      await refreshStatus();
    } catch {
      alert("Erreur lors de la publication.");
    } finally {
      setActionLoading(false);
    }
  }, [teamId, mapId, refreshStatus]);

  const handleUnpublish = useCallback(async () => {
    setShowConfirm(false);
    setActionLoading(true);
    setSuccessMsg(null);
    try {
      await unpublishMap(teamId, mapId);
      setSuccessMsg("Carte rendue privée. Le template n'est plus synchronisé.");
      await refreshStatus();
    } catch {
      alert("Erreur lors du passage en privé.");
    } finally {
      setActionLoading(false);
    }
  }, [teamId, mapId, refreshStatus]);

  const isPublished = status?.published && status?.template?.sync_active === true;
  const isUnpublished = status?.published && status?.template?.sync_active === false;
  const notPublished = !status?.published;

  const canPublish = role === "captain" || role === "cocaptain";

  const confirmTitle = isPublished
    ? "Rendre la carte privée ?"
    : "Rendre la carte publique ?";

  const confirmMessage = isPublished
    ? "Le template restera accessible en lecture par les autres équipes, mais les modifications ne seront plus synchronisées."
    : notPublished
      ? "Toutes les zones de cette carte deviendront publiques et pourront être dupliquées par d'autres équipes. Vous pourrez continuer à modifier vos zones en privé après publication."
      : "Le template existant va être remplacé par les zones actuelles. Cette action est irréversible.";

  if (loading) return null;

  if (isPublished) {
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
        {canPublish && (
          <button
            className={styles.publishButton}
            onClick={() => setShowConfirm(true)}
            disabled={actionLoading}
            aria-label="Rendre la carte privée"
          >
            {actionLoading ? "..." : "Rendre privée"}
          </button>
        )}
        {successMsg && (
          <span className="success-message" style={{ fontSize: "12px" }}>{successMsg}</span>
        )}
        <ConfirmDialog
          open={showConfirm}
          title={confirmTitle}
          message={confirmMessage}
          confirmLabel="Rendre privée"
          cancelLabel="Annuler"
          danger
          onConfirm={handleUnpublish}
          onCancel={() => setShowConfirm(false)}
        />
      </div>
    );
  }

  if (!canPublish) {
    return (
      <span className={styles.badgePrivate} aria-label="Carte privée">
        Privée
      </span>
    );
  }

  return (
    <>
      <div style={{ display: "inline-flex", alignItems: "center", gap: "6px" }}>
        {isUnpublished && (
          <span className={styles.badgePrivate} aria-label="Carte privée">
            Privée
          </span>
        )}
        {notPublished && (
          <span className={styles.badgePrivate} aria-label="Carte privée">
            Privée
          </span>
        )}
        <button
          className={styles.publishButton}
          onClick={() => setShowConfirm(true)}
          disabled={actionLoading}
          aria-label="Rendre la carte publique"
        >
          {actionLoading ? "..." : "Rendre publique"}
        </button>
        {successMsg && (
          <span className="success-message" style={{ fontSize: "12px" }}>{successMsg}</span>
        )}
      </div>
      <ConfirmDialog
        open={showConfirm}
        title={confirmTitle}
        message={confirmMessage}
        confirmLabel="Rendre publique"
        cancelLabel="Annuler"
        danger
        onConfirm={handlePublish}
        onCancel={() => setShowConfirm(false)}
      />
    </>
  );
};
