import { useEffect } from "react";
import styles from "./EditorHelpPopup.module.css";

interface ShortcutEntry {
  command: string;
  shortcut: string;
  context: string;
}

const SHORTCUTS: ShortcutEntry[] = [
  { command: "Sauvegarder la zone", shortcut: "Ctrl+S / Cmd+S", context: "Zone en cours d'édition" },
  { command: "Annuler (undo)", shortcut: "Ctrl+Z / Cmd+Z", context: "Maximum 50 états" },
  { command: "Supprimer polygone(s) sélectionné(s)", shortcut: "Delete / Suppr", context: "Au moins 1 polygone sélectionné" },
  { command: "Confirmer le nom de la zone", shortcut: "Entrée", context: "Champ de saisie du nom" },
  { command: "Valider le renommage", shortcut: "Entrée", context: "Renommage d'une zone" },
  { command: "Annuler le renommage", shortcut: "Échap", context: "Champ de renommage" },
  { command: "Fermer la boîte de dialogue", shortcut: "Échap", context: "Dialogue de nom de zone" },
  { command: "Ajouter un point", shortcut: "Clic gauche", context: "Mode dessin de polygone" },
  { command: "Supprimer le dernier point", shortcut: "Ctrl+Clic sur le canvas", context: "Mode dessin de polygone" },
  { command: "Sélection multiple de polygones", shortcut: "Ctrl+Clic / Cmd+Clic", context: "Sur un polygone" },
  { command: "Supprimer un point", shortcut: "Ctrl+Clic / Cmd+Clic", context: "Sur une poignée de point" },
  { command: "Ajouter/insérer un point", shortcut: "Double-clic", context: "Sur le canvas ou un segment" },
  { command: "Annuler dernière action (menu contextuel)", shortcut: "Clic droit", context: "Mode dessin" },
  { command: "Terminer la forme (menu contextuel)", shortcut: "Clic droit", context: "Mode dessin" },
];

interface EditorHelpPopupProps {
  open: boolean;
  onClose: () => void;
}

export const EditorHelpPopup: React.FC<EditorHelpPopupProps> = ({ open, onClose }) => {
  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [open, onClose]);

  if (!open) return null;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div
        className={`modal-content ${styles.content}`}
        onClick={(e) => e.stopPropagation()}
        role="dialog"
        aria-modal="true"
        aria-label="Raccourcis clavier"
      >
        <h2 className={styles.title}>Raccourcis clavier</h2>
        <table className={styles.table}>
          <thead>
            <tr>
              <th className={styles.th}>Commande</th>
              <th className={styles.th}>Raccourci</th>
              <th className={styles.th}>Contexte</th>
            </tr>
          </thead>
          <tbody>
            {SHORTCUTS.map((entry, index) => (
              <tr key={index}>
                <td className={styles.td}>{entry.command}</td>
                <td className={styles.td}>
                  <kbd className={styles.kbd}>{entry.shortcut}</kbd>
                </td>
                <td className={styles.td}>{entry.context}</td>
              </tr>
            ))}
          </tbody>
        </table>
        <div className={styles.actions}>
          <button className="btn-secondary" onClick={onClose}>
            Fermer
          </button>
        </div>
      </div>
    </div>
  );
};
