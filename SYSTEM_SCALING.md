1. Database
- connection pooling
- database indexes
- slow query identification and optimisation
- monitor database metrics: CPU, memory, active connections, idle in transaction etc.
- database partitioning to break up larger datasets into smaller (reduced I/O ops, parallelism, efficient indexes),
- master-replica to balance the load between master (writing) and replicas (reading)

2. Application scaling in Kubernetes
- use HorizontalPodAutoscaler in Kubernetes and metrics like CPU utilization or custom metrics (requests/sec) exposed e.g. by Prometheus to scale it automatically.
- use the load balancer in front of the application to distribute the traffic between pods
4. Use context with timeouts to 
4. Use cache mechanism:
- Redis
5. Run load testing locally.