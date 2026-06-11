---
name: feedback-fr
description: Génération de stories à partir de feedbacks utilisateurs, en filtrant ceux déjà traités
---

# Skill: Feedback → Stories / Spécification (FR)

Ce guide décrit comment transformer des feedbacks utilisateurs en user stories exploitables, en ne traitant que les retours les plus récents qui n'ont pas encore été pris en compte dans les stories existantes.

## Qu'est-ce que je fais

Je prends le fichier `docs/specification/feedbacks.md`, j'identifie les feedbacks qui n'ont pas encore été traités. Je **regroupe d'abord les feedbacks par thème** — plusieurs feedbacks sur un même sujet donnent une seule unité de travail. Pour chaque groupe :
- **Si le sujet est simple** → je génère une ou plusieurs stories directement (au même format que story-fr) avec le processus itératif de questions/réponses.
- **Si le sujet est complexe** (nouvelle fonctionnalité transverse, multiple sous-sujets, impact architectural) → je crée une sous-spécification dans `docs/specification/features/`, je l'affine via le processus specification-fr, puis je la découpe en stories via story-fr.

**Règle :** plusieurs feedbacks liés → un seul regroupement → une seule story (si simple) ou une seule sous-spécification (si complexe). Pas de doublon.

## Quand m'utiliser

Utilisez cette skill lorsque :
- Vous avez collecté des retours utilisateurs dans `docs/specification/feedbacks.md`
- Vous voulez transformer ces retours en stories structurées et testables
- Vous voulez éviter de retraiter des feedbacks déjà pris en compte

## Prérequis

Le fichier `docs/specification/feedbacks.md` doit être structuré comme suit :

```markdown
## 2025-05-14

* Feedback 1
* Feedback 2
...

## [DATE]

* Feedback N
...
```

Chaque feedback traité reçoit un marqueur `[PROCESSED: story-XXX]` :

```markdown
* Feedback 1 [PROCESSED: story-018]
```

## Workflow

### Étape 1 - Lecture du fichier de feedbacks

- Lire `docs/specification/feedbacks.md`
- Repérer les feedbacks NON marqués `[PROCESSED: story-XXX]` ou `[PROCESSED: feature-XXX.md]` (ce sont ceux à traiter)
- Si TOUS les feedbacks sont marqués : annoncer qu'il n'y a rien de nouveau à traiter → fin du workflow

### Étape 2 - Analyse des stories existantes

- Lire les stories dans `docs/stories/` (les stories terminées sont suffixées `-done`, ex: `story-031-done.md`)
- Vérifier si certains feedbacks non marqués sont déjà couverts par des stories existantes
  - Exemple : "Avoir un écran avec la liste des zones" peut déjà exister dans Story #009
  - Si c'est le cas, ajouter le marqueur `[PROCESSED: story-XXX]` (donc déjà couvert) et ne pas recréer de story
- Les feedbacks restants (ni marqués PROCESSED, ni couverts) sont à transformer en nouvelles stories

### Étape 3 - Regroupement et décision : story directe ou sous-spécification

