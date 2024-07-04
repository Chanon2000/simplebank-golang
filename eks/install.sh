# file นี้แค่เกิด command ที่ใช้ deploy cert-manager และ ingress-nginx เพิ่มเติมเท่านั้นนะ 
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.8.2/deploy/static/provider/aws/deploy.yaml
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.1/cert-manager.yaml
# command ที่เราใช้ deploy cert-manager และ ingress-nginx # สามารถดูได้จาก doc ของ kubernetes
