package cmd

import (
	"bytes"
	//"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/helm/helm-classic/codec"
	"github.com/spf13/cobra"

	"k8s.io/kubernetes/pkg/api/meta"
	kapi "k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/kubectl"
	kubectlcmd "k8s.io/kubernetes/pkg/kubectl/cmd"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
	"k8s.io/kubernetes/pkg/runtime"
	utilerrors "k8s.io/kubernetes/pkg/util/errors"
	//"k8s.io/kubernetes/pkg/watch"
)

// GetOptions is the start of the data required to perform the operation.  As new fields are added, add them here instead of
// referencing the cmd.Flags()
type ExportIntoChartOptions struct {
	kubectlcmd.GetOptions
}

//type GetOptions struct {
//	Filenames []string
//	Recursive bool
//
//	Raw string
//}

const (
	get_long = `Display one or many resources.

` + kubectl.PossibleResourceTypes + `

By specifying the output as 'template' and providing a Go template as the value
of the --template flag, you can filter the attributes of the fetched resource(s).`
	get_example = `# List all pods in ps output format.
kubectl get pods

# List all pods in ps output format with more information (such as node name).
kubectl get pods -o wide

# List a single replication controller with specified NAME in ps output format.
kubectl get replicationcontroller web

# List a single pod in JSON output format.
kubectl get -o json pod web-pod-13je7

# List a pod identified by type and name specified in "pod.yaml" in JSON output format.
kubectl get -f pod.yaml -o json

# Return only the phase value of the specified pod.
kubectl get -o template pod/web-pod-13je7 --template={{.status.phase}}

# List all replication controllers and services together in ps output format.
kubectl get rc,services

# List one or more resources by their type and names.
kubectl get rc/web service/frontend pods/web-pod-13je7`
)

// NewCmdGet creates a command object for the generic "get" action, which
// retrieves one or more resources from a server.
func NewCmdExportIntoChart(f *cmdutil.Factory, out io.Writer) *cobra.Command {
	//func NewCmdGet(f *cmdutil.Factory, out io.Writer) *cobra.Command {
	//	options := &GetOptions{}
	options := &ExportIntoChartOptions{}

	// retrieve a list of handled resources from printer as valid args
	validArgs, argAliases := []string{}, []string{}
	p, err := f.Printer(nil, nil)
	cmdutil.CheckErr(err)
	if p != nil {
		validArgs = p.HandledResources()
		argAliases = kubectl.ResourceAliases(validArgs)
	}

	cmd := &cobra.Command{
		//Use:     "get [(-o|--output=)json|yaml|wide|go-template=...|go-template-file=...|jsonpath=...|jsonpath-file=...] (TYPE [NAME | -l label] | TYPE/NAME ...) [flags]",
		Use:     "export-into-chart [(-o|--output=)json|yaml|wide|go-template=...|go-template-file=...|jsonpath=...|jsonpath-file=...] (TYPE [NAME | -l label] | TYPE/NAME ...) [flags]",
		Short:   "Display one or many resources",
		Long:    get_long,
		Example: get_example,
		Run: func(cmd *cobra.Command, args []string) {
			//err := RunGet(f, out, cmd, args, options)
			err := RunExportIntoChart(f, out, cmd, args, options)
			cmdutil.CheckErr(err)
		},
		SuggestFor: []string{"list", "ps"},
		ValidArgs:  validArgs,
		ArgAliases: argAliases,
	}
	cmdutil.AddPrinterFlags(cmd)
	cmd.Flags().StringP("selector", "l", "", "Selector (label query) to filter on")
	cmd.Flags().BoolP("watch", "w", false, "After listing/getting the requested object, watch for changes.")
	cmd.Flags().Bool("watch-only", false, "Watch for changes to the requested object(s), without listing/getting first.")
	cmd.Flags().Bool("all-namespaces", false, "If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.")
	cmd.Flags().Bool("show-kind", false, "If present, list the kind of each requested resource.")
	cmd.Flags().StringSliceP("label-columns", "L", []string{}, "Accepts a comma separated list of labels that are going to be presented as columns. Names are case-sensitive. You can also use multiple flag statements like -L label1 -L label2...")
	cmd.Flags().Bool("export", false, "If true, use 'export' for the resources.  Exported resources are stripped of cluster-specific information.")
	usage := "Filename, directory, or URL to a file identifying the resource to get from a server."
	kubectl.AddJsonFilenameFlag(cmd, &options.Filenames, usage)
	cmdutil.AddRecursiveFlag(cmd, &options.Recursive)
	cmdutil.AddInclude3rdPartyFlags(cmd)
	cmd.Flags().StringVar(&options.Raw, "raw", options.Raw, "Raw URI to request from the server.  Uses the transport specified by the kubeconfig file.")
	return cmd
}

