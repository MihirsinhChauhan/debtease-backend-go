package handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"DebtEase/internal/api"
	"DebtEase/internal/database"
	"DebtEase/internal/logger"
	"DebtEase/internal/models"
	"DebtEase/internal/utils"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googleidtoken "google.golang.org/api/idtoken"
)

const googleOAuthStateCookie = "google_oauth_state"

type googleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Name          string `json:"name"`
}

func googleOAuth2Config(cfg *api.Config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes:       []string{"openid", "email", "profile"},
		Endpoint:     google.Endpoint,
	}
}

func randomOAuthState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HandleGoogleOAuthStart redirects the browser to Google consent for the configured web redirect_uri.
func HandleGoogleOAuthStart(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" || cfg.GoogleRedirectURL == "" {
			utils.RespondWithError(w, http.StatusServiceUnavailable, "Google sign-in is not configured")
			return
		}
		state, err := randomOAuthState()
		if err != nil {
			logger.Logger.Errorw("oauth state rand failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "OAuth init failed")
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     googleOAuthStateCookie,
			Value:    state,
			Path:     "/",
			MaxAge:   int((10 * time.Minute).Seconds()),
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   r.TLS != nil,
		})
		authURL := googleOAuth2Config(cfg).AuthCodeURL(state)
		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

// HandleGoogleOAuthCallback exchanges the OAuth code, loads user profile, signs the user in, then
// redirects to GOOGLE_LOGIN_SUCCESS_REDIRECT (with token query params) or returns JSON when unset.
func HandleGoogleOAuthCallback(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" || cfg.GoogleRedirectURL == "" {
			utils.RespondWithError(w, http.StatusServiceUnavailable, "Google sign-in is not configured")
			return
		}
		expectedState, err := r.Cookie(googleOAuthStateCookie)
		if err != nil || expectedState.Value == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "OAuth state cookie missing — start login from GET /api/auth/google")
			return
		}
		state := r.URL.Query().Get("state")
		if state == "" || state != expectedState.Value {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid OAuth state")
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     googleOAuthStateCookie,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
		})

		if errMsg := r.URL.Query().Get("error"); errMsg != "" {
			utils.RespondWithError(w, http.StatusBadRequest, fmt.Sprintf("Google OAuth error: %s", errMsg))
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Missing authorization code")
			return
		}

		ctx := r.Context()
		oauthCfg := googleOAuth2Config(cfg)
		token, err := oauthCfg.Exchange(ctx, code)
		if err != nil {
			logger.Logger.Warnw("Google token exchange failed", "error", err)
			utils.RespondWithError(w, http.StatusBadRequest, "Failed to validate Google authorization")
			return
		}

		prof, err := fetchGoogleUserInfo(ctx, oauthCfg.Client(ctx, token))
		if err != nil {
			logger.Logger.Warnw("Google userinfo failed", "error", err)
			utils.RespondWithError(w, http.StatusBadGateway, "Failed to fetch Google profile")
			return
		}

		user, signErr := upsertGoogleUser(ctx, cfg, prof)
		if signErr != nil {
			utils.RespondWithError(w, http.StatusConflict, signErr.Error())
			return
		}

		resp, err := buildLoginSession(ctx, cfg, user)
		if err != nil {
			logger.Logger.Errorw("google oauth session creation failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create session")
			return
		}

		if strings.TrimSpace(cfg.GoogleLoginSuccessRedirect) != "" {
			redirectWithTokens(w, r, cfg.GoogleLoginSuccessRedirect, resp)
			return
		}
		utils.RespondWithJSON(w, http.StatusOK, resp)
	}
}

func redirectWithTokens(w http.ResponseWriter, r *http.Request, redirectBase string, resp models.LoginResponse) {
	u, err := url.Parse(redirectBase)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, "Misconfigured OAuth success redirect URL")
		return
	}
	q := u.Query()
	q.Set("access_token", resp.Token)
	q.Set("refresh_token", resp.RefreshToken)
	q.Set("id", resp.ID.String())
	q.Set("email", resp.Email)
	q.Set("first_name", resp.FirstName)
	q.Set("last_name", resp.LastName)
	q.Set("created_at", resp.CreatedAt)
	q.Set("updated_at", resp.UpdatedAt)
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
}

func fetchGoogleUserInfo(ctx context.Context, client *http.Client) (googleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return googleUserInfo{}, err
	}
	res, err := client.Do(req)
	if err != nil {
		return googleUserInfo{}, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return googleUserInfo{}, err
	}
	if res.StatusCode != http.StatusOK {
		msg := string(body)
		if len(msg) > 200 {
			msg = msg[:200]
		}
		return googleUserInfo{}, fmt.Errorf("userinfo HTTP %d: %s", res.StatusCode, msg)
	}
	var prof googleUserInfo
	if err := json.Unmarshal(body, &prof); err != nil {
		return googleUserInfo{}, err
	}
	return prof, nil
}

