// Code Submitted and Deployment in production environments by:
// Mykola Perehinets (mperehin)
// Tel: +380 67 772 6910
// mailto:mykola.perehinets@gmail.com

package main

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/gorilla/context"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

type user struct {
	Login    string
	Password string
}

var (
	// cookiekey = securecookie.GenerateRandomKey()
	hashKey   = []byte("33446a9dcf9ea060a0a6532b166da32f304af0de")
	blockKey  = []byte("33446a9dcf9ea060a0a6532b166da32f304af0de")
	cookiekey = securecookie.New(hashKey, blockKey)

	store  = sessions.NewCookieStore([]byte("cookiekey"))
	userdb user
	tpl    *template.Template
)

func index(w http.ResponseWriter, req *http.Request) {
	logFunc("Client " + req.RemoteAddr + " visited to /")

	if alreadyLoggedIn(w, req) {
		http.Redirect(w, req, "/todo", http.StatusSeeOther)
	}

	tpl.ExecuteTemplate(w, "index.html", false)
}

func todo(w http.ResponseWriter, req *http.Request) {
	logFunc("Client " + req.RemoteAddr + " visited to /todo")

	if !alreadyLoggedIn(w, req) {
		http.Redirect(w, req, "/login", http.StatusSeeOther)
		return
	}

	if req.Method == http.MethodPost {
		hostname := req.FormValue("hostname")
		ipaddress := req.FormValue("ipaddress")

		fmt.Println(hostname, ipaddress)
		logFunc("Client " + req.RemoteAddr + " deployed Bacula Agent to host (FQDN): " + req.FormValue("hostname") + ", host IP ADDRESS: " + req.FormValue("ipaddress"))

		arg := "-la"
		out, _ := exec.Command("ls", arg).Output()
		io.WriteString(w, fmt.Sprintf("%s", out))

		// tpl.ExecuteTemplate(w, "done.html", nil)
		return
	}

	tpl.ExecuteTemplate(w, "todo.html", nil)
}

func login(w http.ResponseWriter, req *http.Request) {
	logFunc("Client " + req.RemoteAddr + " visited to /login")

	session, err := store.Get(req, "session")
	if err != nil {
		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	if req.Method == http.MethodPost {
		un := req.FormValue("username")
		p := req.FormValue("password")

		if un != userdb.Login && p != userdb.Password {
			// http.Error(w, "", http.StatusForbidden)
			// http.Error(w, "ERROR: Username and/or password do not match. So, please verify it and relogin. Thank you for understanding.", http.StatusForbidden)
			logFunc("Client " + req.RemoteAddr + " visited to /login but ERROR: Username and/or password do not match." + " Username: " + req.FormValue("username") + ", Password: " + req.FormValue("password"))
			timer := time.NewTimer(time.Second * 5)
			<-timer.C
			logFunc("Client " + req.RemoteAddr + " visited to /login but ERROR: Timers expired. Automatic relogin...")
			timer1 := time.NewTimer(time.Second * 10)
			<-timer1.C
			// http.Redirect(w, req, "/logout", http.StatusSeeOther)
			http.Redirect(w, req, "/error", http.StatusSeeOther)
			tpl.ExecuteTemplate(w, "error.html", nil)
			return
		}

		session.Values["username"] = un
		session.Save(req, w)

		http.Redirect(w, req, "/", http.StatusSeeOther)
		return
	}

	tpl.ExecuteTemplate(w, "login.html", nil)
}

func logout(w http.ResponseWriter, req *http.Request) {
	logFunc("Client " + req.RemoteAddr + " visited to /logout")
	session, _ := store.Get(req, "session")

	session.Values["username"] = ""
	session.Save(req, w)

	http.Redirect(w, req, "/logout", http.StatusSeeOther)
	tpl.ExecuteTemplate(w, "logout.html", nil)
}

func alreadyLoggedIn(w http.ResponseWriter, req *http.Request) bool {
	session, err := store.Get(req, "session")
	if err != nil {
		return false
	}

	username, found := session.Values["username"]
	if found && username == userdb.Login {
		return true
	}

	return false
}

func main() {

	userdb.Login = "admin"
	userdb.Password = "admin"

	store.Options = &sessions.Options{
		// Domain:   "localhost.localdomain",
		Path:     "/",
		MaxAge:   10 * 60,
		Secure:   true,
		HttpOnly: true,
	}

	tpl = template.Must(template.ParseGlob("templates/*"))

	http.HandleFunc("/", index)
	http.HandleFunc("/todo", todo)
	http.HandleFunc("/login", login)
	http.HandleFunc("/logout", logout)
	http.HandleFunc("/done", done)
	http.HandleFunc("/error", error)
	http.HandleFunc("favicon.ico", faviconHandler)
	log.Printf("Starting Bacula Agent Deploy Server (BADS) front-end web service...")
	log.Printf("BADS about to listen on 8443. Go to https://127.0.0.1:8443 for verifing...")
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
	err := http.ListenAndServeTLS(":8443", "server.crt", "server.key", context.ClearHandler(http.DefaultServeMux))
	log.Fatal(err)
}

func faviconHandler(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, "/images/favicon.ico")
}

func done(w http.ResponseWriter, req *http.Request) {
	logFunc("Client " + req.RemoteAddr + " visited to /done")

	tpl.ExecuteTemplate(w, "done.html", false)
}

func error(w http.ResponseWriter, req *http.Request) {
	logFunc("Client " + req.RemoteAddr + " visited to /error")
	// session, _ := store.Get(req, "session")

	// session.Values["username"] = ""
	// session.Save(req, w)

	tpl.ExecuteTemplate(w, "error.html", false)
}

func logFunc(l string) {
	log.Println(l)
}
