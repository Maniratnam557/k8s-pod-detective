package explainer

import (
	"fmt"
	"strings"
)

// FailureInfo contains details about a pod failure
type FailureInfo struct {
	PodName       string
	Namespace     string
	ContainerName string
	Reason        string
	Message       string
	ExitCode      int32
	LastLog       string
}

// Explain generates a human-friendly explanation with debug commands
func Explain(info FailureInfo) string {
	var explanation strings.Builder

	explanation.WriteString(fmt.Sprintf("🚨 PROBLEM DETECTED\n"))
	explanation.WriteString(fmt.Sprintf("=====================================\n"))
	explanation.WriteString(fmt.Sprintf("Pod: %s/%s\n", info.Namespace, info.PodName))
	explanation.WriteString(fmt.Sprintf("Container: %s\n\n", info.ContainerName))

	// Analyze based on reason
	switch info.Reason {
	case "CrashLoopBackOff":
		explanation.WriteString(explainCrashLoopBackOff(info))
	case "ImagePullBackOff", "ErrImagePull":
		explanation.WriteString(explainImagePullError(info))
	case "OOMKilled":
		explanation.WriteString(explainOOMKilled(info))
	case "CreateContainerConfigError":
		explanation.WriteString(explainConfigError(info))
	case "RunContainerError":
		explanation.WriteString(explainRunContainerError(info))
	case "InvalidImageName":
		explanation.WriteString(explainInvalidImageName(info))
	default:
		explanation.WriteString(explainGeneric(info))
	}

	return explanation.String()
}

func explainCrashLoopBackOff(info FailureInfo) string {
	explanation := "❌ WHAT HAPPENED:\n"
	explanation += "Your container keeps crashing and restarting.\n\n"

	explanation += "🤔 WHAT THIS MEANS:\n"
	explanation += "The application inside the container starts but then immediately fails.\n"
	explanation += "Kubernetes tried to restart it multiple times but it keeps crashing.\n\n"

	// Analyze exit code
	if info.ExitCode != 0 {
		explanation += fmt.Sprintf("Exit Code: %d\n", info.ExitCode)
		explanation += explainExitCode(info.ExitCode) + "\n\n"
	}

	// Show last error if available
	if info.LastLog != "" {
		explanation += "📝 LAST ERROR MESSAGE:\n"
		explanation += info.LastLog + "\n\n"
	}

	explanation += "🔧 HOW TO FIX:\n"
	explanation += "1. Check application logs for startup errors\n"
	explanation += "2. Verify environment variables and configuration\n"
	explanation += "3. Test the container image locally\n"
	explanation += "4. Check dependencies (database, APIs, etc.)\n\n"

	explanation += "🐛 DEBUG COMMANDS:\n"
	explanation += "-------------------\n\n"

	explanation += "# 1. View recent logs (last 50 lines)\n"
	explanation += fmt.Sprintf("kubectl logs %s -n %s --tail=50\n\n", info.PodName, info.Namespace)

	explanation += "# 2. View logs from previous crash\n"
	explanation += fmt.Sprintf("kubectl logs %s -n %s --previous\n\n", info.PodName, info.Namespace)

	explanation += "# 3. View all logs with timestamps\n"
	explanation += fmt.Sprintf("kubectl logs %s -n %s --timestamps=true --all-containers=true\n\n", info.PodName, info.Namespace)

	explanation += "# 4. Stream logs in real-time\n"
	explanation += fmt.Sprintf("kubectl logs %s -n %s -f\n\n", info.PodName, info.Namespace)

	explanation += "# 5. Get detailed pod information\n"
	explanation += fmt.Sprintf("kubectl describe pod %s -n %s\n\n", info.PodName, info.Namespace)

	explanation += "# 6. Check pod events (last activities)\n"
	explanation += fmt.Sprintf("kubectl get events -n %s --field-selector involvedObject.name=%s --sort-by='.lastTimestamp'\n\n", info.Namespace, info.PodName)

	explanation += "# 7. Get pod YAML configuration\n"
	explanation += fmt.Sprintf("kubectl get pod %s -n %s -o yaml\n\n", info.PodName, info.Namespace)

	explanation += "# 8. Check environment variables\n"
	explanation += fmt.Sprintf("kubectl exec %s -n %s -- env\n\n", info.PodName, info.Namespace)

	explanation += "# 9. Try to exec into container (if it stays up long enough)\n"
	explanation += fmt.Sprintf("kubectl exec -it %s -n %s -- /bin/sh\n\n", info.PodName, info.Namespace)

	explanation += "# 10. Check resource usage\n"
	explanation += fmt.Sprintf("kubectl top pod %s -n %s\n\n", info.PodName, info.Namespace)

	explanation += "📊 COMMON CAUSES:\n"
	explanation += "- Missing required environment variables\n"
	explanation += "- Database connection failures\n"
	explanation += "- External service unavailable\n"
	explanation += "- Configuration file errors\n"
	explanation += "- Application code bugs\n"
	explanation += "- Port already in use\n"
	explanation += "- File system permissions\n"

	return explanation
}