// HandleGoogleLoginIDToken accepts a Google OIDC ID token (e.g. from mobile google_sign_in) and returns JWT + refresh tokens.
func HandleGoogleLoginIDToken(cfg *api.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.GoogleClientID == "" {
			utils.RespondWithError(w, http.StatusServiceUnavailable, "Google sign-in is not configured")
			return
		}
		var req models.GoogleLoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid JSON")
			return
		}
		if strings.TrimSpace(req.IDToken) == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "id_token is required")
			return
		}

		ctx := r.Context()
		payload, err := googleidtoken.Validate(ctx, req.IDToken, cfg.GoogleClientID)
		if err != nil {
			logger.Logger.Warnw("Google id_token validation failed", "error", err)
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid Google token")
			return
		}
		sub := payload.Subject
		if sub == "" {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid Google token (missing subject)")
			return
		}
		claims := payload.Claims
		email, _ := claims["email"].(string)
		if email == "" {
			utils.RespondWithError(w, http.StatusUnauthorized, "Google account has no email on file")
			return
		}
		verified := emailVerifiedClaim(claims["email_verified"])
		first := stringFromClaim(claims["given_name"])
		last := stringFromClaim(claims["family_name"])
		name := stringFromClaim(claims["name"])
		if first == "" && last == "" && name != "" {
			parts := strings.SplitN(strings.TrimSpace(name), " ", 2)
			first = parts[0]
			if len(parts) > 1 {
				last = parts[1]
			}
		}

		prof := googleUserInfo{
			Sub:           sub,
			Email:         email,
			EmailVerified: verified,
			GivenName:     first,
			FamilyName:    last,
			Name:          name,
		}

		user, signErr := upsertGoogleUser(ctx, cfg, prof)
		if signErr != nil {
			utils.RespondWithError(w, http.StatusConflict, signErr.Error())
			return
		}

		resp, err := buildLoginSession(ctx, cfg, user)
		if err != nil {
			logger.Logger.Errorw("google id_token session creation failed", "error", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create session")
			return
		}
		logger.Logger.Infow("Google login OK", "user_id", user.ID)
		utils.RespondWithJSON(w, http.StatusOK, resp)
	}
}

func stringFromClaim(v interface{}) string {
	switch t := v.(type) {
	case string:
		return t
	default:
		return ""
	}
}

func emailVerifiedClaim(v interface{}) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return strings.EqualFold(t, "true") || t == "1"
	default:
		return false
	}
}

func upsertGoogleUser(ctx context.Context, cfg *api.Config, prof googleUserInfo) (database.User, error) {
	if prof.Email == "" {
		return database.User{}, errors.New("Google profile has no email")
	}
	if !prof.EmailVerified {
		return database.User{}, errors.New("Google email address is not verified")
	}

	gid := sql.NullString{String: prof.Sub, Valid: prof.Sub != ""}

	u, err := cfg.DB.GetUserByGoogleID(ctx, gid)
	if err == nil {
		return u, nil
	}
	if err != sql.ErrNoRows {
		return database.User{}, err
	}

	byEmail, err := cfg.DB.GetUserByEmail(ctx, prof.Email)
	if err == nil {
		if byEmail.GoogleID.Valid && byEmail.GoogleID.String != "" {
			if byEmail.GoogleID.String != prof.Sub {
				return database.User{}, errors.New("this email is linked to another Google account")
			}
			return byEmail, nil
		}
		linked, lerr := cfg.DB.LinkUserGoogleIDByEmail(ctx, database.LinkUserGoogleIDByEmailParams{
			Email:    prof.Email,
			GoogleID: gid,
		})
		if lerr == nil {
			return linked, nil
		}
		if lerr == sql.ErrNoRows {
			ref, err2 := cfg.DB.GetUserByGoogleID(ctx, gid)
			if err2 == nil {
				return ref, nil
			}
			return database.User{}, errors.New("cannot link Google account to this email")
		}
		return database.User{}, lerr
	}
	if err != sql.ErrNoRows {
		return database.User{}, err
	}

	first := strings.TrimSpace(prof.GivenName)
	last := strings.TrimSpace(prof.FamilyName)
	if first == "" && strings.TrimSpace(prof.Name) != "" {
		parts := strings.SplitN(strings.TrimSpace(prof.Name), " ", 2)
		first = parts[0]
		if len(parts) > 1 {
			last = parts[1]
		}
	}
	if first == "" {
		first = "Google"
	}
	if last == "" {
		last = "User"
	}

	created, err := cfg.DB.CreateUserGoogle(ctx, database.CreateUserGoogleParams{
		FirstName: first,
		LastName:  last,
		Email:     strings.TrimSpace(prof.Email),
		GoogleID:  gid,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "unique constraint") {
			return database.User{}, errors.New("an account already exists with this email or Google login")
		}
		return database.User{}, err
	}
	return created, nil
}
