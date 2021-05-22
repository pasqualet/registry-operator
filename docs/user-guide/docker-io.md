# Docker.io

## DockerioCredentials

Example:

```yaml
apiVersion: registry.astrokube.com/v1alpha1
kind: DockerioCredentials
metadata:
  name: sample
spec:
  user: XXXXXXXXXXXXXXXXXXXX
  password: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
  imageSelector:
    - myuser/myimage:.*
```