func explainImagePullError(info FailureInfo) string {
	explanation := "❌ WHAT HAPPENED:\n"
	explanation += "Kubernetes cannot download your container image.\n\n"

	explanation += "🤔 WHAT THIS MEANS:\n"
	explanation += "The image specified in your deployment doesn't exist, has the wrong name,\n"
	explanation += "or Kubernetes doesn't have permission to pull it from the registry.\n\n"

	explanation += "🔧 HOW TO FIX:\n"
	explanation += "1. Verify the image name and tag are correct\n"
	explanation += "2. Check if the image exists in the registry\n"
	explanation += "3. Ensure image pull secrets are configured correctly\n"
	explanation += "4. Verify registry credentials are valid\n\n"

	explanation += "🐛 DEBUG COMMANDS:\n"
	explanation += "-------------------\n\n"

	explanation += "# 1. Check pod description for image details\n"
	explanation += fmt.Sprintf("kubectl describe pod %s -n %s | grep -A5 'Image'\n\n", info.PodName, info.Namespace)

	explanation += "# 2. View detailed error message\n"
	explanation += fmt.Sprintf("kubectl describe pod %s -n %s | grep -A10 'Events'\n\n", info.PodName, info.Namespace)

	explanation += "# 3. Get pod events\n"
	explanation += fmt.Sprintf("kubectl get events -n %s --field-selector involvedObject.name=%s\n\n", info.Namespace, info.PodName)

	explanation += "# 4. Check if image pull secret exists\n"
	explanation += fmt.Sprintf("kubectl get secrets -n %s\n\n", info.Namespace)

	explanation += "# 5. Describe the image pull secret\n"
	explanation += fmt.Sprintf("kubectl get secret  -n %s -o yaml\n\n", info.Namespace)

	explanation += "# 6. Test pulling the image locally (if using Docker)\n"
	explanation += fmt.Sprintf("# First, get the image name from pod:\n")
	explanation += fmt.Sprintf("kubectl get pod %s -n %s -o jsonpath='{.spec.containers[*].image}'\n", info.PodName, info.Namespace)
	explanation += "# Then try pulling it:\n"
	explanation += "docker pull \n\n"

	explanation += "# 7. Check deployment/pod spec\n"
	explanation += fmt.Sprintf("kubectl get pod %s -n %s -o yaml | grep -A5 'image:'\n\n", info.PodName, info.Namespace)

	explanation += "# 8. List all image pull secrets in namespace\n"
	explanation += fmt.Sprintf("kubectl get serviceaccount default -n %s -o yaml | grep -A3 'imagePullSecrets'\n\n", info.Namespace)

	explanation += "📊 COMMON CAUSES:\n"
	explanation += "- Typo in image name or tag\n"
	explanation += "- Image doesn't exist in registry\n"
	explanation += "- Private registry without credentials\n"
	explanation += "- Expired or invalid image pull secret\n"
	explanation += "- Wrong registry URL\n"
	explanation += "- Tag 'latest' doesn't exist\n"
	explanation += "- Network issues accessing registry\n\n"

	explanation += "💡 CREATE IMAGE PULL SECRET:\n"
	explanation += "kubectl create secret docker-registry regcred \\\n"
	explanation += "  --docker-server= \\\n"
	explanation += "  --docker-username= \\\n"
	explanation += "  --docker-password= \\\n"
	explanation += "  --docker-email= \\\n"
	explanation += fmt.Sprintf("  -n %s\n", info.Namespace)

	return explanation
}

