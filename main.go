package main

import (
	"flag"
	"net/http"
	"net/url"
	"os"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

var (
	appID = flag.String("app", os.Getenv("APPLICATION_ID"), "Application ID")
	// token        = flag.String("token", os.Getenv("DISCORD_TOKEN"), "Application token")
	clientSecret = flag.String("secret", os.Getenv("SECRET_KEY"), "OAuth2 secret")
	redirectURL  = flag.String("redirect", os.Getenv("REDIRECT_URL"), "OAuth2 Redirect URL")
)

var oauthConfig = oauth2.Config{
	ClientID:     *appID,
	ClientSecret: *clientSecret,
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://discord.com/oauth2/authorize",
		TokenURL: "https://discord.com/api/oauth2/token",
	},
	Scopes: []string{"identify", "role_connections.write", "connections"},
}

func init() {
	flag.Parse()
	// Set OAuth2 Redirect URL
	oauthConfig.RedirectURL, _ = url.JoinPath(*redirectURL, "/callback")
}

func main() {
	// s, _ := discordgo.New("Bot " + *token)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		// Creating state for security
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

	http.ListenAndServe(":8000", nil)
}
