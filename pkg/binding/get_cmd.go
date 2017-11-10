package binding

import (
	"fmt"

	"github.com/Azure/service-catalog-cli/pkg/output"
	"github.com/Azure/service-catalog-cli/pkg/traverse"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

type bindingGetCmd struct {
	cl *clientset.Clientset
	ns string
}

func (b *bindingGetCmd) run(name string) error {
	binding, err := b.cl.Servicecatalog().ServiceBindings(b.ns).Get(name, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("Error getting binding (%s)", err)
	}
	t := output.NewTable()
	output.BindingHeaders(t)
	output.AppendBinding(t, binding)
	t.Render()
	inst, err := traverse.BindingToInstance(b.cl, binding)
	if err != nil {
		return fmt.Errorf("Error traversing binding to its instance (%s)", err)
	}
	logger.Printf("\n\nINSTANCE")
	t = output.NewTable()
	output.InstanceHeaders(t)
	output.AppendInstance(t, inst)
	t.Render()

	logger.Printf("\n\nSERVICE CLASS AND SERVICE PLAN")
	class, _, err := traverse.InstanceToServiceClassAndPlan(b.cl, inst)
	if err != nil {
		return fmt.Errorf("Error traversing instance to its service class and plan (%s)", err)
	}
	t = output.NewTable()
	output.ClusterServiceClassHeaders(t)
	output.AppendClusterServiceClass(t, class)
	t.Render()
	// t = output.NewTable()
	// output.ServicePlanHeaders(t)
	// output.AppendServicePlan(t)
	// t.Render()

	return nil
}

func newBindingGetCmd(cl *clientset.Clientset) *cobra.Command {
	getCmd := bindingGetCmd{cl: cl}
	rootCmd := &cobra.Command{
		Use:   "get",
		Short: "svc-cat binding get -n <namespace> <binding name>",
		Long:  "Get a specific binding along with the instance that it points to",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("Missing binding name")
			}
			bindingName := args[0]
			return getCmd.run(bindingName)
		},
	}

	rootCmd.Flags().StringVarP(
		&getCmd.ns,
		"namespace",
		"n",
		"default",
		"The namespace from which to get the binding",
	)
	return rootCmd
}