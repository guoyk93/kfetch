package main

import (
	"context"
	"github.com/guoyk93/gg"
	"github.com/guoyk93/kfetch/pkg/resources"
	"k8s.io/client-go/kubernetes"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	NameAny      = "-"
	NameWildcard = "*"

	ConfigDir    = ".kfetch"
	ConfigPrefix = "cluster-"
	ConfigSuffix = ".yaml"
)

func DoPush(ctx context.Context, cluster string, namespace string, kind string, _name string) error {
	return IterateCluster(cluster, func(cluster string, client *kubernetes.Clientset) error {
		return IterateNamespace(ctx, client, namespace, func(namespace string) error {
			return IterateKind(kind, func(kind string) (err error) {
				defer gg.Suppress(&err, os.IsNotExist)
				defer gg.Guard(&err)

				dir := filepath.Join(cluster, namespace, kind)

				var names []string
				if _name == NameAny {
					infos := gg.Must(os.ReadDir(dir))
					for _, info := range infos {
						if info.IsDir() {
							log.Println("found unexpected directory in:", dir)
							continue
						}
						if !strings.HasSuffix(info.Name(), ".yaml") {
							log.Println("found unexpected file", info.Name(), "in:", dir, ", for compatible reasons, all YAML files must has extension '.yaml', NOT '.yml'")
							continue
						}
						names = append(names, strings.TrimSuffix(info.Name(), ".yaml"))
					}
				} else {
					names = []string{_name}
				}

				for _, name := range names {
					log.Printf("PUSH: %s/%s/%s/%s", cluster, namespace, kind, name)
					buf := gg.Must(os.ReadFile(filepath.Join(dir, name+".yaml")))
					gg.Must0(resources.Push(ctx, client, resources.PushOptions{
						Namespace: namespace,
						Name:      name,
						Kind:      kind,
						Data:      buf,
					}))
				}
				return
			})
		})
	})
}

func DoPull(ctx context.Context, cluster string, namespace string, kind string, _name string) error {
	return IterateCluster(cluster, func(cluster string, client *kubernetes.Clientset) error {
		return IterateNamespace(ctx, client, namespace, func(namespace string) error {
			return IterateKind(kind, func(kind string) (err error) {
				defer gg.Suppress(&err, os.IsNotExist)
				defer gg.Guard(&err)

				dir := filepath.Join(cluster, namespace, kind)

				var names []string
				if _name == NameAny {
					_ = os.RemoveAll(dir)
					log.Printf("CLEAN: %s/%s/%s", cluster, namespace, kind)
					names = gg.Must(resources.List(ctx, client, resources.ListOptions{
						Kind:      kind,
						Namespace: namespace,
					}))
				} else {
					names = []string{_name}
				}

				if len(names) == 0 {
					return
				}

				if err = os.MkdirAll(dir, 0755); err != nil {
					return
				}

				for _, name := range names {
					log.Printf("PULL: %s/%s/%s/%s", cluster, namespace, kind, name)
					buf := gg.Must(resources.Pull(ctx, client, resources.PullOptions{
						Kind:      kind,
						Namespace: namespace,
						Name:      name,
					}))
					if len(buf) == 0 {
						continue
					}
					gg.Must0(os.WriteFile(filepath.Join(dir, name+".yaml"), buf, 0755))
				}
				return
			})
		})
	})
}
