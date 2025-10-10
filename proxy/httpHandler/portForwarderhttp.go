package httphandler

import (
	"io"
	"net/http"
)

func portForwarder(w http.ResponseWriter, r *http.Request) {
	appName := r.PathValue("appname")
	appPath := r.PathValue("apppath")
	if appName == "" {
		http.Error(w, "Invalid address", http.StatusBadRequest)
	}

	defer r.Body.Close()
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Cant read Request body", http.StatusInternalServerError)
	}

}
