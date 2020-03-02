package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/retailcrm/api-client-go/v5"
)

var (
	tokenCounter         uint32
	credentialsTransport = []string{
		"/api/integration-modules/{code}",
		"/api/integration-modules/{code}/edit",
	}
	markdownSymbols = []string{"*", "_", "`", "["}
)

// GenerateToken function
func GenerateToken() string {
	c := atomic.AddUint32(&tokenCounter, 1)

	return fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d%d", time.Now().UnixNano(), c))))
}

func getAPIClient(url, key string) (*v5.Client, error, int) {
	client := v5.New(url, key)
	client.Debug = config.Debug

	cr, status, e := client.APICredentials()
	if e.RuntimeErr != nil {
		logger.Error(url, status, e.RuntimeErr, cr)
		return nil, e.RuntimeErr, http.StatusInternalServerError

	}

	if !cr.Success {
		logger.Error(url, status, e.ApiErr, cr)
		return nil, errors.New(getLocalizedMessage("incorrect_url_key")), http.StatusBadRequest
	}

	if res := checkCredentials(cr.Credentials); len(res) != 0 {
		logger.Error(url, status, res)
		return nil,
			errors.New(
				getLocalizedTemplateMessage(
					"missing_credentials",
					map[string]interface{}{
						"Credentials": strings.Join(res, ", "),
					},
				),
			),
			http.StatusBadRequest
	}

	return client, nil, 0
}

func checkCredentials(credential []string) []string {
	rc := make([]string, len(credentialsTransport))
	copy(rc, credentialsTransport)

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

//UploadUserAvatar function
func UploadUserAvatar(url string) (picURLs3 string, err error) {
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			config.ConfigAWS.AccessKeyID,
			config.ConfigAWS.SecretAccessKey,
			""),
		Region: aws.String(config.ConfigAWS.Region),
	}

	s := session.Must(session.NewSession(s3Config))
	uploader := s3manager.NewUploader(s)

	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return "", errors.New(fmt.Sprintf("get: %v code: %v", url, resp.StatusCode))
	}

	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(config.ConfigAWS.Bucket),
		Key:         aws.String(fmt.Sprintf("%v/%v.jpg", config.ConfigAWS.FolderName, GenerateToken())),
		Body:        resp.Body,
		ContentType: aws.String(config.ConfigAWS.ContentType),
		ACL:         aws.String("public-read"),
	})
	if err != nil {
		return
	}

	picURLs3 = result.Location

	return
}

func getChannelSettingsHash() (hash string, err error) {
	res, err := json.Marshal(getChannelSettings())

	h := sha1.New()
	h.Write(res)
	hash = fmt.Sprintf("%x", h.Sum(nil))

	return
}

func replaceMarkdownSymbols(s string) string {
	for _, v := range markdownSymbols {
		s = strings.Replace(s, v, "\\"+v, -1)
	}

	return s
}

// shouldMessageBeIgnored returns true if message should not be processed at all
func shouldMessageBeIgnored(m *tgbotapi.Message) bool {
	if m.NewChatMembers != nil ||
		m.LeftChatMember != nil ||
		m.NewChatTitle != "" ||
		m.NewChatPhoto != nil ||
		m.DeleteChatPhoto ||
		m.GroupChatCreated {
		return true
	}

	return false
}
