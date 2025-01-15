package controllers

import (
    "testing"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/client-go/kubernetes/scheme"
    ctrl "sigs.k8s.io/controller-runtime"
    "sigs.k8s.io/controller-runtime/pkg/envtest/printer"
    "sigs.k8s.io/controller-runtime/pkg/log/zap"
    "sigs.k8s.io/controller-runtime/pkg/envtest"
    "sigs.k8s.io/controller-runtime/pkg/client"
    appsv1alpha1 "github.com/dolphy17/cnf-operator/api/v1alpha1"
    "github.com/onsi/ginkgo"
    "github.com/onsi/gomega"
)

var (
    k8sClient client.Client
    testEnv   *envtest.Environment
)

func TestAPIs(t *testing.T) {
    gomega.RegisterFailHandler(ginkgo.Fail)

    ginkgo.RunSpecsWithDefaultAndCustomReporters(t,
        "Controller Suite",
        []ginkgo.Reporter{printer.NewlineReporter{}})
}

var _ = ginkgo.BeforeSuite(func(done ginkgo.Done) {
    zap.New(zap.WriteTo(ginkgo.GinkgoWriter), zap.UseDevMode(true))

    testEnv = &envtest.Environment{
        CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
    }

    cfg, err := testEnv.Start()
    gomega.Expect(err).NotTo(gomega.HaveOccurred())
    gomega.Expect(cfg).NotTo(gomega.BeNil())

    err = appsv1alpha1.AddToScheme(scheme.Scheme)
    gomega.Expect(err).NotTo(gomega.HaveOccurred())

    k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
    gomega.Expect(err).NotTo(gomega.HaveOccurred())
    gomega.Expect(k8sClient).NotTo(gomega.BeNil())

    close(done)
}, 60)

var _ = ginkgo.AfterSuite(func() {
    err := testEnv.Stop()
    gomega.Expect(err).NotTo(gomega.HaveOccurred())
})
