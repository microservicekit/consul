package idp

import (
	"github.com/hashicorp/consul/command/flags"
	"github.com/mitchellh/cli"
)

func New() *cmd {
	return &cmd{}
}

type cmd struct{}

func (c *cmd) Run(args []string) int {
	return cli.RunResultHelp
}

func (c *cmd) Synopsis() string {
	return synopsis
}

func (c *cmd) Help() string {
	return flags.Usage(help, nil)
}

const synopsis = "Manage Consul's ACL Identity Providers"
const help = `
Usage: consul acl idp <subcommand> [options] [args]

  This command has subcommands for managing Consul's ACL Identity Providers.
  Here are some simple examples, and more detailed examples are available in
  the subcommands or the documentation.

  Create a new identity provider:

    $ consul acl idp create -type "kubernetes" \
                            -name "my-idp" \
                            -description "This is an example kube idp" \
                            -kubernetes-host "https://apiserver.example.com:8443" \
                            -kubernetes-ca-file /path/to/kube.ca.crt \
                            -kubernetes-service-account-jwt "JWT_CONTENTS"

  List all identity providers:

    $ consul acl idp list

  Update all editable fields of the identity provider:

    $ consul acl idp update -name "my-idp" \
                            -description "new description" \
                            -kubernetes-host "https://new-apiserver.example.com:8443" \
                            -kubernetes-ca-file /path/to/new-kube.ca.crt \
                            -kubernetes-service-account-jwt "NEW_JWT_CONTENTS"

  Read an identity provider:

    $ consul acl idp read -name my-idp

  Delete an identity provider:

    $ consul acl idp delete -name "my-idp"

  For more examples, ask for subcommand help or view the documentation.
`
