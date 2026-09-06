VERDICT: CHANGES_REQUESTED

## Bericht

Der Service ist grundsätzlich sauber umgesetzt: Body-Limit, kein Query-Logging, keine Speicherung des `user`-Parameters, thread-sicherer Store und definierte JSON-Fehlerobjekte sind vorhanden. Es bestehen jedoch behebbare Lücken bei Transportverschlüsselung, Zugriffsschutz und Nachweisdokumentation.

### 1. DSGVO (GDPR)

**1.1 Transport des `user`-Parameters ohne TLS**  
- **Severity:** high  
- **Befund:** `GET /flags/{key}/evaluate?user=alice` überträgt den `user`-Parameter im Query-String. In `main.go` wird ausschließlich `http.ListenAndServe(":"+port, handler)` verwendet — keine TLS-Konfiguration, keine sichtbare Vorgabe für vorgelagertes TLS. Damit kann der Personenbezug im Klartext über das Netzwerk abfließen.  
- **Remedy:**  
  - In `main.go` entweder `http.ListenAndServeTLS` mit Zertifikaten implementieren oder in `README.md` verbindlich dokumentieren, dass der Dienst ausschließlich hinter einer TLS-terminierenden Komponente (Reverse-Proxy/Load Balancer) betrieben werden darf.  
  - Zusätzlich empfohlen: `user` in einen Request-Header oder den Request-Body verlagern, damit der Identifier nicht in Query-String-Logs vorgelagerter Systeme landet.  
  - Dadurch funktioniert die API weiterhin korrekt; nur der Aufrufvertrag für `evaluate` ändert sich.

**1.2 Fehlende sichtbare Rechtsgrundlage / Datenschutzhinweis**  
- **Severity:** medium  
- **Befund:** Der `user`-Parameter ist ein personenbezogenes Datum. Aus dem sichtbaren Stand geht keine Rechtsgrundlage nach Art. 6 DSGVO und kein Verarbeitungsnachweis hervor.  
- **Remedy:** In `README.md` (oder einer separaten `PRIVACY.md`) einen Abschnitt „Datenschutz“ ergänzen: Verarbeitung des `user`-Identifikators ausschließlich transient zur SHA-256-basierten Rollout-Berechnung, keine Speicherung, keine Ausgabe. Rechtsgrundlage des verantwortlichen Betreibers nennen, z. B. Art. 6 Abs. 1 lit. b/f DSGVO.

**1.3 Logging**  
- **Severity:** low  
- **Befund:** `middleware.go` loggt `r.URL.Path` einschließlich des Flag-Keys. Ein Flag-Key kann in Einzelfällen personenbezogen sein (z. B. `alice-experimental`).  
- **Remedy:** In `middleware.go` nur die registrierte Pfadroute ohne dynamisches Segment loggen oder dynamische Segmente auf eine whitelist-basierte Maske setzen. Die bisherige Umsetzung entspricht bereits AC-13/14, daher gering.

### 2. EU Cyber Resilience Act (CRA)

**2.1 Fehlende Authentifizierung/Authorization**  
- **Severity:** high  
- **Befund:** `newHandler` in `main.go` registriert alle `/flags`-Routen ohne jeden Zugriffsschutz. Jede Partei mit Netzwerkzugriff kann Flags anlegen, ändern oder löschen. Das widerspricht „security by design/default“.  
- **Remedy:** In `main.go` eine Middleware ergänzen, die für `POST /flags`, `PUT /flags/{key}` und `DELETE /flags/{key}` ein Bearer-Token/API-Key prüft; `GET /healthz` ausnehmen. Token aus der Umgebung beziehen, nicht im Code hardcoden. Die produktiven Flows bleiben funktionsfähig; Health-Checks bleiben unauthentifiziert erreichbar.

**2.2 Transportverschlüsselung (siehe 1.1)**  
- **Severity:** high  
- **Befund:** Ungesicherter HTTP-Transport ist zugleich ein CRA-Sicherheitsmangel.  
- **Remedy:** Wie unter 1.1 beschrieben.

**2.3 Fehlende SBOM / Update-Policy / Sicherheitsdokumentation**  
- **Severity:** medium  
- **Befund:** Im sichtbaren Stand fehlt eine maschinenlesbare SBOM und eine dokumentierte Patch-/Update-Richtlinie. Die `go.mod` ist vorhanden, aber ein reines „keine Drittanbieter-Dependencies“-Statement reicht für den CRA-Nachweis in der Regel nicht aus.  
- **Remedy:**  
  - `go list -deps -json` oder ein SBOM-Tool (CycloneDX/SPDX) in die CI aufnehmen und das Ergebnis als Artefakt im Repo bereitstellen.  
  - `SECURITY.md` oder README-Abschnitt ergänzen: Sicherheitseigenschaften, Patch-Intervall, Zuständigkeit für Updates. Da derzeit nur die Go-Standardbibliothek genutzt wird, ist der Aufwand gering, aber der Nachweis muss sichtbar sein.

**2.4 Strict Input Validation (POST /flags)**  
- **Severity:** medium  
- **Befund:** `flags_create.go` dekodiert nur den ersten JSON-Wert und prüft nicht, ob danach noch Daten stehen. Dadurch kann ein Body wie `{"key":"a","enabled":true} garbage` fälschlich akzeptiert werden. `flags_update.go` hat diese Prüfung bereits.  
- **Remedy:** In `flags_create.go` analog zu `flags_update.go` nach dem ersten `Decode` prüfen:  
  `if err := dec.Decode(&struct{}{}); err != io.EOF { writeError(w, http.StatusBadRequest, "invalid JSON"); return }`  
  und einen Test für trailing data ergänzen.

**2.5 Weitere robustheitsbezogene Punkte**  
- **Severity:** low  
- **Befund:** In `flags_update.go` besteht zwischen `store.Get` und `store.Update` ein kleines Race-Fenster bei gleichzeitigem Löschen.  
- **Remedy:** Eine atomare Store-Methode mit Existenzprüfung unter einem einzigen Lock ergänzen oder den Update-Pfad im Store so gestalten, dass er den alten Zustand erwartet (compare-and-swap). Für die aktuelle AC-Abdeckung nicht kritisch, aber für CRA-Robustheit empfehlenswert.

### 3. EU AI Act

Nicht anwendbar. Das Produkt enthält keine KI-Funktion.

### 4. Pflichttexte & UI

Nicht anwendbar. Reines Backend ohne Endbenutzer-UI. Keine Legal-Notice-, Cookie- oder Widerrufstext-Pflichten.

### 5. Barrierefreiheit

Nicht anwendbar. Kein öffentliches Web-UI.

---

**Fazit:** Keine fundamentalen Verstöße, daher keine Sperrung. Die wesentlichen Lücken sind Transportverschlüsselung, Zugriffsschutz und fehlende Dokumentations-/Nachweisartefakte. Nach Umsetzung der High/Medium-Findings ist eine erneute Prüfung mit hoher Wahrscheinlichkeit `APPROVED`.