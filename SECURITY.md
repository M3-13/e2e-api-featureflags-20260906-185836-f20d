# Sicherheit

Dieses Dokument beschreibt die Sicherheitseigenschaften des Feature-Flag-Service, die Vorgehensweise bei Sicherheitsupdates und den Nachweis der Software-Stückliste (SBOM).

## Sicherheitseigenschaften

- **API-Key-Authentifizierung**: Alle fachlichen Endpunkte (mit Ausnahme von `GET /healthz`) sind über einen API-Key geschützt. Der Key wird aus der Umgebungsvariable `FLAG_API_KEY` gelesen und über den Header `Authorization: Bearer <key>` bzw. `X-API-Key: <key>` übermittelt. Der Vergleich erfolgt zeitkonstant über `crypto/subtle.ConstantTimeCompare`. Fehlt der Key oder ist er ungültig, antwortet der Server mit `401`.
- **TLS-Vorgabe**: Der Dienst erzwingt TLS. Die TLS-Terminierung erfolgt entweder direkt über den Server (TLS-Betriebsvorgabe) oder über einen vorgeschalteten TLS-terminierenden Reverse-Proxy. Ein Betrieb über unverschlüsseltes HTTP ist nicht vorgesehen.
- **1-MiB-Body-Limit**: `POST /flags` und `PUT /flags/{key}` begrenzen den Request-Body vor dem vollständigen Einlesen auf 1 MiB. Bei Überschreitung antwortet der Handler mit `413` und liest den Body nicht weiter.
- **Strict-Input-Validierung**: Eingaben werden strikt validiert (fehlender oder leerer `key`, `enabled` kein Bool, `rollout_percent` außerhalb 0–100, ungültiges JSON, fehlender `user`-Parameter). Ungültige Eingaben werden mit `400` abgelehnt. Zusätzliche Daten nach dem JSON-Objekt werden nicht akzeptiert.
- **JSON-Fehlerobjekte ohne Interna**: Fehlerantworten bestehen ausschließlich aus dem definierten JSON-Fehlerobjekt `{"error":"..."}`. Es werden keine Stacktraces, internen Dateipfade, Panic-Texte oder sonstigen Implementierungsdetails zurückgegeben.
- **Logging ohne Query/PII**: Die Zugriffs-Logging-Middleware protokolliert ausschließlich Methode, Pfad ohne Query-String und Statuscode. Der `user`-Parameter aus `GET /flags/{key}/evaluate` wird weder geloggt noch gespeichert oder in einer Antwort zurückgegeben; er fließt ausschließlich in die Hash-Berechnung ein.

## Sicherheitsupdate-Politik

- **Patch-Intervall**: Sicherheitsupdates und Patches werden monatlich geprüft und bei Bedarf umgehend eingespielt.
- **Zuständigkeit**: Für die Einspielung von Sicherheitsupdates ist das Wartungsteam des Feature-Flag-Service zuständig. Sicherheitsrelevante Hinweise werden über die im Repository hinterlegten Kontaktwege an die Maintainer gemeldet.

## SBOM-Nachweis

- Das Modul nutzt **ausschließlich die Go-Standardbibliothek** und hat **keine Drittanbieter-Abhängigkeiten**; `go mod tidy` verändert den Abhängigkeitsbaum nicht.
- Eine maschinenlesbare SBOM wird mit folgendem Befehl erzeugt:

  ```sh
  go list -deps -json > sbom.json
  ```

- Die SBOM wird **vor jedem Release neu generiert** und dem Release beigefügt.
