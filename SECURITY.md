VERDICT: APPROVED

## Zusammenfassung

Der Feature-Flag-Service ist insgesamt sauber umgesetzt. Es gibt keine hartkodierten Secrets, keine klassischen Injection-Schwachstellen (kein SQL/Command/Pfad-Zugriff), keine unsichere Deserialisierung und keine externen Abhängigkeiten. Die Authentifizierung über einen API-Key aus der Umgebung ist konstantzeit implementiert, und der Logging-Filter entfernt Query-Strings.  
Die nachfolgenden Befunde sind niedriger Schwere und betreffen Härtungsdetails; ausnutzbare hoch/kritische Schwachstellen wurden nicht gefunden.

## Scanner-Abdeckung

Für diesen Projekttyp (`go-backend`) wurden keine Security-Scanner ausgeführt. Diese Lücke wird zur Kenntnis genommen, begründet aber keinen Befund, da der Code manuell geprüft wurde. Externe Abhängigkeiten sind laut `go.mod` nicht vorhanden; damit entfallen typische Dependency-Check-Ergebnisse.

## Befunde

### 1. Falscher Statuscode bei zu großem Body in der Trailing-Data-Prüfung (niedrig)

- **Datei/Stelle:** `flags_create.go` zweiter `dec.Decode`, `flags_update.go` zweiter `dec.Decode`
- **Beschreibung:** `http.MaxBytesReader` begrenzt das Einlesen. Tritt der `*http.MaxBytesError` erst beim zweiten `Decode`-Aufruf auf (z. B. erstes gültiges JSON, danach sehr großer Rest), fällt der Code in den allgemeinen 400-Zweig statt 413 zu liefern. Der Speicherschutz greift weiterhin, aber die Vorgabe AC-11 wird in diesem Randfall verletzt.
- **Fix:** Die Fehlerprüfung auch im zweiten Decodier-Aufruf durchführen:

```go
if err := dec.Decode(&struct{}{}); err != io.EOF {
    var maxErr *http.MaxBytesError
    if errors.As(err, &maxErr) {
        writeError(w, http.StatusRequestEntityTooLarge, "request body too large")
        return
    }
    writeError(w, http.StatusBadRequest, "invalid JSON")
    return
}
```

Analog für `handleUpdateFlag`.

### 2. API-Key-Längen-Leak durch frühen Längenvergleich (niedrig)

- **Datei/Stelle:** `auth.go`, Funktion `constantTimeEqual`
- **Beschreibung:** `if len(a) != len(b) { return false }` bricht vor dem konstantzeitigen Vergleich ab. Dadurch kann ein Angreifer über wiederholte Zeitmessungen die Länge des API-Keys ermitteln. Der eigentliche Zeichenvergleich ist konstantzeit, die Länge selbst wird jedoch preisgegeben.
- **Fix:** Beide Werte vor dem Vergleich hashen, damit immer gleich lange Digests verglichen werden:

```go
func constantTimeEqual(a, b string) bool {
    ha := sha256.Sum256([]byte(a))
    hb := sha256.Sum256([]byte(b))
    return subtle.ConstantTimeCompare(ha[:], hb[:]) == 1
}
```

### 3. Unverschlüsselter Transport, falls Host auf nicht-loopback gesetzt wird (niedrig)

- **Datei/Stelle:** `main.go`, Host-/Port-Konfiguration und `ListenAndServe`
- **Beschreibung:** Standardmäßig bindet der Dienst an `127.0.0.1`, was sicher ist. Wird `HOST=0.0.0.0` gesetzt, ist der Dienst ohne TLS im Netz erreichbar; API-Key und Flag-Daten würden im Klartext übertragen.
- **Fix:** Standard-Host `127.0.0.1` belassen und in der Betriebsdokumentation festlegen, dass vor öffentlicher Exposition ein TLS-Terminierungs-Proxy vorgeschaltet werden muss. Alternativ natives TLS (`http.ListenAndServeTLS`) unterstützen. Optional im Code warnen, wenn `HOST != 127.0.0.1` und kein TLS konfiguriert ist.

### 4. Fehlende Rate-Limitierung für Authentifizierungsfehlversuche (niedrig)

- **Datei/Stelle:** `auth.go` / `middleware.go`
- **Beschreibung:** Fehlgeschlagene API-Key-Authentifizierungen (401) sind unbegrenzt möglich. Ein Angreifer kann dadurch den statischen API-Key unbeschränkt raten.
- **Fix:** Einfache Rate-Limitierung pro Client-IP in die Middleware integrieren, z. B. Token-Bucket: nach 5 Fehlversuchen pro Minute `429 Too Many Requests` zurückgeben. Der Zustand muss thread-sicher sein und nur bei Fehlversuchen zählen.

### 5. Konkreter Flag-Key in Logs bei nicht authentifizierten Anfragen (niedrig)

- **Datei/Stelle:** `main.go` (Middleware-Reihenfolge) / `middleware.go` (Fallback `r.URL.Path`)
- **Beschreibung:** Da `logMiddleware` die äußerste Schicht ist, wird bei einem 401 aus `authMiddleware` der `http.NewServeMux` nie erreicht. `r.Pattern` bleibt daher leer, und der Logger fällt auf `r.URL.Path` zurück. Dadurch werden z. B. `/flags/alice-experimental/evaluate` im Klartext geloggt, während authentifizierte Anfragen das Pattern `/flags/{key}/evaluate` loggen. Der `user`-Query-Parameter wird weiterhin nicht geloggt, aber Flag-Keys können in Logs sichtbar werden.
- **Fix:** Eine Pfad-Maskierungsfunktion im Logger einsetzen, die dynamische Segmente ersetzt, z. B.:

```go
func maskPath(path string) string {
    parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
    if len(parts) == 2 && parts[0] == "flags" {
        return "/flags/{key}"
    }
    if len(parts) == 3 && parts[0] == "flags" && parts[2] == "evaluate" {
        return "/flags/{key}/evaluate"
    }
    return path
}
```

Im `logMiddleware` dann:

```go
path := r.Pattern
if path == "" {
    path = maskPath(r.URL.Path)
}
```

---

**Fazit:** Keine hoch/kritisch ausnutzbaren Schwachstellen erkennbar. Die genannten Punkte sind Härtungsempfehlungen, die das Produkt sicherer machen, aber keinen sofortigen Auslieferungsstopp erfordern.