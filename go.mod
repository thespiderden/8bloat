module bloat

go 1.13

require (
	github.com/gorilla/mux v1.7.3
	github.com/mattn/go-sqlite3 v2.0.2+incompatible // indirect
	mastodon v0.0.0-00010101000000-000000000000
)

replace mastodon => ./mastodon
