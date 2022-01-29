package main

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

//go:embed templates/*
var resources embed.FS

var t = template.Must(template.ParseFS(resources, "templates/*"))

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	slackToken := os.Getenv("SLACK_TOKEN")
	if slackToken == "" {
		fmt.Println("Env SLACK_TOKEN is required")
		os.Exit(1)
	}
	slackAddress := os.Getenv("SLACK_ADDRESS")
	if slackAddress == "" {
		fmt.Println("Env SLACK_ADDRESS is required. Example: pymi.slack.com")
		os.Exit(1)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := map[string]string{}

		t.ExecuteTemplate(w, "index.html.tmpl", data)
	})
	http.HandleFunc("/invite", func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Form: %v", r.Form)
		email := strings.Join(r.Form["email"], "")
		log.Printf("Got email %s", email)
		resp, err := http.PostForm(fmt.Sprintf("https://%s/api/users.admin.invite", slackAddress), url.Values{
			"email":      {email},
			"token":      {slackToken},
			"set_active": {"true"},
		})
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		log.Printf("Response body: %s", body)

		data := map[string]string{}
		t.ExecuteTemplate(w, "invite.html.tmpl", data)
	})

	log.Println("listening on", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
