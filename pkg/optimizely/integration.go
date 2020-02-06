package optimizely

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/optimizely/go-sdk/pkg/notification"
	"github.com/optimizely/go-sdk/pkg/registry"
	"github.com/optimizely/go-sdk/pkg/utils"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type Message struct {
	Env     map[string]string
	Message interface{}
}

const (
	NOTIFICATION = iota
	LOG
)

var Env = map[string]string{}

func init() {
	viper.SetEnvPrefix("optimizely")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		Env[pair[0]] = pair[1]
	}
}

type Integration struct {
	class   int
	notType notification.Type
	tplPath string
	urlConf string
	headers map[string]string
}

var requestor = utils.NewHTTPRequester(utils.Headers()) // TODO have a global requestor

var defaultIntegrations = map[string][]Integration{
	"amplitude": {
		{
			class:   NOTIFICATION,
			notType: notification.Track,
			tplPath: "./templates/track/amplitude_body.tmpl",
			urlConf: "amplitude.url",
		},
	},
	"slack": {
		{
			class:   NOTIFICATION,
			notType: notification.Decision,
			tplPath: "./templates/decision/slack_body.tmpl",
			urlConf: "slack.url",
		},
		{
			class:   LOG,
			notType: notification.Decision,
			tplPath: "./templates/log/slack_body.tmpl",
			urlConf: "slack.url",
		},
	},
	"aws_sqs": {
		{
			class:   NOTIFICATION,
			notType: notification.Decision,
			tplPath: "./templates/decision/sqs_body.tmpl",
			urlConf: "aws.sqs.url",
			headers: map[string]string{
				"Content-Type": "application/x-www-form-urlencoded",
				"Accept":       "application/x-www-form-urlencoded",
			},
		},
	},
	"pagerduty": {
		{
			class:   LOG,
			tplPath: "./templates/log/pd_body.tmpl",
			urlConf: "pd.url",
		},
	},
}

func AddIntegration(sdkKey string, name string) error {
	ins, ok := defaultIntegrations[name]
	if !ok {
		return fmt.Errorf(`"%s" integration not supported`, name)
	}

	for _, in := range ins {
		err := addIntegration(sdkKey, in)
		if err != nil {
			return err
		}
	}

	return nil
}

func addIntegration(sdkKey string, in Integration) error {
	url := viper.GetString(in.urlConf)
	if url == "" {
		return fmt.Errorf("cannot find integration url: %s", in.urlConf)
	}

	filename := in.tplPath
	if filename == "" {
		return errors.New("must specify template file")
	}

	dl := NewTemplateListener(requestor, filename, url, in.headers)

	if in.class == LOG {
		id, err := LogManager.Add(dl.Listen)
		if err != nil {
			log.Debug().Int("handlerId", id).Msg("successfully added integration")
		}

		return err
	}

	nc := registry.GetNotificationCenter(sdkKey)
	id, err := nc.AddHandler(in.notType, dl.Listen)
	if err != nil {
		log.Debug().Int("handlerId", id).Msg("successfully added integration")
	}

	return err
}
