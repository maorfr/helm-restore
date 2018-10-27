package restore

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/golang/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	rspb "k8s.io/helm/pkg/proto/hapi/release"
)

// Restore performs a restore of a release
func Restore(releaseName, tillerNamespace, label string) error {
	releases, err := listReleases(releaseName, tillerNamespace, label)
	if err != nil {
		return err
	}
	if len(releases) != 1 {
		return fmt.Errorf("%s has no deployed releases", releaseName)
	}

	fileName := "manifests.yaml"
	os.Remove(fileName)
	if err := ioutil.WriteFile(fileName, []byte(releases[0].manifest), 0644); err != nil {
		return err
	}
	applyCmd := fmt.Sprintf("kubectl apply --namespace %s -f %s", releases[0].namespace, fileName)
	execute(applyCmd)
	os.Remove(fileName)
	return nil
}

func execute(cmd string) {
	args := strings.Split(cmd, " ")
	binary := args[0]
	_, err := exec.LookPath(binary)
	if err != nil {
		log.Fatal(err)
	}

	output, err := exec.Command(binary, args[1:]...).CombinedOutput()
	if err != nil {
		log.Println("Error: command execution failed:", cmd)
		log.Fatal(string(output))
	}
	fmt.Print((string)(output))
}

type releaseData struct {
	namespace string
	manifest  string
}

func listReleases(releaseName, tillerNamespace, label string) ([]releaseData, error) {
	clientSet := getClientSet()
	var releasesData []releaseData
	coreV1 := clientSet.CoreV1()
	storage := getTillerStorage(clientSet, tillerNamespace)
	switch storage {
	case "secrets":
		secrets, err := coreV1.Secrets(tillerNamespace).List(metav1.ListOptions{
			LabelSelector: label + ",NAME=" + releaseName,
		})
		if err != nil {
			return nil, err
		}
		for _, item := range secrets.Items {
			releaseData := getReleaseData((string)(item.Data["release"]))
			if releaseData == nil {
				continue
			}
			releasesData = append(releasesData, *releaseData)
		}
	case "cfgmaps":
		configMaps, err := coreV1.ConfigMaps(tillerNamespace).List(metav1.ListOptions{
			LabelSelector: label + ",NAME=" + releaseName,
		})
		if err != nil {
			return nil, err
		}
		for _, item := range configMaps.Items {
			releaseData := getReleaseData(item.Data["release"])
			if releaseData == nil {
				continue
			}
			releasesData = append(releasesData, *releaseData)
		}
	}

	return releasesData, nil
}

func getReleaseData(itemReleaseData string) *releaseData {
	data, _ := decodeRelease(itemReleaseData)
	releaseData := releaseData{
		namespace: data.Namespace,
		manifest:  data.Manifest,
	}
	return &releaseData
}

var b64 = base64.StdEncoding
var magicGzip = []byte{0x1f, 0x8b, 0x08}

func decodeRelease(data string) (*rspb.Release, error) {
	// base64 decode string
	b, err := b64.DecodeString(data)
	if err != nil {
		return nil, err
	}

	// For backwards compatibility with releases that were stored before
	// compression was introduced we skip decompression if the
	// gzip magic header is not found
	if bytes.Equal(b[0:3], magicGzip) {
		r, err := gzip.NewReader(bytes.NewReader(b))
		if err != nil {
			return nil, err
		}
		b2, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		b = b2
	}

	var rls rspb.Release
	// unmarshal protobuf bytes
	if err := proto.Unmarshal(b, &rls); err != nil {
		return nil, err
	}
	return &rls, nil
}

func getClientSet() *kubernetes.Clientset {
	var kubeconfig string
	if kubeConfigPath := os.Getenv("KUBECONFIG"); kubeConfigPath != "" {
		kubeconfig = kubeConfigPath
	} else {
		kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

	config, err := buildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatal(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err.Error())
	}

	return clientset
}

func buildConfigFromFlags(context, kubeconfigPath string) (*rest.Config, error) {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfigPath},
		&clientcmd.ConfigOverrides{
			CurrentContext: context,
		}).ClientConfig()
}

func getTillerStorage(clientset *kubernetes.Clientset, tillerNamespace string) string {
	coreV1 := clientset.CoreV1()
	listOptions := metav1.ListOptions{
		LabelSelector: "name=tiller",
	}
	pods, err := coreV1.Pods(tillerNamespace).List(listOptions)
	if err != nil {
		log.Fatal(err)
	}

	if len(pods.Items) == 0 {
		log.Fatal("Found 0 tiller pods")
	}

	storage := "cfgmaps"
	for _, c := range pods.Items[0].Spec.Containers[0].Command {
		if strings.Contains(c, "secret") {
			storage = "secrets"
		}
	}

	return storage
}
