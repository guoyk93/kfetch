package main

import (
	"context"
	"github.com/guoyk93/gg"
	"github.com/guoyk93/kfetch/pkg/resources"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
	"strings"
)

var (
	IgnoredNamespacePrefixes = []string{
		"argocd",
		"fleet-",
		"rancher-",
		"cert-manager",
		"cattle-",
		"kube-system",
		"kube-public",
		"kube-ingress",
		"kube-storage",
		"kube-node-lease",
		"nginx-ingress",
		"ingress-nginx",
		"nfs-client-provisioner",
		"security-scan",
		"nfs-provisioner",
		"autoops",
		"istio-system",
		"p-",
		"u-",
		"user-",
	}
)

func IterateCluster(_cluster string, fn func(cluster string, client *kubernetes.Clientset) error) (err error) {
	defer gg.Guard(&err)

	dir := filepath.Join(gg.Must(os.UserHomeDir()), ConfigDir)

	var clusters []string

	if _cluster == NameAny {
		infos := gg.Must(os.ReadDir(dir))
		for _, info := range infos {
			if info.IsDir() {
				continue
			}
			if !strings.HasPrefix(info.Name(), ConfigPrefix) {
				continue
			}
			if !strings.HasSuffix(info.Name(), ConfigSuffix) {
				continue
			}
			clusters = append(clusters, strings.TrimSuffix(strings.TrimPrefix(info.Name(), ConfigPrefix), ConfigSuffix))
		}
	} else {
		clusters = []string{_cluster}
	}

	for _, cluster := range clusters {
		client := gg.Must(
			kubernetes.NewForConfig(
				gg.Must(
					clientcmd.BuildConfigFromFlags(
						"",
						filepath.Join(dir, ConfigPrefix+cluster+ConfigSuffix),
					),
				),
			),
		)

		if err = fn(cluster, client); err != nil {
			return
		}
	}
	return
}

func IterateNamespace(ctx context.Context, client *kubernetes.Clientset, _namespace string, fn func(namespace string) error) (err error) {
	defer gg.Guard(&err)

	var namespaces []string

	if _namespace == NameAny || strings.HasSuffix(_namespace, NameWildcard) {
		items := gg.Must(client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{}))
	outerLoop:
		for _, item := range items.Items {
			for _, ignored := range IgnoredNamespacePrefixes {
				if strings.HasPrefix(strings.ToLower(item.Name), ignored) {
					continue outerLoop
				}
			}
			if _namespace == NameAny || strings.HasPrefix(item.Name, strings.TrimSuffix(_namespace, NameWildcard)) {
				namespaces = append(namespaces, item.Name)
			}
		}
	} else {
		item := gg.Must(client.CoreV1().Namespaces().Get(ctx, _namespace, metav1.GetOptions{}))
		namespaces = []string{item.Name}
	}

	for _, namespace := range namespaces {
		if err = fn(namespace); err != nil {
			return
		}
	}
	return
}

func IterateKind(_kind string, fn func(kind string) error) (err error) {
	var kinds []string
	if _kind == NameAny {
		kinds = resources.knownResourceNames
	} else {
		kinds = []string{_kind}
	}
	for _, kind := range kinds {
		if err = fn(kind); err != nil {
			return
		}
	}
	return
}
