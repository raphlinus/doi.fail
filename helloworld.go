// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// [START gae_go111_app]

// Sample helloworld is an App Engine app.
package main

// [START import]
import (
	"bytes"
	"context"
	"fmt"
	"html"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"cloud.google.com/go/datastore"
	"google.golang.org/api/iterator"
)

// [END import]
// [START main_func]

type Doifail struct {
	Client *datastore.Client
}

type Doi struct {
	Doi string
	Url string
}

func main() {
	ctx := context.Background()
	projectID := "savvy-courage-259416"

	// Creates a client.
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
			log.Fatalf("Failed to create client: %v", err)
	}

	doifail := Doifail{
		Client: client,
	}

	http.HandleFunc("/", doifail.indexHandler)

	// [START setting_port]
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
	// [END setting_port]
}

// [END main_func]

func validateDoi(s string) bool {
	// This is a very restrictive regex, almost all Unicode sequences are technically
	// valid, but we prefer to be conservative.
	var validDoi = regexp.MustCompile(`[0-9.]+(/[0-9a-zA-Z.\-]+)+$`)
	return validDoi.MatchString(s)
}

// [START indexHandler]

func fetchCrossref(ctx context.Context, doi string) (string, error) {
	query_url := "https://api.crossref.org/works/" + url.PathEscape(doi)
	log.Printf("query_url: %s", query_url)
	resp, err := http.Get(query_url)
	if err != nil {
		return "", err
	}
	json := new(bytes.Buffer)
	// TODO: handle err?
	json.ReadFrom(resp.Body)
	return json.String(), nil
}

// indexHandler responds to requests with our greeting.
func (d *Doifail) indexHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	header := w.Header()
	header.Set("Content-Type", "text/html")
	fmt.Fprint(w, "<!doctype html><html>\n")
	// TODO: validate path, otherwise it could be a security nightmare
	fmt.Fprintf(w, "<p>URL path: %s</p>\n", html.EscapeString(r.URL.Path))
	doi := r.URL.Path[1:]
	link := fmt.Sprintf("https://sci-hub.se/%s", doi)
	xref_link := fmt.Sprintf("https://doi.org/%s", doi)
	if validateDoi(doi) {
		fmt.Fprintf(w, "<p>Sci-hub link: <a href=\"%s\">%s</a></p>\n", link, html.EscapeString(link))
		fmt.Fprintf(w, "<p>Crossref link: <a href=\"%s\">%s</a></p>\n", xref_link, html.EscapeString(xref_link))
	} else {
		fmt.Fprintf(w, "<p>Not valid!</p>")
		return
	}


	query := datastore.NewQuery("doi").
		Filter("doi =", doi)
	it := d.Client.Run(ctx, query)

	count := 0
	for {
		var doi Doi
		_, err := it.Next(&doi)
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Error fetching next doi: %v", err)
		}
		count += 1
		link := doi.Url
		fmt.Fprintf(w, "<p>Url: <a href=\"%s\">%s</a></p>", link, html.EscapeString(link))
	}
	log.Printf("Found %d items for %v", count, doi)
	if false {
		xref, err := fetchCrossref(ctx, doi)
		if err == nil {
			fmt.Fprintf(w, "<p>json response: %s</p>\n", html.EscapeString(xref))
		} else {
			log.Printf("Error fetching json: %v", err)
		}
	}
}

// [END indexHandler]
// [END gae_go111_app]
