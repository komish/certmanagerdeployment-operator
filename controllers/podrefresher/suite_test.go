/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package podrefresher

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	corev1 "k8s.io/api/core/v1"

	operatorsv1alpha1 "github.com/opdev/certmanagerdeployment-operator/api/v1alpha1"
	"github.com/opdev/certmanagerdeployment-operator/cmdoputils"
	"github.com/opdev/certmanagerdeployment-operator/controllers/componentry"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var identifier int64
var cr = operatorsv1alpha1.CertManagerDeployment{
	ObjectMeta: metav1.ObjectMeta{
		Name: "cluster",
	},
	Spec: operatorsv1alpha1.CertManagerDeploymentSpec{
		Version: cmdoputils.GetStringPointer(componentry.CertManagerDefaultVersion),
	},
}

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"PodRefresher Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func(done Done) {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	identifier = time.Now().Unix()

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
		// documentation indicates that setting USE_EXISTING_CLUSTER should
		// cause envtest to not try and spin up a cluster when testEnv.Start() is run.
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	// TODO: can scheme and client build-out be consolidated to similar logic
	// or the same logic as done for the controller?
	err = corev1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = apiextv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// err = appsv1.AddToScheme(scheme.Scheme)
	// Expect(err).NotTo(HaveOccurred())

	err = operatorsv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	// This suite requires an instance of CertManagerDeployment
	Expect(k8sClient.Create(context.TODO(), &cr)).To(Succeed())
	// TODO we need certificate and issuer for these tests
	// but they're not handled currently

	close(done)
}, 60)

var _ = AfterSuite(func() {
	By("tearing down dependencies")
	Expect(k8sClient.Delete(context.TODO(), &cr)).To(Succeed())

	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})
