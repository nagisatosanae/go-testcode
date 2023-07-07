package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func ExecuteInstall(number int, hp *helmParam, hd helmDapr) {
	result, res, err := hd.Install(number, hp)

	fmt.Println(result, res, err)
}

func main() {
	// 새로운 fake 클라이언트를 생성합니다.
	client := fake.NewSimpleClientset()

	// fake 클라이언트를 사용하여 새로운 네임스페이스를 생성합니다.

	gg, err := client.CoreV1().Namespaces().Create(
		context.Background(),
		&corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "testing",
			},
		},
		metav1.CreateOptions{},
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(gg.Name)

	// 생성한 네임스페이스를 가져와 출력합니다.
	namespaces, err := client.CoreV1().Namespaces().List(
		context.Background(),
		metav1.ListOptions{},
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(gg.Name)

	for _, ns := range namespaces.Items {
		fmt.Printf("Name: %s\n", ns.Name)
	}
}