// RunGet implements the generic Get command
// TODO: convert all direct flag accessors to a struct and pass that instead of cmd
func RunExportIntoChart(f *cmdutil.Factory, out io.Writer, cmd *cobra.Command, args []string, options *ExportIntoChartOptions) error {
	//func RunGet(f *cmdutil.Factory, out io.Writer, cmd *cobra.Command, args []string, options *GetOptions) error {
	if len(options.Raw) > 0 {
		client, err := f.Client()
		if err != nil {
			return err
		}

		stream, err := client.RESTClient.Get().RequestURI(options.Raw).Stream()
		if err != nil {
			return err
		}
		defer stream.Close()

		for {
			buffer := make([]byte, 1024, 1024)
			bytesRead, err := stream.Read(buffer)
			if bytesRead > 0 {
				fmt.Printf("%s", string(buffer[:bytesRead]))
			}
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
		}
	}

	selector := cmdutil.GetFlagString(cmd, "selector")
	allNamespaces := cmdutil.GetFlagBool(cmd, "all-namespaces")
	showKind := cmdutil.GetFlagBool(cmd, "show-kind")
	mapper, typer := f.Object(cmdutil.GetIncludeThirdPartyAPIs(cmd))

	cmdNamespace, enforceNamespace, err := f.DefaultNamespace()
	if err != nil {
		return err
	}

	if allNamespaces {
		enforceNamespace = false
	}

	if len(args) == 0 && len(options.Filenames) == 0 {
		fmt.Fprint(out, "You must specify the type of resource to get. ", valid_resources)
		return cmdutil.UsageError(cmd, "Required resource not specified.")
	}

	// always show resources when getting by name or filename
	argsHasNames, err := resource.HasNames(args)
	if err != nil {
		return err
	}
	if len(options.Filenames) > 0 || argsHasNames {
		cmd.Flag("show-all").Value.Set("true")
	}
	export := cmdutil.GetFlagBool(cmd, "export")

	/*// handle watch separately since we cannot watch multiple resource types
	isWatch, isWatchOnly := cmdutil.GetFlagBool(cmd, "watch"), cmdutil.GetFlagBool(cmd, "watch-only")
	if isWatch || isWatchOnly {
		r := resource.NewBuilder(mapper, typer, resource.ClientMapperFunc(f.ClientForMapping), f.Decoder(true)).
			NamespaceParam(cmdNamespace).DefaultNamespace().AllNamespaces(allNamespaces).
			FilenameParam(enforceNamespace, options.Recursive, options.Filenames...).
			SelectorParam(selector).
			ExportParam(export).
			ResourceTypeOrNameArgs(true, args...).
			SingleResourceType().
			Latest().
			Do()
		err := r.Err()
		if err != nil {
			return err
		}
		infos, err := r.Infos()
		if err != nil {
			return err
		}
		if len(infos) != 1 {
			return fmt.Errorf("watch is only supported on individual resources and resource collections - %d resources were found", len(infos))
		}
		info := infos[0]
		mapping := info.ResourceMapping()
		printer, err := f.PrinterForMapping(cmd, mapping, allNamespaces)
		if err != nil {
			return err
		}

		obj, err := r.Object()
		if err != nil {
			return err
		}

		// watching from resourceVersion 0, starts the watch at ~now and
		// will return an initial watch event.  Starting form ~now, rather
		// the rv of the object will insure that we start the watch from
		// inside the watch window, which the rv of the object might not be.
		rv := "0"
		isList := meta.IsListType(obj)
		if isList {
			// the resourceVersion of list objects is ~now but won't return
			// an initial watch event
			rv, err = mapping.MetadataAccessor.ResourceVersion(obj)
			if err != nil {
				return err
			}
		}

		// print the current object
		if !isWatchOnly {
			if err := printer.PrintObj(obj, out); err != nil {
				return fmt.Errorf("unable to output the provided object: %v", err)
			}
		}

		// print watched changes
		w, err := r.Watch(rv)
		if err != nil {
			return err
		}

		first := true
		kubectl.WatchLoop(w, func(e watch.Event) error {
			if !isList && first {
				// drop the initial watch event in the single resource case
				first = false
				return nil
			}
			return printer.PrintObj(e.Object, out)
		})
		return nil
	}

	r := resource.NewBuilder(mapper, typer, resource.ClientMapperFunc(f.ClientForMapping), f.Decoder(true)).
		NamespaceParam(cmdNamespace).DefaultNamespace().AllNamespaces(allNamespaces).
		FilenameParam(enforceNamespace, options.Recursive, options.Filenames...).
		SelectorParam(selector).
		ExportParam(export).
		ResourceTypeOrNameArgs(true, args...).
		ContinueOnError().
		Latest().
		Flatten().
		Do()
	err = r.Err()
	if err != nil {
		return err
	}

	printer, generic, err := cmdutil.PrinterForCommand(cmd)
	if err != nil {
		return err
	}

	if generic {
		clientConfig, err := f.ClientConfig()
		if err != nil {
			return err
		}

		allErrs := []error{}
		singular := false
		infos, err := r.IntoSingular(&singular).Infos()
		if err != nil {
			if singular {
				return err
			}
			allErrs = append(allErrs, err)
		}

		// the outermost object will be converted to the output-version, but inner
		// objects can use their mappings
		version, err := cmdutil.OutputVersion(cmd, clientConfig.GroupVersion)
		if err != nil {
			return err
		}

		obj, err := resource.AsVersionedObject(infos, !singular, version, f.JSONEncoder())
		if err != nil {
			return err
		}

		if err := printer.PrintObj(obj, out); err != nil {
			allErrs = append(allErrs, err)
		}
		return utilerrors.NewAggregate(allErrs)
	}

	allErrs := []error{}
	infos, err := r.Infos()
	if err != nil {
		allErrs = append(allErrs, err)
	}

	objs := make([]runtime.Object, len(infos))
	for ix := range infos {
		objs[ix] = infos[ix].Object
	}

	sorting, err := cmd.Flags().GetString("sort-by")
	var sorter *kubectl.RuntimeSort
	if err == nil && len(sorting) > 0 && len(objs) > 1 {
		clientConfig, err := f.ClientConfig()
		if err != nil {
			return err
		}

		version, err := cmdutil.OutputVersion(cmd, clientConfig.GroupVersion)
		if err != nil {
			return err
		}

		for ix := range infos {
			objs[ix], err = infos[ix].Mapping.ConvertToVersion(infos[ix].Object, version)
			if err != nil {
				allErrs = append(allErrs, err)
				continue
			}
		}

		// TODO: questionable
		if sorter, err = kubectl.SortObjects(f.Decoder(true), objs, sorting); err != nil {
			return err
		}
	}

	// use the default printer for each object
	printer = nil
	var lastMapping *meta.RESTMapping
	w := kubectl.GetNewTabWriter(out)
	defer w.Flush()

	if mustPrintWithKinds(objs, infos, sorter) {
		showKind = true
	}

	for ix := range objs {
		var mapping *meta.RESTMapping
		var original runtime.Object
		if sorter != nil {
			mapping = infos[sorter.OriginalPosition(ix)].Mapping
			original = infos[sorter.OriginalPosition(ix)].Object
		} else {
			mapping = infos[ix].Mapping
			original = infos[ix].Object
		}
		if printer == nil || lastMapping == nil || mapping == nil || mapping.Resource != lastMapping.Resource {
			printer, err = f.PrinterForMapping(cmd, mapping, allNamespaces)
			if err != nil {
				allErrs = append(allErrs, err)
				continue
			}
			lastMapping = mapping
		}
		if resourcePrinter, found := printer.(*kubectl.HumanReadablePrinter); found {
			resourceName := resourcePrinter.GetResourceName()
			if mapping != nil {
				if resourceName == "" {
					resourceName = mapping.Resource
				}
				if alias, ok := kubectl.ResourceShortFormFor(mapping.Resource); ok {
					resourceName = alias
				} else if resourceName == "" {
					resourceName = "none"
				}
			} else {
				resourceName = "none"
			}

			if showKind {
				resourcePrinter.EnsurePrintWithKind(resourceName)
			}

			if err := printer.PrintObj(original, w); err != nil {
				allErrs = append(allErrs, err)
			}
			continue
		}
		if err := printer.PrintObj(original, w); err != nil {
			allErrs = append(allErrs, err)
			continue
		}
	}
	return utilerrors.NewAggregate(allErrs)*/

	r := resource.NewBuilder(mapper, typer, resource.ClientMapperFunc(f.ClientForMapping), f.Decoder(true)).
		NamespaceParam(cmdNamespace).DefaultNamespace().AllNamespaces(allNamespaces).
		FilenameParam(enforceNamespace, options.Recursive, options.Filenames...).
		SelectorParam(selector).
		ExportParam(export).
		ResourceTypeOrNameArgs(true, args...).
		ContinueOnError().
		Latest().
		Flatten().
		Do()
	err = r.Err()
	if err != nil {
		return err
	}

	allErrs := []error{}
	infos, err := r.Infos()
	if err != nil {
		allErrs = append(allErrs, err)
	}

	objs := make([]runtime.Object, len(infos))
	for ix := range infos {
		objs[ix] = infos[ix].Object
	}

	sorting, err := cmd.Flags().GetString("sort-by")
	var sorter *kubectl.RuntimeSort
	if err == nil && len(sorting) > 0 && len(objs) > 1 {
		clientConfig, err := f.ClientConfig()
		if err != nil {
			return err
		}

		version, err := cmdutil.OutputVersion(cmd, clientConfig.GroupVersion)
		if err != nil {
			return err
		}

		for ix := range infos {
			objs[ix], err = infos[ix].Mapping.ConvertToVersion(infos[ix].Object, version)
			if err != nil {
				allErrs = append(allErrs, err)
				continue
			}
		}

		// TODO: questionable
		if sorter, err = kubectl.SortObjects(f.Decoder(true), objs, sorting); err != nil {
			return err
		}
	}

	// use the default printer for each object
	//printer = nil //printer -> kubectl.ResourcePrinter
	var lastMapping *meta.RESTMapping

	if mustPrintWithKinds(objs, infos, sorter) {
		showKind = true
	}

	for ix := range objs {
		var mapping *meta.RESTMapping
		var original runtime.Object
		if sorter != nil {
			mapping = infos[sorter.OriginalPosition(ix)].Mapping
			original = infos[sorter.OriginalPosition(ix)].Object
		} else {
			mapping = infos[ix].Mapping
			original = infos[ix].Object
		}
		if lastMapping == nil || mapping == nil || mapping.Resource != lastMapping.Resource {
			lastMapping = mapping
		}
		resourceName := ""
		if mapping != nil && resourceName == "" {
			resourceName = mapping.Resource
			if resourceName == "" {
				resourceName = "none"
			}
		}

		var buf *bytes.Buffer = &bytes.Buffer{}
		//var b []byte
		//b, err = json.Marshal(original)
		//if err != nil {
		if err = codec.JSON.Encode(buf).One(original); err != nil {
			allErrs = append(allErrs, err)
			continue
		}
		if showKind {
			fmt.Printf("Convert %s from %s\n", resourceName /*string(b)*/, buf.String())
		}
		var hco *codec.Object
		if hco, err = codec.JSON.Decode(buf.Bytes()).One(); err != nil {
			allErrs = append(allErrs, err)
			continue
		}
		buf.Reset()
		buf.WriteString("# Default value for\n")
		switch strings.ToLower(resourceName) {
		case "services":
			tgt := new(kapi.Service)
			if err = hco.Object(tgt); err != nil {
				allErrs = append(allErrs, err)
				continue
			}
			var k, v string
			k = "name"
			v = tgt.Name
			if _, ok := metadata_chart_value[k]; ok {
				metadata_chart_value[k] = v
			}
			k = "namespace"
			v = tag.Namespace
			if _, ok := metadata_chart_value[k]; ok {
				metadata_chart_value[k] = v
			}
			k = "labels"
			v = tag.Labels
			if _, ok := metadata_chart_value[k]; ok {
				metadata_chart_value[k] = v
			}
			k = "annotations"
			v = tag.Annotations
			if _, ok := metadata_chart_value[k]; ok {
				metadata_chart_value[k] = v
			}

			service_chart_value = map[string]interface{}{
				"metadata": metadata_chart_value,
				"spec":     map[string]interface{}{},
			}
			service_spec := service_chart_value["spec"]
			k = "clusterIP"
			v = tgt.Spec.ClusterIP
			if false == strings.Compare(v, "None") {
				v = ""
			}
			if _, ok := service_chart_value[k]; ok {
				service_spec[k] = v
			}
			for i := range tgt.Spec.Ports {
				var p int32
				k = "name"
				v = tgt.Spec.Ports[i].Name
				if _, ok := serviceport_chart_value[k]; ok {
					serviceport_chart_value[k] = v
				}
				k = "nodePort"
				p = tgt.Spec.Ports[i].NodePort
				if _, ok := serviceport_chart_value[k]; ok {
					serviceport_chart_value[k] = p
				}
				k = "port"
				p = tag.Spec.Ports[i].Port
				if _, ok := serviceport_chart_value[k]; ok {
					serviceport_chart_value[k] = p
				}
				k = "protocol"
				v = tag.Spec.Ports[i].Protocol
				if _, ok := serviceport_chart_value[k]; ok {
					serviceport_chart_value[k] = v
				}
				k = "targetPort"
				v = int32(tag.Spec.Ports[i].TargetPort.IntValue())
				if _, ok := serviceport_chart_value[k]; ok {
					serviceport_chart_value[k] = v
				}
			}
			service_spec["ports"] = append(service_spec["ports"], serviceport_chart_value)
		case "deployments":
		default:
			allErrs = append(allErrs, fmt.Errorf("Not implemented or unknown resource type"))
		}
	}

	return utilerrors.NewAggregate(allErrs)
}

