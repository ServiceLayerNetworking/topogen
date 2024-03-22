# Microservice Topology Generator

### Handcrafting Benchmarks Sucks
The aim of this tool is to automatically generate code and config for given microservice topologies. We find that the functionality of microservices
doesn't really matter for system performance, and implementing these features for benchmarks is a waste of time. What we really care about is the service topology,
potential call graphs, call sizes, and differences in compute between services.



Go from

```yaml
services:
  - name: frontend
    methods:
      - method: POST
        path: /composepost
        computes: 0
        calls:
          - name: posts
            method: POST
            path: /posts
            size: 5242880
          - name: users
            method: GET
            path: /users
            size: 0
      - method: GET
        path: /posts/feed
        computes: 0
        calls:
          - name: posts
            method: GET
            path: /posts
            size: 5120
  - name: posts
    methods:
      - method: GET
        path: /posts
        computes: 0
        returnSize: 10240
      - method: POST
        path: /posts
        computes: 0
  - name: users
    methods:
      - method: GET
        path: /users
        computes: 0
        returnSize: 5120
```
to this in one command:

![alt text](image.png)
