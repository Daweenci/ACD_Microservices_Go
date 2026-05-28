# Microservices mit Go und Docker

Auth- und Warenkorb-Service als Docker-Compose-Setup.

## Architektur (4 Container)

```
┌─────────────────────────────────────────────────────────────┐
│                      Docker Network                         │
│                                                             │
│  ┌──────────────┐      ┌────────────────────────────────┐   │
│  │   auth-db    │◄────►│       auth-service :8081       │   │
│  │  PostgreSQL  │      │  /register  /login  /validate  │   │
│  │    :5432     │      │  /user  /user/username /health │   │
│  └──────────────┘      └────────────────────────────────┘   │
│                                                             │
│  ┌──────────────┐      ┌────────────────────────────────┐   │
│  │   cart-db    │◄────►│       cart-service :8082       │   │
│  │  PostgreSQL  │      │  /cart  /cart/item/{id} /health│   │
│  │    :5433     │      │  (JWT lokal validiert)         │   │
│  └──────────────┘      └────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
```

Der Cart-Service validiert JWTs **lokal** anhand des öffentlichen EC-Schlüssels (`JWT_PUBLIC_KEY`) – ohne Netzwerkaufruf an den Auth-Service. Der Auth-Service signiert Token mit dem privaten EC-Schlüssel (`JWT_PRIVATE_KEY`). Kein geteiltes Secret – der Cart-Service kann Token verifizieren, aber nicht selbst ausstellen.

## Dateistruktur

```
microservices-go/
├── docker-compose.yml
├── .env                          ← nicht eingecheckt
├── .env.example                  ← Vorlage
├── auth-service/
│   ├── Dockerfile
│   ├── go.mod
│   └── cmd/
│       └── main.go               ← Einstiegspunkt + Router
│   └── internal/
│       ├── core/
│       │   ├── user.go           ← Domain-Typen + Interface
│       │   ├── auth_service.go   ← Geschäftslogik
│       │   └── jwt.go            ← Token ausstellen + validieren (private key)
│       ├── repository/
│       │   └── postgres_user_repo.go
│       └── handler/
│           └── auth_handler.go
└── cart-service/
    ├── Dockerfile
    ├── go.mod
    └── cmd/
        └── main.go
    └── internal/
        ├── core/
        │   ├── cart.go           ← Domain-Typen + Interface
        │   ├── cart_service.go   ← Geschäftslogik
        │   └── jwt.go            ← lokale Token-Validierung (public key)
        ├── repository/
        │   └── postgres_cart_repo.go
        └── handler/
            └── cart_handler.go
```

## Setup

### 1. Schlüsselpaar generieren

```bash
openssl ecparam -name prime256v1 -genkey -noout -out private.pem
openssl ec -in private.pem -pubout -out public.pem
```

Ausgabe für die `.env` als einzeiligen String konvertieren:

```bash
awk 'NF {sub(/\r/, ""); printf "%s\\n",$0;}' private.pem
awk 'NF {sub(/\r/, ""); printf "%s\\n",$0;}' public.pem
```

### 2. `.env` anlegen

```bash
cp .env.example .env
```

`JWT_PRIVATE_KEY` und `JWT_PUBLIC_KEY` mit den generierten Schlüsseln befüllen. Die `private.pem` und `public.pem` Dateien können danach gelöscht werden.

### 3. Starten

```bash
docker-compose up --build
```

### 4. Im Hintergrund

```bash
docker-compose up --build -d
docker-compose down
```

## API

### Auth-Service (Port 8081)

#### Registrieren

```bash
curl -X POST http://localhost:8081/register \
  -H "Content-Type: application/json" \
  -d '{"username":"max","email":"max@example.com","password":"geheim123"}'
```

#### Einloggen

```bash
curl -X POST http://localhost:8081/login \
  -H "Content-Type: application/json" \
  -d '{"identifier":"max","password":"geheim123"}'

# Antwort: {"token":"eyJhbGci...","message":"login erfolgreich"}
```

`identifier` kann E-Mail-Adresse oder Benutzername sein.

#### Token speichern

```bash
TOKEN="eyJhbGci..."
```

#### Benutzernamen ändern

```bash
curl -X PATCH http://localhost:8081/user/username \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"new_username":"maxneu"}'
```

#### Benutzer löschen

```bash
curl -X DELETE http://localhost:8081/user \
  -H "Authorization: Bearer $TOKEN"
```

#### Token validieren (Debug)

```bash
curl http://localhost:8081/validate \
  -H "Authorization: Bearer $TOKEN"
```

#### Health-Check

```bash
curl http://localhost:8081/health
# Antwort: {"status":"ok"}
```

---

### Cart-Service (Port 8082)

Alle Endpunkte erfordern einen gültigen JWT im `Authorization: Bearer`-Header.

#### Warenkorb abrufen

```bash
curl http://localhost:8082/cart \
  -H "Authorization: Bearer $TOKEN"
```

#### Item hinzufügen

```bash
curl -X POST http://localhost:8082/cart \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"item_id":42,"quantity":2}'
```

Existiert das Item bereits, wird die Menge erhöht.

#### Menge anpassen (Delta)

```bash
# +1 hinzufügen
curl -X PATCH http://localhost:8082/cart/item/42 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"delta":1}'

# 1 entfernen
curl -X PATCH http://localhost:8082/cart/item/42 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"delta":-1}'
```

Fällt die Menge auf 0 oder darunter, wird der Eintrag automatisch gelöscht.

#### Item komplett entfernen

```bash
curl -X DELETE http://localhost:8082/cart/item/42 \
  -H "Authorization: Bearer $TOKEN"
```

#### Warenkorb leeren

```bash
curl -X DELETE http://localhost:8082/cart \
  -H "Authorization: Bearer $TOKEN"
```

#### Health-Check

```bash
curl http://localhost:8082/health
# Antwort: {"status":"ok"}
```

## go.sum erzeugen

Falls nötig (passiert automatisch beim Docker-Build):

```bash
cd auth-service && go mod tidy
cd ../cart-service && go mod tidy
```