package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/retailcrm/api-client-go/v5"
	"golang.org/x/text/language"
)

func setLocale(al string) {
	tag, _ := language.MatchStrings(matcher, al)
	localizer = i18n.NewLocalizer(bundle, tag.String())
}

func getLocale() map[string]string {
	return map[string]string{
		"ButtonSave":  getLocalizedMessage("button_save"),
		"ApiKey":      getLocalizedMessage("api_key"),
		"TabSettings": getLocalizedMessage("tab_settings"),
		"TabBots":     getLocalizedMessage("tab_bots"),
		"TableName":   getLocalizedMessage("table_name"),
		"TableToken":  getLocalizedMessage("table_token"),
		"AddBot":      getLocalizedMessage("add_bot"),
		"TableDelete": getLocalizedMessage("table_delete"),
		"Title":       getLocalizedMessage("title"),
	}
}

func getLocalizedMessage(messageID string) string {
	return localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: messageID})
}

// GenerateToken function
func GenerateToken() string {
	c := atomic.AddUint32(&tokenCounter, 1)

	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d%d", time.Now().UnixNano(), c))))
}

func getAPIClient(url, key string) (*v5.Client, error, int) {
	client := v5.New(url, key)

	cr, status, e := client.APICredentials()
	if e.RuntimeErr != nil {
		logger.Error(url, status, e.RuntimeErr, cr)
		return nil, errors.New(getLocalizedMessage("not_found_account")), http.StatusInternalServerError

	}

	if !cr.Success {
		logger.Error(url, status, e.ApiErr, cr)
		return nil, errors.New(getLocalizedMessage("incorrect_url_key")), http.StatusBadRequest
	}

	if res := checkCredentials(cr.Credentials); len(res) != 0 {
		logger.Error(url, status, res)
		return nil,
			errors.New(localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "missing_credentials",
				TemplateData: map[string]interface{}{
					"Credentials": strings.Join(res, ", "),
				},
			})),
			http.StatusBadRequest
	}

	return client, nil, 0
}

func checkCredentials(credential []string) []string {
	rc := config.Credentials

	for _, vc := range credential {
		for kn, vn := range rc {
			if vn == vc {
				if len(rc) == 1 {
					rc = rc[:0]
					break
				}
				rc = append(rc[:kn], rc[kn+1:]...)
			}
		}
	}

	return rc
}
