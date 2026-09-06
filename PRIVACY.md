# Datenschutzhinweis Feature-Flag-Service

## Datenschutz

Der Endpunkt `GET /flags/{key}/evaluate` akzeptiert über den Query-Parameter `user`
einen Nutzer-Identifikator. Dieser Identifikator ist ein personenbezogenes Datum im
Sinne der Datenschutz-Grundverordnung (DSGVO).

Der Service verarbeitet diesen Identifikator ausschließlich transient im
Arbeitsspeicher: Er fließt unmittelbar in die SHA-256-basierte, deterministische
Rollout-Berechnung ein (stabiler Hash aus Flag-Key und user) und wird danach
verworfen. Der Identifikator wird dabei

- weder persistiert noch anderweitig gespeichert,
- nicht in einer Antwort zurückgegeben,
- und erscheint nicht im Zugriffslog (die Logging-Middleware protokolliert
  ausschließlich Methode, Pfad ohne Query-String und Statuscode; siehe
  AC-13/AC-14).

Die Rechtsgrundlage für die Verarbeitung durch den verantwortlichen Betreiber ist
Art. 6 Abs. 1 lit. b DSGVO (Verarbeitung zur Durchführung eines Vertrags bzw. auf
Anfrage der betroffenen Person vor Vertragsschluss) sowie Art. 6 Abs. 1 lit. f
DSGVO (berechtigtes Interesse am sicheren und zuverlässigen Betrieb des
Feature-Flag-Dienstes).

Der Betrieb dieses Dienstes muss zwingend hinter TLS erfolgen, damit der
Nutzer-Identifikator und alle übertragenen Daten auf dem Transportweg verschlüsselt
sind (siehe README).
