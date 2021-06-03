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
	"k8s.io/client-go/util/retry"
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

func (k *k8s) UpdateInstance(ctx context.Context, args UpdateArgs) error {
	if err := k.updateWafInstance(ctx, args); err != nil {
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
				"rules": wafv1.Rules{
					CustomRules:           args.Rules.CustomRules,
					EnableDefaultHoneyPot: args.Rules.EnableDefaultHoneyPot,
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

func setReplicas(wafObject map[string]interface{}, replicas int64) error {
	if err := unstructured.SetNestedField(wafObject, int64(replicas), "spec", "replicas"); err != nil {
		return err
	}
	return nil
}

func setPlan(wafObject map[string]interface{}, planName string) error {
	if err := unstructured.SetNestedField(wafObject, planName, "spec", "planName"); err != nil {
		return err
	}
	return nil
}

func setBind(wafObject map[string]interface{}, bind Bind, protocol string) error {
	if err := unstructured.SetNestedField(wafObject, bind.ServiceName, "spec", "bind", "name"); err != nil {
		return err
	}
	hostname := fmt.Sprintf("%s%s.%s.svc.cluster.local", protocol, bind.ServiceName, bind.Namespace)
	if err := unstructured.SetNestedField(wafObject, hostname, "spec", "bind", "name"); err != nil {
		return err
	}

	return nil
}

func setRules(wafObject map[string]interface{}, rules Rules) error {
	sliceInterface := make([]interface{}, len(rules.CustomRules))
	for i, rule := range rules.CustomRules {
		sliceInterface[i] = rule
	}
	if err := unstructured.SetNestedSlice(wafObject, sliceInterface, "spec", "rules", "customRules"); err != nil {
		return err
	}
	if err := unstructured.SetNestedField(wafObject, rules.EnableDefaultHoneyPot, "spec", "rules", "enableDefaultHoneyPot"); err != nil {
		return err
	}

	return nil
}

func (k *k8s) updateWafInstance(ctx context.Context, args UpdateArgs) error {
	protoPrefix := "http://"
	if strings.Compare("HTTPS", args.Bind.Protocol) == 0 {
		protoPrefix = "https://"
	}
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		waf, getErr := k.dynamicClient.Resource(wafGVR).Namespace(args.Namespace).Get(ctx, args.Name, metav1.GetOptions{})
		if getErr != nil {
			return getErr
		}

		// update replicas
		if err := setReplicas(waf.Object, int64(args.Replicas)); err != nil {
			return err
		}

		// update plan
		if err := setPlan(waf.Object, args.PlanName); err != nil {
			return err
		}

		// update bind
		if err := setBind(waf.Object, args.Bind, protoPrefix); err != nil {
			return err
		}

		// update rules
		if err := setRules(waf.Object, args.Rules); err != nil {
			return err
		}
		_, updateErr := k.dynamicClient.Resource(wafGVR).Namespace(args.Namespace).Update(ctx, waf, metav1.UpdateOptions{})
		return updateErr
	})

	if retryErr != nil {
		return retryErr
	}

	return nil
}

func (k *k8s) deleteWafInstance(ctx context.Context, args DeleteArgs) error {
	if err := k.dynamicClient.Resource(wafGVR).Namespace(args.Namespace).Delete(ctx, args.Name, metav1.DeleteOptions{}); err != nil {
		return err
	}

	return nil
}
