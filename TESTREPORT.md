VERDICT: BUGS_FOUND

- **Titel:** Go-Testsuite kompiliert nicht wegen doppelter Deklarationen
- **Symptom:** `go test ./...` bricht mit Build-Fehler ab, bevor irgendein Test ausgeführt wird. Damit ist AC-10 („go test läuft grün und deckt alle Handler sowie die deterministische Rollout-Entscheidung ab“) nicht erfüllt: Das Produkt liefert keinerlei erfolgreich laufende automatisierte Verifikation seiner Funktionalität, obwohl `go build ./...` selbst fehlerfrei durchläuft.
- **Repro:** Im Projektverzeichnis `go test ./...` ausführen.
- **Evidence:**
  ```
  FAIL	featureflags [build failed]
  FAIL

  # featureflags [featureflags.test]
  .\flags_create_test.go:16:6: newTestHandler redeclared in this block
  	.\api_test.go:13:6: other declaration of newTestHandler
  .\rollout_test.go:27:6: TestRolloutDecisionDeterministic redeclared in this block
  	.\evaluate_test.go:21:6: other declaration of TestRolloutDecisionDeterministic
  ```
- **Suspected file(s):** Symbolkollision im Package `main` zwischen `flags_create_test.go` / `api_test.go` (`newTestHandler`) und `rollout_test.go` / `evaluate_test.go` (`TestRolloutDecisionDeterministic`). Die doppelten Definitionen müssen konsolidiert oder die Dateien einem eigenen Testpaket zugeordnet werden.
- **Severity:** high