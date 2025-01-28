package uadmin

import (
	"net/http"
	"strings"
)

// loginHandler HTTP handeler for verifying login data and creating sessions for users
func loginHandler(w http.ResponseWriter, r *http.Request) {
	type Context struct {
		Err         string
		ErrExists   bool
		SiteName    string
		Languages   []Language
		RootURL     string
		OTPRequired bool
		Language    Language
		Username    string
		Password    string
		Logo        string
		FavIcon     string
		SSOURL      string
	}

	c := Context{}
	c.SiteName = SiteName
	c.RootURL = RootURL
	c.Language = getLanguage(r)
	c.Logo = Logo
	c.FavIcon = FavIcon
	c.SSOURL = SSOURL

	if session := IsAuthenticated(r); session != nil {
		session = session.User.GetActiveSession()
		SetSessionCookie(w, r, session)
		if r.URL.Query().Get("next") != "" {
			http.Redirect(w, r, r.URL.Query().Get("next"), 303)
		}
	}

	if r.Method == cPOST {
		if r.FormValue("save") == "Send Request" {
			// This is a password reset request
			IncrementMetric("uadmin/security/passwordreset/request")
			email := r.FormValue("email")
			user := User{}
			Get(&user, "Email = ?", email)
			if user.ID != 0 {
				IncrementMetric("uadmin/security/passwordreset/emailsent")
				c.ErrExists = true
				c.Err = "Password recovery request sent. Please check email to reset your password"
				forgotPasswordHandler(&user, r, "", "")
			} else {
				IncrementMetric("uadmin/security/passwordreset/invalidemail")
				c.ErrExists = true
				c.Err = "Please check email address. Email address must be associated with the account to be recovered."
			}
		} else {
			// This is a login request
			username := r.PostFormValue("username")
			username = strings.TrimSpace(strings.ToLower(username))
			password := r.PostFormValue("password")
			otp := r.PostFormValue("otp")
			lang := r.PostFormValue("language")

			session := Login2FA(r, username, password, otp)
			if session == nil || !session.User.Active {
				c.ErrExists = true
				c.Err = "Invalid username/password or inactive user"
			} else {
				// Set session cookie
				SetSessionCookie(w, r, session)

				// set language cookie
				cookie, _ := r.Cookie("language")
				if cookie == nil {
					cookie = &http.Cookie{}
				}
				cookie.Name = "language"
				cookie.Value = lang
				cookie.Path = "/"
				http.SetCookie(w, cookie)

				// Check for OTP
				if session.PendingOTP {
					c.Username = username
					c.Password = password
					c.OTPRequired = true
				} else {
					if r.URL.Query().Get("next") == "" {
						http.Redirect(w, r, strings.TrimSuffix(r.RequestURI, "logout"), http.StatusSeeOther)
						return
					}
					http.Redirect(w, r, r.URL.Query().Get("next"), http.StatusSeeOther)
					return
				}
			}
		}
	}
	c.Languages = ActiveLangs
	RenderHTML(w, r, "./templates/uadmin/"+Theme+"/login.html", c)
}
