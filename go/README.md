# Go Resource Monitor Application

A Go application that continuously monitors and prints allocated CPU and memory resources from cgroup information. This application is designed to run in containerized environments, particularly on Kubernetes.

## Features

- Detects and supports both cgroup v1 and v2
- Displays Go runtime memory statistics (Alloc, Sys, HeapAlloc, etc.)
- Shows cgroup CPU limits and usage statistics
- Shows cgroup memory limits and usage statistics
- Continuous monitoring with 5-second intervals
- Lightweight and efficient Go implementation
- Very small container image size (~10MB)

## Prerequisites

- Go 1.21 or later (for local development)
- Docker (for building container image)
- Kubernetes cluster (for deployment)

## Project Structure

```
go/
├── main.go           # Main application
├── go.mod            # Go module file
├── k8s/
│   └── pod.yaml      # Kubernetes Pod manifest
├── Dockerfile        # Multi-stage Docker build
├── .gitignore        # Git ignore file
├── .dockerignore     # Docker ignore file
└── README.md         # This file
```

## Building

### Local Build

Build the binary:

```bash
go build -o resource-monitor main.go
```

### Running Locally

```bash
./resource-monitor
```

Or directly with Go:

```bash
go run main.go
```

### Building Docker Image

```bash
docker build -t go-resource-monitor:latest .
```

## Deployment on Kubernetes

### 1. Build and Load the Image

If using Minikube:

```bash
# Build the image
docker build -t go-resource-monitor:latest .

# Load into Minikube
minikube image load go-resource-monitor:latest
```

If using a remote registry:

```bash
# Tag the image
docker tag go-resource-monitor:latest your-registry/go-resource-monitor:latest

# Push to registry
docker push your-registry/go-resource-monitor:latest

# Update k8s/pod.yaml with the new image name
```

### 2. Deploy the Pod

```bash
kubectl apply -f k8s/pod.yaml
```

### 3. View the Output

```bash
# View logs
kubectl logs go-resource-monitor -f

# Check pod status
kubectl get pod go-resource-monitor

# Describe pod for more details
kubectl describe pod go-resource-monitor
```

### 4. Clean Up

```bash
kubectl delete -f k8s/pod.yaml
```

## Configuration

### Resource Limits

The Pod manifest (`k8s/pod.yaml`) includes resource requests and limits:

- Memory Request: 64Mi
- Memory Limit: 128Mi
- CPU Request: 250m (0.25 cores)
- CPU Limit: 500m (0.5 cores)

You can modify these values in the manifest according to your needs.

### Go Runtime Configuration

The application uses the following environment variables:

- `GOMAXPROCS=0`: Automatically uses all available CPUs (respects cgroup limits)
- `GOGC=100`: Default garbage collection target percentage

## Output Example

The application will print output similar to:

```
================================================================================
Resource Monitor Started (Go)
================================================================================
Detected cgroup version: V2
================================================================================

[2025-11-06 10:30:15]
--------------------------------------------------------------------------------
CPU Resources:
  Available CPUs (Go runtime): 4
  GOMAXPROCS: 4
  cgroup v2 cpu.max: 50000 100000
  CPU Limit: 0.50 cores
  cgroup v2 cpu.stat:
    usage_usec 123456
    user_usec 98765
    system_usec 24691

Memory Resources:
  Go Alloc: 2.50 MB
  Go TotalAlloc: 5.00 MB
  Go Sys: 15.00 MB
  Go NumGC: 3
  Go HeapAlloc: 2.50 MB
  Go HeapSys: 8.00 MB
  Go HeapInuse: 4.00 MB
  cgroup v2 memory.max: 128.00 MB
  cgroup v2 memory.current: 20.00 MB
  cgroup v2 memory.stat (selected):
    anon: 15.00 MB
    file: 5.00 MB
    kernel_stack: 256.00 KB
    slab: 2.00 MB

================================================================================
```

## Advantages of Go Implementation

Compared to the Java version, the Go implementation offers:

1. **Smaller Memory Footprint**: Go uses significantly less memory than the JVM
2. **Faster Startup**: Near-instantaneous startup time
3. **Smaller Container Image**: ~10MB vs ~200MB+ for Java
4. **Native Binary**: No runtime dependencies
5. **Better Container Integration**: Native cgroup support

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

### Build Issues

If you encounter build issues:

1. Ensure you have Go 1.21 or later installed: `go version`
2. Run `go mod tidy` to sync dependencies
3. Clear the Go build cache: `go clean -cache`

## Development

### Adding Dependencies

```bash
# Add a new dependency
go get <package>

# Update all dependencies
go get -u ./...

# Tidy up go.mod and go.sum
go mod tidy
```

### Testing Locally

To test the application locally without containers:

```bash
go run main.go
```

Note: On non-containerized systems, some cgroup paths may not exist, and the application will report that it cannot read those files.

## License

This project is provided as-is for educational and testing purposes.
