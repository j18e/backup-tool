package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"git.bouvet.no/t16r/backup-tool/azurestorage"
	"git.bouvet.no/t16r/backup-tool/grafana"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

type StorageType int

const (
	LOCAL StorageType = iota
	AZURE
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
	flag.Parse()

	// check the storage type
	var storType StorageType
	switch *flgStorType {
	case "local":
		storType = LOCAL
	case "azure":
		storType = AZURE
	default:
		fmt.Fprintf(os.Stderr, "ERROR - valid -storage.type options: [azure local]")
		os.Exit(1)
	}

	// get dashboards from Grafana
	buf, err := grafanaCmd()
	if err != nil {
		logger.Fatal(err)
	}

	now := time.Now()
	outFile := fmt.Sprintf("%d/%d/%d/dashboards-%s.tgz", now.Year(), now.Month(), now.Day(), now.Format("1504"))

	logger.Infof("writing archive to %s storage as %s", *flgStorType, outFile)
	if err := putArchive(buf, storType, outFile); err != nil {
		logger.Fatalf("writing archive: %v", err)
	}
	logger.Info("done")
}

// putArchive writes the newly created archive to the chosen storage service.
func putArchive(reader io.Reader, storType StorageType, fileName string) error {
	switch storType {
	case LOCAL:
		path := filepath.Dir(fileName)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("creating directory path %s: %v", path, err)
		}
		file, err := os.Create(fileName)
		if err != nil {
			return fmt.Errorf("creating file %s: %w", fileName, err)
		}
		defer file.Close()
		if _, err := io.Copy(file, reader); err != nil {
			return fmt.Errorf("writing to file %s: %w", fileName, err)
		}
		return nil
	case AZURE:
		if err := azurestorage.PutAzureBlob(reader, fileName); err != nil {
			return fmt.Errorf("putting azure blob: %w", err)
		}
		return nil
	}
	return nil
}

// grafanaDashboards initializes a client connection to Grafana, fetches all of
// the dashboards stored in its database and creates a compressed tar archive
// from them.
func grafanaCmd() (io.Reader, error) {
	logger := logrus.WithField("component", "grafana-cmd")

	logger.Info("archiving dashboards from Grafana's database")

	// enumerate azure storage environment variables
	var conf struct {
		URL   string `envconfig:"GRAFANA_URL" required:"true"`
		Token string `envconfig:"GRAFANA_TOKEN" required:"true"`
	}
	if err := envconfig.Process("", &conf); err != nil {
		return nil, err
	}

	// connect to Grafana
	cli, err := grafana.NewClient(conf.URL, conf.Token)
	if err != nil {
		return nil, fmt.Errorf("connecting to Grafana: %w", err)
	}

	// get a listing of all dashboards in Grafana
	dashboards, err := cli.SearchDashboards()
	if err != nil {
		return nil, fmt.Errorf("searching dashboards: %w", err)
	}
	logger.Debugf("found %d dashboards", len(dashboards))

	// create output file, tar and gzip writers
	buf := new(bytes.Buffer)
	gzw := gzip.NewWriter(buf)
	defer gzw.Close()
	tw := tar.NewWriter(gzw)
	defer tw.Close()

	// go through the dashboards
	failed := 0
	for _, dash := range dashboards {
		bs, err := cli.GetDashboard(dash.UID)
		if err != nil {
			logger.Errorf("getting dashboard %s: %v. Skipping...", dash.UID, err)
			failed++
			continue
		}

		hdr := &tar.Header{
			Name: dash.UID + ".json",
			Mode: 0644,
			Size: int64(len(bs)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return nil, fmt.Errorf("writing tar header: %w", err)
		}
		if _, err := tw.Write(bs); err != nil {
			return nil, fmt.Errorf("writing byte slice to tar archive: %w", err)
		}
	}
	logger.Infof("successfully archived %d of %d dashboards", len(dashboards)-failed, len(dashboards))
	return buf, nil
}
