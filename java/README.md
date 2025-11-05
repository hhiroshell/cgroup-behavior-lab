# Java Resource Monitor Application

A Java application that continuously monitors and prints allocated CPU and memory resources from cgroup information. This application is designed to run in containerized environments, particularly on Kubernetes.

## Features

- Detects and supports both cgroup v1 and v2
- Displays JVM memory information (max, total, used, free)
- Shows cgroup CPU limits and usage statistics
- Shows cgroup memory limits and usage statistics
- Continuous monitoring with 5-second intervals
- Container-aware JVM configuration

## Prerequisites

- Java 17 or later
- Docker (for building container image)
- Kubernetes cluster (for deployment)

## Project Structure

```
java/
├── src/
│   └── main/
│       └── java/
│           └── com/example/resourcemonitor/
│               └── ResourceMonitorApp.java
├── k8s/
│   └── pod.yaml
├── .mvn/
│   └── wrapper/
│       └── maven-wrapper.properties
├── mvnw
├── pom.xml
├── Dockerfile
└── README.md
```

## Building

### Using Maven Wrapper

Build the JAR file:

```bash
./mvnw clean package
```

### Running Locally

```bash
java -jar target/resource-monitor.jar
```

### Building Docker Image

```bash
docker build -t resource-monitor:latest .
```

## Deployment on Kubernetes

### 1. Build and Load the Image

If using Minikube:

```bash
# Build the image
docker build -t resource-monitor:latest .

# Load into Minikube
minikube image load resource-monitor:latest
```

If using a remote registry:

```bash
# Tag the image
docker tag resource-monitor:latest your-registry/resource-monitor:latest

# Push to registry
docker push your-registry/resource-monitor:latest

# Update k8s/pod.yaml with the new image name
```

### 2. Deploy the Pod

```bash
kubectl apply -f k8s/pod.yaml
```

### 3. View the Output

```bash
# View logs
kubectl logs java-resource-monitor -f

# Check pod status
kubectl get pod java-resource-monitor

# Describe pod for more details
kubectl describe pod java-resource-monitor
```

### 4. Clean Up

```bash
kubectl delete -f k8s/pod.yaml
```

## Configuration

### Resource Limits

The Pod manifest (`k8s/pod.yaml`) includes resource requests and limits:

- Memory Request: 256Mi
- Memory Limit: 512Mi
- CPU Request: 500m (0.5 cores)
- CPU Limit: 1000m (1 core)

You can modify these values in the manifest according to your needs.

### JVM Options

The application uses the following JVM options for container awareness:

- `-XX:+UseContainerSupport`: Enables container support
- `-XX:MaxRAMPercentage=75.0`: Uses up to 75% of available memory
- `-XX:InitialRAMPercentage=50.0`: Starts with 50% of available memory

## Output Example

The application will print output similar to:

```
================================================================================
Resource Monitor Started
================================================================================
Detected cgroup version: V2
================================================================================

[2025-11-06 10:30:15]
--------------------------------------------------------------------------------
CPU Resources:
  Available Processors (JVM): 4
  cgroup v2 cpu.max: 100000 100000
  CPU Limit: 1.00 cores
  cgroup v2 cpu.stat:
    usage_usec 1234567
    user_usec 987654
    system_usec 246913

Memory Resources:
  JVM Max Memory: 384.00 MB
  JVM Total Memory: 256.00 MB
  JVM Used Memory: 128.00 MB
  JVM Free Memory: 128.00 MB
  cgroup v2 memory.max: 512.00 MB
  cgroup v2 memory.current: 256.00 MB
  cgroup v2 memory.stat (selected):
    anon: 128.00 MB
    file: 64.00 MB
    kernel_stack: 1.50 MB
    slab: 8.00 MB

================================================================================
```

## Troubleshooting

### Pod ImagePullBackOff

If you see `ImagePullBackOff` error:

1. Ensure the image is properly loaded (for Minikube: `minikube image ls`)
2. Check `imagePullPolicy` is set to `IfNotPresent` in the manifest
3. Verify the image name matches exactly

### Permission Issues

If you see permission errors reading cgroup files:

1. Ensure the container has proper security context
2. Check that the Kubernetes node is using cgroup v2 or v1
3. Verify the Pod is running with appropriate privileges

## License

This project is provided as-is for educational and testing purposes.