func explainOOMKilled(info FailureInfo) string {
	explanation := "❌ WHAT HAPPENED:\n"
	explanation += "Your container ran out of memory (OOM = Out Of Memory).\n\n"

	explanation += "🤔 WHAT THIS MEANS:\n"
	explanation += "The application used more memory than the limit you set.\n"
	explanation += "Kubernetes killed it to prevent affecting other pods on the node.\n\n"

	explanation += "🔧 HOW TO FIX:\n"
	explanation += "1. Increase memory limits in your deployment\n"
	explanation += "2. Fix memory leaks in your application\n"
	explanation += "3. Optimize memory usage\n"
	explanation += "4. Use memory profiling tools\n\n"

	explanation += "🐛 DEBUG COMMANDS:\n"
	explanation += "-------------------\n\n"

	explanation += "# 1. Check current memory limits\n"
	explanation += fmt.Sprintf("kubectl get pod %s -n %s -o jsonpath='{.spec.containers[*].resources}'\n\n", info.PodName, info.Namespace)

	explanation += "# 2. View actual memory usage (if metrics-server is installed)\n"
	explanation += fmt.Sprintf("kubectl top pod %s -n %s\n\n", info.PodName, info.Namespace)

	explanation += "# 3. Check historical resource usage\n"
	explanation += fmt.Sprintf("kubectl describe pod %s -n %s | grep -A5 'Limits\\|Requests'\n\n", info.PodName, info.Namespace)

	explanation += "# 4. View OOM events\n"
	explanation += fmt.Sprintf("kubectl get events -n %s --field-selector reason=OOMKilling\n\n", info.Namespace)

	explanation += "# 5. Check node memory pressure\n"
	explanation += "kubectl describe nodes | grep -A5 'Memory'\n\n"

	explanation += "# 6. Get pod restart count\n"
	explanation += fmt.Sprintf("kubectl get pod %s -n %s -o jsonpath='{.status.containerStatuses[*].restartCount}'\n\n", info.PodName, info.Namespace)

	explanation += "# 7. View logs before OOM kill\n"
	explanation += fmt.Sprintf("kubectl logs %s -n %s --previous --tail=100\n\n", info.PodName, info.Namespace)

	explanation += "📊 HOW TO INCREASE MEMORY:\n\n"
	explanation += "Edit your deployment/pod spec:\n\n"
	explanation += "resources:\n"
	explanation += "  requests:\n"
	explanation += "    memory: \"256Mi\"  # Minimum guaranteed\n"
	explanation += "  limits:\n"
	explanation += "    memory: \"512Mi\"  # Maximum allowed (INCREASE THIS)\n\n"

	explanation += "Then apply changes:\n"
	explanation += "kubectl edit deployment  -n " + info.Namespace + "\n\n"

	explanation += "📊 COMMON CAUSES:\n"
	explanation += "- Memory limit set too low\n"
	explanation += "- Memory leak in application\n"
	explanation += "- Loading too much data at once\n"
	explanation += "- Inefficient caching\n"
	explanation += "- Large file processing\n"

	return explanation
}

func explainConfigError(info FailureInfo) string {
	explanation := "❌ WHAT HAPPENED:\n"
	explanation += "There's a problem with your container configuration.\n\n"

	explanation += "🤔 WHAT THIS MEANS:\n"
	explanation += "Kubernetes found an error in your pod/container configuration\n"
	explanation += "before it could even start the container.\n\n"

	explanation += "🔧 HOW TO FIX:\n"
	explanation += "1. Verify all ConfigMaps and Secrets exist\n"
	explanation += "2. Check volume mount paths are correct\n"
	explanation += "3. Ensure environment variables reference valid resources\n"
	explanation += "4. Validate YAML syntax\n\n"

	explanation += "🐛 DEBUG COMMANDS:\n"
	explanation += "-------------------\n\n"

	explanation += "# 1. Get detailed error description\n"
	explanation += fmt.Sprintf("kubectl describe pod %s -n %s\n\n", info.PodName, info.Namespace)

	explanation += "# 2. Check if referenced ConfigMaps exist\n"
	explanation += fmt.Sprintf("kubectl get configmaps -n %s\n\n", info.Namespace)

	explanation += "# 3. Check if referenced Secrets exist\n"
	explanation += fmt.Sprintf("kubectl get secrets -n %s\n\n", info.Namespace)

	explanation += "# 4. View pod YAML to find configuration issues\n"
	explanation += fmt.Sprintf("kubectl get pod %s -n %s -o yaml\n\n", info.PodName, info.Namespace)

	explanation += "# 5. Check volume mounts\n"
	explanation += fmt.Sprintf("kubectl get pod %s -n %s -o jsonpath='{.spec.volumes}'\n\n", info.PodName, info.Namespace)

	explanation += "📊 COMMON CAUSES:\n"
	explanation += "- Missing ConfigMap or Secret\n"
	explanation += "- Wrong ConfigMap/Secret key name\n"
	explanation += "- Invalid volume mount path\n"
	explanation += "- Incorrect environment variable reference\n"

	return explanation
}

