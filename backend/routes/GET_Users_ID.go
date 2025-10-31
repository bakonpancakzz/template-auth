package routes

import (
	"net/http"
	"strconv"

	"github.com/bakonpancakz/template-auth/tools"

	"github.com/jackc/pgx/v5"
)

func GET_Users_ID(w http.ResponseWriter, r *http.Request) {

	snowflake, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_USER)
		return
	}
	ctx, cancel := tools.NewContext()
	defer cancel()

	// Fetch Relevant Profile
	var profile tools.DatabaseProfile
	err = tools.Database.QueryRow(ctx,
		`SELECT 
			id, created, username, displayname, 
			biography, subtitle, avatar_hash, banner_hash,
			accent_banner, accent_border, accent_background
		FROM auth.profiles 
		WHERE id = $1`,
		snowflake,
	).Scan(
		&profile.ID, &profile.Created, &profile.Username, &profile.Displayname,
		&profile.Biography, &profile.Subtitle, &profile.AvatarHash, &profile.BannerHash,
		&profile.AccentBanner, &profile.AccentBorder, &profile.AccentBackground,
	)
	if err == pgx.ErrNoRows {
		tools.SendClientError(w, r, tools.ERROR_UNKNOWN_USER)
		return
	}
	if err != nil {
		tools.SendServerError(w, r, err)
		return
	}

	// Organize Profile
	tools.SendJSON(w, r, map[string]any{
		"id":                profile.ID,
		"created":           profile.Created,
		"username":          profile.Username,
		"displayname":       profile.Displayname,
		"biography":         profile.Biography,
		"subtitle":          profile.Subtitle,
		"avatar":            profile.AvatarHash,
		"banner":            profile.BannerHash,
		"accent_banner":     profile.AccentBanner,
		"accent_border":     profile.AccentBorder,
		"accent_background": profile.AccentBackground,
	})
}
