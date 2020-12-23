package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

type raiderIOResponse struct {
	Region         string         `json:"region"`
	Title          string         `json:"title"`
	LeaderboardURL string         `json:"leaderboard_url"`
	AffixDetails   []affixDetails `json:"affix_details"`
}

type affixDetails struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	WowheadURL  string `json:"wowhead_url"`
}

// Shoutout to raider.io for providing the free API!
func getRaiderIOData() (*raiderIOResponse, error) {
	resp, err := http.Get("https://raider.io/api/v1/mythic-plus/affixes?region=us&locale=en")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rioAffixes := new(raiderIOResponse)
	json.NewDecoder(resp.Body).Decode(rioAffixes)

	return rioAffixes, nil
}

func (r raiderIOResponse) hasAffix(str string) bool {
	for _, v := range r.AffixDetails {
		if v.Name == str {
			return true
		}
	}

	return false
}

func (r raiderIOResponse) affixDescription(str string) string {
	for _, v := range r.AffixDetails {
		if v.Name == str {
			return v.Description
		}
	}

	return ""
}

func respond(w http.ResponseWriter, m string) {
	log.Debugf("respond(): %q", m)
	fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?><Response><Message><Body>`)
	fmt.Fprintf(w, m)
	fmt.Fprintf(w, "</Body></Message></Response>")
}

var helpText string = `WoW Affix SMS!

The following commands are available:
- Current: list this week's affixes
- [AffixName]: description of a specific affix`

func main() {
	log.SetLevel(log.DebugLevel)
	log.Info("wow-affixes-sms started...")

	http.HandleFunc("/sms", func(w http.ResponseWriter, r *http.Request) {
		body := r.FormValue("Body")

		rioAffixes, err := getRaiderIOData()
		if err != nil {
			log.Fatalln(err)
		}

		if len(strings.Split(body, " ")) > 1 {
			log.Debug("request: body content more than 1 word")
			respond(w, fmt.Sprintf("Sorry! You must use 1-word commands.\n\n%s", helpText))
		} else if body == "Current" {
			log.Debug("request: current")
			respond(w, fmt.Sprintf("This week's affixes: %s", rioAffixes.Title))
		} else if rioAffixes.hasAffix(body) {
			log.Debugf("request: specific affix - %q", body)
			respond(w, fmt.Sprintf("%s: %s", body, rioAffixes.affixDescription(body)))
		} else {
			log.Debugf("request: %q", body)
			respond(w, helpText)
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
