package bindingrule

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

const synopsis = "Manage Consul's ACL Binding Rules"
const help = `
Usage: consul acl binding-rule <subcommand> [options] [args]

  This command has subcommands for managing Consul's ACL Binding Rules.
  Here are some simple examples, and more detailed examples are available
  in the subcommands or the documentation.

  Create a new binding rules:

      $ consul acl binding-rule create \
             -idp-name=minikube \
             -role-name='k8s-${serviceaccount.name}' \
             -match-selector='serviceaccount.namespace=default,serviceaccount.name=web' \
             -match-selector='serviceaccount.namespace=default,serviceaccount.name=db'

  List all binding rules:

      $ consul acl binding-rule list

  Update a binding rule:

      $ consul acl binding-rule update -id=43cb72df-9c6f-4315-ac8a-01a9d98155ef \
             -role-name='k8s-${serviceaccount.name}'

  Read a binding rule:

      $ consul acl binding-rule read -id 0479e93e-091c-4475-9b06-79a004765c24

  Delete a binding rule:

      $ consul acl binding-rule delete -name -id 0479e93e-091c-4475-9b06-79a004765c24

  For more examples, ask for subcommand help or view the documentation.
`
