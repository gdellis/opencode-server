package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const cookieName = "opencode_session"

type Middleware struct {
	passwordHash    []byte
	sessionSecret   []byte
	sessionExpiry   time.Duration
	rateLimit       int
	rateLimitWindow time.Duration

	mu          sync.RWMutex
	attempts    map[string]int
	lastCleanup time.Time
}

func NewMiddleware(passwordHash, sessionSecret string, sessionExpiry time.Duration, rateLimit int, rateLimitWindow time.Duration) *Middleware {
	return &Middleware{
		passwordHash:    []byte(passwordHash),
		sessionSecret:   []byte(sessionSecret),
		sessionExpiry:   sessionExpiry,
		rateLimit:       rateLimit,
		rateLimitWindow: rateLimitWindow,
		attempts:        make(map[string]int),
		lastCleanup:     time.Now(),
	}
}

func (m *Middleware) LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("error") == "1" {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(loginPageWithError))
		return
	}
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(loginPage))
}

func (m *Middleware) Login(w http.ResponseWriter, r *http.Request) {
	ip := getIP(r)

	if !m.checkRateLimit(ip) {
		http.Error(w, "Too many login attempts", http.StatusTooManyRequests)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	password := r.FormValue("password")
	if password == "" {
		http.Redirect(w, r, "/login?error=1", http.StatusFound)
		return
	}

	if !m.validatePassword(password) {
		m.recordAttempt(ip)
		http.Redirect(w, r, "/login?error=1", http.StatusFound)
		return
	}

	m.clearAttempts(ip)

	sessionToken := m.generateSessionToken()
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    sessionToken,
		Path:     "/",
		MaxAge:   int(m.sessionExpiry.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

func (m *Middleware) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/login", http.StatusFound)
}

func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		if !m.validateSession(cookie.Value) {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *Middleware) validatePassword(password string) bool {
	return bcrypt.CompareHashAndPassword(m.passwordHash, []byte(password)) == nil
}

func (m *Middleware) generateSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (m *Middleware) validateSession(token string) bool {
	return len(token) == 64 && isHex(token)
}

func isHex(s string) bool {
	for _, r := range s {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') {
			return false
		}
	}
	return true
}

func (m *Middleware) checkRateLimit(ip string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if time.Since(m.lastCleanup) > m.rateLimitWindow {
		m.attempts = make(map[string]int)
		m.lastCleanup = time.Now()
	}

	return m.attempts[ip] < m.rateLimit
}

func (m *Middleware) recordAttempt(ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.attempts[ip]++
}

func (m *Middleware) clearAttempts(ip string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.attempts, ip)
}

func getIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	return strings.Split(r.RemoteAddr, ":")[0]
}

const loginPage = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>OpenCode Server - Login</title>
	<style>
		* { margin: 0; padding: 0; box-sizing: border-box; }
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
			background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
			min-height: 100vh;
			display: flex;
			align-items: center;
			justify-content: center;
		}
		.login-container {
			background: #fff;
			padding: 40px;
			border-radius: 12px;
			box-shadow: 0 20px 60px rgba(0,0,0,0.3);
			width: 100%;
			max-width: 400px;
		}
		h1 {
			text-align: center;
			color: #1a1a2e;
			margin-bottom: 30px;
			font-size: 28px;
		}
		.subtitle {
			text-align: center;
			color: #666;
			margin-bottom: 30px;
			font-size: 14px;
		}
		.form-group {
			margin-bottom: 20px;
		}
		label {
			display: block;
			margin-bottom: 8px;
			color: #333;
			font-weight: 500;
		}
		input[type="password"] {
			width: 100%;
			padding: 12px 16px;
			border: 2px solid #e0e0e0;
			border-radius: 8px;
			font-size: 16px;
			transition: border-color 0.2s;
		}
		input[type="password"]:focus {
			outline: none;
			border-color: #4a6fa5;
		}
		button {
			width: 100%;
			padding: 14px;
			background: linear-gradient(135deg, #4a6fa5 0%, #1a1a2e 100%);
			color: white;
			border: none;
			border-radius: 8px;
			font-size: 16px;
			font-weight: 600;
			cursor: pointer;
			transition: transform 0.2s, box-shadow 0.2s;
		}
		button:hover {
			transform: translateY(-1px);
			box-shadow: 0 4px 12px rgba(74, 111, 165, 0.4);
		}
		.error {
			background: #fee2e2;
			color: #dc2626;
			padding: 12px;
			border-radius: 8px;
			margin-bottom: 20px;
			text-align: center;
		}
	</style>
</head>
<body>
	<div class="login-container">
		<h1>OpenCode Server</h1>
		<p class="subtitle">Enter your password to access the dashboard</p>
		<form method="POST" action="/login">
			<div class="form-group">
				<label for="password">Password</label>
				<input type="password" id="password" name="password" placeholder="Enter your password" required autofocus>
			</div>
			<button type="submit">Sign In</button>
		</form>
	</div>
</body>
</html>`

const loginPageWithError = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>OpenCode Server - Login</title>
	<style>
		* { margin: 0; padding: 0; box-sizing: border-box; }
		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
			background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
			min-height: 100vh;
			display: flex;
			align-items: center;
			justify-content: center;
		}
		.login-container {
			background: #fff;
			padding: 40px;
			border-radius: 12px;
			box-shadow: 0 20px 60px rgba(0,0,0,0.3);
			width: 100%;
			max-width: 400px;
		}
		h1 {
			text-align: center;
			color: #1a1a2e;
			margin-bottom: 30px;
			font-size: 28px;
		}
		.subtitle {
			text-align: center;
			color: #666;
			margin-bottom: 30px;
			font-size: 14px;
		}
		.form-group {
			margin-bottom: 20px;
		}
		label {
			display: block;
			margin-bottom: 8px;
			color: #333;
			font-weight: 500;
		}
		input[type="password"] {
			width: 100%;
			padding: 12px 16px;
			border: 2px solid #e0e0e0;
			border-radius: 8px;
			font-size: 16px;
			transition: border-color 0.2s;
		}
		input[type="password"]:focus {
			outline: none;
			border-color: #4a6fa5;
		}
		button {
			width: 100%;
			padding: 14px;
			background: linear-gradient(135deg, #4a6fa5 0%, #1a1a2e 100%);
			color: white;
			border: none;
			border-radius: 8px;
			font-size: 16px;
			font-weight: 600;
			cursor: pointer;
			transition: transform 0.2s, box-shadow 0.2s;
		}
		button:hover {
			transform: translateY(-1px);
			box-shadow: 0 4px 12px rgba(74, 111, 165, 0.4);
		}
		.error {
			background: #fee2e2;
			color: #dc2626;
			padding: 12px;
			border-radius: 8px;
			margin-bottom: 20px;
			text-align: center;
		}
	</style>
</head>
<body>
	<div class="login-container">
		<h1>OpenCode Server</h1>
		<p class="subtitle">Enter your password to access the dashboard</p>
		<div class="error">Invalid password. Please try again.</div>
		<form method="POST" action="/login">
			<div class="form-group">
				<label for="password">Password</label>
				<input type="password" id="password" name="password" placeholder="Enter your password" required autofocus>
			</div>
			<button type="submit">Sign In</button>
		</form>
	</div>
</body>
</html>`
