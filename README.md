# Kubernetes Cluster Manager

## Build instructions
To build and run the frontend app, run:
``` sh
cd cluster-manager/frontend
yarn add
yarn start
```

To build and run the backend code, run:
``` sh
cd cluster-manager/backend
go run main.go
```

To manage your Kubernetes cluster, you will need to make sure that you have one running on your computer. 
For example, you can set up a cluster quickly using `minikube`:
``` sh
minikube start
```

Once the app is running, you can manage your cluster's deployments and pods through the web interface:
![image](https://github.com/user-attachments/assets/238a3867-c681-4aac-873e-31cc666a9c8a)
