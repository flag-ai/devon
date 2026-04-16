// Package web provides the embedded SPA frontend assets.
package web

import "embed"

// Dist contains the built React SPA files from web/dist/.
//
// The web/dist directory is produced by `npm run build` inside web/.
// When the frontend has not been built, the .gitkeep file is the only
// entry — that's fine, the server simply returns 404 for SPA routes.
//
//go:embed all:dist
var Dist embed.FS
