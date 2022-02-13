package main

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
)

var startMagazynZnicze = 100
var startMagazynWiazanki = 50
var magazynZnicze = startMagazynZnicze
var magazyn_wiazanki = startMagazynWiazanki
var iloscBabekZnicze = 2
var iloscBabekWiazanki = 2
var maxIloscPoslancow = 5
var maxKoszNaZnicze = 10
var maxKoszNaWiazanki = 10
var koszNaZnicze = 0
var koszNaWiazanki = 0
var maxPoslaniecWiazanki = 1
var maxPoslaniecZnicze = 2

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/magazyn", magazyn)
	mux.HandleFunc("/kosz", kosz)
	mux.HandleFunc("/babka/znicze/", babkaZnicze)
	mux.HandleFunc("/babka/wiazanki/", babkaWiazanki)
	mux.HandleFunc("/poslaniec/", poslaniec)
	log.Fatal(http.ListenAndServe(":3000", recoverMw(mux, true)))
}

func recoverMw(app http.Handler, dev bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
				stack := debug.Stack()
				log.Println(string(stack))
				if !dev {
					http.Error(w, "Something went wrong :(", http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "<h1>panic: %v</h1><pre>%s</pre>", err, string(stack))
			}
		}()

		nw := &responseWriter{ResponseWriter: w}
		app.ServeHTTP(nw, r)
		nw.flush()
	}
}

type responseWriter struct {
	http.ResponseWriter
	writes [][]byte
	status int
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.writes = append(rw.writes, b)
	return len(b), nil
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.status = statusCode
}

func (rw *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := rw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the ResponseWriter does not support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (rw *responseWriter) Flush() {
	flusher, ok := rw.ResponseWriter.(http.Flusher)
	if !ok {
		return
	}
	flusher.Flush()
}

func (rw *responseWriter) flush() error {
	if rw.status != 0 {
		rw.ResponseWriter.WriteHeader(rw.status)
	}
	for _, write := range rw.writes {
		_, err := rw.ResponseWriter.Write(write)
		if err != nil {
			return err
		}
	}
	return nil
}

func magazyn(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Stan magazynu:\nZnicze: ", magazynZnicze, "\nWiazanki: ", magazyn_wiazanki)
}
func kosz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Stan kosza:\nZnicze: ", koszNaZnicze, "\nWiazanki: ", koszNaWiazanki)
}

func babkaZnicze(w http.ResponseWriter, r *http.Request) {
	babka_numer, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/babka/znicze/"))

	if err != nil {
		panic("Błąd: [Podczas pobierania numeru babki]")
	}
	if babka_numer > iloscBabekZnicze {
		msg := fmt.Sprintf("Błąd: [dopuszczalna ilość babek od zniczy to: %d]", iloscBabekZnicze)
		panic(msg)
	}
	if magazynZnicze == 0 {
		panic("Błąd: [Wyczerpały się znicze z magazuny]")
	}
	if koszNaZnicze == maxKoszNaZnicze {
		msg := fmt.Sprintf("Błąd: [Kosz na znicze jest pelny, maksymalna ilosc to: %d]", maxKoszNaZnicze)
		panic(msg)
	}

	koszNaZnicze = koszNaZnicze + 1
	magazynZnicze = magazynZnicze - 1

	fmt.Fprintln(w, "Babka", babka_numer, "pobrala znicz", math.Abs(float64(magazynZnicze-startMagazynZnicze)), "z magazynu do kosza.")
}

func babkaWiazanki(w http.ResponseWriter, r *http.Request) {
	babka_numer, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/babka/wiazanki/"))

	if err != nil {
		panic("Błąd: [Podczas pobierania numeru babki]")
	}
	if babka_numer > iloscBabekWiazanki {
		msg := fmt.Sprintf("Błąd: [dopuszczalna ilosc babek od wiazanek to: %d]", iloscBabekWiazanki)
		panic(msg)
	}
	if magazyn_wiazanki == 0 {
		panic("Błąd: [Wyczerpały się wiązankiu z magazynu]")
	}
	if koszNaWiazanki == maxKoszNaWiazanki {
		msg := fmt.Sprintf("Błąd: [Kosz na wiazanki jest pelny, maksymalna ilosc to: %d]", maxKoszNaWiazanki)
		panic(msg)
	}

	koszNaWiazanki = koszNaWiazanki + 1
	magazyn_wiazanki = magazyn_wiazanki - 1

	fmt.Fprintln(w, "Babka", babka_numer, "pobrala wiazanke", math.Abs(float64(magazyn_wiazanki-startMagazynWiazanki)), "z magazynu do kosza.")
}

func poslaniec(w http.ResponseWriter, r *http.Request) {
	poslaniec_numer, err := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/poslaniec/"))

	if err != nil {
		panic("Błąd: [Podczas pobierania numeru babki]")
	}
	if poslaniec_numer > maxIloscPoslancow {
		msg := fmt.Sprintf("Błąd: [Mamy tylko : %d posłanców]", maxIloscPoslancow)
		panic(msg)
	}
	if koszNaWiazanki == 0 && koszNaZnicze == 0 {
		panic("Błąd:  [Kosz jest pusty!]")
	}
	msg := fmt.Sprintf("")
	if koszNaZnicze >= maxPoslaniecZnicze {
		koszNaZnicze = koszNaZnicze - maxPoslaniecZnicze
		msg = msg + fmt.Sprintf("Poslaniec %d pobiera %d znicze", poslaniec_numer, maxPoslaniecZnicze)
	} else {
		msg = msg + fmt.Sprintf("Poslaniec %d pobiera %d znicze", poslaniec_numer, koszNaZnicze)
		koszNaZnicze = 0
	}

	if koszNaWiazanki >= maxPoslaniecWiazanki {
		koszNaWiazanki = koszNaWiazanki - maxPoslaniecWiazanki
		msg = msg + fmt.Sprintf("\nPoslaniec %d pobiera %d wiazanki", poslaniec_numer, maxPoslaniecWiazanki)
	} else {
		msg = msg + fmt.Sprintf("\nPoslaniec %d pobiera %d wiazanki", poslaniec_numer, koszNaWiazanki)
		koszNaWiazanki = 0
	}

	fmt.Fprintln(w, msg)
}
