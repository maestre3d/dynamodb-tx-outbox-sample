package messaging

import (
	"context"
	logging "log"
	"os"
	"sync"

	"github.com/maestre3d/dynamodb-tx-outbox/domain/event"
	"github.com/maestre3d/dynamodb-tx-outbox/infrastructure/cloud"
	"github.com/neutrinocorp/gluon"
	"github.com/neutrinocorp/gluon/gaws"
	"github.com/rs/zerolog/log"
)

var defaultBusMu sync.Once
var DefaultEventBus *gluon.Bus

func init() {
	defaultBusMu.Do(func() {
		osLogger := logging.New(os.Stdout, "", 0)
		DefaultEventBus = gluon.NewBus("aws_sns_sqs",
			gluon.WithDriverConfiguration(gaws.SnsSqsConfig{
				AwsConfig:                 cloud.DefaultAwsConfig,
				AccountID:                 os.Getenv("NEUTRINO_AWS_ACCOUNT_ID"),
				MaxNumberOfMessagesPolled: 0,
				VisibilityTimeout:         0,
				WaitTimeSeconds:           0,
				MaxBatchPollingRetries:    0,
				FailedPollingBackoff:      0,
			}),
			gluon.WithLogger(osLogger),
			gluon.WithPublisherMiddleware(gluonProducerLogger))
		RegisterDomainEvents()
	})
}

func gluonProducerLogger(next gluon.PublisherFunc) gluon.PublisherFunc {
	return func(ctx context.Context, message *gluon.TransportMessage) (err error) {
		defer func() {
			if err != nil {
				log.Print(err)
			}
		}()

		log.Info().
			Str("topic", message.Type).
			Str("message_id", message.ID).
			Str("causation_id", message.CausationID).
			Str("correlation_id", message.CorrelationID).
			Str("subject", message.Subject).
			Msg("gluon: publishing message")
		err = next(ctx, message)
		return
	}
}

func RegisterDomainEvents() {
	// Neutrino Resource Name (NRN)
	// nrn:ORGANIZATION:PLATFORM:SERVICE
	DefaultEventBus.RegisterSchema(event.StudentRegistered{},
		gluon.WithSource("nrn:ncorp:workspaces:students"),
		gluon.WithTopic("ncorp.workspaces.iam.1.student.registered"))
}
