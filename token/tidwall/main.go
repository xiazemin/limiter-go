package main

import (
	"io/ioutil"
	"net/http"

	"github.com/tidwall/limiter"
)

func main() {

	// Create a limiter for a maximum of 10 concurrent operations
	l := limiter.New(10)

	http.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		input, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
		defer r.Body.Close()

		var result []byte
		func() {
			l.Begin()
			defer l.End()
			// Do some intensive work here. It's guaranteed that only a maximum of ten
			// of these operations will run at the same time.
			result = []byte("rad!")
		}()

		w.Write(result.([]byte))
	})

	http.ListenAndServe(":8080", nil)
}