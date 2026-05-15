# Tâche #003 - Story #023 : Tests et documentation

## Objectif
Ajouter les tests unitaires, d'intégration et E2E pour la popup d'aide, ainsi que la documentation utilisateur.

## Contexte
- Story #023 : `docs/stories/story-023.md`
- Dépend de : Tâche #001 et Tâche #002 (les composants et l'intégration doivent être terminés)
- Nécessaire pour : Aucune

## Prompt

En tant qu'agent de code, tu dois implémenter ce qui suit.

### Description fonctionnelle

**Tests unitaires :**
- Vérifier que la popup rend correctement les 14 entrées de raccourcis
- Vérifier que la popup est masquée quand `open={false}`
- Vérifier que la popup est visible quand `open={true}`
- Vérifier que le bouton Fermer est visible
- Vérifier la structure du tableau (3 colonnes : Commande, Raccourci, Contexte)

**Tests E2E :**
- Naviguer vers l'éditeur → cliquer sur le bouton "?" → popup affichée → raccourcis visibles → fermer avec Échap
- Vérifier que la popup se ferme en cliquant sur l'overlay (en dehors du contenu)

### Fichiers concernés

| Fichier | Action | Description |
|---------|--------|-------------|
| `frontend/src/components/MapEditor/EditorHelpPopup.test.tsx` | Créer | Tests unitaires du composant |
| `frontend/e2e/editor/help-popup.spec.ts` | Créer | Tests E2E de la popup d'aide |
| `docs/editor-help.md` | Créer | Documentation utilisateur |

### Tests unitaires

#### Fichier : `frontend/src/components/MapEditor/EditorHelpPopup.test.tsx`

```typescript
import { describe, it, expect } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { EditorHelpPopup } from './EditorHelpPopup';

describe('EditorHelpPopup', () => {
  it('ne rend rien quand open est false', () => {
    const { container } = render(<EditorHelpPopup open={false} onClose={() => {}} />);
    expect(container.innerHTML).toBe('');
  });

  it('rend le titre et le tableau quand open est true', () => {
    render(<EditorHelpPopup open={true} onClose={() => {}} />);
    expect(screen.getByText(/Raccourcis clavier/i)).toBeTruthy();
    expect(screen.getByText(/Commande/i)).toBeTruthy();
    expect(screen.getByText(/Raccourci/i)).toBeTruthy();
    expect(screen.getByText(/Contexte/i)).toBeTruthy();
  });

  it('affiche toutes les entrées de raccourcis', () => {
    render(<EditorHelpPopup open={true} onClose={() => {}} />);
    expect(screen.getByText(/Sauvegarder la zone/i)).toBeTruthy();
    expect(screen.getByText(/Ctrl\+S/i)).toBeTruthy();
    expect(screen.getByText(/Annuler \(undo\)/i)).toBeTruthy();
    expect(screen.getByText(/Delete \/ Suppr/i)).toBeTruthy();
    expect(screen.getByText(/Clic gauche/i)).toBeTruthy();
    expect(screen.getByText(/Clic droit/i)).toBeTruthy();
    expect(screen.getByText(/Double-clic/i)).toBeTruthy();
  });

  it('appelle onClose quand le bouton Fermer est cliqué', () => {
    let closed = false;
    render(<EditorHelpPopup open={true} onClose={() => { closed = true; }} />);
    const closeBtn = screen.getByRole('button', { name: /fermer/i });
    fireEvent.click(closeBtn);
    expect(closed).toBe(true);
  });

  it('appelle onClose quand Escape est pressé', () => {
    let closed = false;
    render(<EditorHelpPopup open={true} onClose={() => { closed = true; }} />);
    fireEvent.keyDown(window, { key: 'Escape' });
    expect(closed).toBe(true);
  });

  it('appelle onClose quand on clique sur l\'overlay', () => {
    let closed = false;
    const { container } = render(<EditorHelpPopup open={true} onClose={() => { closed = true; }} />);
    // L'overlay est le premier enfant direct avec role dialog ou la div d'overlay
    const overlay = container.querySelector('[class*="overlay"]') || container.firstChild;
    if (overlay) {
      fireEvent.click(overlay);
      expect(closed).toBe(true);
    }
  });
});
```

### Tests E2E

#### Fichier : `frontend/e2e/editor/help-popup.spec.ts`

