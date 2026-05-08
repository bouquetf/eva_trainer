# Configuration

## Principe de chargement

Le package [`pkg/config`](../pkg/config/config.go) charge la configuration au démarrage de l'application :

1. **Fichier `.env`** : chargé via [godotenv](https://github.com/joho/godotenv) s'il existe à la racine du projet
2. **Variables système** : les variables déjà définies dans l'environnement ne sont PAS écrasées par `.env` (les vars système ont priorité)
3. **Valeurs par défaut** : chaque variable optionnelle a une valeur par défaut documentée ci-dessous
4. **Validation** : les champs obligatoires sont vérifiés immédiatement ; le serveur refuse de démarrer avec un message listant les champs manquants

### Ordre de priorité

```
Variables système (priorité haute)
        ↓
Fichier .env (priorité moyenne)
        ↓
Valeurs par défaut (priorité basse)
```

## Variables obligatoires

| Variable | Type | Description |
|----------|------|-------------|
| `JWT_SECRET` | string | Clé secrète pour signer les tokens JWT (min. 32 caractères recommandé) |
| `DISCORD_CLIENT_ID` | string | Client ID de l'application Discord OAuth |
| `DISCORD_CLIENT_SECRET` | string | Client Secret de l'application Discord OAuth |

Si une ou plusieurs de ces variables sont absentes ou vides au démarrage, le serveur affiche une erreur et s'arrête :

```
Configuration error: missing required configuration: JWT_SECRET, DISCORD_CLIENT_ID, DISCORD_CLIENT_SECRET
```

## Variables optionnelles

| Variable | Type | Défaut | Description |
|----------|------|--------|-------------|
| `PORT` | string | `8080` | Port d'écoute du serveur HTTP |
| `GIN_MODE` | string | `debug` | Mode Gin (`debug` ou `release`) |
| `CORS_ORIGINS` | string | `*` | Origines CORS autorisées (séparées par des virgules, `*` pour toutes) |
| `JWT_EXPIRY` | duration | `24h` | Durée de validité des tokens JWT ([format Go](https://pkg.go.dev/time#ParseDuration)) |
| `REFRESH_TOKEN_EXPIRY` | duration | `168h` (7j) | Durée de validité des refresh tokens |
| `DB_HOST` | string | `localhost` | Hôte PostgreSQL |
| `DB_PORT` | string | `5432` | Port PostgreSQL |
| `DB_USER` | string | `eva_user` | Utilisateur PostgreSQL |
| `DB_PASSWORD` | string | `eva_pass` | Mot de passe PostgreSQL |
| `DB_NAME` | string | `eva_trainer` | Nom de la base de données |
| `DB_SSLMODE` | string | `disable` | Mode SSL PostgreSQL |
| `DB_TIMEOUT` | int | `10` | Timeout de connexion en secondes |
| `DATABASE_URL` | string | — | URL de connexion PostgreSQL. Si définie, écrase toutes les vars `DB_*` individuelles |

### Notes

- **`DATABASE_URL`** : si cette variable est définie, les variables `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME` et `DB_SSLMODE` sont ignorées. Format : `postgres://user:pass@host:port/db?sslmode=disable`
- **`GIN_MODE=release`** : active le mode production. Le binaire Go sert alors les fichiers statiques du frontend compilé (`/assets/*`), la page d'accueil (`/`), et le fallback SPA pour toute route non-API. En mode `debug`, le frontend est servi par Vite (port 5173).
- **`JWT_EXPIRY`** et **`REFRESH_TOKEN_EXPIRY`** : utilisent le format de durée Go (`24h`, `7d`, `60m`, `30s`). En cas d'erreur de parsing, la valeur par défaut est utilisée.

## Fichier `.env`

Copier le fichier `.env.example` à la racine du projet :

```bash
cp .env.example .env
```

### Exemple complet

```env
# Server
PORT=8080
GIN_MODE=debug

# Database (individual)
DB_HOST=localhost
DB_PORT=5432
DB_USER=eva_user
DB_PASSWORD=eva_pass
DB_NAME=eva_trainer
DB_SSLMODE=disable
DB_TIMEOUT=10

# Database (URL - prioritaire sur les vars individuelles)
DATABASE_URL=postgres://eva_user:eva_pass@localhost:5432/eva_trainer?sslmode=disable

# CORS
CORS_ORIGINS=*

# Discord OAuth
DISCORD_CLIENT_ID=
DISCORD_CLIENT_SECRET=

# JWT
JWT_SECRET=change-me-in-production
JWT_EXPIRY=24h
REFRESH_TOKEN_EXPIRY=7d
```

## Validation au démarrage

Le serveur appelle la méthode `Validate()` du package `config` après le chargement. Si des champs obligatoires manquent, le processus s'arrête immédiatement avec un message explicite :

```
Configuration error: missing required configuration: JWT_SECRET, DISCORD_CLIENT_ID, DISCORD_CLIENT_SECRET
```

Aucune variable optionnelle invalide ne bloque le démarrage (les valeurs par défaut sont utilisées en cas d'erreur de parsing).

## Exemples

### Développement

```env
PORT=8080
GIN_MODE=debug
CORS_ORIGINS=*
DISCORD_CLIENT_ID=my_dev_id
DISCORD_CLIENT_SECRET=my_dev_secret
JWT_SECRET=dev-secret-key-32-chars-minimum!!
```

### Production

```env
PORT=8080
GIN_MODE=release
CORS_ORIGINS=https://mon-domaine.com
DISCORD_CLIENT_ID=prod_id
DISCORD_CLIENT_SECRET=prod_secret
JWT_SECRET=une-clé-très-longue-et-sécurisée-32-caractères-minimum
DATABASE_URL=postgres://user:password@host:5432/db?sslmode=require
```
