package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ed-henrique/suk"
)

// This is an example of using suk to store server-side sessions for your users
// with a syncMap. Most error checking is skipped for brevity, but you should
// check in your application.
//
// We define three simple handlers for an HTTP server:
// - getCookie (creates a new key for the given session as in an user login);
// - getResource (checks if the session is valid for the given key, and if so,
// returns the resource);
// - removeCookie (removes the key and the cookie, as in an user logout);

// server holds our server, session storage and more importantly, our SUPER
// SECRET resource.
type server struct {
	mux      *http.ServeMux
	resource string // Our top-tier SECRET, which normally would be a
	// DB resource or a file that the user may want to retrieve
	sessionStorage *suk.SessionStorage
}

// newAccessTokenCookie creates a new authentication cookie with the value and
// maxAge given, adding some sane defaults.
func newAccessTokenCookie(value string, maxAge int) *http.Cookie {
	cookie := &http.Cookie{
		Name:     "access-token",
		Value:    value,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
	}

	return cookie
}

// getCookie generates a new "access-token" cookie, as in an user login. In an
// actual application, you would check the user credentials before handing the
// access to him.
func (s *server) getCookie(w http.ResponseWriter, r *http.Request) {
	token, _ := s.sessionStorage.Set(s.resource)
	http.SetCookie(w, newAccessTokenCookie(token, 0))
	fmt.Fprint(w, "Cookie created")
}

// getResource checks if the session for the given key is valid, and if so,
// returns the resource to the user.
func (s *server) getResource(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("access-token")

	if err == http.ErrNoCookie {
		http.Error(w, "No cookie, no resource", http.StatusUnauthorized)
		return
	}

	resourceRaw, newToken, err := s.sessionStorage.Get(cookie.Value)

	if err == suk.ErrNoKeyFound {
		http.Error(w, "No key in storage", http.StatusNotFound)
		return
	} else if err == suk.ErrKeyWasExpired {
		http.Error(w, "The given key is expired", http.StatusUnauthorized)
		return
	}

	resource, ok := resourceRaw.(string)
	if !ok {
		http.Error(
			w,
			"The server could not infer the resource type correctly",
			http.StatusInternalServerError,
		)
		return
	}

	http.SetCookie(w, newAccessTokenCookie(newToken, 0))
	fmt.Fprintf(w, "%s", resource)
}

// removeCookie removes the reference key to the session, if there's any, and
// returns a blank "access-token" cookie, as in an user logout. In an actual
// application, you may also perform some custom logout tasks.
func (s *server) removeCookie(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("access-token")

	if err == http.ErrNoCookie {
		http.Error(w, "No cookie, why remove?", http.StatusBadRequest)
		return
	}

	err = s.sessionStorage.Remove(cookie.Value)

	http.SetCookie(w, newAccessTokenCookie("", -1))
	fmt.Fprint(w, "Cookie deleted, reference to resource lost")
}

func main() {
	// We are using the default syncMap
	ss, err := suk.NewSessionStorage(
		suk.WithKeyLength(10),
		suk.WithKeyDuration(5*time.Minute),
		suk.WithAutoClearExpiredKeys(),
	)
	if err != nil {
		panic(err)
	}

	s := &server{
		mux:            http.NewServeMux(),
		resource:       "SECRET",
		sessionStorage: ss,
	}

	s.mux.HandleFunc("GET /get_cookie", s.getCookie)
	s.mux.HandleFunc("GET /access_resource", s.getResource)
	s.mux.HandleFunc("DELETE /remove_cookie", s.removeCookie)

	if err := http.ListenAndServe(":8080", s.mux); err != nil {
		fmt.Printf("Server could not start. err=%s\n", err.Error())
	}
}
