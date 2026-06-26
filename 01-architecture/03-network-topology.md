# Network Topology & Gateway Configuration

To host services securely on a private bare-metal network without exposing public router ports, CNP implements a **Zero-Trust Ingress Topology** using Cloudflare Tunnels (`cloudflared`) and the Kubernetes Gateway API (powered by Envoy Gateway).

---

## 1. Global Ingress Traffic Flow

No inbound ports (such as `80` or `443`) are opened on the local network router. Instead, a lightweight `cloudflared` daemon running inside the cluster initiates secure outbound TCP connections to the Cloudflare Edge network.

```text
[ Developer Browser ] ──(HTTPS)──► [ Cloudflare Edge ] ◄──(Secure Tunnel)──► [ cloudflared Pod ]
                                                                                   │
                                                                                   ▼ (HTTP Port 80)
                                                                           [ Envoy Gateway ]
                                                                                   │
                                                                           [ HTTPRoute Rules ]
                                                                                   │
                                                                                   ▼
                                                                           [ Application Pod ]
```

---

## 2. Gateway API Configuration

CNP uses the standardized Kubernetes Gateway API (`gateway.networking.k8s.io/v1`) rather than legacy Ingress resources.

### A. The Ingress Gateway (`gateway.yaml`)
A single, shared Gateway resource handles wildcard traffic for `*.3istor.com` and routes requests to namespaces labeled with `prod-gateway-access: "true"`.

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: shared-gateway
  namespace: gateway-infra
spec:
  gatewayClassName: eg # Powered by Envoy Gateway
  listeners:
    - name: http-wildcard
      protocol: HTTP
      port: 80
      hostname: "*.3istor.com"
      allowedRoutes:
        namespaces:
          from: Selector
          selector:
            matchLabels:
              prod-gateway-access: "true"
```

### B. Client IP Preservation (`ClientTrafficPolicy`)
Because HTTP requests transit through the Cloudflare proxy and the local tunnel, the client's real IP address would normally be lost (replaced by the local tunnel IP). 

CNP resolves this by deploying a `ClientTrafficPolicy` targeting the Gateway, instructing Envoy to read the `X-Forwarded-For` header and trust the upstream proxy hop:

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: ClientTrafficPolicy
metadata:
  name: trust-cloudflare
  namespace: gateway-infra
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: shared-gateway
  clientIPDetection:
    xForwardedFor:
      numTrustedHops: 1 # Considers Cloudflare as 1 trusted proxy hop
```

---

## 3. Local Development Port Forwarding (SSH)

Since the Kubernetes API server (`6443`) is hidden behind a private network and cannot be accessed over the public internet, developers access the cluster control planes securely using SSH Local Port Forwarding tunnels.

### Developer `~/.ssh/config` Template:
```text
Host pae-node-1
    Hostname 10.0.0.1
    User pae-node-1
    Port 22
    IdentityFile ~/.ssh/id_rsa

    # Local port forwarding tunnels
    LocalForward 8080 192.168.1.210:80     # Horizon Dashboard Access
    LocalForward 5000 192.168.1.210:5000   # Keystone Identity API
    LocalForward 9876 192.168.1.210:9876   # LoadBalancer Octavia
    LocalForward 6443 192.168.1.212:6443   # K3s Main API Server (Production)
    LocalForward 6444 192.168.1.217:6443   # K3s Lab API Server (Testing)
```

By keeping an active SSH session to `pae-node-1` open, developers can execute standard commands locally as if they were directly in the cluster:
```bash
# Point local kubectl to forwarded port
export KUBECONFIG=~/.kube/config-forwarded
kubectl get nodes --server=https://127.0.0.1:6443
```

---

## 4. IP Allocations & CIDR Strategy

To prevent routing conflicts across the hybrid environment, the platform strictly segregates IP address spaces:

| Network Plane | CIDR Range | Allocation Targets |
| :--- | :--- | :--- |
| **WireGuard Mesh VPN** | `10.0.0.0/24` | local nodes (.1, .2, .3) |
| **OpenStack External (Local LAN)**| `192.168.1.0/24` | Gateway (.254), Kolla VIP (.210), FIPs (.211-.230) |
| **OpenStack Tenant Network** | `172.16.0.0/24` | Isolated virtual machine tenants |
| **AWS VPC** | `10.1.0.0/16` | Public/Private subnets (Non-overlapping with local mesh) |

---
**Next Step**: Continue to [CMP Core Dashboard & API](../02-core-components/01-cmp-dashboard.md) (or return to the [Project Overview](../README.md)).
