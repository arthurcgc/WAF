package manager

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
	"strings"

	wafv1 "github.com/arthurcgc/waf-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var wafGVR = schema.GroupVersionResource{Group: "waf.arthurcgc.waf-operator", Version: "v1", Resource: "wafs"}

type k8s struct {
	dynamicClient dynamic.Interface
}

func NewInCluster() (Manager, error) {
	mgr := &k8s{}
	config, err := rest.InClusterConfig()
	if err != nil {
		return mgr, err
	}

	mgr.dynamicClient, err = dynamic.NewForConfig(config)
	if err != nil {
		return mgr, err
	}

	return mgr, nil
}

func NewOutsideCluster() (Manager, error) {
	mgr := &k8s{}
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return mgr, err
	}

	mgr.dynamicClient, err = dynamic.NewForConfig(config)
	if err != nil {
		return mgr, err
	}

	return mgr, nil
}

func (k *k8s) CreateInstance(ctx context.Context, args CreateArgs) error {
	if err := k.createWafInstance(ctx, args); err != nil {
		return err
	}

	return nil
}

func (k *k8s) DeleteInstance(ctx context.Context, args DeleteArgs) error {
	if err := k.deleteWafInstance(ctx, args); err != nil {
		return err
	}

	return nil
}

func (k *k8s) createWafInstance(ctx context.Context, args CreateArgs) error {
	protoPrefix := "http://"
	if strings.Compare("HTTPS", args.Bind.Protocol) == 0 {
		protoPrefix = "https://"
	}
	waf := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "waf.arthurcgc.waf-operator/v1",
			"kind":       "Waf",
			"metadata": map[string]interface{}{
				"name": args.Name,
			},
			"spec": map[string]interface{}{
				"replicas": args.Replicas,
				"planName": args.PlanName,
				"bind": wafv1.Bind{
					Name:     args.Bind.ServiceName,
					Hostname: fmt.Sprintf("%s%s.%s.svc.cluster.local", protoPrefix, args.Bind.ServiceName, args.Bind.Namespace),
				},
				"service": map[string]interface{}{
					"type": "NodePort",
				},
			},
		},
	}

	_, err := k.dynamicClient.Resource(wafGVR).Namespace(args.Namespace).Create(ctx, waf, metav1.CreateOptions{})
	if err != nil {
		return err
	}

	return nil
}

func (k *k8s) deleteWafInstance(ctx context.Context, args DeleteArgs) error {
	if err := k.dynamicClient.Resource(wafGVR).Namespace(args.Namespace).Delete(ctx, args.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}
