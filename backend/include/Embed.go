package include

import "embed"

//go:embed archives/geolocation.kani.gz
var ArchiveGeolocation []byte

//go:embed templates/*.html
var EmailTemplates embed.FS

//go:embed schema.sql
var DatabaseSchema string
