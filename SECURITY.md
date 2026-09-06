VERDICT: BLOCKED

## Sicherheitsbericht Feature-Flag-Service

Es wurde kein automatisierter Security-Scanner geliefert (`no applicable security scanners for this project type`); die Beurteilung basiert daher ausschließlich auf dem sichtbaren Go-Quellcode. Diese Scanner-Lücke ist dokumentiert, aber kein eigener Befund.

### Zusammenfassung
Positiv fallen auf:

- Keine hartkodierten Secrets, Passwörter, Token oder internen URLs.
- `POST /flags` und `PUT /flags/{key}` begrenzen den Request-Body auf 1 MiB und antworten bei Überschreitung mit 413.
- Fehlerantworten enthalten nur das definierte JSON-Fehlerobjekt, keine Stacktraces, Pfade oder Panic-Texte.
- Die Logging-Middleware protokolliert nur Methode, Pfad ohne Query-String und Statuscode; der `user`-Parameter wird nicht geloggt.
- Der `user`-Parameter wird ausschließlich für den Hash verwendet, nicht gespeichert und nicht zurückgegeben.
- Der In-Memory-Store ist über `sync.RWMutex` korrekt gegen parallele Zugriffe geschützt.
- Es sind keine Drittanbieter-Abhängigkeiten sichtbar; dadurch keine bekannten ausnutzbaren CVEs.

Der Blocker ist das vollständige Fehlen von Authentifizierung/Autorisierung in Verbindung mit einem Listener auf allen Netzwerkschnittstellen.

---

### 1. Hoch — Fehlende Authentifizierung und Autorisierung

**Ort:** `main.go` (`newHandler`, `http.ListenAndServe(":"+port, handler)`), `middleware.go`

**Beschreibung:**  
Die gesamte REST-API ist ohne jede Zugriffskontrolle erreichbar. `http.ListenAndServe(":"+port, handler)` lauscht standardmäßig auf allen IPv4-/IPv6-Interfaces und nicht nur auf `127.0.0.1`. Jeder, der den Port erreicht, kann:

- Flags lesen (`GET /flags`, `GET /flags/{key}`),
- Flags anlegen (`POST /flags`),
- Flags verändern (`PUT /flags/{key}`),
- Flags löschen (`DELETE /flags/{key}`).

Damit kann ein Angreifer Feature-Auslieferungen manipulieren, Rollout-Prozentsätze auf 0 oder 100 setzen, Flags löschen oder den In-Memory-Store mit beliebigen Daten füllen. Das ist eine schwerwiegende Verletzung der Zugriffskontrolle.

**Fix:**  
Vor alle fachlichen Routen — mit Ausnahme von `GET /healthz` — eine Authentifizierungs-/Autorisierungs-Middleware schalten. Konkret z. B.:

- API-Key aus Umgebungsvariable `FLAG_API_KEY` lesen,
- Client-Header `Authorization: Bearer <key>` oder `X-API-Key: <key>` verifizieren,
- Vergleich mit `crypto/subtle.ConstantTimeCompare` durchführen,
- bei fehlendem/ungültigem Key mit `401` antworten.

Alternativ oder ergänzend: Das Produkt nachweislich nur hinter einem authentifizierenden Reverse-Proxy/Service-Mesh betreiben. Zusätzlich sollte der Server explizit an eine konfigurierbare Adresse binden; ein rein lokaler Betrieb kann auf `127.0.0.1` beschränkt werden, sofern keine echten Remote-Clients ohne Proxy zugelassen werden sollen.

---

### 2. Mittel — HTTP-Server ohne Timeouts

**Ort:** `main.go`

**Beschreibung:**  
`http.ListenAndServe` erzeugt intern einen `http.Server` mit `ReadTimeout`, `ReadHeaderTimeout`, `WriteTimeout` und `IdleTimeout` gleich `0`. Dadurch sind langsame Header-/Body-Übertragungen (Slowloris) und lange offene Verbindungen möglich; ein Angreifer kann Serverressourcen erschöpfen.

**Fix:**  
Einen eigenen `http.Server` verwenden:

```go
srv := &http.Server{
    Addr:              ":" + port,
    Handler:           handler,
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       10 * time.Second,
    WriteTimeout:      15 * time.Second,
    IdleTimeout:       60 * time.Second,
    MaxHeaderBytes:    1 << 20,
}
log.Fatal(srv.ListenAndServe())
```

---

### 3. Mittel — Unbegrenztes In-Memory-Wachstum und große Schlüssel

**Ort:** `store.go`, `flags_create.go`

**Beschreibung:**  
Das Body-Limit von 1 MiB begrenzt nur den einzelnen Request. Ein einzelner `key` oder eine `description` kann dennoch fast 1 MiB groß sein; die Anzahl der Flags ist unbegrenzt. Da der Store rein im Speicher liegt, kann ein erreichbarer Angreifer durch viele große oder viele kleine Flags den Speicher des Prozesses füllen.

**Fix:**  
- Maximale Länge für `key` validieren, z. B. 128 oder 256 Bytes/Zeichen.
- Maximale Länge für `description` validieren, z. B. 4096 Zeichen.
- Maximale Anzahl Flags im Store begrenzen, z. B. `maxFlags = 10_000`.
- Optional ein Rate-Limit auf die schreibenden Endpunkte legen.

---

### 4. Niedrig — `POST /flags` akzeptiert zusätzliche Daten nach dem JSON-Objekt

**Ort:** `flags_create.go`

**Beschreibung:**  
Nach dem ersten `Decode` wird im Gegensatz zu `PUT /flags/{key}` nicht geprüft, ob danach noch weitere Nicht-Whitespace-Daten folgen. Ein Body wie `{"key":"a","enabled":true} garbage` wird daher als gültig akzeptiert. Das ist kein direkt ausnutzbarer Angriff, aber eine inkonsistente und zu lockere Eingabevalidierung.

**Fix:**  
Analog zu `handleUpdateFlag` nach dem ersten Decode prüfen:

```go
if err := dec.Decode(&struct{}{}); err != io.EOF {
    writeError(w, http.StatusBadRequest, "invalid JSON")
    return
}
```

---

### 5. Niedrig — Transport nur über HTTP

**Ort:** `main.go`

**Beschreibung:**  
Der Dienst verwendet ausschließlich HTTP. Falls Flag-Beschreibungen, API-Keys oder andere vertrauliche Daten übertragen werden, geschieht dies unverschlüsselt. Wenn der Dienst ohne vorgeschalteten Proxy betrieben wird, ist das ein Risiko für Abhörung und Manipulation.

**Fix:**  
TLS-Terminierung vor dem Dienst erzwingen oder den Server direkt mit `ListenAndServeTLS` betreiben. Bei rein internem Betrieb ohne TLS muss dies bewusst und dokumentiert sein.