```typescript
import { test, expect } from '../fixtures';
import { loginAsUser, generateJWT } from '../helpers/auth';
import { seedTeam, seedMap, queryDB, createZoneWithLock } from '../helpers/db';

test.describe('Feature: Editor Help Popup', () => {
  test('1.1: Cliquer sur le bouton ? ouvre la popup d\'aide', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Help Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Help-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);
    const zone = await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Test Zone',
      [{ points: [{ x: 10, y: 10 }, { x: 50, y: 10 }, { x: 30, y: 50 }] }],
    );

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${team.id}/editor/${map.id}`);

    await page.waitForSelector('[data-testid="help-btn"]');
    await page.click('[data-testid="help-btn"]');

    await expect(page.getByText(/Raccourcis clavier/i)).toBeVisible();

    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('1.2: La popup liste les commandes et leurs raccourcis', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Help Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Help-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);
    const zone = await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Test Zone',
      [{ points: [{ x: 10, y: 10 }, { x: 50, y: 10 }, { x: 30, y: 50 }] }],
    );

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${team.id}/editor/${map.id}`);

    await page.click('[data-testid="help-btn"]');

    await expect(page.getByText(/Sauvegarder la zone/i)).toBeVisible();
    await expect(page.getByText(/Annuler \(undo\)/i)).toBeVisible();
    await expect(page.getByText(/Supprimer polygone/i)).toBeVisible();
    await expect(page.getByText(/Ajouter un point/i)).toBeVisible();

    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });

  test('1.3: La popup se ferme avec Échap', async ({ page, adminUser }) => {
    const team = await seedTeam('E2E Help Team ' + Date.now(), adminUser.id);
    const map = await seedMap('E2E-Help-Map-' + Date.now());
    const secret = process.env.JWT_SECRET || 'test-secret-key-for-manual-validation-32chars';
    const token = await generateJWT(adminUser.id, secret);
    const zone = await createZoneWithLock(
      process.env.E2E_API_URL || 'http://localhost:8080',
      team.id, map.id, token,
      'Test Zone',
      [{ points: [{ x: 10, y: 10 }, { x: 50, y: 10 }, { x: 30, y: 50 }] }],
    );

    await page.goto('/login');
    await loginAsUser(page, adminUser);
    await page.goto(`/teams/${team.id}/editor/${map.id}`);

    await page.click('[data-testid="help-btn"]');
    await expect(page.getByText(/Raccourcis clavier/i)).toBeVisible();

    await page.keyboard.press('Escape');
    await expect(page.getByText(/Raccourcis clavier/i)).not.toBeVisible();

    await queryDB('DELETE FROM zones WHERE map_id = $1', [map.id]);
    await queryDB('DELETE FROM maps WHERE id = $1', [map.id]);
    await queryDB('DELETE FROM team_members WHERE team_id = $1', [team.id]);
    await queryDB('DELETE FROM teams WHERE id = $1', [team.id]);
  });
});

// Note au développeur :
// - Utiliser `seedTeam` + `seedMap` (SQL, pas d'API) + `generateJWT` + `createZoneWithLock` (API)
// - Le helper `createZoneWithLock` acquiert un verrou et crée une zone via API
// - Le bouton "?" utilise `data-testid="help-btn"` (CSS Modules masque les classes réelles)
// - Utiliser `adminUser` plutot que `regularUser` car le helper JWT nécessite un userId existant en base
```

### Documentation

#### Fichier : `docs/editor-help.md`

```markdown
# Aide de l'éditeur de zones

L'éditeur de zones polygonales permet de créer et modifier des zones d'entraînement sur les cartes.

## Ouvrir l'aide

- Cliquer sur le bouton **?** dans la barre d'outils
- Appuyer sur **F1** ou **Ctrl+? / Cmd+?** au clavier

## Raccourcis clavier

| Commande | Raccourci | Contexte |
|----------|-----------|----------|
| Sauvegarder la zone | Ctrl+S / Cmd+S | Zone en cours d'édition |
| Annuler (undo) | Ctrl+Z / Cmd+Z | Maximum 50 états |
| Supprimer polygone(s) sélectionné(s) | Delete / Suppr | Au moins 1 polygone sélectionné |
| Confirmer le nom de la zone | Entrée | Champ de saisie du nom |
| Valider le renommage | Entrée | Renommage d'une zone |
| Annuler le renommage | Échap | Champ de renommage |
| Fermer la boîte de dialogue | Échap | Dialogue de nom de zone |
| Ajouter un point | Clic gauche sur le canvas | Mode dessin de polygone |
| Supprimer le dernier point | Ctrl+Clic sur le canvas | Mode dessin de polygone |
| Sélection multiple de polygones | Ctrl+Clic / Cmd+Clic | Sur un polygone |
| Supprimer un point | Ctrl+Clic / Cmd+Clic | Sur une poignée de point |
| Ajouter/insérer un point | Double-clic | Sur le canvas ou un segment |
| Menu contextuel | Clic droit | Mode dessin |
| Terminer la forme (menu contextuel) | Clic droit | Mode dessin |

## Fermer l'aide

- Appuyer sur **Échap**
- Cliquer en dehors de la popup
- Cliquer sur le bouton **Fermer**
```

### Contraintes techniques

- **Framework tests** : Vitest + React Testing Library pour les tests unitaires
- **Framework E2E** : Playwright (config existante dans le projet)
- **Fixtures** : Utiliser `adminUser` des fixtures existantes (`frontend/e2e/fixtures.ts`) car `seedTeam` nécessite un `captain_id` existant en base
- **Helpers** : Utiliser `seedTeam`, `seedMap`, `createZoneWithLock`, `generateJWT` de `frontend/e2e/helpers/db.ts` et `frontend/e2e/helpers/auth.ts`
- **Sélecteur CSS** : Ne pas utiliser de classe CSS (les CSS Modules hashent les noms). Utiliser `data-testid="help-btn"` sur le bouton dans `Toolbar.tsx`
- **Exécution** : Lancer les tests unitaires avec `cd frontend && npx vitest run` ; tests E2E avec `cd frontend && npm run test:e2e`
- **Rédaction** : Créer la documentation dans `docs/editor-help.md` — reprendre exactement le tableau de la story #023

### Exemples d'API (non applicable)
