package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"git.bouvet.no/t16r/backup-tool/azurestorage"
	"git.bouvet.no/t16r/backup-tool/grafana"
	"git.bouvet.no/t16r/backup-tool/localstorage"
	"github.com/sirupsen/logrus"
)

func main() {
	// initialize logging
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
	logger := logrus.WithField("component", "main")

	// parse command line flags
	flgStorType := flag.String("storage.type", "", "destination storage service to use")
	flgDataSource := flag.String("datasource", "", "datasource to retrieve data from")
	flgPathPrefix := flag.String("output.prefix", "", "path prefix to use, if any, in front of yyyy/mm/dd/filename")
	flag.Parse()

	// set the datasource
	var src DataSource
	switch *flgDataSource {
	case "file":
		src = new(file.DataSource)
	case "grafana":
		src = new(grafana.DataSource)
	default:
		failopts("valid options for -datasource are: [grafana]")
	}
	if err := src.Init(); err != nil {
		logger.Fatalf("initializing datasource: %v", err)
	}

	// set the storage type
	var stor Storage
	switch *flgStorType {
	case "local":
		stor = new(localstorage.Storage)
	case "azure":
		stor = new(azurestorage.Storage)
	default:
		failopts("valid options for -storage.type are: [azure local]")
	}
	if err := stor.Init(); err != nil {
		logger.Fatal("initializing storage: %v", err)
	}

	// create the archive
	logger.Info("archiving the datasource")
	reader, err := src.Archive()
	if err != nil {
		logger.Fatal(err)
	}

	// write the archive to our storage
	fullPath := filepath.Join(storagePath(*flgPathPrefix), "archive.tgz")
	logger.Infof("writing archive to %s storage as %s", *flgStorType, fullPath)
	if err := stor.Write(reader, fullPath); err != nil {
		logger.Fatalf("writing archive: %v", err)
	}
	logger.Info("done")
}

type DataSource interface {
	Init() error
	Archive() (string, io.Reader, error)
}

type Storage interface {
	Init() error
	Write(io.Reader, string) error
}

func failopts(msg string) {
	fmt.Fprintf(os.Stderr, "ERROR - %s\n", msg)
	os.Exit(1)
}

func storagePath(prefix string) string {
	now := time.Now()
	path := filepath.Join(
		prefix,
		strconv.Itoa(now.Year()),
		strconv.Itoa(int(now.Month())),
		strconv.Itoa(now.Day()),
	)
	return path
}
