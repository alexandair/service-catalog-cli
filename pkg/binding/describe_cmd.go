package binding

import (
	"fmt"

	"github.com/Azure/service-catalog-cli/pkg/output"
	"github.com/Azure/service-catalog-cli/pkg/traverse"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/spf13/cobra"
)

type describeCmd struct {
	cl       *clientset.Clientset
	ns       string
	traverse bool
}

// NewDescribeCmd builds a "svc-cat describe binding" command
func NewDescribeCmd(cl *clientset.Clientset) *cobra.Command {
	describeCmd := &describeCmd{cl: cl}
	cmd := &cobra.Command{
		Use:     "binding NAME",
		Aliases: []string{"bindings", "bnd"},
		Short:   "Show details of a specific binding",
		Example: `
  svc-cat describe binding wordpress-mysql-binding
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return describeCmd.run(args)
		},
	}
	cmd.Flags().StringVarP(
		&describeCmd.ns,
		"namespace",
		"n",
		"default",
		"The namespace in which to get the binding",
	)
	cmd.Flags().BoolVarP(
		&describeCmd.traverse,
		"traverse",
		"t",
		false,
		"Whether or not to traverse from binding -> instance -> class/plan -> broker",
	)
	return cmd
}

func (c *describeCmd) run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("name is required")
	}

	key := args[0]
	return c.describe(key)
}

func (c *describeCmd) describe(name string) error {
	binding, err := retrieveByName(c.cl, c.ns, name)
	if err != nil {
		return err
	}

	output.WriteBindingDetails(binding)

	if !c.traverse {
		return nil
	}

	// Traverse from binding to instance
	inst, err := traverse.BindingToInstance(c.cl, binding)
	if err != nil {
		return fmt.Errorf("Error traversing binding to its instance (%s)", err)
	}
	logger.Printf("\n\nINSTANCE")
	t := output.NewTable()
	output.InstanceHeaders(t)
	output.AppendInstance(t, inst)
	t.Render()

	// Traverse from instance to service class and plan
	class, plan, err := traverse.InstanceToServiceClassAndPlan(c.cl, inst)
	if err != nil {
		return fmt.Errorf("Error traversing instance to its service class and plan (%s)", err)
	}
	logger.Printf("\n\nSERVICE CLASS")
	t = output.NewTable()
	output.ClusterServiceClassHeaders(t)
	output.AppendClusterServiceClass(t, class)
	t.Render()

	logger.Printf("\n\nSERVICE PLAN")
	t = output.NewTable()
	output.ClusterServicePlanHeaders(t)
	output.AppendClusterServicePlan(t, plan)
	t.Render()

	// traverse from service class to broker
	broker, err := traverse.ServiceClassToBroker(c.cl, class)
	if err != nil {
		return fmt.Errorf("Error traversing service class to broker (%s)", err)
	}
	logger.Printf("\n\nBROKER")
	t = output.NewTable()
	output.ClusterServiceBrokerHeaders(t)
	output.AppendClusterServiceBroker(t, broker)
	t.Render()

	return nil
}