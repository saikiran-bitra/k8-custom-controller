package controller

import (
        "context"
        "strings"
        "time"

        corev1 "k8s.io/api/core/v1"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

        "k8s.io/client-go/kubernetes"
        "k8s.io/apimachinery/pkg/runtime"
        "k8s.io/client-go/rest"
        ctrl "sigs.k8s.io/controller-runtime"
        "sigs.k8s.io/controller-runtime/pkg/client"
        "sigs.k8s.io/controller-runtime/pkg/log"
        "sigs.k8s.io/controller-runtime/pkg/reconcile"
        "github.com/go-logr/logr"

        monitoringv1 "mydomain.com/m/api/v1"
)

// PodLogMonitorReconciler reconciles a PodLogMonitor object
type PodLogMonitorReconciler struct {
        client.Client
        Scheme *runtime.Scheme
        Log logr.Logger
        clientset *kubernetes.Clientset
}

// +kubebuilder:rbac:groups=monitoring.mydomain.com,resources=podlogmonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=monitoring.mydomain.com,resources=podlogmonitors/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=monitoring.mydomain.com,resources=podlogmonitors/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PodLogMonitor object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.0/pkg/reconcile
func (r *PodLogMonitorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
        _ = log.FromContext(ctx)

        log := r.Log.WithValues("podlogmonitor", req.NamespacedName)
        log.Info("Line - 1")
        // Fetch the PodLogMonitor resource
        var podLogMonitor monitoringv1.PodLogMonitor
        if err := r.Get(ctx, req.NamespacedName, &podLogMonitor); err != nil {
                log.Error(err, "unable to fetch PodLogMonitor")
                return reconcile.Result{}, client.IgnoreNotFound(err)
        }

        log.Info("Line - 2")
        // Fetch Pod logs from the specified namespace
        podList := &corev1.PodList{}
        if err := r.List(ctx, podList, client.MatchingLabels{"someLabel": podLogMonitor.Spec.Namespace}); err != nil {
        //if err := r.List(ctx, podList, client.InNamespace(podLogMonitor.Spec.Namespace)); err != nil {
                log.Error(err, "unable to list pods")
                return reconcile.Result{}, err
        }

        log.Info("Line - 3")
        // Iterate over the pods and check logs
        for _, pod := range podList.Items {
                log.Info("In the for loop")
                if err := r.checkPodLogs(ctx, &pod, podLogMonitor.Spec.LogMessage, &podLogMonitor); err != nil {
                log.Error(err, "unable to check pod logs", "pod", pod.Name)
                continue
                }
        }

        log.Info("Line - 4")
        // Save the updated status
        if err := r.Status().Update(ctx, &podLogMonitor); err != nil {
                log.Error(err, "unable to update PodLogMonitor status")
                return reconcile.Result{}, err
        }
        log.Info("Line - 5")
        // Return after processing
        return reconcile.Result{}, nil
}


func (r *PodLogMonitorReconciler) checkPodLogs(ctx context.Context, pod *corev1.Pod, logMessage string, podLogMonitor *monitoringv1.PodLogMonitor) error {
   log := r.Log.WithValues("pod", pod.Name)
   log.Info("Line - 6")
   // Check if we have the clientset; if not, create it
    if r.clientset == nil {
        log.Info("Creating a new clientset...")
        config, err := rest.InClusterConfig()
        if err != nil {
            log.Error(err, "unable to create in-cluster config")
            return err
        }
        clientset, err := kubernetes.NewForConfig(config)
        if err != nil {
            log.Error(err, "unable to create Kubernetes clientset")
            return err
        }
        r.clientset = clientset
    }
   // Get the logs of the pod using the Kubernetes API
   req := r.clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{})
   podLogs, err := req.Stream(ctx)
   if err != nil {
       log.Error(err, "unable to stream pod logs")
       return err
   }
   defer podLogs.Close()
   // Check if the log message exists in the pod logs
   buf := make([]byte, 2000)
   _, err = podLogs.Read(buf)
   if err != nil {
       log.Error(err, "error reading pod logs")
       return err
   }
   logStr := string(buf)
   if strings.Contains(logStr, logMessage) {
       log.Info("Found the specified log message, restarting pod")
       // Restart the pod by deleting it (will be recreated by deployment/statefulset/etc.)
       err = r.clientset.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
       if err != nil {
           log.Error(err, "unable to delete pod", "pod", pod.Name)
           return err
       }
       log.Info("Pod restarted", "pod", pod.Name)
       // Update the PodLogMonitor status
       podLogMonitor.Status.LastRestartedPodName = pod.Name
       podLogMonitor.Status.LastRestartTime = metav1.NewTime(time.Now())

                if err := r.Status().Update(ctx, podLogMonitor); err != nil {
           log.Error(err, "unable to update PodLogMonitor status")
           return err
       }
   }
   return nil
}


// SetupWithManager sets up the controller with the Manager.
func (r *PodLogMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
        return ctrl.NewControllerManagedBy(mgr).
                For(&monitoringv1.PodLogMonitor{}).
                Complete(r)
}
