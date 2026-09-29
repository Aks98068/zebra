// osint-cli: a small personal-recon command framework.
//
// Scope and boundaries (read this before extending it):
//   - Every command here only ever touches PUBLIC, UNAUTHENTICATED endpoints
//     (DNS resolution, Gravatar's public avatar API, HIBP's official breach
//     API, and plain profile-URL HTTP checks).
//   - Nothing here logs in, scrapes authenticated pages, or pulls private
//     data (photos, likes, comments, follower graphs). That data sits behind
//     OAuth/login walls on every major platform by design, and scraping
//     around that violates their Terms of Service — that's true no matter
//     what language or framework is used, and true whether you're checking
//     your own footprint or someone else's.
//   - The "dork" command generates search-query URLs for you to open in a
//     browser. It deliberately does NOT auto-scrape Google/Bing results,
//     since doing that at any volume gets you rate-limited/blocked and edges
//     into ToS violation territory for marginal benefit over just clicking
//     the link yourself.
//
// This file is package osint: a subcommand module meant to be imported by a
// parent CLI's command registry (e.g. via Register(Command{Name: "osint", ...})),
// not run standalone.
package osint

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------------------------
// Command registry (same pattern you showed)
// ---------------------------------------------------------------------

type Context struct {
	Timeout time.Duration
}

type Command struct {
	Name        string
	Aliases     []string
	Description string
	Usage       string
	Run         func(args []string, ctx *Context) bool
}

// subRegistry is this module's own command table, kept separate from the
// parent CLI's registry so registering "osint" as one command up top doesn't
// collide with the names of the commands nested inside it (e.g. both could
// have a command called "info" without conflict).
var subRegistry = map[string]*Command{}

func Register(cmd Command) {
	c := cmd
	subRegistry[c.Name] = &c
	for _, alias := range c.Aliases {
		subRegistry[alias] = &c
	}
}

// Dispatch runs the named sub-command (e.g. "gravatar", "hibp", "username",
// "dork") with the args that follow it. The parent CLI's "osint" command
// should call this with the args it received minus its own name — see the
// wiring example at the bottom of this file's comments.
func Dispatch(name string, args []string, ctx *Context) bool {
	cmd, ok := subRegistry[name]
	if !ok {
		fmt.Printf("unknown osint subcommand: %s\n", name)
		PrintHelp()
		return false
	}
	return cmd.Run(args, ctx)
}

// PrintHelp lists the sub-commands this module registers.
func PrintHelp() {
	fmt.Println("osint subcommands:")
	seen := map[string]bool{}
	for _, cmd := range subRegistry {
		if seen[cmd.Name] {
			continue
		}
		seen[cmd.Name] = true
		fmt.Printf("  %-10s %s\n", cmd.Name, cmd.Description)
		fmt.Printf("             usage: %s\n", cmd.Usage)
	}
}

// ---------------------------------------------------------------------
// gravatar — public avatar + (if opted-in) profile info lookup
// ---------------------------------------------------------------------

func gravatarLookup(args []string) bool {
	if len(args) == 0 {
		fmt.Println("Usage: gravatar <email>")
		return false
	}
	email := strings.ToLower(strings.TrimSpace(args[0]))
	sum := md5.Sum([]byte(email))
	hash := hex.EncodeToString(sum[:])

	client := &http.Client{Timeout: 8 * time.Second}

	avatarURL := fmt.Sprintf("https://www.gravatar.com/avatar/%s?d=404", hash)
	resp, err := client.Get(avatarURL)
	if err != nil {
		fmt.Println("error:", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("No public Gravatar registered for this email.")
		return true
	}
	fmt.Printf("Avatar : https://www.gravatar.com/avatar/%s\n", hash)

	// Gravatar also exposes an opt-in public profile JSON if the user filled
	// one out. Purely public, no auth.
	profileURL := fmt.Sprintf("https://www.gravatar.com/%s.json", hash)
	presp, err := client.Get(profileURL)
	if err == nil && presp.StatusCode == http.StatusOK {
		defer presp.Body.Close()
		var data map[string]interface{}
		if json.NewDecoder(presp.Body).Decode(&data) == nil {
			fmt.Println("Public profile JSON found:", profileURL)
		}
	}
	return true
}

// ---------------------------------------------------------------------
// hibp — Have I Been Pwned breach check (requires your own API key)
// ---------------------------------------------------------------------

type hibpBreach struct {
	Title      string `json:"Title"`
	BreachDate string `json:"BreachDate"`
}

func hibpCheck(args []string) bool {
	var email, apiKey string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-key":
			if i+1 < len(args) {
				apiKey = args[i+1]
				i++
			}
		default:
			if email == "" {
				email = args[i]
			}
		}
	}
	if email == "" || apiKey == "" {
		fmt.Println("Usage: hibp <email> -key <hibp_api_key>")
		fmt.Println("Get a key at https://haveibeenpwned.com/API/Key")
		return false
	}

	reqURL := fmt.Sprintf("https://haveibeenpwned.com/api/v3/breachedaccount/%s?truncateResponse=false", url.PathEscape(email))
	req, _ := http.NewRequest(http.MethodGet, reqURL, nil)
	req.Header.Set("hibp-api-key", apiKey)
	req.Header.Set("User-Agent", "osint-cli/1.0")

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("error:", err)
		return false
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		fmt.Println("No known breaches for this email.")
	case http.StatusOK:
		var breaches []hibpBreach
		json.NewDecoder(resp.Body).Decode(&breaches)
		for _, b := range breaches {
			fmt.Printf("- %s (%s)\n", b.Title, b.BreachDate)
		}
	case http.StatusUnauthorized:
		fmt.Println("HIBP rejected the API key.")
	default:
		fmt.Printf("Unexpected HIBP status: %d\n", resp.StatusCode)
	}
	return true
}

