/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
)

var files = []string{
	"manifests/deploy/crds/submariner.io_brokers.yaml",
	"manifests/deploy/crds/submariner.io_submariners.yaml",
	"manifests/deploy/crds/submariner.io_servicediscoveries.yaml",
	"manifests/deploy/submariner/crds/submariner.io_clusters.yaml",
	"manifests/deploy/submariner/crds/submariner.io_endpoints.yaml",
	"manifests/deploy/submariner/crds/submariner.io_gateways.yaml",
	"manifests/deploy/mcsapi/crds/multicluster.x_k8s.io_serviceexports.yaml",
	"manifests/deploy/mcsapi/crds/multicluster.x_k8s.io_serviceimports.yaml",
	"manifests/config/broker/broker-admin/service_account.yaml",
	"manifests/config/broker/broker-admin/role.yaml",
	"manifests/config/broker/broker-admin/role_binding.yaml",
	"manifests/config/broker/broker-client/service_account.yaml",
	"manifests/config/broker/broker-client/role.yaml",
	"manifests/config/broker/broker-client/role_binding.yaml",
	"manifests/config/rbac/submariner-operator/service_account.yaml",
	"manifests/config/rbac/submariner-operator/role.yaml",
	"manifests/config/rbac/submariner-operator/role_binding.yaml",
	"manifests/config/rbac/submariner-operator/cluster_role.yaml",
	"manifests/config/rbac/submariner-operator/cluster_role_binding.yaml",
	"manifests/config/rbac/submariner-gateway/service_account.yaml",
	"manifests/config/rbac/submariner-gateway/role.yaml",
	"manifests/config/rbac/submariner-gateway/role_binding.yaml",
	"manifests/config/rbac/submariner-gateway/cluster_role.yaml",
	"manifests/config/rbac/submariner-gateway/cluster_role_binding.yaml",
	"manifests/config/rbac/submariner-route-agent/service_account.yaml",
	"manifests/config/rbac/submariner-route-agent/role.yaml",
	"manifests/config/rbac/submariner-route-agent/role_binding.yaml",
	"manifests/config/rbac/submariner-route-agent/cluster_role.yaml",
	"manifests/config/rbac/submariner-route-agent/cluster_role_binding.yaml",
	"manifests/config/rbac/submariner-globalnet/service_account.yaml",
	"manifests/config/rbac/submariner-globalnet/role.yaml",
	"manifests/config/rbac/submariner-globalnet/role_binding.yaml",
	"manifests/config/rbac/submariner-globalnet/cluster_role.yaml",
	"manifests/config/rbac/submariner-globalnet/cluster_role_binding.yaml",
	"manifests/config/rbac/lighthouse-agent/service_account.yaml",
	"manifests/config/rbac/lighthouse-agent/cluster_role.yaml",
	"manifests/config/rbac/lighthouse-agent/cluster_role_binding.yaml",
	"manifests/config/rbac/lighthouse-coredns/service_account.yaml",
	"manifests/config/rbac/lighthouse-coredns/cluster_role.yaml",
	"manifests/config/rbac/lighthouse-coredns/cluster_role_binding.yaml",
	"manifests/config/rbac/networkplugin_syncer/service_account.yaml",
	"manifests/config/rbac/networkplugin_syncer/cluster_role.yaml",
	"manifests/config/rbac/networkplugin_syncer/cluster_role_binding.yaml",
	"manifests/fix/crds/discovery.k8s.io_endpointslices.yaml",
	"manifests/fix/calico/calico-config.yaml",
	"manifests/fix/calico/calico-kube-controllers-clusterrole.yaml",
	"manifests/fix/calico/calico-kube-controllers-clusterrolebinding.yaml",
	"manifests/fix/calico/calico-kube-controllers-pdb.yaml",
	"manifests/fix/calico/calico-kube-controllers-sa.yaml",
	"manifests/fix/calico/calico-kube-controllers.yaml",
	"manifests/fix/calico/calico-node-clusterrole.yaml",
	"manifests/fix/calico/calico-node-clusterrolebinding.yaml",
	"manifests/fix/calico/calico-node-sa.yaml",
	"manifests/fix/calico/calico-node.yaml",
	"manifests/fix/calico/crds/crd.projectcalico.org_bgpconfigurations.yaml",
	"manifests/fix/calico/crds/crd.projectcalico.org_bgppeers.yaml",
	"manifests/fix/calico/crds/crd.projectcalico.org_blockaffinities.yaml",
	"manifests/fix/calico/crds/crd.projectcalico.org_clusterinformations.yaml",
	"manifests/fix/calico/crds/crd.projectcalico.org_felixconfigurations.yaml",
	"manifests/fix/calico/crds/crd.projectcalico.org_globalnetworkpolicies.yaml",
	"manifests/fix/calico/crds/crd.projectcalico.org_globalnetworksets.yaml",
	"manifests/fix/calico/crds/crd.projectcalico.org_hostendpoints.yaml",
	"manifests/fix/calico/crds/crd.projectcalico.org_ipamblocks.yaml",
	"manifests/fix/calico/crds/crd.projectcalico.org_ipamconfigs.yaml",
	"manifests/fix/calico/crds/crd.projectcalico.org_ipamhandles.yaml",
	"manifests/fix/calico/crds/crd.projectcalico.org_ippools.yaml",
	"manifests/fix/calico/crds/crd.projectcalico.org_kubecontrollersconfigurations.yaml",
	"manifests/fix/calico/crds/crd.projectcalico.org_networkpolicies.yaml",
	"manifests/fix/calico/crds/crd.projectcalico.org_networksets.yaml",
}

