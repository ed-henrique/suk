package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ed-henrique/suk"
)

type server struct {
	mux            *http.ServeMux
	resource       string
	sessionStorage *suk.SessionStorage
}

func newCookie(value string, maxAge int) *http.Cookie {
	cookie := &http.Cookie{
		Name:     "access-token",
		Value:    value,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteDefaultMode,
		MaxAge:   maxAge,
	}

	return cookie
}

func (s *server) getCookie(w http.ResponseWriter, r *http.Request) {
	token, _ := s.sessionStorage.Set(s.resource)
	http.SetCookie(w, newCookie(token, 0))
	fmt.Fprint(w, "Cookie created")
}

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
		http.Error(w, "The server could not infer the resource type correctly", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, newCookie(newToken, 0))
	fmt.Fprintf(w, "%s", resource)
}

func (s *server) removeCookie(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("access-token")

	if err == http.ErrNoCookie {
		http.Error(w, "No cookie, why remove?", http.StatusBadRequest)
		return
	}

	err = s.sessionStorage.Remove(cookie.Value)

	http.SetCookie(w, newCookie("", -1))
	fmt.Fprint(w, "Cookie deleted, reference to resource lost")
}

func main() {
	ss, err := suk.NewSessionStorage(
		suk.WithCustomKeyLength(10),
		suk.WithTokenDuration(5*time.Minute),
		suk.WithAutoExpiredClear(),
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
