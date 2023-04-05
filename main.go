package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

var (
	appID = flag.String("app", "", "Application ID")
	// token        = flag.String("token", os.Getenv("DISCORD_TOKEN"), "Application token")
	clientSecret = flag.String("secret", os.Getenv("SECRET_KEY"), "OAuth2 secret")
	redirectURL  = flag.String("redirect", os.Getenv("REDIRECT_URL"), "OAuth2 Redirect URL")
)

var oauthConfig = oauth2.Config{
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://discord.com/oauth2/authorize",
		TokenURL: "https://discord.com/api/oauth2/token",
	},
	Scopes: []string{"identify", "role_connections.write", "connections"},
}

func init() {
	flag.Parse()

	// Set credentials.
	oauthConfig.ClientID = *appID
	oauthConfig.ClientSecret = *clientSecret
	// Set OAuth2 Redirect URL.
	oauthConfig.RedirectURL, _ = url.JoinPath(*redirectURL, "/callback")
}

func main() {
	// s, _ := discordgo.New("Bot " + *token)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		// Creating state for security.
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

		// Safeguard verification.
		cookie, _ := r.Cookie("state")
		if q["state"][0] != cookie.Value {
			return
		}

		// Fetch the tokens with code we've received.
		tokens, err := oauthConfig.Exchange(r.Context(), q["code"][0])
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		// Construct a temporary session with user's OAuth2 access_token.
		ts, _ := discordgo.New("Bearer " + tokens.AccessToken)

		// Retrive the user connections.
		conns, err := ts.UserConnections()
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}
		creator := findGitHubAccount(conns)
		if creator == "" {
			w.Write([]byte("You need to link your GitHub account to Discord."))
			return
		}

		// And show it to the user.
		w.Write([]byte(fmt.Sprintf("Your GitHub account: https://github.com/%s", creator)))
	})
	port := ":8000"
	log.Printf("server listening at http://localhost%s\n", port)
	http.ListenAndServe(port, nil)
}

func findGitHubAccount(conns []*discordgo.UserConnection) string {
	for _, conn := range conns {
		if conn.Type == "github" {
			return conn.Name
		}
	}
	return ""
}
