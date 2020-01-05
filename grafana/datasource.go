package grafana

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"

	"git.bouvet.no/t16r/backup-tool/grafana/api"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
)

var logger = logrus.WithField("component", "grafana-datasource")

type DataSource struct {
	client api.Client
}

// Init initializes a client connection to Grafana.
func (ds *DataSource) Init() error {
	// check for required env vars
	var conf struct {
		URL   string `envconfig:"GRAFANA_URL" required:"true"`
		Token string `envconfig:"GRAFANA_TOKEN" required:"true"`
	}
	if err := envconfig.Process("", &conf); err != nil {
		return err
	}

	// connect to Grafana
	cli, err := api.NewClient(conf.URL, conf.Token)
	if err != nil {
		return fmt.Errorf("connecting to Grafana: %w", err)
	}
	ds.client = cli

	return nil
}

// Archive fetches the dashboards stored in Grafana's database and archives
// them in a gzipped tarball. It returns a reader linked to the bytes of the
// archive and a (hopefully nil) error.
func (ds *DataSource) Archive() (io.Reader, error) {
	logger.Info("archiving dashboards from Grafana's database")

	// get a listing of all dashboards in Grafana
	dashboards, err := ds.client.SearchDashboards()
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
		bs, err := ds.client.GetDashboard(dash.UID)
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
