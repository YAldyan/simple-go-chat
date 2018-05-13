package main

import (
	gomniauthcommon "github.com/stretchr/gomniauth/common"
	"net/http"
)

type ChatUser interface {
	UniqueID() string
	AvatarURL() string
}

type chatUser struct {
	gomniauthcommon.User
	uniqueID string
}

func (u chatUser) UniqueID() string {
	return u.uniqueID
}

type authHandler struct {
	next http.Handler
}

func (h *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, err := r.Cookie("auth")

	if err == http.ErrNoCookie {
		// not authenticated
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusTemporaryRedirect)
		return
	}

	if err != nil {
		// some other error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// success - call the next handler
	h.next.ServeHTTP(w, r)
}

func MustAuth(handler http.Handler) http.Handler {
	return &authHandler{next: handler}
}

/*
Keterangan

Tipe authHandler tidak hanya mengimplement method serveHTTP (yang mana adalah extends
dari interface http.Handler), tapi juga menyiapkan field next dengan tipe http.handler.

Fungsi MustAuth sederhananya adalah instance dari Tipe authHandler yang melewatkan variabel
next dengan tipe http.Handler, untuk memberitahukan eksekusi berikutnya setelah autetifikasi.

Pola seperti ini membuat kita dengan mudah menambahkan otorisasi pada code di main program.
*/

// loginHandler handles the third-party login process.
// format: /auth/{action}/{provider}
func loginHandler(w http.ResponseWriter, r *http.Request) {
	segs := strings.Split(r.URL.Path, "/")
	action := segs[2]
	provider := segs[3]

	switch action {
	case "login":

		provider, err := gomniauth.Provider(provider)

		if err != nil {
			http.Error(w, fmt.Sprintf("Error when trying to get provider %s: %s", provider, err), http.StatusBadRequest)
			return
		}

		loginUrl, err := provider.GetBeginAuthURL(nil, nil)

		if err != nil {
			http.Error(w, fmt.Sprintf("Error when trying to GetBeginAuthURL for %s:%s", provider, err),
				http.StatusInternalServerError)
			return
		}

		w.Header.Set("Location", loginUrl)
		w.WriteHeader(http.StatusTemporaryRedirect)

	case "callback":
		provider, err := gomniauth.Provider(provider)

		if err != nil {
			http.Error(w, fmt.Sprintf("Error when trying to get provider %s: %s", provider, err), http.StatusBadRequest)
			return
		}

		creds, err := provider.CompleteAuth(objx.MustFromURLQuery(r.URL.RawQuery))

		if err != nil {
			http.Error(w, fmt.Sprintf("Error when trying to complete auth for %s: %s", provider, err), http.StatusInternalServerError)
			return
		}

		user, err := provider.GetUser(creds)

		if err != nil {
			http.Error(w, fmt.Sprintf("Error when trying to get user from %s: %s", provider, err), http.StatusInternalServerError)
			return
		}

		// authCookieValue := objx.New(map[string]interface{}{
		// 	"name": user.Name(),
		// }).MustBase64()
		chatUser := &chatUser{User: user}
		m := md5.New()
		io.WriteString(m, strings.ToLower(user.Email()))
		// userId := fmt.Sprintf("%x", m.Sum(nil))

		chatUser.uniqueID = fmt.Sprintf("%x", m.Sum(nil))
		avatarURL, err := avatars.GetAvatarURL(chatUser)
		if err != nil {
			log.Fatalln("Error when trying to GetAvatarURL", "-", err)
		}

		authCookieValue := objx.New(map[string]interface{}{

			/*
				In order to uniquely identify our users, we are going to copy Gravatar's approach by
				hashing their e-mail address and using the resulting string as an identifier. We will
				store the user ID in the cookie along with the rest of the user-specific data

				Here, we have hashed the e-mail address and stored the resulting value in the userid
				field at the point at which the user logs in
			*/
			"userid": chatUser.uniqueID,
			"name":   user.Name(),
			/*
				The AvatarURL field called in the preceding code will return the appropriate
				URL value and store it in our avatar_url field, which we then put into the
				cookie
			*/
			"avatar_url": avatarURL,
			// "email":      user.Email(),
		}).MustBase64()

		http.SetCookie(w, &http.Cookie{
			Name:  "auth",
			Value: authCookieValue,
			Path:  "/"})
		w.Header().Set("Location", "/chat")
		w.WriteHeader(http.StatusTemporaryRedirect)

		/*
			In our case, the cookie value is eyJuYW1lIjoiTWF0IFJ5ZXIifQ==, which is a Base64- encoded
			version of {"name":"Mat Ryer"}. Remember, we never typed in a name in our chat application;
			instead, Gomniauth asked Google for a name when we opted to sign in with Google. Storing
			non-signed cookies like this is fine for incidental information, such as a user's name;
			however, you should avoid storing any sensitive information using nonsigned cookies as it's
			easy for people to access and change the data.
		*/
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "Auth action %s not supported", action)
	}
}
