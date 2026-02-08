Gator is a command‑line RSS aggregator that lets you manage multiple users and RSS feeds, and browse the fetched posts.

In order for the software to work, the user will need:

1. Install Go
   - # for linux/mac
   - curl -sS https://webi.sh/golang | sh
   - # for windows
   - curl.exe https://webi.ms/golang | powershell
2. Install PostgreSQL
   - # macOS with brew
   - brew install postgresql@15
   - # Linux / WSL
   - sudo apt update
   - sudo apt install postgresql postgresql-contrib
   - Ensure Installation with psql --version
3. Create a config file
   - Create a config in your home directory `~/.gatorconfig.json`

```json
{
  "db_url": "postgres://user:password@localhost:5432/gator?sslmode=disable"
}
```

4. Install the software
   - `go install github.com/Seva-Sh/blogaggregator@latest`

These steps cover all the required initial setups.
To run the software, just type `gator ...` in your command-line and let the magic flow.

Currently the software includes the following commands

1. Register
   - `gator register username` expects a username. Creates a new user
2. Login
   - `gator login username` expects a username. Logs you in as an entered user
3. Reset
   - `gator reset` resets the session by deleting all the tables and entries
4. Users
   - `gator users` lists all the registered users
5. Agg
   - `gator agg seconds` expects a number of seconds. Fetches the RSS feeds, parses them and prints the posts to the console
6. AddFeed
   - `gator addfeed username url` expects a username and a url. Adds feed to the user
7. Feeds
   - `gator feeds` lists all the feeds for the current user
8. Follow
   - `gator follow url` expects a url. Follows the feed for the current user
9. Following
   - `gator following` lists all the feeds followed by a current user
10. Unfollow
    - `gator unfollow url` expects a url. Unfollows the url for the current user
11. Browse
    - `gator browse num` allows optional selection of browsed feeds (2 by default). Allows user to browse feeds of the current user

## Quick start

```bash
gator register <username>
gator login <username>
gator addfeed <url>
gator agg 30s
gator browse 5
```
