# Workflow d'implémentation des User Stories

Ce document définit l'ordre recommandé pour l'implémentation des user stories, en identifiant les dépendances et les opportunités de parallélisation.

## Légende
- 🔴 : Blocant (doit être fait avant)
- 🟡 : Peut être fait en parallèle
- 🟢 : Dépendance levée, peut commencer

---

## Phase 0 : Initialisation (Sprint 0 - 1 semaine)
*Mise en place de l'ossature du projet*

| Story | Titre | Dépendances | Tests E2E |
|-------|-------|-------------|------------|
| #000 | Project Initialization and Scaffolding | Aucune | - |

**Livrables :**
- Structure Go + Gin créée
- Frontend React + Vite + librairies installées
- docker-compose.yml fonctionnel
- Schéma de base de données initial créé
- Endpoint de santé `/health` opérationnel

---

## Phase 1 : Fondations (Sprint 1)
*Mise en place de l'infrastructure et de l'authentification de base*

### Parallélisation possible : 2 équipes/branches

| Story | Titre | Dépendances | Tests E2E |
|-------|-------|-------------|------------|
| #001 | Configuration and Deployment | #000 | - |
| #002 | User Authentication via Discord | #000 | Playwright |

**Livrables :**
- Environnement de développement configuré (Docker Compose)
- Authentification Discord fonctionnelle
- JWT + Refresh token avec rotation

---

## Phase 2 : Gestion des équipes - Base (Sprint 2)
*Création d'équipe et rôles de base*

| Story | Titre | Dépendances | Tests E2E |
|-------|-------|-------------|------------|
| #003 | Team Creation and Management | #002 (Auth) | Playwright |
| #004 | Team Roles and Permissions | #003 (Team) | Playwright |
| #005 | Map Management (Admin) | #002 (Auth) | Playwright |

**Notes :** 
- #003 et #005 peuvent démarrer en parallèles après #002
- #004 dépend de #003
- #002 dépend de #000 (initialisation)

---

## Phase 3 : Fonctionnalités équipe avancées (Sprint 3)
*Invitations, Splits, Sécurité API, Zones de base*

### Parallélisation possible : 4 équipes/branches

| Story | Titre | Dépendances | Tests E2E |
|-------|-------|-------------|------------|
| #006 | Team Invitation and Join System | #003, #004 | Playwright |
| #007 | Split Management and Filtering | #005 (Maps) | Playwright |
| #008 | API Security and Role Verification | #002, #004 | Playwright |
| #009 | Zone Creation and Editing with Polygons | #005, #004 | Playwright |

**Notes :**
- Ces 4 stories sont indépendantes entre elles une fois les dépendances levées
- Recommandé : 4 développeurs/branches différentes

---

## Phase 4 : Édition avancée et Publication (Sprint 4)
*Verrouillage, Templates, Historique, Gestion d'erreurs*

### Parallélisation possible : 4 équipes/branches

| Story | Titre | Dépendances | Tests E2E |
|-------|-------|-------------|------------|
| #010 | Map Locking Mechanism | #009 (Zones) | Playwright |
| #011 | Zone Publishing and Template System | #009, #004 | Playwright |
| #012 | Edit Mode with History and Rollback | #009, #004 | Playwright |
| #013 | Error Handling and Logging | #008 (API Security) | - |

**Notes :**
- #010, #011, #012 peuvent être développés en parallèle
- #013 peut être intégré progressivement dans les autres stories

---

## Phase 5 : Révision et Templates (Sprint 5)
*Duplication de templates et mode révision de base*

### Parallélisation possible : 2 équipes/branches

| Story | Titre | Dépendances | Tests E2E |
|-------|-------|-------------|------------|
| #014 | Template Duplication by Other Teams | #011 (Publish) | Playwright |
| #015 | Revision Mode with Validation | #009, #011 | Playwright |

**Notes :**
- #015 est le cœur de l'application, tester exhaustivement
- #014 complète le système de partage entre équipes

---

## Phase 6 : Révision avancée (Sprint 6)
*Algorithme de poids pour l'entraînement optimal*

| Story | Titre | Dépendances | Tests E2E |
|-------|-------|-------------|------------|
| #016 | Revision Mode with Weighted Probability | #015 (Revision) | Playwright |

**Notes :**
- Dernière story fonctionnelle
- Tests de charge recommandés pour valider l'algorithme de poids

---

## Phase 7 : Finalisation des tests (Sprint 7)
*Couverture de tests à 100%*

| Story | Titre | Dépendances | Tests E2E |
|-------|-------|-------------|------------|
| #017 | Test Automation Strategy | Toutes les stories | Playwright |

**Notes :**
- Cette story est transversale et doit être intégrée tout au long
- Sprint dédié pour atteindre les 100% de couverture fonctionnelle

---

## Résumé du graphe de dépendances

