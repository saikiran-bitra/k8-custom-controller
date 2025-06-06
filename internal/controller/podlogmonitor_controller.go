/*
Copyright 2025 Saikiran Bitra.

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
package controller

import (
	"context"
	"strings"
	"bufio"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	monitoringv1 "mydomain.com/m/api/v1"
)

// PodLogMonitorReconciler reconciles a PodLogMonitor object
type PodLogMonitorReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Log       logr.Logger
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
	logger := log.FromContext(ctx).WithValues("podlogmonitor", req.NamespacedName)

	// Fetch the PodLogMonitor resource
	var podLogMonitor monitoringv1.PodLogMonitor
	if err := r.Get(ctx, client.ObjectKey{Name: req.Name}, &podLogMonitor); err != nil {
		logger.Error(err, "unable to fetch PodLogMonitor")
		return reconcile.Result{}, client.IgnoreNotFound(err)
	}

	// Fetch Pod logs from the specified namespace
	var podList corev1.PodList
	if err := r.List(ctx, &podList, client.InNamespace(podLogMonitor.Spec.Namespace)); err != nil {
		logger.Error(err, "unable to list pods")
		return reconcile.Result{}, err
	}

	// Iterate over the pods and check logs
	for _, pod := range podList.Items {
		if strings.HasPrefix(pod.Name, "pod-log-monitor") {
                        continue // Skip controller pods or any pods starting with 'pod-log-monitor'
                }
                if pod.Status.Phase != corev1.PodRunning {
                        logger.Info("Skipping non-running pod", "pod", pod.Name, "phase", pod.Status.Phase)
                        continue
                }
		logger.Info("found pod", "pod", pod.Name)

		if err := r.checkPodLogs(ctx, &pod, podLogMonitor.Spec.LogMessage, &podLogMonitor); err != nil {
			logger.Error(err, "unable to check pod logs", "pod", pod.Name)
			continue
		}
	}

	// Return after processing
	return reconcile.Result{}, nil
}

func (r *PodLogMonitorReconciler) checkPodLogs(ctx context.Context, pod *corev1.Pod, logMessage string, podLogMonitor *monitoringv1.PodLogMonitor) error {
	logger := log.FromContext(ctx).WithValues("podlogmonitor", pod.Name)

	// Check if we have the clientset; if not, create it
	if r.clientset == nil {
		logger.Info("Creating a new clientset...")
		config, err := rest.InClusterConfig()
		if err != nil {
			logger.Error(err, "unable to create in-cluster config")
			return err
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			logger.Error(err, "unable to create Kubernetes clientset")
			return err
		}
		r.clientset = clientset
	}
	// Get the logs of the pod using the Kubernetes API
	req := r.clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{Follow: true,})
	podLogs, err := req.Stream(ctx)
	if err != nil {
		logger.Error(err, "unable to stream pod logs")
		return err
	}
	defer podLogs.Close()
	// Check if the log message exists in the pod logs
	scanner := bufio.NewScanner(podLogs)
        for scanner.Scan() {
                logStr := scanner.Text()

	if strings.Contains(logStr, logMessage) {
		logger.Info("Found the specified log message, restarting pod")
		// Restart the pod by deleting it (will be recreated by deployment/statefulset/etc.)
		err = r.clientset.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{})
		if err != nil {
			logger.Error(err, "unable to delete pod", "pod", pod.Name)
			return err
		}
		logger.Info("Pod restarted", "pod", pod.Name)
		// Patch the PodLogMonitor status
		orig := podLogMonitor.DeepCopy()
		patch := client.MergeFrom(orig)
		podLogMonitor.Status.LastRestartedPodName = pod.Name
		podLogMonitor.Status.LastRestartTime = metav1.Now()

		logger.Info("Updating PodLogMonitor status", "namespace", podLogMonitor.Namespace, "name", podLogMonitor.Name)

		if err := r.Status().Patch(ctx, podLogMonitor, patch); err != nil {
			logger.Error(err, "unable to update PodLogMonitor status")
			return err
		}
	}
	}
	return nil
}

func (r *PodLogMonitorReconciler) startPodWatcher() {
        config, err := rest.InClusterConfig()
        if err != nil {
                r.Log.Error(err, "failed to get in-cluster config")
                return
        }

        clientset, err := kubernetes.NewForConfig(config)
        if err != nil {
                r.Log.Error(err, "failed to create clientset")
                return
        }
        r.clientset = clientset

        watcher, err := clientset.CoreV1().Pods("").Watch(context.TODO(), metav1.ListOptions{})
        if err != nil {
                r.Log.Error(err, "failed to start pod watcher")
                return
        }

        for event := range watcher.ResultChan() {
                pod, ok := event.Object.(*corev1.Pod)
                if !ok {
                        continue
                }
                if pod.Status.Phase == corev1.PodRunning && isPodReady(pod) {
                        // check if this pod matches any CR
                        var monitors monitoringv1.PodLogMonitorList
                        if err := r.Client.List(context.TODO(), &monitors); err != nil {
                                r.Log.Error(err, "failed to list PodLogMonitor resources")
                                continue
                        }
                        for _, monitor := range monitors.Items {
                                if monitor.Spec.Namespace == pod.Namespace {
                                        // Call log checking logic
                                        go r.checkPodLogs(context.TODO(), pod, monitor.Spec.LogMessage, &monitor)
                                }
                        }
                }
        }
}

func isPodReady(pod *corev1.Pod) bool {
        if pod.Status.Phase != corev1.PodRunning {
                return false
        }
        for _, cond := range pod.Status.Conditions {
                if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
                        return true
                }
        }
        return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodLogMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	go r.startPodWatcher()
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitoringv1.PodLogMonitor{}).
		Complete(r)
}
