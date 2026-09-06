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

```sh
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

| Variable | Bedeutung           | Default |
|----------|---------------------|---------|
| `PORT`   | HTTP-Port des Servers | `8080` |