// ---------------------------------------------------------------------
// username — profile-URL existence check across popular platforms
// ---------------------------------------------------------------------

var usernameSites = map[string]string{
	"GitHub":    "https://github.com/%s",
	"GitLab":    "https://gitlab.com/%s",
	"Reddit":    "https://www.reddit.com/user/%s",
	"Twitter/X": "https://x.com/%s",
	"Instagram": "https://www.instagram.com/%s/",
	"TikTok":    "https://www.tiktok.com/@%s",
	"Medium":    "https://medium.com/@%s",
	"Dev.to":    "https://dev.to/%s",
	"Telegram":  "https://t.me/%s",
	"Steam":     "https://steamcommunity.com/id/%s",
	"Twitch":    "https://www.twitch.tv/%s",
	"LinkedIn":  "https://www.linkedin.com/in/%s",
}

func usernameCheck(args []string) bool {
	if len(args) == 0 {
		fmt.Println("Usage: username <handle>")
		return false
	}
	handle := args[0]
	client := &http.Client{Timeout: 8 * time.Second}

	var wg sync.WaitGroup
	var mu sync.Mutex
	for site, pattern := range usernameSites {
		wg.Add(1)
		go func(site, pattern string) {
			defer wg.Done()
			u := fmt.Sprintf(pattern, handle)
			req, _ := http.NewRequest(http.MethodGet, u, nil)
			req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; osint-cli/1.0)")
			resp, err := client.Do(req)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				fmt.Printf("%-10s error: %v\n", site, err)
				return
			}
			defer resp.Body.Close()
			status := "not found"
			if resp.StatusCode == http.StatusOK {
				status = "FOUND"
			}
			fmt.Printf("%-10s %-9s %s\n", site, status, u)
		}(site, pattern)
	}
	wg.Wait()
	return true
}

// ---------------------------------------------------------------------
// dork — generate search-engine dork URLs (you click them; nothing is
// auto-scraped, which is what keeps this the legitimate, ToS-safe version)
// ---------------------------------------------------------------------

func dorkLinks(args []string) bool {
	if len(args) == 0 {
		fmt.Println("Usage: dork <email|username>")
		return false
	}
	target := args[0]
	q := url.QueryEscape("\"" + target + "\"")

	links := []string{
		"https://www.google.com/search?q=" + q,
		"https://www.google.com/search?q=" + q + "+site%3Alinkedin.com",
		"https://www.google.com/search?q=" + q + "+site%3Agithub.com",
		"https://www.google.com/search?q=" + q + "+site%3Apastebin.com",
		"https://www.bing.com/search?q=" + q,
		"https://duckduckgo.com/?q=" + q,
	}
	fmt.Println("Open these to see what's publicly indexed for:", target)
	for _, l := range links {
		fmt.Println(" ", l)
	}
	return true
}

// ---------------------------------------------------------------------
// wiring it all up
// ---------------------------------------------------------------------

func init() {
	Register(Command{
		Name:        "gravatar",
		Description: "check public Gravatar avatar/profile for an email",
		Usage:       "gravatar <email>",
		Run: func(args []string, ctx *Context) bool {
			return gravatarLookup(args)
		},
	})

	Register(Command{
		Name:        "hibp",
		Description: "check Have I Been Pwned breach data for an email",
		Usage:       "hibp <email> -key <hibp_api_key>",
		Run: func(args []string, ctx *Context) bool {
			return hibpCheck(args)
		},
	})

	Register(Command{
		Name:        "username",
		Aliases:     []string{"user"},
		Description: "check if a username exists across popular platforms",
		Usage:       "username <handle>",
		Run: func(args []string, ctx *Context) bool {
			return usernameCheck(args)
		},
	})

	Register(Command{
		Name:        "dork",
		Description: "generate search-engine dork links for an email/username",
		Usage:       "dork <email|username>",
		Run: func(args []string, ctx *Context) bool {
			return dorkLinks(args)
		},
	})
}