- Regrouper les feedbacks par thème commun (ex : "éditeur", "révision", "UI/UX", "archives")
- Un feedback seul peut devenir une story s'il est suffisamment riche
- Plusieurs feedbacks liés peuvent être regroupés dans une même story
- **Décider du niveau de complexité** de chaque groupe (ou feedback seul) :
  - **Sujet simple** → aller en **Étape 4a** (génération directe de stories)
  - **Sujet complexe** → aller en **Étape 4b** (création d'une sous-spécification)
- Critères indicatifs de complexité (un seul suffit) :
  - Plusieurs rôles ou permissions distincts
  - Impacts multiples sur des fonctionnalités existantes
  - Sous-sujets nombreux et indépendants
  - Changement architectural ou nouvelle table en base
  - Nécessite plusieurs stories pour couvrir le périmètre

---

### Chemin A — Sujet simple : Génération directe de stories

#### Étape 4a - Génération des stories

- Lire les stories dans `docs/stories/` (suffixées `-done` ou non) pour déterminer le prochain numéro de story
- Créer les nouvelles stories dans `docs/stories/` en suivant le format story-fr :

```markdown
# Story #XXX : [Titre]

## Description
En tant que [rôle], je veux [action] afin de [bénéfice].

## Origine
Feedback du [DATE] : "[texte du feedback]"

## Critères d'acceptation
- [ ] Critère 1
- [ ] Critère 2

## Tests automatisés
- Test unitaire : [Description]
- Test d'intégration : [Description]
- Test E2E : [Description]

## Documentation
- [Nouvelle documentation à créer]

## Valeur utilisateur
[Description de la valeur probante pour l'utilisateur final]

## Dépendances
- [Story #YYY si applicable]
```

- Ajouter une section `## Origine` pour tracer le feedback source
- Marquer chaque feedback traité dans `docs/specification/feedbacks.md` avec `[PROCESSED: story-XXX]`
- Exemple : `* Ne pas afficher d'accueil [PROCESSED: story-018]`

#### Étape 5a - Création du fichier de questions

- Créer `docs/stories/questions.md` (ou s'il existe déjà, le compléter)
- Contient les questions de clarification pour les nouvelles stories générées
- Même format que story-fr (contexte, question, propositions)

#### Étape 6a - Analyse et clarification

- Analyser les nouvelles stories
- Vérifier la cohérence avec la spécification `docs/specification/specification.md`
- Vérifier la cohérence avec les stories existantes dans `docs/stories/` (y compris les `*-done.md`)
- Ajouter des questions dans `docs/stories/questions.md` si nécessaire

#### Étape 7a - Mise à jour

- L'utilisateur répond aux questions dans `docs/stories/questions.md`
- L'utilisateur tape : `J'ai répondu aux questions, mets à jour les stories`
- Mettre à jour les stories selon les réponses
- Retourner à l'Étape 6a

---

### Chemin B — Sujet complexe : Sous-spécification puis découpage en stories

#### Étape 4b - Création de la sous-spécification

- Créer le répertoire `docs/specification/features/` s'il n'existe pas
- Créer le fichier `docs/specification/features/feedback-XXX.md` (XXX = nom court du sujet, ex: `feedback-archives-matchs.md`)
- Copier le texte du feedback dans le fichier comme base de spécification
- Ajouter la section `## Origine` avec la date et le texte du feedback source
- Structurer le fichier avec des placeholders `[À_CLARIFIER: ...]` sur les points flous
- Créer `docs/specification/features/questions.md` avec :
  - Le contexte (description du feedback)
  - Les questions de clarification (ouvertes, avec propositions)
  - Les inconsistances détectées
- **Marquer le feedback** dans `docs/specification/feedbacks.md` avec le marqueur `[PROCESSED: feature-XXX.md]`

#### Étape 5b - Itération de spécification

- L'utilisateur répond aux questions dans `docs/specification/features/questions.md`
- L'utilisateur tape : `J'ai répondu aux questions, continue`
- Lire les réponses, mettre à jour `feedback-XXX.md` en remplaçant les placeholders par les réponses
- Résoudre les inconsistances identifiées
- Vérifier s'il reste des `[À_CLARIFIER: ...]` ou `[INCONSISTANCE: ...]` :
  - **S'il en reste** → générer un nouveau `questions.md` avec les questions restantes → retour au début de l'Étape 5b
  - **Sinon** → supprimer `questions.md` → passer à l'Étape 6b

#### Étape 6b - Découpage en stories (story-fr)

- La sous-spécification est complète (plus de placeholders)
- Exécuter le workflow story-fr sur `docs/specification/features/feedback-XXX.md` :
  - Lire la spécification
  - Découper en user stories dans `docs/stories/` (chacune avec la section `## Origine` pointant vers `feature-XXX.md`)
  - Créer `docs/stories/questions.md` si des points restent flous
  - Appliquer le cycle itératif de questions/réponses propre à story-fr
- Une fois les stories finalisées (plus de questions) :
  - Mettre à jour `feedbacks.md` si des sous-feedbacks doivent être marqués individuellement
  - Annoncer les stories créées

---

### Étape 8 - Fin du workflow

- Quand il n'y a plus rien à clarifier (pour tous les chemins)
- Supprimer le fichier `questions.md` s'il existe
- Les nouvelles stories restent dans `docs/stories/` ; elles seront renommées en `story-XXX-done.md` par la skill `tasks` après découpage en tâches
- Mettre à jour `AGENTS.md` et `workflow.md` si nécessaire
- Annoncer la fin du workflow

## Format du fichier feedbacks.md

```markdown
## [DATE]

* [Texte du feedback] [PROCESSED: story-XXX]
* [Texte du feedback] [PROCESSED: feature-XXX.md]
* [Texte du feedback]

## [AUTRE DATE]

* [Texte du feedback] [PROCESSED: story-XXX]
```

Les feedbacks marqués `[PROCESSED: story-XXX]` ou `[PROCESSED: feature-XXX.md]` sont ignorés.
Les feedbacks sans marqueur sont à traiter.

## Exemple

### Fichier feedbacks.md initial
```markdown
## 2025-05-14

* Ne pas afficher d'accueil : on tombe direct sur "Mes équipes"
* Afficher un message d'erreur lorsque l'on tente d'ouvrir l'éditeur et que quelqu'un d'autre a le lock [PROCESSED: story-010]
* Popup d'aide sur l'éditeur : affiche la liste de commandes et les raccourcis clavier associés
```

### Stories générées
- `docs/stories/story-018.md` : "Suppression de la page d'accueil" (basé sur feedback #1)
- `docs/stories/story-019.md` : "Popup d'aide dans l'éditeur" (basé sur feedback #3)

### Fichier feedbacks.md après traitement
```markdown
## 2025-05-14

* Ne pas afficher d'accueil : on tombe direct sur "Mes équipes" [PROCESSED: story-018]
* Afficher un message d'erreur lorsque l'on tente d'ouvrir l'éditeur et que quelqu'un d'autre a le lock [PROCESSED: story-010]
* Popup d'aide sur l'éditeur : affiche la liste de commandes et les raccourcis clavier associés [PROCESSED: story-019]
```

## Comment passer d'une étape à l'autre

### Chemin A — Sujet simple

| Transition | Action de l'utilisateur | Action de l'agent |
|------------|------------------------|-------------------|
| Étape 1 → 2 | Donner le chemin du fichier de feedbacks en argument | Lit les feedbacks, analyse les stories existantes |
| Étape 2 → 3 | Automatique | Regroupe les feedbacks non traités par thème et décide simple/complexe |
| Étape 3 → 4a | Automatique (sujet simple) | Génère les nouvelles stories dans docs/stories/ |
| Étape 4a → 5a | Automatique après génération | Crée questions.md avec les points à clarifier |
| Étape 5a → 6a | Automatique | Analyse les stories et ajoute les questions |
| Étape 6a → 7a | Taper "J'ai répondu aux questions, mets à jour les stories" | Met à jour les stories selon les réponses |
| Étape 7a → 6a | Automatique après mise à jour | Re-vérifie la cohérence |
| Étape 7a → 8 | Automatique (plus de questions) | Supprime questions.md, marque les feedbacks traités, annonce la fin |

### Chemin B — Sujet complexe

| Transition | Action de l'utilisateur | Action de l'agent |
|------------|------------------------|-------------------|
| Étape 1 → 2 | Donner le chemin du fichier de feedbacks en argument | Lit les feedbacks, analyse les stories existantes |
| Étape 2 → 3 | Automatique | Regroupe les feedbacks non traités par thème et décide simple/complexe |
| Étape 3 → 4b | Automatique (sujet complexe) | Crée docs/specification/features/feedback-XXX.md + questions.md |
| Étape 4b → 5b | Automatique | Attend les réponses de l'utilisateur |
| Étape 5b → 5b | Taper "J'ai répondu aux questions, continue" | Met à jour la spec, vérifie les placeholders restants |
| Étape 5b → 6b | Automatique (plus de placeholders) | Supprime questions.md, exécute story-fr sur la sous-spécification |
| Étape 6b → 8 | Automatique (stories finalisées) | Annonce les stories créées |

**Important :** L'agent ne surveille pas les modifications de fichiers en temps réel. L'utilisateur doit toujours donner une instruction explicite pour que l'agent passe à l'étape suivante.

## Exemples d'utilisation

### Cas simple — Génération directe de stories

```
/ feedback feedbacks.md
```

L'agent lit `docs/specification/feedbacks.md`, identifie les feedbacks non marqués, vérifie les stories existantes, et génère les nouvelles stories avec le processus itératif de clarification.

### Cas complexe — Sous-spécification puis stories

```
/ feedback feedbacks.md
```

L'agent lit `docs/specification/feedbacks.md`, détecte qu'un feedback (ex: "Archives des matchs") est trop complexe pour une story unique, puis :

1. Crée `docs/specification/features/feedback-archives-matchs.md`
2. Génère `docs/specification/features/questions.md` avec les points à clarifier (rôles, relations entre données, etc.)
3. Itère avec l'utilisateur jusqu'à spécification complète
4. Supprime `questions.md`
5. Exécute story-fr sur la spécification pour la découper en plusieurs stories
6. Marque le feedback original comme `[PROCESSED: feature-archives-matchs.md]`

### Gestion des doublons avec les stories existantes

Si un feedback non marqué est déjà couvert par une story existante :

1. L'agent l'identifie lors de l'Étape 2
2. Il ajoute le marqueur `[PROCESSED: story-XXX]` dans le fichier feedbacks.md
3. Il n'en génère pas de nouvelle story
4. Il peut notifier l'utilisateur du doublon dans questions.md
