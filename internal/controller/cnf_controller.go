package controllers

import (
    "context"
    "net/http"
    "io/ioutil"
    "os/exec"
    "bytes"
    "io"

    appsv1alpha1 "github.com/dolphy17/cnf-operator/api/v1alpha1"
    "k8s.io/apimachinery/pkg/api/errors"
    "k8s.io/apimachinery/pkg/runtime"
    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "k8s.io/apimachinery/pkg/util/yaml"
    "sigs.k8s.io/controller-runtime/pkg/client"
    "sigs.k8s.io/controller-runtime/pkg/log"
    ctrl "sigs.k8s.io/controller-runtime"
)

// CNFReconciler reconciles a CNF object
type CNFReconciler struct {
    client.Client
    Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *CNFReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := log.FromContext(ctx)

    // Fetch the CNF instance
    cnf := &appsv1alpha1.CNF{}
    err := r.Get(ctx, req.NamespacedName, cnf)
    if err != nil {
        if errors.IsNotFound(err) {
            // Request object not found, could have been deleted after reconcile request.
            // Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
            // Return and don't requeue
            return ctrl.Result{}, nil
        }
        // Error reading the object - requeue the request.
        return ctrl.Result{}, err
    }

    // Fetch Helm chart and convert to manifests
    helmChartURL := cnf.Spec.HelmChartURL
    manifests, err := fetchAndConvertHelmChart(helmChartURL)
    if err != nil {
        log.Error(err, "unable to fetch and convert Helm chart")
        return ctrl.Result{}, err
    }

    // Apply manifests to the cluster
    for _, manifest := range manifests {
        if err := r.applyManifest(ctx, manifest); err != nil {
            log.Error(err, "unable to apply manifest")
            return ctrl.Result{}, err
        }
    }

    // Update status
    cnf.Status.Deployed = true
    if err := r.Status().Update(ctx, cnf); err != nil {
        log.Error(err, "unable to update CNF status")
        return ctrl.Result{}, err
    }

    return ctrl.Result{}, nil
}

func fetchAndConvertHelmChart(url string) ([]unstructured.Unstructured, error) {
    // Fetch the Helm chart
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    // Read the Helm chart
    chartData, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }

    // Save the chart to a temporary file
    tmpFile, err := ioutil.TempFile("", "helm-chart-*.tgz")
    if err != nil {
        return nil, err
    }
    defer tmpFile.Close()

    if _, err := tmpFile.Write(chartData); err != nil {
        return nil, err
    }

    // Run helm template to convert the chart to manifests
    cmd := exec.Command("helm", "template", tmpFile.Name())
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }

    // Parse the manifests
    var manifests []unstructured.Unstructured
    decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader(output), 4096)
    for {
        var manifest unstructured.Unstructured
        if err := decoder.Decode(&manifest); err != nil {
            if err == io.EOF {
                break
            }
            return nil, err
        }
        manifests = append(manifests, manifest)
    }

    return manifests, nil
}

func (r *CNFReconciler) applyManifest(ctx context.Context, manifest unstructured.Unstructured) error {
    // Apply the manifest to the cluster
    err := r.Client.Create(ctx, &manifest)
    if err != nil && !errors.IsAlreadyExists(err) {
        return err
    }
    return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CNFReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&appsv1alpha1.CNF{}).
        Complete(r)
}
