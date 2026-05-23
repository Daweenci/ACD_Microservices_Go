# Microservices mit Go – Projektübersicht

## Architektur (4 Container)

```
┌─────────────────────────────────────────────────────────┐
│                   Docker Network                        │
│                                                         │
│  ┌──────────────┐      ┌──────────────────────────┐     │
│  │   auth-db    │◄────►│     auth-service :8081   │     │
│  │ PostgreSQL   │      │  /register  /login       │     │
│  │   :5432      │      │  /validate               │     │
│  └──────────────┘      └──────────┬───────────────┘     │
│                                   │ prüft JWT           │
│  ┌──────────────┐      ┌──────────▼───────────────┐     │
│  │  shop-db     │◄────►│    shop-service :8082    │     │
│  │ PostgreSQL   │      │  GET/POST /shop          │     │
│  │   :5433      │      │  (braucht gültiges JWT)  │     │
│  └──────────────┘      └──────────────────────────┘     │
└─────────────────────────────────────────────────────────┘
```

## Dateistruktur

```
microservices-project/
├── docker-compose.yml
├── auth-service/
│   ├── Dockerfile
│   ├── go.mod
│   └── cmd/
│       └── main.go                  ← Einstiegspunkt + Router
│   └── internal/
│       ├── core/
│       │   ├── user.go              ← Domain-Typen + Interface
│       │   └── auth_service.go      ← Geschäftslogik (Onion-Kern)
│       ├── repository/
│       │   └── postgres_user_repo.go← DB-Implementierung
│       └── handler/
│           └── auth_handler.go      ← HTTP-Handler
└── shop-service/
    ├── Dockerfile
    ├── go.mod
    └── cmd/
        └── main.go
    └── internal/
        ├── core/
        │   ├── shop.go
        │   └── shop_service.go
        ├── repository/
        │   └── postgres_shop_repo.go
        └── handler/
            └── shop_handler.go     ← validiert JWT via auth-service
```

## Architektur-Pattern

Das Projekt folgt der **Onion-Architektur**:

- **`core/`** – Geschäftslogik ohne technische Abhängigkeiten
- **`repository/`** – Datenbankzugriff implementiert das Interface aus core
- **`handler/`** – HTTP-Schicht, kennt nur core
- **`cmd/main.go`** – Startpunkt, Router (identisch zum Artikel-Pattern)

Der Order-Service nutzt das **Proxy-Pattern**: Er kennt die JWT-Logik nicht
selbst, sondern delegiert die Validierung an den Auth-Service (`/validate`).

## Starten

```bash
# Alle 4 Container starten
docker-compose up --build

# Nur im Hintergrund
docker-compose up --build -d
```

## API-Verwendung (curl-Beispiele)

### 1. Registrieren
```bash
curl -X POST http://localhost:8081/register \
  -H "Content-Type: application/json" \
  -d '{"username":"max","email":"max@example.com","password":"geheim123"}'
```

### 2. Einloggen → Token erhalten
```bash
curl -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{"email":"max@example.com","password":"geheim123"}'

# Antwort: {"token":"eyJhbGci...","message":"Login erfolgreich"}
```

### 3. Token speichern
```bash
TOKEN="eyJhbGci..."  # aus der Login-Antwort
```

### 4. Bestellung anlegen (braucht Token)
```bash
curl -X POST http://localhost:8082/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"item":"Laptop","quantity":1}'
```

### 5. Bestellungen abrufen
```bash
curl http://localhost:8082/orders \
  -H "Authorization: Bearer $TOKEN"
```

### 6. Token direkt validieren
```bash
curl http://localhost:8081/validate \
  -H "Authorization: Bearer $TOKEN"
```

## go.sum erzeugen

Da go.sum-Dateien nicht eingecheckt sind, einmalig ausführen:

```bash
cd auth-service && go mod tidy
cd ../order-service && go mod tidy
```

Oder direkt mit Docker bauen – das passiert automatisch im Builder-Stage.