func explainRunContainerError(info FailureInfo) string {
	explanation := "❌ WHAT HAPPENED:\n"
	explanation += "Kubernetes couldn't start your container.\n\n"

	explanation += "🐛 DEBUG COMMANDS:\n"
	explanation += "-------------------\n\n"

	explanation += fmt.Sprintf("kubectl describe pod %s -n %s\n\n", info.PodName, info.Namespace)
	explanation += fmt.Sprintf("kubectl get events -n %s --field-selector involvedObject.name=%s\n\n", info.Namespace, info.PodName)

	return explanation
}

func explainInvalidImageName(info FailureInfo) string {
	explanation := "❌ WHAT HAPPENED:\n"
	explanation += "The container image name is invalid or malformed.\n\n"

	explanation += "🐛 DEBUG COMMANDS:\n"
	explanation += "-------------------\n\n"

	explanation += "# Check the image name\n"
	explanation += fmt.Sprintf("kubectl get pod %s -n %s -o jsonpath='{.spec.containers[*].image}'\n\n", info.PodName, info.Namespace)

	return explanation
}

func explainGeneric(info FailureInfo) string {
	explanation := fmt.Sprintf("❌ WHAT HAPPENED:\n%s\n\n", info.Reason)

	if info.Message != "" {
		explanation += "📝 ERROR MESSAGE:\n" + info.Message + "\n\n"
	}

	explanation += "🐛 DEBUG COMMANDS:\n"
	explanation += "-------------------\n\n"

	explanation += "# 1. Get detailed pod information\n"
	explanation += fmt.Sprintf("kubectl describe pod %s -n %s\n\n", info.PodName, info.Namespace)

	explanation += "# 2. View logs\n"
	explanation += fmt.Sprintf("kubectl logs %s -n %s\n\n", info.PodName, info.Namespace)

	explanation += "# 3. View previous logs (if restarted)\n"
	explanation += fmt.Sprintf("kubectl logs %s -n %s --previous\n\n", info.PodName, info.Namespace)

	explanation += "# 4. Check events\n"
	explanation += fmt.Sprintf("kubectl get events -n %s --field-selector involvedObject.name=%s --sort-by='.lastTimestamp'\n\n", info.Namespace, info.PodName)

	return explanation
}

func explainExitCode(code int32) string {
	switch code {
	case 0:
		return "→ Exit code 0: Success (but should not crash)"
	case 1:
		return "→ Exit code 1: Application error - check your code for bugs"
	case 2:
		return "→ Exit code 2: Misuse of shell command"
	case 126:
		return "→ Exit code 126: Command cannot execute (permission problem?)"
	case 127:
		return "→ Exit code 127: Command not found (binary doesn't exist?)"
	case 130:
		return "→ Exit code 130: Terminated by Ctrl+C (SIGINT)"
	case 137:
		return "→ Exit code 137: Killed by SIGKILL (usually OOM or forced termination)"
	case 139:
		return "→ Exit code 139: Segmentation fault (memory access violation)"
	case 143:
		return "→ Exit code 143: Terminated by SIGTERM (graceful shutdown)"
	case 255:
		return "→ Exit code 255: Exit status out of range"
	default:
		return fmt.Sprintf("→ Exit code %d: Check application documentation", code)
	}
}