// Reads all .yaml files in the crdDirectory
// and encodes them as constants in crdyamls.go
func main() {
	if len(os.Args) < 3 {
		fmt.Println("yamls2go needs two arguments, the base directory containing the YAML files, and the target directory")
		os.Exit(1)
	}

	yamlsDirectory := os.Args[1]
	goDirectory := os.Args[2]

	fmt.Println("Generating yamls.go")
	out, err := os.Create(goDirectory + string(os.PathSeparator) + "yamls.go")
	panicOnErr(err)

	_, err = out.WriteString("// This file is auto-generated by yamls2go.go\n" +
		"package embeddedyamls\n\nconst (\n")
	panicOnErr(err)

	// Raw string literals can’t contain backticks (which enclose the literals)
	// and there’s no way to escape them. Some YAML files we need to embed include
	// backticks... To work around this, without having to deal with all the
	// subtleties of wrapping arbitrary YAML in interpreted string literals, we
	// split raw string literals when we encounter backticks in the source YAML,
	// and add the backtick-enclosed string as an interpreted string:
	//
	// `resourceLock:
	//    description: The type of resource object that is used for locking
	//      during leader election. Supported options are ` + "`configmaps`" + ` (default)
	//      and ` + "`endpoints`" + `.
	//    type: string`

	re := regexp.MustCompile("`([^`]*)`")
	reNS := regexp.MustCompile(`(?s)\s*namespace:\s*placeholder\s*`)

	for _, f := range files {
		_, err = out.WriteString("\t" + constName(f) + " = `")
		panicOnErr(err)

		fmt.Println(f)
		contents, err := ioutil.ReadFile(path.Join(yamlsDirectory, f))
		panicOnErr(err)

		_, err = out.Write(re.ReplaceAll(reNS.ReplaceAll(contents, []byte("\n")),
			[]byte("` + \"`$1`\" + `")))
		panicOnErr(err)

		_, err = out.WriteString("`\n")
		panicOnErr(err)
	}
	_, err = out.WriteString(")\n")
	panicOnErr(err)

	err = out.Close()
	panicOnErr(err)
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}

func constName(filename string) string {
	return strings.Title(strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(filename,
				"-", "_"),
			".", "_"),
		"/", "_"))
}
