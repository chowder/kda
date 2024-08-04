package main

import (
	"github.com/chowder/kda/server"
	"github.com/chowder/kda/tokensource"
	"github.com/chowder/kda/validator"
	"github.com/urfave/cli/v2"
	"log"
	"net/url"
	"os"
)

const (
	BackendFlag        = "backend"
	AddressFlag        = "address"
	NamespaceFlag      = "namespace"
	ServiceAccountFlag = "serviceAccount"
	HtpasswdFlag       = "htpasswd"
)

func main() {
	app := cli.NewApp()

	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:     BackendFlag,
			Usage:    "URL of the backend service",
			Required: true,
		},
		&cli.StringFlag{
			Name:  AddressFlag,
			Usage: "Address of the reverse proxy",
			Value: ":8080",
		},
		&cli.StringFlag{
			Name:  NamespaceFlag,
			Usage: "Namespace of the Kubernetes Service Account",
			Value: "default",
		},
		&cli.StringFlag{
			Name:     ServiceAccountFlag,
			Usage:    "Kubernetes Service Account",
			Required: true,
		},
		&cli.StringFlag{
			Name:     HtpasswdFlag,
			Usage:    "Path to htpasswd file",
			Required: true,
		},
	}

	app.Action = entrypoint

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func entrypoint(c *cli.Context) error {
	backendURL, err := url.Parse(c.String(BackendFlag))
	if err != nil {
		return cli.Exit("Failed to parse backend URL: "+err.Error(), 1)
	}

	ns := c.String(NamespaceFlag)
	sa := c.String(ServiceAccountFlag)
	ts, err := tokensource.NewKubernetesTokenSource(ns, sa)
	if err != nil {
		return cli.Exit("Failed to create token source: "+err.Error(), 1)
	}

	htp := c.String(HtpasswdFlag)
	v, err := validator.NewHtpasswdValidator(htp)
	if err != nil {
		return cli.Exit("Failed to create validator: "+err.Error(), 1)
	}

	s := server.NewServer(ts, v)
	addr := c.String(AddressFlag)
	err = s.Serve(addr, backendURL)
	if err != nil {
		return cli.Exit("Reverse proxy server failed: "+err.Error(), 1)
	}

	return nil
}
