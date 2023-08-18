package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth/providers/auth0"
	"github.com/markbates/goth/providers/bitbucket"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gitlab"
	"net/http"
	"os"

	"github.com/igk1972/netlify-cms-oauth-provider-go/internal/dotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

var (
	externalHost = "localhost:3000"
	listenHost   = "localhost:3000"
)

const (
	script = `
<html><body><script>
    (function() {
      function recieveMessage(e) {
        console.log("recieveMessage", e)
        // send message to main window with da app
        window.opener.postMessage(
          	%s
			,
          e.origin
        )
      }
      window.addEventListener("message", recieveMessage, false)
      // Start handshare with parent
      console.log("Sending message")
      window.opener.postMessage("authorizing:github", "*")
      })()
    </script></body></html>
	`
)

// GET /
func handleMain(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Printf("handleMain...\n")
	res.WriteHeader(http.StatusOK)
	res.Write([]byte(``))
}

// GET /auth/provider  Initial page redirecting by provider
func handleAuthProvider(res http.ResponseWriter, req *http.Request) {
	fmt.Printf("handleAuthProvider: calling begin\n")
	gothic.BeginAuthHandler(res, req)
}

// GET /callback/{provider}  Called by provider after authorization is granted
func handleCallbackProvider(res http.ResponseWriter, req *http.Request) {

	var (
		status string
		result string
	)
	provider, errProvider := gothic.GetProviderName(req)
	fmt.Printf("handleCallbackProvider: dealing with callback, provider=%s, errProvider=%s\n", provider, errProvider)

	user, errAuth := gothic.CompleteUserAuth(res, req)
	status = "error"
	if errProvider != nil {
		fmt.Printf("provider failed with '%s'\n", errProvider)
		result = fmt.Sprintf("%s", errProvider)
	} else if errAuth != nil {
		fmt.Printf("auth failed with '%s'\n", errAuth)
		result = fmt.Sprintf("%s", errAuth)
	} else {
		fmt.Printf("Logged in as %s user: %s (%s)\n", user.Provider, user.Email, user.AccessToken)
		status = "success"
		result = fmt.Sprintf(`'authorization:github:success:{"token":"%s", "provider":"%s"}'`, user.AccessToken, user.Provider)

	}
	res.Header().Set("Content-Type", "text/html; charset=utf-8")
	res.WriteHeader(http.StatusOK)

	fmt.Printf("Plug in later, status=%s\n", status)

	// WARNING eric: Not handling error cases
	//message := fmt.Sprintf(`authorization:github:success:%s`, result)

	script := fmt.Sprintf(script, result)

	res.Write([]byte(script))
	//res.Write([]byte(fmt.Sprintf(script, status, provider, result)))
}

// GET /refresh
func handleRefresh(res http.ResponseWriter, req *http.Request) {
	fmt.Printf("handleRefresh: refresh with '%s'\n", req)
	res.Write([]byte(""))
}

// GET /success
func handleSuccess(res http.ResponseWriter, req *http.Request) {
	fmt.Printf("handleSuccess: success with '%s'\n", req)
	res.Write([]byte(""))
}

func init() {
	dotenv.File(".env")
	if hostEnv, ok := os.LookupEnv("LISTEN_HOST"); ok {
		listenHost = hostEnv
	}

	if redirectEnv, ok := os.LookupEnv("EXTERNAL_HOST"); ok {
		externalHost = redirectEnv
	}
	var (
		gitlabProvider goth.Provider
	)
	if gitlabServer, ok := os.LookupEnv("GITLAB_SERVER"); ok {
		gitlabProvider = gitlab.NewCustomisedURL(
			os.Getenv("GITLAB_KEY"), os.Getenv("GITLAB_SECRET"),
			fmt.Sprintf("https://%s/callback/gitlab", externalHost),
			fmt.Sprintf("https://%s/oauth/authorize", gitlabServer),
			fmt.Sprintf("https://%s/oauth/token", gitlabServer),
			fmt.Sprintf("https://%s/api/v3/user", gitlabServer),
		)
	} else {
		gitlabProvider = gitlab.New(
			os.Getenv("GITLAB_KEY"), os.Getenv("GITLAB_SECRET"),
			fmt.Sprintf("https://%s/callback/gitlab", externalHost),
		)
	}

	goth.UseProviders(
		github.New(
			os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"),
			fmt.Sprintf("https://%s/callback/github", externalHost),
			"repo",
		),
		bitbucket.New(
			os.Getenv("BITBUCKET_KEY"), os.Getenv("BITBUCKET_SECRET"),
			fmt.Sprintf("https://%s/callback//bitbucket", externalHost),
		),
		gitlabProvider,
		auth0.New(
			os.Getenv("AUTH0_KEY"), os.Getenv("AUTH0_SECRET"),
			//fmt.Sprintf("https://%s/auth/callback/auth0", externalHost),
			fmt.Sprintf("https://%s/callback/auth0", externalHost),
			os.Getenv("AUTH0_DOMAIN"),
		),
	)
}

func main() {
	//router := pat.New()
	router := mux.NewRouter()

	session_secret := "asdfasdfasdf asdflkqwerwqerqwe23434343"
	os.Setenv("SESSION_SECRET", session_secret)
	gothic.Store = sessions.NewCookieStore([]byte(session_secret))

	//router := r.PathPrefix("/auth").Subrouter()
	router.HandleFunc("/callback/{provider}", handleCallbackProvider).Methods("GET")
	router.HandleFunc("/auth/{provider}", handleAuthProvider).Methods("GET")
	router.HandleFunc("/auth/", handleAuthProvider).Methods("GET")
	router.HandleFunc("/auth", handleAuthProvider).Methods("GET")
	//router.HandleFunc("/auth/", handleAuth).Methods("GET")
	//router.HandleFunc("/", handleAuth).Methods("GET")
	router.HandleFunc("/refresh", handleRefresh).Methods("GET")
	router.HandleFunc("/success", handleSuccess).Methods("GET")
	//router.HandleFunc("/", handleMain).Methods("GET")
	router.NotFoundHandler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Printf(fmt.Sprintf("NOT FOUND! %s\n", request.URL))
	})
	router.MethodNotAllowedHandler = http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		fmt.Printf(fmt.Sprintf("Method Not Allowed! %s\n", request.Method))
	})
	//
	http.Handle("/", router)
	//http.Handle("/", router)
	//
	fmt.Printf("Started running on %s, with externalHost=%s\n", listenHost, externalHost)
	fmt.Println(http.ListenAndServe(listenHost, nil))
}
