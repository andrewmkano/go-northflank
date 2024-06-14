package dictionary

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/leapkit/core/render"
)

var (
	wordsAPIHost = "wordsapiv1.p.rapidapi.com"
	wordsAPIKey  = cmp.Or(os.Getenv("RAPID_API_KEY"), "-")
)

type (
	result struct {
		Definition   string
		PartOfSpeech string
		Synonyms     []string
		Types        []string
	}

	results []result

	syllables struct {
		Count int
		List  []string
	}

	WordsAPIResponse struct {
		Word      string
		Results   results
		Syllables syllables
	}
)

// AllDefinitions provides all definitions in a single string.
func (rs results) AllDefinitions() string {
	if len(rs) == 1 {
		return rs[0].Definition
	}

	var groupedDefinition string
	for idx, definition := range rs {
		groupedDefinition += fmt.Sprintf("%d. %s \n", idx, definition)
	}

	return groupedDefinition
}

func RandomWord(w http.ResponseWriter, r *http.Request) {
	rw := render.FromCtx(r.Context())

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s/words/?random=true", wordsAPIHost), nil)
	if err != nil {
		errMsg := "unable to generate request to obtain random word: %v"
		slog.ErrorContext(r.Context(), errMsg, err)

		http.Error(w, fmt.Sprintf(errMsg, err.Error()), http.StatusServiceUnavailable)

		return
	}

	req.Header.Add("x-rapidapi-key", wordsAPIKey)
	req.Header.Add("x-rapidapi-host", wordsAPIHost)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		errMsg := "failed to get random word from API: %v"
		slog.ErrorContext(r.Context(), errMsg, err)

		http.Error(w, fmt.Sprintf(errMsg, err.Error()), http.StatusServiceUnavailable)

		return
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		errMsg := "failed to read API response body: %v"
		slog.ErrorContext(r.Context(), errMsg, err)

		http.Error(w, fmt.Sprintf(errMsg, err.Error()), http.StatusServiceUnavailable)

		return
	}

	apiRes := WordsAPIResponse{}
	if err := json.Unmarshal(body, &apiRes); err != nil {
		errMsg := "failed to unmarshall API response: %v"
		slog.ErrorContext(r.Context(), errMsg, err)

		http.Error(w, fmt.Sprintf(errMsg, err.Error()), http.StatusServiceUnavailable)

		return
	}

	rw.Set("res", apiRes)

	err = rw.Render("dictionary/new_word.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
