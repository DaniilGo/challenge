package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/DaniilGo/challenge/internal/app"
	"github.com/DaniilGo/challenge/internal/merge"
	"github.com/DaniilGo/challenge/internal/source"
)

var (
	u           = "https://seg.halo.ad.gt/api/v1/rtd"
	requestBody = []byte(`{"userIds":{"haloId": "TEST"},"config":{"publisherId": 999999000001}}`)
	sprintfTpl  = `{
    "returned_http_payload": %s
}`
)

func main() {
	httpClient := &http.Client{
		Timeout: time.Second * 30,
	}

	s := source.NewHttpPostSource(u, requestBody, httpClient)
	m := merge.NewJSONSprintfMerger(sprintfTpl)
	a := app.NewApp(s, m, os.Stdout)

	if err := a.Run(context.Background()); err != nil {
		fmt.Printf("failed to run app: %s\n", err.Error())
		os.Exit(1)
	}
}
