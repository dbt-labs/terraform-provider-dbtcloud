package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"

	dbt_cloud_new "github.com/gthesheep/terraform-provider-dbt-cloud/dbt_cloud"
	"github.com/gthesheep/terraform-provider-dbt-cloud/pkg/provider"
)

// Provider documentation generation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name dbt-cloud

func main() {
	sdkProvider, err := tf5to6server.UpgradeServer(
		context.Background(),
		provider.Provider().GRPCProvider,
	)
	if err != nil {
		log.Fatal(err)
	}

	providers := []func() tfprotov6.ProviderServer{
		// Old provider
		func() tfprotov6.ProviderServer {
			return sdkProvider
		},
		// New provider
		func() tfprotov6.ProviderServer {
			return providerserver.NewProtocol6(dbt_cloud_new.New())()
		},
	}

	muxServer, err := tf6muxserver.NewMuxServer(context.Background(), providers...)
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf6server.ServeOpt
	err = tf6server.Serve(
		"registry.terraform.io/gthesheep/dbt-cloud",
		muxServer.ProviderServer,
		serveOpts...,
	)
	if err != nil {
		log.Fatal(err)
	}
}
