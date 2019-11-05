package azurestorage

import (
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
)

var logger = log.WithField("component", "azurestorage")

type Storage struct {
	container *storage.Container
}

func (s *Storage) Init() error {
	// enumerate azure storage environment variables
	var conf struct {
		Account   string `envconfig:"AZURE_STORAGE_ACCOUNT" required:"true"`
		Container string `envconfig:"AZURE_STORAGE_CONTAINER" required:"true"`
		AccessKey string `envconfig:"AZURE_STORAGE_KEY" required:"true"`
	}
	if err := envconfig.Process("", &conf); err != nil {
		return err
	}

	// initialize azure storage client
	logger.Debug("connecting to azure storageaccount ", conf.Account)
	cli, err := storage.NewBasicClient(conf.Account, conf.AccessKey)
	if err != nil {
		return fmt.Errorf("connecting to azure storage: %w", err)
	}
	blobCli := cli.GetBlobService()

	// verify container exists
	logger.Debug("connecting to container ", conf.Container)
	container := blobCli.GetContainerReference(conf.Container)
	if exists, err := container.Exists(); err != nil {
		return fmt.Errorf("looking up container %s: %w", conf.Container, err)
	} else if !exists {
		return fmt.Errorf("container %s not found", conf.Container)
	}

	s.container = container

	return nil
}

func (s *Storage) Write(reader io.Reader, dest string) error {

	// verify blob does not exist
	logger.Debug("preparing to write to new blob ", dest)
	blob := s.container.GetBlobReference(dest)
	if exists, err := blob.Exists(); err != nil {
		return fmt.Errorf("initializing blob: %w", err)
	} else if exists {
		return fmt.Errorf("blob %s already exists", dest)
	}

	// write to blob
	logger.Debug("writing to new blob ", dest)
	if err := blob.CreateBlockBlobFromReader(reader, &storage.PutBlobOptions{}); err != nil {
		return fmt.Errorf("writing to blob: %w", err)
	}
	logger.Debug("successfully wrote to new blob ", dest)

	return nil
}
