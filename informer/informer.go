package main

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	var kubeconfig string

	kubeconfig = "/Users/omkard/Downloads/omkar-dev.yaml"

	// get client
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset := kubernetes.NewForConfigOrDie(config)

	// Create the shared informer factory and use the client to connect to
	// Kubernetes
	factory := informers.NewSharedInformerFactory(clientset, 0)

	// Get the informer for the right resource, in this case a Node
	informer := factory.Core().V1().Nodes().Informer()

	// Create a channel to stops the shared informer gracefully
	stopper := make(chan struct{})
	defer close(stopper)

	// Kubernetes serves an utility to handle API crashes
	defer runtime.HandleCrash()

	// add event catchers
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		// When a new Node gets created
		AddFunc: func(obj interface{}) { fmt.Println("Created a node") },
		// When a Node gets updated
		UpdateFunc: onUpdate,
		// When a Node gets deleted
		DeleteFunc: func(interface{}) { fmt.Println("Deleted a node") },
	})

	// start informer
	informer.Run(stopper)
}

func onUpdate(oldObj interface{}, obj interface{}) {

	var kubeconfig string

	kubeconfig = "/Users/omkard/Downloads/omkar-dev.yaml"

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset := kubernetes.NewForConfigOrDie(config)

	// Cast the obj as node
	newNode := obj.(*v1.Node)
	oldNode := oldObj.(*v1.Node)

	// check if modification is relevant, else update node objects
	if oldNode.Spec.Unschedulable == newNode.Spec.Unschedulable {
		return
	}

	// If the node is drined by nodelet, then 'isCordonedByPf9' Annotation will be present
	// That means this modification is a nodelet initiated action and we do not need to
	// add 'isCordonedByUser' annotation.
	// fmt.Println(newNode.ObjectMeta.Annotations["isCordonedByPf9"])
	if newNode.ObjectMeta.Annotations["isCordonedByPF9"] == "yes" {
		fmt.Println(newNode.ObjectMeta.Annotations["isCordonedByPF9"])
		return
	}
	if newNode.Spec.Unschedulable == true {
		newNode, err = clientset.CoreV1().Nodes().Get(context.TODO(), newNode.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			fmt.Println(err.Error())
		}
		newNode.Annotations["isCordonedByUser"] = "yes"
		_, err := clientset.CoreV1().Nodes().Update(context.TODO(), newNode, metav1.UpdateOptions{})
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Printf("Node annotation updated: %s\n", newNode.ObjectMeta.Name)
	} else {
		newNode, err = clientset.CoreV1().Nodes().Get(context.Background(), newNode.ObjectMeta.Name, metav1.GetOptions{})
		if err != nil {
			fmt.Println(err.Error())
		}
		newNode.Annotations["isCordonedByUser"] = "no"
		_, err := clientset.CoreV1().Nodes().Update(context.TODO(), newNode, metav1.UpdateOptions{})
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		fmt.Printf("Node annotation updated: %s\n", newNode.ObjectMeta.Name)
	}
}
