package detector

import (
	"context"
	"fmt"
	"time"

	"github.com/Maniratnam557/k8s-pod-detective/pkg/explainer"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PodDetector struct {
	clientset *kubernetes.Clientset
	seen      map[string]bool
	options   Options
}

type Options struct {
	PodName       string
	LabelSelector string
}

func New(clientset *kubernetes.Clientset, opts Options) *PodDetector {
	return &PodDetector{
		clientset: clientset,
		seen:      make(map[string]bool),
		options:   opts,
	}
}

// WatchPods monitors pods for failures
func (d *PodDetector) WatchPods(namespace string) error {
	fmt.Printf("🔍 Watching pods in namespace: %s\n\n", namespace)

	for {

		listOptions := metav1.ListOptions{}

		if d.options.LabelSelector != "" {
			listOptions.LabelSelector = d.options.LabelSelector
		}

		if d.options.PodName != "" {
			listOptions.FieldSelector = fmt.Sprintf("metadata.name=%s", d.options.PodName)
		}

		fmt.Printf("[DEBUG] Querying namespace='%s'\n", namespace)

		pods, err := d.clientset.CoreV1().Pods(namespace).List(
			context.TODO(),
			listOptions,
		)

		if err != nil {
			return fmt.Errorf("failed to list pods: %w", err)
		}

		// Check each pod
		for _, pod := range pods.Items {
			d.checkPod(&pod)
		}

		// Wait before next check
		time.Sleep(10 * time.Second)
	}
}

func (d *PodDetector) checkPod(pod *corev1.Pod) {
	podKey := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)

	// Check container statuses
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.State.Waiting != nil {
			waiting := containerStatus.State.Waiting

			// Detect failure reasons
			if d.isFailureReason(waiting.Reason) {
				// Create unique key to avoid duplicate reports
				statusKey := fmt.Sprintf("%s-%s-%s", podKey, containerStatus.Name, waiting.Reason)

				if !d.seen[statusKey] {
					info := d.gatherFailureInfo(pod, containerStatus, waiting)
					explanation := explainer.Explain(info)
					fmt.Println(explanation)
					fmt.Println("=====================================\n")
					d.seen[statusKey] = true
				}
			}
		}

		// Check terminated state
		if containerStatus.State.Terminated != nil {
			terminated := containerStatus.State.Terminated
			if terminated.ExitCode != 0 {
				statusKey := fmt.Sprintf("%s-%s-terminated-%d", podKey, containerStatus.Name, terminated.ExitCode)

				if !d.seen[statusKey] {
					info := d.gatherTerminationInfo(pod, containerStatus, terminated)
					explanation := explainer.Explain(info)
					fmt.Println(explanation)
					fmt.Println("=====================================\n")
					d.seen[statusKey] = true
				}
			}
		}
	}
}

func (d *PodDetector) isFailureReason(reason string) bool {
	failureReasons := []string{
		"CrashLoopBackOff",
		"ImagePullBackOff",
		"ErrImagePull",
		"CreateContainerConfigError",
		"InvalidImageName",
		"RunContainerError",
	}

	for _, fr := range failureReasons {
		if reason == fr {
			return true
		}
	}
	return false
}

func (d *PodDetector) gatherFailureInfo(
	pod *corev1.Pod,
	status corev1.ContainerStatus,
	waiting *corev1.ContainerStateWaiting,
) explainer.FailureInfo {

	// Get last logs if available
	lastLog := d.getLastLog(pod.Namespace, pod.Name, status.Name)

	return explainer.FailureInfo{
		PodName:       pod.Name,
		Namespace:     pod.Namespace,
		ContainerName: status.Name,
		Reason:        waiting.Reason,
		Message:       waiting.Message,
		ExitCode:      0,
		LastLog:       lastLog,
	}
}

func (d *PodDetector) gatherTerminationInfo(
	pod *corev1.Pod,
	status corev1.ContainerStatus,
	terminated *corev1.ContainerStateTerminated,
) explainer.FailureInfo {

	reason := terminated.Reason
	if reason == "" {
		reason = "CrashLoopBackOff"
	}

	lastLog := d.getLastLog(pod.Namespace, pod.Name, status.Name)

	return explainer.FailureInfo{
		PodName:       pod.Name,
		Namespace:     pod.Namespace,
		ContainerName: status.Name,
		Reason:        reason,
		Message:       terminated.Message,
		ExitCode:      terminated.ExitCode,
		LastLog:       lastLog,
	}
}

func (d *PodDetector) getLastLog(namespace, podName, containerName string) string {
	tailLines := int64(10)
	logOptions := &corev1.PodLogOptions{
		Container: containerName,
		TailLines: &tailLines,
	}

	logs, err := d.clientset.CoreV1().Pods(namespace).
		GetLogs(podName, logOptions).Do(context.TODO()).Raw()

	if err != nil {
		return ""
	}

	return string(logs)
}
