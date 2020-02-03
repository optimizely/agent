package integrations

import (
	"bytes"
	"io"
	"text/template"

	"github.com/optimizely/go-sdk/pkg/utils"
	"github.com/rs/zerolog/log"
)

type TemplateListener struct {
	tpl       *template.Template
	requester *utils.HTTPRequester
	headers   []utils.Header
	url       string
}

func NewTemplateListener(requester *utils.HTTPRequester, filename string, url string) *TemplateListener {
	tpl, err := template.ParseFiles(filename)
	if err != nil {
		log.Error().Err(err).Msg("error loading template")
		return &TemplateListener{}
	}

	headers := []utils.Header{{"Content-Type", "application/json"}, {"Accept", "application/json"}}

	return &TemplateListener{tpl: tpl, requester: requester, url: url, headers: headers}

}

func (l *TemplateListener) parse(message interface{}) (io.Reader, error) {
	buf := new(bytes.Buffer)
	err := l.tpl.Execute(buf, message)
	if err != nil {
		return buf, err
	}

	return buf, err
}

func (l *TemplateListener) Listen(message interface{}) {
	body, err := l.parse(message)
	if err != nil {
		log.Error().Err(err).Msg("error parsing request")
	}

	_, _, _, err = l.requester.Do(l.url, "POST", body, l.headers)
	if err != nil {
		log.Error().Err(err).Msg("error submitting request")
	}
}
