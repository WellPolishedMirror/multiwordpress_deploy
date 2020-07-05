kubectl create -f crds/example.com_wordpresses_crd.yaml 
kubectl create -f service_account.yaml
kubectl create -f role.yaml
kubectl create -f role_binding.yaml
kubectl create -f operator.yaml
kubectl create -f crds/example.com_v1_wordpress_cr.yaml
