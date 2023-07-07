package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/release"
)

func TestSimpleFakeClient(t *testing.T) {
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

func TestReactor(t *testing.T) {
	client := fake.NewSimpleClientset()

	client.PrependReactor("list", "pods", func(action ktesting.Action) (handled bool, ret runtime.Object, err error) {
		// 원하는 조건에 따라 handled, ret, err를 설정하여 반환합니다.
		// 이 예제에서는 모든 "list pods" 요청에 대해 빈 PodList를 반환하도록 설정하였습니다.
		return true, &corev1.PodList{Items: []corev1.Pod{
			{Spec: corev1.PodSpec{NodeName: "aaa"}},
			{Spec: corev1.PodSpec{NodeName: "bbb"}},
			{Spec: corev1.PodSpec{NodeName: "ccc"}},
		}}, nil
	})

	// 이제 client.CoreV1().Pods("test").List()를 호출하면 우리가 설정한 reactor가 처리하게 됩니다.
	pods, err := client.CoreV1().Pods("test").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	// 우리는 빈 PodList를 반환하도록 설정했으므로, pods.Items의 길이는 0이어야 합니다.
	fmt.Println(len(pods.Items)) // "0"
}

func TestHttpTestNewServer(t *testing.T) {
	// Fake server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/test" {
			w.Write([]byte(`You hit the test endpoint!`))
			return
		}

		if r.Method == http.MethodPost {
			body, _ := ioutil.ReadAll(r.Body)
			w.Write([]byte("You sent a POST request with body: " + string(body)))
			return
		}

		w.Write([]byte(`OK`))
	}))
	defer ts.Close()

	fmt.Println("##ts.URL : ", ts.URL)

	// Make a request to the test endpoint on our fake server
	res, err := http.Get(ts.URL + "/test")

	if err != nil {
		log.Fatal(err)
	}

	greeting, err := ioutil.ReadAll(res.Body)
	res.Body.Close()

	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%s", greeting) // "You hit the test endpoint!"

	// POST request example
	resp, err := http.Post(ts.URL, "application/json", strings.NewReader(`{"foo":"bar"}`))
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("%s", body) // "You sent a POST request with body: {"foo":"bar"}"
}

func TestHelmInstall(t *testing.T) {
	if true {
		return
	}

	settings := cli.New()

	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		log.Fatalf("Failed to init action configuration: %v", err)
	}

	client := action.NewInstall(actionConfig)
	client.Namespace = settings.Namespace()
	client.ReleaseName = "my-release"
	client.Wait = true

	chartPathOptions := &client.ChartPathOptions
	chartPath, err := chartPathOptions.LocateChart("/path/to/your/chart", settings)
	if err != nil {
		log.Fatalf("Failed to locate chart: %v", err)
	}

	chart, err := loader.Load(chartPath)
	if err != nil {
		log.Fatalf("Failed to load chart: %v", err)
	}

	_, err = client.Run(chart, nil)
	if err != nil {
		log.Fatalf("Failed to install chart: %v", err)
	}

	log.Println("Chart installed successfully!")
}

type HelmClientInterface interface {
	InstallRelease(chartPath, ns string, values map[string]interface{}) (*release.Release, error)
}

type FakeHelmClient struct{}

func (c *FakeHelmClient) InstallRelease(chartPath, ns string, values map[string]interface{}) (*release.Release, error) {
	// 테스트에 필요한 가짜 응답을 반환합니다.
	return &release.Release{
		Name:      "fake-release",
		Namespace: "default",
		Info: &release.Info{
			Status: release.StatusDeployed,
		},
	}, nil
}

type MyMockedObject struct {
	mock.Mock
}

func (m *MyMockedObject) Install(number int, param *helmParam) (bool, *helmResponse, error) {
	args := m.Called(number, param)
	fmt.Println(args)

	return args.Bool(0), args.Get(1).(*helmResponse), args.Error(2)
}

func (m *MyMockedObject) UnInstall(number int) error {

	return nil
}

func TestMockHelm(t *testing.T) {
	testObj := new(MyMockedObject)
	testObj.
		On("Install",
			100,
			&helmParam{name: "aabbcc"},
		).
		Return(
			true,
			&helmResponse{message: "mock installed"},
			nil,
		)
	ExecuteInstall(100, &helmParam{name: "aabbcc"}, testObj)

	// assert that the expectations were met
	// testObj.AssertExpectations(t)

	// ExecuteInstall()
}

func TestMockHelm2(t *testing.T) {
	testObj := new(MyMockedObject)
	testObj.
		On("Install",
			mock.Anything,
		).
		Return(
			false,
			&helmResponse{message: "mock222 ??"},
			errors.New("failed install error!!!"),
		)
	ExecuteInstall(111, &helmParam{name: "aabb33cc"}, testObj)

	// assert that the expectations were met
	// testObj.AssertExpectations(t)

	// ExecuteInstall()
}
