// Thema: Wie baue ich einen Webserver in go?
// Zweck: Mögliche Anwendung für dynamische Views auf eine Datenbank über einen Browser.
//	Go ab 1.22 (!), aktueller Browser (es geht auch mit go 1.18, ist aber etwas mehr Arbeit)
//
//	Datum: 02.06.2024

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
)

// Hier würde man für jede Aufgabe einen eigenen Handler schreiben
// Ein "Handler" erfüllt das Interface http.Handler.

// Ein Beispiel-Handler für alles
func echoHandler(tag string, logger *log.Logger) http.Handler {

	// Erzeuge den Handler als anonyme Funktion und gib diese zurück. Die ist
	// hier als closure implementiert, um verschiedene zusätzliche Objekte, wie
	// Datenbankverbindungen, Logger und dergleichen mit einzubinden.
	var handler = func(w http.ResponseWriter, r *http.Request) {

		// In diesem Handler werden keine Datenbank-Inhalte gezeigt, sondern eine einfache Nachricht.
		msg := fmt.Sprintf("Ich handle %s.\n"+
			"Du hast den Server mit der %s Methode unter %s erreicht.",
			tag, r.Method, r.URL.Path)

		logger.Println(msg) // -> Nachricht zur Info loggen

		// Die Response (z.B. eine Web-Page) für den Browser bzw. Client erzeugen.
		// Tipp: Benutze html/template zur dynamischen Erzeugung von HTML-Seiten.
		htmlContent := `
<html>
	<head>
    	<title>WebEcho</title>
	</head>
	<body>
    	<div>` + msg + `</div>
	</body>
</html>`

		//Im Header sagt man dem Browser, was man da anliefert ...
		w.Header().Set("Content-Type", "text/html")
		// ...und schreibt die Response -> Fehler sind zu behandeln.
		if _, err := w.Write([]byte(htmlContent)); err != nil {
			http.Error(w, fmt.Sprintf("interner Serverfehler %s", err), http.StatusInternalServerError)
			return
		}
	}

	// Einmal das geforderte Interface drüberziehen und die Handler-Funktion zurückgeben.
	return http.HandlerFunc(handler)
}

func main() {
	logger := log.New(os.Stdout, "**", 0)

	// ein neuer Multiplexer für das Routing
	var mux *http.ServeMux = http.NewServeMux()

	// Registriere einige Methoden und Routen, die Deine WebApp anbieten soll.
	// Hier wird jeder Methode/Route ein Handler zugeordnet.
	// Der jeweilige Handler enthält die "Modelle und Logik" Deiner WebApp.
	// Der Mux ruft den passenden Handler auf, wenn jemand die HTTP-Route anfragt.

	// HINWEIS: das funktioniert so erst mit go 1.22 - sonst kurz umbauen:
	// Vor go 1.22 muss man die HTTP-Methode händisch aus dem Request extrahieren.
	// (fehleranfällig und unschön)
	// Der neue ServeMux seit 1.22 nimmt einem diese Arbeiten ab ...
	// siehe https://eli.thegreenplace.net/2023/better-http-server-routing-in-go-122

	// Wir bedienen hier die folgende sehr einfache Benutzerschnittstelle.
	// Es wird derzeit in jedem Fall lediglich eine Webpage ausgeliefert,
	// die die Methode und den Pfad noch einmal zurückgibt (Echo).
	// Andere Handler würden hier echte Aufgaben erfüllen ...
	// GET   /
	// GET   /meinPfad
	// POST  /meinPfad
	// ...   /... alle anderen
	mux.Handle("GET /meinPfad", echoHandler("Get an /meinPfad", logger))
	mux.Handle("GET /{$}", echoHandler("Get an der Wurzel", logger))
	mux.Handle("POST /meinPfad", echoHandler("Posts an /meinPfad", logger))
	// Ein öffentliches Verzeichnis für Bilder und Downloads benutzen -
	// beispielsweise für das Tab-Icon.
	mux.Handle("GET /favicon.ico", http.FileServer(http.Dir("static")))
	// Hier sollte man eigentlich nie rauskommen.
	mux.Handle("/", echoHandler("Fallbacks", logger))

	// Baue einen Web-Server zusammen und starte ihn
	server := &http.Server{
		// Hier würde man die eigene IP-Adresse angeben und müsste
		// die eigene Firewall für ankommendes TCP freigeben,
		// damit der Server von anderen im Netz erreichbar wird.
		Addr:    net.JoinHostPort("localhost", "8081"),
		Handler: http.Handler(mux),
	}

	logger.Printf("Starte Web-Server unter http://%s\n", server.Addr)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Printf("Server-Fehler: %s\n", err)
	}
}
