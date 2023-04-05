package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

var oauthConfig = oauth2.Config{
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://discord.com/oauth2/authorize",
		TokenURL: "https://discord.com/api/oauth2/token",
	},
	Scopes: []string{"identify", "role_connections.write", "connections"},
}

var (
	appID = flag.String("app", os.Getenv("APPLICATION_ID"), "Application ID")
	token = flag.String("token", os.Getenv("DISCORD_TOKEN"), "Application token")
)

func init() {
	flag.Parse()
}

func main() {
	s, _ := discordgo.New("Bot " + *token)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		state := uuid.NewString()
		cookie := http.Cookie{
			Name:     "state",
			Value:    state,
			Path:     "/",
			MaxAge:   300, // 5 minutes
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, &cookie)
		// Redirect the user to Discord OAuth2 page.
		http.Redirect(w, r, oauthConfig.AuthCodeURL(state), http.StatusMovedPermanently)
	})

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		cookie, _ := r.Cookie("state")

		if q["state"][0] != cookie.Value {
			return
		}

		tokens, err := oauthConfig.Exchange(r.Context(), q["code"][0])
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		ts, _ := discordgo.New("Bearer " + tokens.AccessToken)

		u, err := ts.User("@me")
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
	})

	http.ListenAndServe(":8000", nil)
}
