package broker

import "errors"

type Config struct {
	QueueMain        string
	QueueInvalid     string
	QueueError       string
	RepositoryConfig *RepositoryConfig
}

type RepositoryConfig struct {
	Backend          string
	DynamoDBTLS      bool
	DynamoDBRegion   string
	DynamoDBEndpoint string
	DynamoDBTable    string
}

func (c *Config) Validate() error {
	if c.QueueError == "" {
		return errors.New("error queue name is undefined")
	}
	if c.QueueInvalid == "" {
		return errors.New("invalid queue name is undefined")
	}
	if c.QueueMain == "" {
		return errors.New("main queue name is undefined")
	}
	if c.RepositoryConfig == nil || c.RepositoryConfig.Backend == "" {
		return errors.New("repository config is undefined")
	} else {
		if c.RepositoryConfig.Backend == "dynamodb" && c.RepositoryConfig.DynamoDBTable == "" {
			return errors.New("repository config is missing details")
		}
	}
	return nil
}
