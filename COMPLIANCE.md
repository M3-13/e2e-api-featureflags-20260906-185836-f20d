VERDICT: APPROVED

# Prüfbericht – Feature-Flag-Service REST-API in Go (`go-backend`)

Prüfgegenstand ist der vollständig zusammengeführte Produktstand (Quellcode + Spec). Der Projekttyp ist ein reiner Backend-Dienst ohne Endbenutzer-Web-UI. Daher sind Pflichttexte-/Cookie-/Consent-Pflichten sowie Barrierefreiheitspflichten für eine öffentliche Web-UI nicht einschlägig.

## 1. DSGVO

### Verarbeitete personenbezogene Daten
- Der Query-Parameter `user` in `GET /flags/{key}/evaluate?user=...` ist potenziell personenbezogen, sofern er eine natürliche Person identifiziert.
- Der Wert wird ausschließlich in `evaluate.go` als Eingabe für den SHA-256-Hash verwendet.
- Es erfolgt keine Speicherung, keine Protokollierung und keine Rückgabe des `user`-Werts in der Antwort.

### Logging
- `middleware.go` protokolliert ausschließlich Methode, gemusterten Pfad bzw. Pfad ohne Query-String und Statuscode.
- Konkrete Flag-Keys werden bei Mustern wie `GET /flags/{key}` maskiert.
- Der `user`-Parameter und andere Query-Werte erscheinen nachweislich nicht im Log.
- Der API-Key wird ebenfalls nicht geloggt.

### Bewertung
Die Grundsätze der Datenminimierung und Vertraulichkeit nach Art. 5 und Art. 32 DSGVO sind im Code umgesetzt. Es sind keine kritischen DSGVO-Verstöße erkennbar.

### Hinweise

- **F1 – low – Transportverschlüsselung**  
  `main.go` startet den Server ohne TLS (`httpServer.ListenAndServe()`). Der Standard-Host ist `127.0.0.1`, wodurch der Datenverkehr lokal bleibt. Wird der Dienst öffentlich betrieben, würde der `user`-Parameter unverschlüsselt übertragen.  
  **Abhilfe:** In `SECURITY.md`/`README.md` festschreiben, dass Produktion ausschließlich hinter einer TLS-Terminierung (Reverse Proxy) erfolgt. Optional in `main.go` TLS-Zertifikate aus Umgebungsvariablen unterstützen.

- **F2 – low – Rechtsgrundlage/AVV dokumentieren**  
  Die transiente Verarbeitung des `user`-Parameters bedarf im Betrieb einer klaren Rechtsgrundlage, z. B. Art. 6 Abs. 1 lit. b oder f DSGVO. Diese ist im Code nicht sichtbar und muss organisatorisch dokumentiert sein.  
  **Abhilfe:** In `PRIVACY.md` einen Abschnitt „Feature-Flag-Auswertung“ ergänzen: Zweck, Rechtsgrundlage, Speicherdauer „keine“, Empfänger, ggf. Hinweis auf einen Auftragsverarbeitungsvertrag.

## 2. EU Cyber Resilience Act (CRA)

### Sicherheit by design / by default
- API-Key-Authentifizierung mit konstantem Zeitvergleich (`auth.go`).
- Body-Limit von 1 MiB für `POST /flags` und `PUT /flags/{key}` mit 413-Antwort.
- Validierung für Schlüssel, Beschreibung und Rollout-Prozent.
- Fehlerantworten ohne Stacktraces, Dateipfade oder interne Implementierungsdetails.
- Server-Timeouts und Header-Limit in `main.go`.
- Thread-sicherer In-Memory-Store (`store.go`).

Damit sind zentrale CRA-Anforderungen an sichere Voreinstellungen erfüllt.

### Update-/Patch-Fähigkeit und SBOM
- Das Modul nutzt die Go-Standardbibliothek; sichtbare externe Abhängigkeiten bestehen nicht.
- Eine explizite SBOM-Erzeugung ist im Code nicht sichtbar. Bei ausschließlicher Standardbibliothek ist das Risiko niedrig.
- `SECURITY.md`, `COMPLIANCE.md` und `PRIVACY.md` sind vorhanden; ihre Inhalte sind im Prüfkontext nicht vollständig enthalten und daher nicht abschließend bewertbar.

### Hinweise

- **F3 – low – SBOM-/Vulnerability-Prozess**  
  **Abhilfe:** In `COMPLIANCE.md`/`README.md` dokumentieren, dass bei künftigen externen Modulen `go.sum` gepflegt und ein SBOM-/Vulnerability-Scan (`govulncheck`, `syft`) in die Pipeline aufgenommen wird.

- **F4 – low – Sicherheitsdokumentation**  
  **Abhilfe:** Sicherstellen, dass `SECURITY.md` Sicherheitseigenschaften, Update-Weg und Meldestelle für Schwachstellen enthält. Falls noch nicht geschehen, dort einen Abschnitt „Security properties & update process“ ergänzen.

## 3. EU AI Act

Nicht einschlägig: Das Produkt enthält keine KI-Funktion.

## 4. Pflichttexte & UI

Nicht einschlägig: Reiner Go-Backend-Dienst ohne öffentliche Web-UI. Keine Cookie-, Consent- oder Legal-Notice-Pflichten. Eine Datenschutzdokumentation ist mit `PRIVACY.md` vorhanden.

## 5. Barrierefreiheit

Nicht einschlägig: Keine öffentliche Web-Oberfläche.

## Gesamtergebnis

Die verbindlichen Sprint-Anforderungen AC-01 bis AC-14 sind im Code umgesetzt. Es bestehen keine offenen rechtlichen Blocker. Die genannten Hinweise betreffen Betriebsdokumentation, TLS-Betrieb und künftige CRA-Vorsorge und sind nicht release-verhindernd.