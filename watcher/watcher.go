package main

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func runWatcher() {
	var kubeconfig string

	kubeconfig = "/Users/omkard/Downloads/omkar-dev.yaml"

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset := kubernetes.NewForConfigOrDie(config)

	nodeList, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	for _, n := range nodeList.Items {
		fmt.Println(n.Name)
	}

	watchNode, err := clientset.CoreV1().Nodes().Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}
	// watchlist := cache.NewListWatchFromClient(
	// 	clientset.CoreV1().RESTClient(),
	// 	string(v1.ResourceServices),
	// 	v1.NamespaceAll,
	// 	fields.Everything(),
	// )
	// watchNode, err := watch.NewRetryWatcher("5395", watchlist)
	// if err != nil {
	// 	panic(err)
	// }
	for event := range watchNode.ResultChan() {
		nde := event.Object.(*v1.Node)
		anotation := make(map[string]string)
		anotation["idCordonedByUser"] = "yes"
		result, getErr := clientset.CoreV1().Nodes().Get(context.TODO(), nde.Name, metav1.GetOptions{})
		if getErr != nil {
			panic(fmt.Errorf("Failed to get latest version of Deployment: %v", getErr))
		}
		if result.Spec.Unschedulable == true {
			result.Annotations["isCordonedByUser"] = "yes"
			_, err := clientset.CoreV1().Nodes().Update(context.TODO(), result, metav1.UpdateOptions{})
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			fmt.Printf("Node annotation updated: %s", nde.ObjectMeta.Name)
		} else {
			result.Annotations["isCordonedByUser"] = "no"
			_, err := clientset.CoreV1().Nodes().Update(context.TODO(), result, metav1.UpdateOptions{})
			if err != nil {
				fmt.Println(err.Error())
				return
			}

			fmt.Printf("Node annotation updated: %s", nde.ObjectMeta.Name)
		}

		fmt.Println(nde.Name, event.Type, nde.DeepCopy().Annotations)
	}
}
