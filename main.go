package main

import (
	"simple-go-chat/trace"
	"flag"
	"fmt"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/github"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"
)

type templateHandler struct {

	// compile the template once
	once sync.Once

	// take a filename string
	filename string

	// keep the reference to the compiled template, and then respond to HTTP requests
	templ *template.Template
}

// ServeHTTP handles the HTTP request.
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// memastikan web server hanya akan di-compile sekali saja
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("template/", t.filename)))
	})

	/*
		Presenting User Data

		Tampilkan User Information yang diperoleh setelah Sign In menggunakan User Information
		Sosial Media tertentu
	*/

	data := map[string]interface{}{
		"Host": r.Host,
	}
	if authCookie, err := r.Cookie("auth"); err == nil {
		data["UserData"] = objx.MustFromBase64(authCookie.Value)
	}

	t.templ.Execute(w, data)

	// t.templ.Execute(w, r)
}

func main() {

	/*
		setup gomniauth

		setelah mendaftarkan website kita pada social media tertentu, maka kita
		akan memperoleh security key dan secret.

		You should replace the key and secret placeholders with the actual values you noted
		down earlier. The third argument represents the callback URL that should match the ones
		you provided when creating your clients on the provider's website. Notice the second path
		segment is callback; while we haven't implemented this yet, this is where we handle the
		response from the authorization process.

		Gomniauth requires the SetSecurityKey call because it sends state data between
		the client and server along with a signature checksum, which ensures that the
		state values are not tempered with while being transmitted. The security key is
		used when creating the hash in a way that it is almost impossible to recreate
		the same hash without knowing the exact security key. You should replace some
		long key with a security hash or phrase of your choice
	*/

	gomniauth.SetSecurityKey("PUT YOUR AUTH KEY HERE")

	gomniauth.WithProviders(
		facebook.New("key", "secret",
			"http://localhost:8080/auth/callback/facebook"),
		github.New("key", "secret",
			"http://localhost:8080/auth/callback/github"),
		google.New("key", "secret",
			"http://localhost:8080/auth/callback/google"),
	)

	/*
		The call to flag.String returns a type of *string, which is to say it returns
		the address of a string variable where the value of the flag is stored. To get
		the value itself (and not the address of the value), we must use the pointer
		indirection operator, *.
	*/
	var addr = flag.String("addr", ":8081", "The addr of the application.")
	flag.Parse() // parse the flags

	// create a new room
	// r := newRoom()

	// r := newRoom(UseGravatar)

	// r := newRoom(UseFileSystemAvatar)
	// r.tracer = tracer.New(os.Stdout)

	// set the active Avatar implementation
	// var avatars Avatar = UseFileSystemAvatar

	var avatars Avatar = TryAvatars{
		UseFileSystemAvatar,
		UseAuthAvatar,
		UseGravatar
	}

	/*
		The templateHandler structure is a valid http.Handler type so we can pass it directly to
		the http.Handle function and ask it to handle requests that match the specified pattern
	*/
	// http.Handle("/", &templateHandler{filename: "chat.html"})

	/*
		Setelah penambahan autentifikasi

		templateHandler yang seharusnya langsung reload page chat.html, akan melalui sebuah handler
		lainnya, yaitu authHandler. MustAuth memaksa templateHandler dieksekui setelah authHandler.
	*/
	http.Handle("/chat", MustAuth(&templateHandler{filename: "chat.html"}))

	/*
		Login Page
	*/
	http.Handle("/login", &templateHandler{filename: "login.html"})

	/*
		Our loginHandler is only a function and not an object that implements the http.Handler
		interface.  This is because, unlike other handlers, we don't need it to store any state.
		The Go standard library supports this, so we can use the http.HandleFunc function to map
		it in a way similar to how we used http.Handle earlier

		Apakah beda http.Handle dan http.HandleFunc ??
	*/
	http.HandleFunc("/auth/", loginHandler)

	http.Handle("/room", r)

	http.Handle("/upload", MustAuth(&templateHandler{filename: "upload.html"}))

	http.HandleFunc("/uploader", uploaderHandler)

	http.Handle("/avatars/",
		http.StripPrefix("/avatars/",
			http.FileServer(http.Dir("./avatars"))))

	/*
		Logout Handler

		Untuk mereset user yang login, agara bisa login ulang
	*/
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:   "auth",
			Value:  "",
			Path:   "/",
			MaxAge: -1,
		})
		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)
	})

	// get the room going
	go r.run()

	// start the web server
	fmt.Println("starting web server at ", *addr)
	log.Println("Starting web server on ", *addr)

	// http.ListenAndServe(":8081", nil)
	err := http.ListenAndServe(*addr, nil)

	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
