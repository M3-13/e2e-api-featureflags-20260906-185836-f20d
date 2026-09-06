# Feature-Flag-Service

Ein Feature-Flag-Service als REST-API in Go. Flags werden angelegt, aufgelistet, gelesen, geändert und gelöscht und pro Nutzer deterministisch ausgewertet. Die Daten liegen in einem thread-sicheren In-Memory-Store; Persistenz über einen Neustart hinaus gibt es nicht.

## Tech Stack

- **Language**: Go (1.22+)
- **Runtime**: ausschließlich `net/http` aus der Standardbibliothek (kein Framework)
- **Build**: `go build`
- **Tests**: `go test`

## Installation

Voraussetzung ist eine Go-Installation (>= 1.22). Es gibt keine externen Abhängigkeiten — alle Imports kommen aus der Standardbibliothek, ein `go mod tidy` ändert nichts.

## Start (Dev)

Alle `/flags`-Routen sind geschützt und verlangen `FLAG_API_KEY`; ohne gesetzten
Schlüssel antworten sie mit `503`. Erzeuge vor dem Start einen Schlüssel und
exportiere ihn (die CI-Pipeline rollt denselben Wert über `RUN.json`):

```sh
export FLAG_API_KEY="$(openssl rand -hex 32)"
go run .
```

Unter Windows (PowerShell):

```powershell
$env:FLAG_API_KEY = -join ((1..32) | ForEach-Object { '{0:x2}' -f (Get-Random -Max 256) })
go run .
```

Der Server bindet an den Port aus der Umgebungsvariable `PORT`; ist sie nicht gesetzt, wird `8080` verwendet:

```sh
PORT=9090 go run .
```

## Build (Produktion)

```sh
go build -o featureflags .
./featureflags
```

## Endpunkte

| Methode | Pfad                    | Beschreibung                                   |
|---------|-------------------------|------------------------------------------------|
| GET     | `/healthz`              | Health-Check, antwortet `200 {"status":"ok"}`  |
| POST    | `/flags`                | Flag anlegen                                    |
| GET     | `/flags`                | Alle Flags auflisten                            |
| GET     | `/flags/{key}`          | Einzelnes Flag lesen                            |
| PUT     | `/flags/{key}`          | Flag ändern (nur angegebene Felder)            |
| DELETE  | `/flags/{key}`          | Flag löschen                                    |
| GET     | `/flags/{key}/evaluate` | Flag deterministisch pro Nutzer auswerten       |

## Features

- Thread-sicherer In-Memory-Store mit `sync.RWMutex`
- Health-Endpoint (`GET /healthz`)
- Route-Registrierung mit Methoden-Patterns (Go 1.22 `ServeMux`)
- Fehlerantworten ausschließlich als `{"error":"..."}` ohne interne Details

## Konfiguration

| Variable       | Bedeutung                                                              | Default      |
|----------------|------------------------------------------------------------------------|--------------|
| `PORT`         | HTTP-Port des Servers                                                  | `8080`       |
| `HOST`         | Bind-Adresse des Servers                                               | `127.0.0.1`  |
| `FLAG_API_KEY` | API-Schlüssel für den Zugriff auf alle `/flags`-Routen (siehe unten)   | *(keiner)*   |

## Authentifizierung

Alle `/flags`-Routen (GET/POST/PUT/DELETE/evaluate) sind geschützt. Der Schlüssel
wird über `FLAG_API_KEY` bereitgestellt und pro Anfrage auf eine der beiden Arten
übermittelt:

- `Authorization: Bearer <key>`
- `X-API-Key: <key>`

Ist `FLAG_API_KEY` nicht gesetzt, antworten alle `/flags`-Routen mit `503`
(`{"error":"authentication not configured"}`); `GET /healthz` bleibt dann weiter
ohne Schlüssel erreichbar. Ist ein Schlüssel gesetzt, antworten die Routen bei
fehlendem oder ungültigem Schlüssel mit `401`. Der Vergleich erfolgt in
konstanter Zeit (`crypto/subtle`).

## Betrieb & TLS

Der Dienst spricht ausschließlich klares HTTP und terminiert selbst **kein TLS**.
In Produktion ist er deshalb **verbindlich ausschließlich hinter einer
TLS-terminierenden Komponente** (Reverse-Proxy oder Load Balancer) zu betreiben,
die `https` nach außen anbietet und intern an diesen Dienst weiterleitet.

Direktes, unverschlüsseltes HTTP (`http://`) ist **nur lokal** (z. B. `127.0.0.1`
während der Entwicklung) zulässig. Der Server bindet standardmäßig an
`127.0.0.1:8080` und setzt Lese-/Schreib-/Header-Timeouts, um langsamen und
hängen bleibenden Verbindungen vorzubeugen.
