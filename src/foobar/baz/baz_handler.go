package baz

import (
	"github.com/Gamma169/go-server-helpers/server"
	"net/http"
)

func BazHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"baz-id": bazId,
	}

	server.WriteModelToResponseJSON(response, http.StatusOK, w)
}
