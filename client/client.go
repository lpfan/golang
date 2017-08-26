package main

import (
    "net/http"

    "github.com/zenazn/goji"
)

func main() {
    goji.Get("/", Root)
    goji.Serve()
}

func Root(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "templates/index.html")
}