// mustPrintWithKinds determines if printer is dealing
// with multiple resource kinds, in which case it will
// return true, indicating resource kind will be
// included as part of printer output
func mustPrintWithKinds(objs []runtime.Object, infos []*resource.Info, sorter *kubectl.RuntimeSort) bool {
	var lastMap *meta.RESTMapping

	for ix := range objs {
		var mapping *meta.RESTMapping
		if sorter != nil {
			mapping = infos[sorter.OriginalPosition(ix)].Mapping
		} else {
			mapping = infos[ix].Mapping
		}

		// display "kind" only if we have mixed resources
		if lastMap != nil && mapping.Resource != lastMap.Resource {
			return true
		}
		lastMap = mapping
	}

	return false
}

var (
	metadata_chart_value = map[string]interface{}{
		"name":        "",
		"namespace":   "default",
		"labels":      map[string]string{},
		"annotations": map[string]string{},
	}
	serviceport_chart_value = map[string]interface{}{
		"name":       "",
		"nodePort":   0,
		"port":       0,
		"protocol":   "TCP",
		"targetPort": 0,
	}
	service_chart_value = map[string]interface{}{
		"metadata": metadata_chart_value,
		"spec": map[string]interface{}{
			"clusterIP":       "",
			"ports":           []map[string]interface{}{serviceport_chart_value},
			"selector":        map[string]string{},
			"sessionAffinity": "None",
			"type":            "ClusterIP",
		},
	}
	env_chart_value = map[string]string{
		"name":  "",
		"value": "",
	}
	container_port_chart_value = map[string]interface{}{
		"containerPort": 0,
		"protocol":      "TCP",
	}
	container_chart_value = map[string]string{
		"args":            []string{},
		"command":         []string{},
		"env":             []map[string]string{env_chart_value},
		"imageRepo":       "",
		"imageTag":        "latest",
		"imagePullPolicy": "Always",
		"name":            "",
		"ports":           []map[string]interface{}{container_port_chart_value},
		"resources":       {},
		"workingDir":      "",
	}
	localobjectreference_chart_value = map[string]string{
		"name": "",
	}
	volume_chart_value = map[string]interface{}{
		"name":                  "",
		"hostPath":              *(map[string]interface{})(nil),
		"emptyDir":              *(map[string]interface{})(nil),
		"nfs":                   *(map[string]interface{})(nil),
		"iscsi":                 *(map[string]interface{})(nil),
		"glusterfs":             *(map[string]interface{})(nil),
		"persistentVolumeClaim": *(map[string]interface{})(nil),
		"rbd":        *(map[string]interface{})(nil),
		"flexVolume": *(map[string]interface{})(nil),
		"cinder":     *(map[string]interface{})(nil),
		"cephFS":     *(map[string]interface{})(nil),
		"flocker":    *(map[string]interface{})(nil),
		"fc":         *(map[string]interface{})(nil),
	}
	podspec_chart_value = map[string]interface{}{
		"containers":                    []map[string]interface{}{container_chart_value},
		"dnsPolicy":                     "ClusterFirst",
		"imagePullSecrets":              []map[string]interface{}{localobjectreference_chart_value},
		"restartPolicy":                 "Always",
		"securityContext":               map[string]interface{}{},
		"serviceAccountName":            "",
		"terminationGracePeriodSeconds": 30,
		"volumes":                       []map[string]interface{}{volume_chart_value},
	}
	deployment_chart_value = map[string]interface{}{
		"metadata": metadata_chart_value,
		"spec": map[string]interface{}{
			"replicas": 1,
			"strategy": "RollingUpdate",
			"template": map[string]interface{}{
				"metadata": metadata_chart_value,
				"spec":     podspec_chart_value,
			},
		},
	}
)