```mermaid
graph TD
    #000[Story #000<br/>Project Initialization]
    #001[Story #001<br/>Configuration and Deployment]
    #002[Story #002<br/>User Authentication via Discord]
    
    #000 --> #001
    #000 --> #002
    
    #001 --> #002
    
    #002 --> #003[Story #003<br/>Team Creation and Management]
    #002 --> #005[Story #005<br/>Map Management Admin]
    #002 --> #008[Story #008<br/>API Security and Role Verification]
    
    #003 --> #004[Story #004<br/>Team Roles and Permissions]
    #004 --> #006[Story #006<br/>Team Invitation and Join System]
    
    #005 --> #007[Story #007<br/>Split Management]
    #005 --> #009[Story #009<br/>Zone Creation and Editing]
    
    #009 --> #010[Story #010<br/>Map Locking Mechanism]
    #009 --> #011[Story #011<br/>Zone Publishing and Template System]
    #009 --> #012[Story #012<br/>Edit Mode with History and Rollback]
    
    #011 --> #014[Story #014<br/>Template Duplication]
    #011 --> #015[Story #015<br/>Revision Mode with Validation]
    
    #015 --> #016[Story #016<br/>Revision Mode with Weighted Probability]
    
    #008 --> #013[Story #013<br/>Error Handling and Logging]
    
    #004 -.-> #012
    #004 -.-> #011
    
    #006 -.-> #015
    #007 -.-> #015
    
    #000 --> #017[Story #017<br/>Test Automation Strategy]
    #002 --> #017
    #003 -.-> #017
    #004 -.-> #017
    #006 -.-> #017
    #007 -.-> #017
    #008 -.-> #017
    #009 -.-> #017
    #010 -.-> #017
    #011 -.-> #017
    #012 -.-> #017
    #013 -.-> #017
    #014 -.-> #017
    #015 -.-> #017
    #016 -.-> #017
    
    style #000 fill:#90EE90
    style #001 fill:#90EE90
    style #002 fill:#90EE90
    style #017 fill:#FFB6C1
```

---

## Planning recommandé (8 Sprints : 0 à 7)

| Sprint | Stories | Équipe recommandée | Jalons |
|--------|---------|-------------------|--------|
| S0 | #000 | 1-2 devs | Project scaffolded |
| S1 | #001, #002 | 2 devs | Auth OK, Env OK |
| S2 | #003, #004, #005 | 3 devs | Teams OK, Maps OK |
| S3 | #006, #007, #008, #009 | 4 devs | Invitations, Splits, Zones |
| S4 | #010, #011, #012, #013 | 4 devs | Lock, Publish, History |
| S5 | #014, #015 | 2 devs | Templates, Revision |
| S6 | #016 | 1 dev | Weighted Revision |
| S7 | #017 + Tests | 2 devs | 100% Coverage |

**Total : 16 semaines (environ 4 mois)**

### Diagramme de Gantt (Mermaid)

```mermaid
gantt
    title Planning d'implémentation - 8 Sprints
    dateFormat  D
    axisFormat  Semaine %d
    
    section Phase 0 - Initialisation
    #000 Project Init                       :done, s0, 0, 10d
    
    section Phase 1 - Fondations
    #001 Configuration & Deployment        :done, s1a, after s0, 10d
    #002 Authentification Discord          :done, s1b, after s0, 10d
    
    section Phase 2 - Équipes base
    #003 Team Creation                     :active, s2a, after s1b, 10d
    #004 Team Roles                        :s2b, after s2a, 10d
    #005 Map Management                    :s2c, after s1b, 10d
    
    section Phase 3 - Fonctionnalités avancées
    #006 Invitation System                 :s3a, after s2b, 10d
    #007 Split Management                  :s3b, after s2c, 10d
    #008 API Security                      :s3c, after s1b, 10d
    #009 Zone Creation                     :s3d, after s2c, 10d
    
    section Phase 4 - Édition & Publication
    #010 Map Locking                       :s4a, after s3d, 10d
    #011 Zone Publishing                   :s4b, after s3d, 10d
    #012 Edit History & Rollback           :s4c, after s3d, 10d
    #013 Error Handling                    :s4d, after s3c, 10d
    
    section Phase 5 - Révision & Templates
    #014 Template Duplication              :s5a, after s4b, 10d
    #015 Revision Mode                     :s5b, after s4b, 10d
    
    section Phase 6 - Révision avancée
    #016 Weighted Probability              :s6, after s5b, 10d
    
    section Phase 7 - Tests
    #017 Test Automation                   :s7, after s6, 10d
```

### Parallélisation maximale

```mermaid
graph LR
    subgraph "Sprint 3 - Parallélisation max (4 équipes)"
        A[#006 Invitation]
        B[#007 Splits]
        C[#008 API Security]
        D[#009 Zones]
    end
    
    subgraph "Sprint 4 - Parallélisation max (4 équipes)"
        E[#010 Locking]
        F[#011 Publishing]
        G[#012 History]
        H[#013 Error Handling]
    end
    
    A --> E
    B --> F
    C --> H
    D --> F
    D --> G
    D --> E
    
    style A fill:#FFD700
    style B fill:#FFD700
    style C fill:#FFD700
    style D fill:#FFD700
    style E fill:#87CEEB
    style F fill:#87CEEB
    style G fill:#87CEEB
    style H fill:#87CEEB
```

---

## Notes importantes

0. **Story #000 - Ossature du projet** : Cette story est CRUCIALE et doit être terminée avant toute autre. Elle créée la structure de base (Go/Gin, React/Vite, Docker, PostgreSQL) sans laquelle les stories #001 et #002 ne peuvent pas être développées. Prévoyez une revue de code stricte sur cette ossature.

1. **Tests continus** : Story #017 ne doit pas attendre la fin. Chaque story doit inclure ses tests unitaires et d'intégration dès le Sprint 1.

2. **Démo après chaque Sprint** : Prévoir une démonstration fonctionnelle après chaque phase.

3. **Code Reviews** : Les stories en parallèle nécessitent des reviews croisées pour éviter les conflits de merge.

4. **Base de données** : Le schéma (`docs/specification/database_schema.md`) doit être validé avant le Sprint 2.

5. **Framework E2E** : Playwright doit être configuré dès le Sprint 0 (Story #000) pour être utilisé dès la Story #002.
