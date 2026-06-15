# MiniDocker 🚀

MiniDocker is a high-performance, daemonless container engine built completely from scratch in Go. 

By bypassing the traditional, heavy background daemons (`dockerd`) used by mainstream container software, MiniDocker interfaces directly with raw Linux kernel subsystems. It allows users to pull official images directly from the public Docker Hub registry, extract them into localized storage layers, spin up an isolated filesystem jail, and route live internet access directly into a secure sandbox process—all with near-zero idle memory footprint.

---

## ⚡ What MiniDocker Does

MiniDocker provides an end-to-end container virtualization pipeline in a single, lightweight CLI binary. When you execute an image, the repository orchestrates the following low-level operations natively:

1. **OCI Registry Synchronization:** Handshakes with the official Docker Hub API to download layered base OS files (`.tar.gz` blobs) based on standardized content digests.
2. **Dynamic Layer Stacking:** Combines those compressed template layers on-the-fly into a single unified root directory using the Linux `OverlayFS` subsystem.
3. **Subsystem Isolation:** Spawns a dedicated process container wrapped inside secure custom Linux namespace walls and restricts resource theft (RAM/CPU) using control groups.
4. **Network Multiplexing:** Provisions virtual network bridges and ethernet pairs, mapping physical host connectivity straight into the sandboxed cell using iptables Network Address Translation (NAT).

---

## 🏆 Key Advantages

* **Near-Zero Resource Overhead:** Traditional engines maintain a persistent, memory-hungry background daemon running 24/7. MiniDocker is entirely daemonless—consuming system resources *only* while a container is actively executing.
* **No Single Point of Failure:** In standard setups, if the central container daemon crashes, every running container on the server dies with it. MiniDocker operates via a clean fork-and-forget execution model; containers run independently of any parent management service.
* **Zero External Dependencies:** Built without relying on massive runtime wrappers (like `runc` or `containerd`), the entire codebase uses native Go packages and direct kernel system calls, offering maximum educational clarity into how containment primitives actually function.
* **Hardened Security Perimeter:** Combines modern Cgroups v2 constraints with filesystem isolation to ensure a compromised containerized application cannot leak resources, access the host directory structure, or interfere with adjacent host tasks.

---

## 💎 Features

### 🛡️ Core Runtime & Sandbox Isolation (`container/`)
* **Namespace Jailing:** Leverages Linux `CLONE_NEWPID`, `CLONE_NEWUTS`, `CLONE_NEWNS`, and `CLONE_NEWNET` flags to decouple the sandbox completely from the host's process tree, hostname configuration, storage mount tables, and network stack.
* **Absolute Root Pivoting:** Implements the `pivot_root` system call to dismount the host file layout inside the container space, substituting it entirely with the sandboxed environment.
* **Cgroups v2 Resource Throttling:** Programmatically interfaces with `/sys/fs/cgroup/` to impose hard limitations on memory utilization (`memory.max`) and compute cycles (`cpu.max`).

### 🗄️ Layer-Stacking Filesystem (`storage/`)
* **OverlayFS Integration:** Programs a multi-layered mount infrastructure compiling an immutable `lowerdir` sequence (the download template) and a dynamic `upperdir` scratch space into a seamless `merged/` runtime directory.
* **Streaming Archive Extraction:** Features an optimized `.tar.gz` data extraction loop that reads layer tarballs and maps file permissions natively to standard Linux directory definitions.

### 🌐 Autonomic Networking (`network/`)
* **Virtual Bridge Engine:** Automatically provisions and scales a virtual software switch interface named `md-br0` directly inside the host kernel.
* **Veth Cable Injection:** Instantiates dual-ended Virtual Ethernet (`veth`) pipe pairs, anchoring one endpoint to the host bridge while safely projecting the secondary side directly through the running container's isolated network namespace wall.
* **Masqueraded NAT Routing:** Dynamically writes local `iptables` rule matrices to seamlessly perform Network Address Translation, routing external data signals cleanly to the container.

### 🌐 OCI Registry Core client (`registry/`)
* **Token Authentication:** Programmatically negotiates secure handshake tokens across `https://auth.docker.io`.
* **Manifest Manifest Schema Parsing:** Interrogates standard OCI distribution endpoints to cleanly decode JSON digest matrices, calculating exact layer composition trees before initiating data transfer streams.

---

## 📂 Repository Layout

```text
Mini-Docker/
│
├── main.go               # Entry point: Parses root commands and flags
├── go.mod                # Go module project configuration
├── go.sum                # Automatically managed dependency checksums
│
├── cmd/
│   └── run.go            # Central Pipeline Controller: Orchestrates all packages
│
├── container/
│   ├── namespace.go      # Configures isolation namespaces and executes pivot_root
│   └── cgroup.go         # Enforces hardware constraints via the /sys/fs/cgroup filesystem
│
├── storage/
│   ├── extract.go        # Decompresses registry cache tarballs into static layers
│   ├── overlay.go        # Executes system mount calls for runtime OverlayFS layering
│   └── storage.go        # Low-level directory tree provisioner and path cleaner
│
├── network/
│   ├── bridge.go         # Provisions host-side virtual software switch (md-br0)
│   ├── veth.go           # Generates virtual network pairs and connects container interfaces
│   └── nat.go            # Deploys masqueraded iptables network sharing rules
│
├── registry/
│   ├── auth.go           # Handles secure bearer credential exchanges with Docker Hub
│   ├── manifest.go       # Requests and parses standard image layer manifest structures
│   └── download.go       # Streams container blobs down into localized disk caches
│
└── var/lib/minidocker/   # Data Directory (Requires root permissions)
    ├── cache/            # Storage location for pristine download blobs
    ├── layers/           # Extracted static image file templates
    └── merged/           # Active target windows for container operating spaces