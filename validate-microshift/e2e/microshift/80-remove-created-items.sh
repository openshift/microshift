#!/bin/bash
kubectl delete ns guestbook
kubectl delete application -n argocd guestbook
kubectl delete ns uploader
kubectl delete -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml -n argocd || true
kubectl delete ns argocd 
