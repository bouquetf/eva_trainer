# Paramètres utilisateur

## Endpoints

### GET /api/v1/auth/settings
Retourne les préférences de l'utilisateur connecté.
Champs non définis → valeurs par défaut.

### PUT /api/v1/auth/settings
Met à jour les préférences. Body partiel accepté.

## Structure UserSettings

| Champ | Type | Défaut | Description |
|-------|------|--------|-------------|
| `revision_correct_timer` | number | 3 | Temps d'affichage après bonne réponse (s) |
| `revision_incorrect_timer` | number | 5 | Temps d'affichage après mauvaise réponse (s) |
| `revision_question_count` | number | 20 | Nombre de questions par session |

## Interface utilisateur
Disponible dans `/settings` → section "Révision".
