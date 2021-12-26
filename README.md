# Go实现的简单Http Server
1. 接收客户端 request，并将 request 中带的 header 写入 response header
2. 读取当前系统的环境变量中的 VERSION 配置，并写入 response header
3. Server 端记录访问日志包括客户端 IP，HTTP 返回码，输出到 server 端的标准输出
4. 当访问 localhost/healthz时，返回200

# k8s部署
```shell
kubectl apply -f deploy.yaml
```

# istio部署
```shell
cd istio/
kubectl create ns demo
kubectl apply -f nginx.yaml -n demo
kubectl apply -f service2.yaml -n demo
kubectl apply -f httpserver.yaml -n demo
kubectl apply -f istio-specs.yaml -n demo
kubectl apply -f secret.yaml -n istio-system
```